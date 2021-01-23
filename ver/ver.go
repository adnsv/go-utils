package ver

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Quad struct {
	Major uint32
	Minor uint32
	Patch uint32
	Build uint32
}

var ErrInvalidVersionQuad = errors.New("invalid version quad")

func (q *Quad) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", q.Major, q.Minor, q.Patch, q.Build)
}

var quadRegex = regexp.MustCompile(`^([0-9]|[1-9][0-9]*)\.([0-9]|[1-9][0-9]*)\.([0-9]|[1-9][0-9]*)(?:.([0-9A-Za-z-]))?$`)

func ParseQuad(s string) (Quad, error) {
	parts := strings.Split(s, ".")
	if len(parts) < 2 || len(parts) > 4 {
		return Quad{}, ErrInvalidVersionQuad
	}
	major, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return Quad{}, err
	}
	minor, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return Quad{}, err
	}
	q := Quad{
		Major: uint32(major),
		Minor: uint32(minor),
	}
	if len(parts) == 3 {
		patch, err := strconv.ParseUint(parts[2], 10, 32)
		if err != nil {
			return Quad{}, err
		}
		q.Patch = uint32(patch)
	}
	if len(parts) == 4 {
		build, err := strconv.ParseUint(parts[3], 10, 32)
		if err != nil {
			return Quad{}, err
		}
		q.Build = uint32(build)
	}
	return q, nil
}

func RelaxedParseQuad(s string) (Quad, error) {
	if len(s) > 0 && (s[0] == 'v' || s[0] == 'V') {
		s = s[1:]
	}
	return ParseQuad(s)
}

func (v *Quad) Compare(o Quad) int {
	if v.Major != o.Major {
		if v.Major > o.Major {
			return 1
		}
		return -1
	}
	if v.Minor != o.Minor {
		if v.Minor > o.Minor {
			return 1
		}
		return -1
	}
	if v.Patch != o.Patch {
		if v.Patch > o.Patch {
			return 1
		}
		return -1
	}
	if v.Build == o.Build {
		return 0
	}
	if v.Build > o.Build {
		return 1
	}
	return -1
}
