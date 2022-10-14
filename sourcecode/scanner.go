package sourcecode

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

// Scanner is a generic helper for scraping source code for translation strings
type Scanner struct {
	Filepath  string
	buffer    string
	end       int // length of buffer
	lineindex int // 1-based, zero for empty inputs
	linestart int
	cur       int
}

func NewScanner[T StringLikeContent](b T) *Scanner {
	scn := &Scanner{
		buffer: string(b),
		end:    len(b),
	}

	// skip BOM
	if strings.HasPrefix(scn.buffer, "\uFEFF") {
		scn.cur = 3
	}
	scn.linestart = scn.cur
	return scn
}

func NewFileScanner(filepath string) (*Scanner, error) {
	b, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	ret := NewScanner(b)
	ret.Filepath = filepath
	return ret, nil
}

func (scn *Scanner) Anchor() Anchor {
	return Anchor{
		lineIndex:       scn.lineindex,
		lineStartOffset: scn.linestart,
		offset:          scn.cur,
	}
}

func (scn *Scanner) LineContentAt(anchor Anchor) string {
	return LineContentAt(scn.buffer, anchor)
}

func (scn *Scanner) LocationAt(anchor Anchor) Location {
	return LocationAt(scn.buffer, anchor)
}

func (scn *Scanner) MakeErrorAt(anchor Anchor, err error) error {
	loc := LocationAt(scn.buffer, anchor)
	if scn.Filepath == "" {
		return NewLocationError(loc, err)
	}
	floc := FileLocation{scn.Filepath, loc}
	return NewFileLocationError(floc, err)
}

func (scn *Scanner) IsWS() bool {
	return scn.cur < scn.end && isWS(scn.buffer[scn.cur])
}

func (scn *Scanner) IsEOF() bool {
	return scn.cur >= scn.end
}

func (scn *Scanner) IsEOL() bool {
	return scn.cur < scn.end && isEOL(scn.buffer[scn.cur])
}

func (scn *Scanner) IsByte(c byte) bool {
	return scn.cur < scn.end && scn.buffer[scn.cur] == c
}

func (scn *Scanner) IsByteFunc(f func(byte) bool) bool {
	return scn.cur < scn.end && f(scn.buffer[scn.cur])
}

func (scn *Scanner) IsNextByte(offset int, c byte) bool {
	return scn.cur+offset < scn.end && scn.buffer[scn.cur+offset] == c
}

func (scn *Scanner) IsNextByteFunc(offset int, f func(byte) bool) bool {
	return scn.cur+offset < scn.end && f(scn.buffer[scn.cur+offset])
}

func (scn *Scanner) IsAnyOf(chars string) bool {
	if scn.cur >= scn.end {
		return false
	}
	return strings.IndexByte(chars, scn.buffer[scn.cur]) >= 0
}

func (scn *Scanner) IsSequence(s string) bool {
	n := len(s)
	return scn.cur+n < scn.end && scn.buffer[scn.cur:scn.cur+n] == s
}

func (scn *Scanner) Skip() bool {
	if scn.cur >= scn.end {
		return false
	}
	if !scn.SkipEOL() {
		scn.cur++
	}
	return true
}

func (scn *Scanner) SkipWS() bool {
	for scn.cur < scn.end && isWS(scn.buffer[scn.cur]) {
		scn.cur++
		return true
	}
	return false
}

func (scn *Scanner) SkipByte(c byte) bool {
	if scn.cur < scn.end && scn.buffer[scn.cur] == c {
		scn.cur++
		return true
	}
	return false
}

func (scn *Scanner) SkipByteFunc(f func(byte) bool) bool {
	if scn.cur < scn.end && f(scn.buffer[scn.cur]) {
		scn.cur++
		return true
	}
	return false
}

func (scn *Scanner) SkipSequence(s string) bool {
	n := len(s)
	if scn.cur+n < scn.end && scn.buffer[scn.cur:scn.cur+n] == s {
		scn.cur += n
		return true
	}
	return false
}

func (scn *Scanner) SkipEOL() bool {
	if scn.cur >= scn.end {
		return false
	}
	c := scn.buffer[scn.cur]
	if c == '\r' {
		scn.cur++
		scn.SkipByte('\n')
	} else if c == '\n' {
		scn.cur++
	} else {
		return false
	}
	scn.lineindex++
	scn.linestart = scn.cur
	return true
}

func (scn *Scanner) SkipWSEOL() bool {
	save := scn.cur
	for scn.SkipWS() || scn.SkipEOL() {
	}
	return save != scn.cur
}

var ErrHexDigitExpected = errors.New("a valid hex digit expected")

func (scn *Scanner) ReadHexUint() (uint64, error) {
	i := scn.cur
	for i < scn.end && isHexDigit(scn.buffer[i]) {
		i++
	}
	if i == scn.cur {
		return 0, ErrHexDigitExpected
	}
	v, err := strconv.ParseUint(scn.buffer[scn.cur:i], 16, 64)
	if err == nil {
		scn.cur = i
	}
	return v, err
}

func (scn *Scanner) ReadLineUntil(c byte) string {
	i := scn.cur
	for i < scn.end {
		cp := scn.buffer[i]
		if isEOL(cp) || cp == c {
			break
		} else {
			i++
		}
	}
	ret := scn.buffer[scn.cur:i]
	scn.cur = i
	return ret
}

var ErrDecDigitExpected = errors.New("a valid decimal digit expected")

func (scn *Scanner) ReadDecUint() (uint64, error) {
	i := scn.cur
	for i < scn.end && isDecDigit(scn.buffer[i]) {
		i++
	}
	if i == scn.cur {
		return 0, ErrDecDigitExpected
	}
	v, err := strconv.ParseUint(scn.buffer[scn.cur:i], 10, 64)
	if err == nil {
		scn.cur = i
	}
	return v, err
}

func (scn *Scanner) ReadSequenceFunc(isFirst, isOther func(byte) bool) string {
	start := scn.cur
	if !scn.IsByteFunc(isFirst) {
		return ""
	}
	for scn.SkipByteFunc(isOther) {
	}
	return scn.buffer[start:scn.cur]
}

func isWS(c byte) bool {
	return c == ' ' || c == '\t' // || c == '\r' || c == '\n'
}

func isEOL(c byte) bool {
	return c == '\r' || c == '\n'
}

func isDecDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return isDecDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}
