package git

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/adnsv/go-utils/runner"
	"github.com/adnsv/go-utils/ver"
)

// Stats contains a set of git statistics
type Stats struct {
	Branch     string // result of `git branch --show-current`
	Describe   string // result of `git describe --long` command
	Hash       string // result of `git rev-parse HEAD` command
	ShortHash  string // result of `git rev-parse --short HEAD` command
	AuthorDate string // result of `git log -n1 --date=format:"%Y-%m-%dT%H:%M:%S" --format=%ad`
}

// Stat obtains git stats for a specified local directory
func Stat(dir string) (*Stats, error) {
	ret := &Stats{}
	var err error
	ret.Branch, err = runner.WDTrimmedOutput(dir, "git", "branch", "--show-current")
	if err != nil {
		return nil, err
	}
	ret.Describe, err = runner.WDTrimmedOutput(dir, "git", "describe", "--long")
	if err != nil {
		return nil, err
	}
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

// DescVersion contains the results of parsing the git describe output with
// RetrieveSemanticVersionV4 (see below)
type DescVersion struct {
	LastTag           string
	AdditionalCommits string // number of additional commits after the last tag
	Numeric           string
	Quad              ver.Quad // semantic version extracted from tag: "0.1.2.3"
	Combined          string
}

// RetrieveSemanticVersionV4 parses the result of git describe and extracts the
// application version number assuming the repo is tagged with v1.2.3.4 tags. If
// the repo had evolved since the last tag, the git describe adds a suffix in a
// -<nsteps>-<hash> form that is reported in the Suffix field of the returned
// struct
func (s *Stats) RetrieveSemanticVersionV4() (*DescVersion, error) {
	reDescribe := regexp.MustCompile(`^(.*)-(\d+)-g([0-9,a-f]{7})$`)
	parts := reDescribe.FindStringSubmatch(s.Describe)
	if len(parts) != 4 {
		return nil, errors.New("failed to parse `git describe` result")
	}

	ret := &DescVersion{
		LastTag:           parts[1],
		AdditionalCommits: parts[2],
		Combined:          parts[1],
	}

	v := ret.LastTag
	if len(v) > 0 && (v[0] == 'v' || v[0] == 'V') {
		v = v[1:]
	}
	ret.Numeric = v
	var err error
	ret.Quad, err = ver.ParseQuad(v)
	if err != nil {
		return nil, fmt.Errorf("failed to extract semantic version from tag %s", ret.LastTag)
	}

	ret.Combined = v
	if ret.AdditionalCommits != "0" {
		ret.Combined += "+" + ret.AdditionalCommits
	}

	return ret, nil
}
