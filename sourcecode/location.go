package sourcecode

import (
	"fmt"
	"unicode/utf8"
)

// Location specifies line:column position within a text file
//
//   - both numers are 1-indexed
//   - zero values are used when line or column is unknown
//
// note: that the column number here may differ from what your text editor is
// showing, due to the following simplifications:
//
//   - we can't expand tabs `\t` because we don't have a clue what's the tab size is
//   - we can't account for zero-width and wide (double-column) characters because unicode
//     character width lookup is outside of the scope of this library
type Location struct {
	LineNumber   int // 1-based
	ColumnNumber int // 1-based
}

func (loc *Location) Valid() bool {
	return loc.LineNumber > 0
}

func (loc *Location) String() string {
	if loc.ColumnNumber > 0 {
		return fmt.Sprintf("%d:%d", loc.LineNumber, loc.ColumnNumber)
	} else if loc.LineNumber > 0 {
		return fmt.Sprintf("%d", loc.LineNumber)
	} else {
		return ""
	}
}

type FileLocation struct {
	Filename string
	Location
}

func (loc *FileLocation) String() string {
	if loc.ColumnNumber > 0 {
		return fmt.Sprintf("%s:%d:%d", loc.Filename, loc.LineNumber, loc.ColumnNumber)
	} else if loc.LineNumber > 0 {
		return fmt.Sprintf("%s:%d", loc.Filename, loc.LineNumber)
	} else {
		return loc.Filename
	}
}

type Anchor struct {
	lineIndex       int // this is 0-indexed (not 1-indexed like Location.LineNumber)
	lineStartOffset int // anchored line offset in buffer, in codeunits
	offset          int // anchor offset in buffer, in codeunits
}

func (a *Anchor) LineNumber() int {
	return a.lineIndex + 1
}

func (a *Anchor) Offset() int {
	return a.offset
}

type StringLikeContent interface {
	~string | []byte
}

func CalcAnchor[T StringLikeContent](buf T) Anchor {
	i, n := 0, len(buf)
	if n == 0 {
		return Anchor{0, 0, 0}
	}
	a := Anchor{0, 0, n}
	for i < n {
		if isEOL(buf[i]) {
			if buf[i] == '\r' {
				i++
				if i < n && buf[i] == '\n' {
					i++
				}
			} else {
				i++
			}
			a.lineIndex++
			a.lineStartOffset = i
		} else {
			i++
		}
	}
	return a
}

func LocationAt[T StringLikeContent](buf T, anchor Anchor) Location {
	b, e := anchor.lineStartOffset, anchor.offset
	if b < 0 || e < b || e > len(buf) {
		panic("invalid location anchor")
	}
	return Location{
		LineNumber:   1 + anchor.lineIndex,
		ColumnNumber: 1 + utf8.RuneCount([]byte(buf[b:e])),
	}
}

func LineContentAt[T StringLikeContent](buf T, anchor Anchor) T {
	b := anchor.lineStartOffset
	if b < 0 || b > len(buf) {
		panic("invalid location anchor")
	}

	e := b
	n := len(buf)
	for e < n {
		if isEOL(buf[e]) {
			break
		} else {
			e++
		}
	}
	if anchor.offset > e {
		panic("invalid location anchor")
	}
	return buf[b:e]
}

type LocationError struct {
	Location Location
	Err      error
}

type FileLocationError struct {
	Filename string
	Location Location
	Err      error
}

func NewLocationError(loc Location, err error) *LocationError {
	return &LocationError{loc, err}
}

func NewFileLocationError(fl FileLocation, err error) *FileLocationError {
	return &FileLocationError{fl.Filename, fl.Location, err}
}

func (e *LocationError) Error() string {
	if e.Location.Valid() {
		return fmt.Sprintf("[%s] %s", e.Location.String(), e.Err.Error())
	} else {
		return e.Err.Error()
	}
}

func (e *FileLocationError) Error() string {
	fl := FileLocation{e.Filename, e.Location}
	return fmt.Sprintf("[%s] %s", fl.String(), e.Err.Error())
}
