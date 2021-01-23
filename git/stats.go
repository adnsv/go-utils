package git

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/adnsv/go-utils/runner"
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
