package git

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/adnsv/go-utils/runner"
	"github.com/adnsv/go-utils/version"
	"github.com/blang/semver/v4"
)

// Stats contains a set of git statistics
type Stats struct {
	Branch      string      // result of `git branch --show-current`
	Description Description // result of `git describe --long` command
	Hash        string      // result of `git rev-parse HEAD` command
	ShortHash   string      // result of `git rev-parse --short HEAD` command
	AuthorDate  string      // result of `git log -n1 --date=format:"%Y-%m-%dT%H:%M:%S" --format=%ad`
	Dirty       bool        // repo returns non-empty `git status --porcelain`
}

var ErrNotInsideWorktree = errors.New("dir is outside of git worktree")
var ErrNoTags = errors.New("repository has no tags (git describe failed)")

type StatFlags int

const (
	StatNoPorcelain = StatFlags(1 << iota) // exclude the `--porcelain` flag from `git status` query
)

// Stat obtains git stats for a specified local directory
func Stat(dir string, flags ...StatFlags) (*Stats, error) {
	ret := &Stats{}
	var err error

	ff := StatFlags(0)
	for _, f := range flags {
		ff |= f
	}

	_, err = exec.Command("git", "rev-parse", "--is-inside-work-tree").Output()
	if err != nil {
		return nil, ErrNotInsideWorktree
	}

	ret.Branch, err = runner.WDTrimmedOutput(dir, "git", "branch", "--show-current")
	if err != nil {
		ret.Branch = ""
		//return nil, fmt.Errorf("while running 'git branch --show-current': %w", err)
	}
	ret.Hash, err = runner.WDTrimmedOutput(dir, "git", "rev-parse", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("while running 'git rev-parse HEAD': %w", err)
	}
	ret.ShortHash, err = runner.WDTrimmedOutput(dir, "git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("while running 'git rev-parse --short HEAD': %w", err)
	}
	ret.AuthorDate, err = runner.WDTrimmedOutput(dir, "git", "log", "-n1", "--date=format:%Y-%m-%dT%H:%M:%S", "--format=%ad")
	if err != nil {
		return nil, fmt.Errorf("while running '%s': %w", `git log -n1 --date=format:%Y-%m-%dT%H:%M:%S --format=%ad`, err)
	}

	status_flags := []string{"status"}
	if ff&StatNoPorcelain == 0 {
		status_flags = append(status_flags, "--porcelain")
	}
	s, err := runner.WDTrimmedOutput(dir, "git", status_flags...)
	if err != nil && err != runner.ErrEmptyOutput {
		return nil, fmt.Errorf("while running 'git %s': %w", strings.Join(status_flags, " "), err)
	}
	ret.Dirty = !(s == "" || s == "\n" || s == "\r\n")
	s, err = runner.WDTrimmedOutput(dir, "git", "describe", "--tags", "--long")
	if err != nil {
		return ret, ErrNoTags
	}
	d, err := ParseDescription(s)
	if err != nil {
		return ret, fmt.Errorf("while parsing 'git describe' output: %w", err)
	}
	ret.Description = *d

	return ret, nil
}

// Description contains the results of parsing the git describe output
type Description struct {
	Tag               string
	AdditionalCommits int // number of additional commits after the last tag
	ShortHash         string
}

// ParseDescription parses the result of `git describe --long`
func ParseDescription(s string) (*Description, error) {
	re := regexp.MustCompile(`^(.*)-(\d+)-g([0-9,a-f]+)$`)
	parts := re.FindStringSubmatch(s)
	if len(parts) != 4 {
		return nil, errors.New("failed to parse `git describe` result")
	}
	n, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, err
	}
	return &Description{
		Tag:               parts[1],
		AdditionalCommits: n,
		ShortHash:         parts[3],
	}, nil
}

func LastSemanticTag(dir string) (string, *VersionInfo, error) {
	out, err := exec.Command("git", "tag", "--sort=-creatordate").Output()
	if err != nil {
		return "", nil, err
	}

	scn := bufio.NewScanner(bytes.NewReader(out))
	for scn.Scan() {
		ln := scn.Text()

		v, err := semver.ParseTolerant(ln)
		if err != nil {
			continue
		}

		out, err := exec.Command("git", "describe", ln, "--long").Output()
		if err != nil {
			continue
		}
		d, err := ParseDescription(strings.TrimSpace(string(out)))
		if err != nil {
			continue
		}
		ret := &VersionInfo{
			SemanticTag:       v,
			Semantic:          v,
			AdditionalCommits: d.AdditionalCommits,
		}
		if d.AdditionalCommits > 0 {
			// add additional commits (if any) as build suffix
			ret.Semantic.Build = append([]string{strconv.Itoa(d.AdditionalCommits)}, ret.Semantic.Build...)
		}
		ret.Quad, err = version.MakeQuad(v, d.AdditionalCommits)
		if err != nil {
			return "", nil, err
		}
		if len(v.Pre) > 0 {
			ret.Pre = v.Pre[0].String()
			for _, p := range v.Pre[1:] {
				ret.Pre += "." + p.String()
			}
		}
		if len(v.Build) > 0 {
			ret.Build = strings.Join(v.Build, ".")
		}
		return ln, ret, nil
	}
	return "", nil, errors.New("failed to parse `git describe` result")
}

// VersionInfo contains version information in various formats
type VersionInfo struct {
	SemanticTag       version.Semantic // as parsed from tag
	Semantic          version.Semantic // with additional commits, if != 0
	AdditionalCommits int
	Quad              version.Quad
	Pre               string // semantic pre-release suffix
	Build             string // semantic build suffix
}

// ParseVersion extracts useful version info from git.Stat description
func ParseVersion(d Description) (*VersionInfo, error) {
	ret := &VersionInfo{}
	v, err := semver.ParseTolerant(d.Tag)
	if err != nil {
		return nil, err
	}
	ret.AdditionalCommits = d.AdditionalCommits
	ret.SemanticTag = v
	ret.Semantic = v
	if d.AdditionalCommits > 0 {
		// add additional commits (if any) as build suffix
		ret.Semantic.Build = append([]string{strconv.Itoa(d.AdditionalCommits)}, ret.Semantic.Build...)
	}
	ret.Quad, err = version.MakeQuad(v, d.AdditionalCommits)
	if err != nil {
		return ret, err
	}
	if len(v.Pre) > 0 {
		ret.Pre = v.Pre[0].String()
		for _, p := range v.Pre[1:] {
			ret.Pre += "." + p.String()
		}
	}
	if len(v.Build) > 0 {
		ret.Build = strings.Join(v.Build, ".")
	}
	return ret, nil
}
