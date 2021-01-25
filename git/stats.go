package git

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/adnsv/go-utils/runner"
	"github.com/blang/semver/v4"
)

// Stats contains a set of git statistics
type Stats struct {
	Branch      string      // result of `git branch --show-current`
	Description Description // result of `git describe --long` command
	Hash        string      // result of `git rev-parse HEAD` command
	ShortHash   string      // result of `git rev-parse --short HEAD` command
	AuthorDate  string      // result of `git log -n1 --date=format:"%Y-%m-%dT%H:%M:%S" --format=%ad`
}

// Stat obtains git stats for a specified local directory
func Stat(dir string) (*Stats, error) {
	ret := &Stats{}
	var err error
	ret.Branch, err = runner.WDTrimmedOutput(dir, "git", "branch", "--show-current")
	if err != nil {
		return nil, err
	}
	s, err := runner.WDTrimmedOutput(dir, "git", "describe", "--long")
	if err != nil {
		return nil, err
	}
	d, err := ParseDescription(s)
	if err != nil {
		return nil, err
	}
	ret.Description = *d
	ret.Hash, err = runner.WDTrimmedOutput(dir, "git", "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}
	ret.ShortHash, err = runner.WDTrimmedOutput(dir, "git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return nil, err
	}
	ret.AuthorDate, err = runner.WDTrimmedOutput(dir, "git", "log", "-n1", "--date=format:%Y-%m-%dT%H:%M:%S", "--format=%ad")
	if err != nil {
		return nil, err
	}
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
	re := regexp.MustCompile(`^(.*)-(\d+)-g([0-9,a-f]{7})$`)
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

// VersionInfo contains version information in various formats
type VersionInfo struct {
	SemanticTag semver.Version // as parsed from tag
	Semantic    semver.Version // with additional commits, if != 0
	Triplet     string         // Major.Minor.Patch
	Quad        string         // dot-separated quad (Major.Minor.Patch.GitAdditionalCommits)
	NNNN        string         // comma-separated quad, useful for windows RC building
	Pre         string         // semantic pre-release suffix
	Build       string         // semantic build suffix
}

// ParseVersion extracts useful version info from git.Stat description
func ParseVersion(d Description) (*VersionInfo, error) {
	ret := &VersionInfo{}
	v, err := semver.ParseTolerant(d.Tag)
	if err != nil {
		return nil, err
	}
	ret.SemanticTag = v
	ret.Semantic = v
	if d.AdditionalCommits > 0 {
		// add additional commits (if any) as build suffix
		ret.Semantic.Build = append([]string{strconv.Itoa(d.AdditionalCommits)}, ret.Semantic.Build...)
	}
	ret.Triplet = fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	ret.Quad = fmt.Sprintf("%d.%d.%d.%d", v.Major, v.Minor, v.Patch, d.AdditionalCommits)
	ret.NNNN = fmt.Sprintf("%d,%d,%d,%d", v.Major, v.Minor, v.Patch, d.AdditionalCommits)
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
