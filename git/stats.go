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
	s, err := runner.WDTrimmedOutput(dir, "git", "describe", "--long")
	if err != nil {
		return nil, errors.New("git describe failure, repo has no tags")
	}
	d, err := ParseDescription(s)
	if err != nil {
		return nil, err
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

// VersionQuad can be used to form version quads from semantic versions.
// These are required for embedding numeric versions into Windows resources.
//
type VersionQuad struct {
	Major int
	Minor int
	Patch int
	Build int
}

const (
	quadPatchAlphaBase    = 10000
	quadPatchBetaBase     = 20000
	quadPatchRCBase       = 30000
	quadPatchReleaseBase  = 50000
	quadPatchPRMultiplier = 100
)

// ErrVersionNumberIsTooLarge is produced by MakeVersionQuad when one of the
// numbers is out of range.
var ErrVersionNumberIsTooLarge = errors.New("version number is too large")

// ErrReleaseNumberIsTooLarge is produced by MakeVersionQuad when numeric suffix
// is out of range.
var ErrReleaseNumberIsTooLarge = errors.New("release number is too large")

// ErrNumberOfAdditionalCommitsIsTooLarge is produced by MakeVersionQuad when
// the number of additional commits is out of range.
var ErrNumberOfAdditionalCommitsIsTooLarge = errors.New("number of additional commits is too large")

// ErrUnsupportedPR is produced by MakeVersionQuad for unsupported pre-release
// suffixes.
var ErrUnsupportedPR = errors.New("unsupported pre-release content")

// MakeVersionQuad constructs a version quad from a semantic version and the
// number of additional commits.
//
// To keep the numbers consistent with Win32 storage (2-bytes per number), this
// function will reject certain inputs where numbers are too large
//
// The first triplet is taken from the semantic version, the Build number is
// formed as:
//
//   - alphas ('alpha', 'a' PR suffixes): are 10000+ numbers
//   - betas ('beta', 'b' PR suffixes): are 20000+ numbers
//   - release candidates ('rc' suffixes): are 30000+ numbers
//   - releases (no suffixes): are 50000+ numbers
//
// The .<number> after the PR suffix, if present is multiplied by a 100 and
// added to the Build number. Same works for releases with -<number> suffix.
//
// The number of additional commits is added to the Build number.
//
// Examples:
//   - v1.0.0-alpha -> 1.0.0.10000
//   - v1.0.0-alpha.1 -> 1.0.0.10100
//   - v1.0.0-alpha.1 + 7 commits -> 1.0.0.10107
//   - v1.0.0-rc.5 -> 1.0.0.30500
//   - v1.0.0 -> 1.0.0.50000
//   - v1.0.0-5 -> 1.0.0.50005
//
// see stats-test.go for more examples
//
func MakeVersionQuad(v semver.Version, additionalCommits int) (VersionQuad, error) {
	if v.Major > 65535 || v.Minor > 65535 || v.Patch > 65535 {
		return VersionQuad{0, 0, 0, 0}, ErrVersionNumberIsTooLarge
	}

	ret := VersionQuad{
		Major: int(v.Major),
		Minor: int(v.Minor),
		Patch: int(v.Patch),
	}

	if len(v.Pre) == 0 {
		ret.Build = quadPatchReleaseBase
	} else if v.Pre[0].IsNum {
		ret.Build = quadPatchReleaseBase
		if v.Pre[0].VersionNum > 99 {
			return ret, ErrReleaseNumberIsTooLarge
		}
		ret.Build += int(v.Pre[0].VersionNum) * quadPatchPRMultiplier
	} else if !v.Pre[0].IsNum {
		switch v.Pre[0].VersionStr {
		case "alpha":
			ret.Build = quadPatchAlphaBase
		case "a":
			ret.Build = quadPatchAlphaBase
		case "beta":
			ret.Build = quadPatchBetaBase
		case "b":
			ret.Build = quadPatchBetaBase
		case "rc":
			ret.Build = quadPatchRCBase
		default:
			return ret, ErrUnsupportedPR
		}
		if len(v.Pre) >= 2 && v.Pre[1].IsNum {
			if v.Pre[1].VersionNum > 99 {
				return ret, ErrReleaseNumberIsTooLarge
			}
			ret.Build += int(v.Pre[1].VersionNum) * quadPatchPRMultiplier
		}
	}

	if additionalCommits >= quadPatchPRMultiplier {
		ret.Build += quadPatchPRMultiplier - 1
		return ret, ErrNumberOfAdditionalCommitsIsTooLarge
	}

	ret.Build += additionalCommits
	return ret, nil
}

// String implements a stringer interface, producing #.#.#.# output
func (v VersionQuad) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", v.Major, v.Minor, v.Patch, v.Build)
}

// CommaSeparatedString produces #,#,#,# output, this is useful when
// code-generating Win32 resources
func (v VersionQuad) CommaSeparatedString() string {
	return fmt.Sprintf("%d,%d,%d,%d", v.Major, v.Minor, v.Patch, v.Build)
}

// VersionInfo contains version information in various formats
type VersionInfo struct {
	SemanticTag       semver.Version // as parsed from tag
	Semantic          semver.Version // with additional commits, if != 0
	AdditionalCommits int
	Quad              VersionQuad
	Triplet           string // Major.Minor.Patch
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
	ret.Triplet = fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	ret.Quad, err := MakeVersionQuad(v, d.AdditionalCommits)
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
