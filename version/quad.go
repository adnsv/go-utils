package version

import (
	"errors"
	"fmt"

	"github.com/josephspurrier/goversioninfo"
)

// Quad can be used to form version quads from semantic versions.
// These are required for embedding numeric versions into Windows resources.
type Quad goversioninfo.FileVersion

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

// ErrUnsupportedPR is produced by MakeVersionQuad for unsupported pre-release
// suffixes.
var ErrUnsupportedPR = errors.New("unsupported pre-release content")

// MakeQuad constructs a version quad from a semantic version and the
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
func MakeQuad(v Semantic, additionalCommits int) (Quad, error) {
	if v.Major > 65535 || v.Minor > 65535 || v.Patch > 65535 {
		return Quad{0, 0, 0, 0}, ErrVersionNumberIsTooLarge
	}

	ret := Quad{
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
		additionalCommits = quadPatchPRMultiplier - 1
	}

	ret.Build += additionalCommits
	return ret, nil
}

// String implements a stringer interface, producing #.#.#.# output
func (v Quad) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", v.Major, v.Minor, v.Patch, v.Build)
}

// CommaSeparatedString produces #,#,#,# output, this is useful when
// code-generating Win32 resources
func (v Quad) CommaSeparatedString() string {
	return fmt.Sprintf("%d,%d,%d,%d", v.Major, v.Minor, v.Patch, v.Build)
}
