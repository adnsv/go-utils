package fs

import (
	"fmt"
	"io/ioutil"
	"os"
)

// CheckFileHasContent returns true if the specified
// file exists and has content that matches buf.
func CheckFileHasContent(fn string, buf []byte) bool {
	if _, err := os.Stat(fn); os.IsNotExist(err) {
		return false
	}
	oldBytes, err := ioutil.ReadFile(fn)
	if err != nil || len(oldBytes) != len(buf) {
		return false
	}
	for i, b := range oldBytes {
		if buf[i] != b {
			return false
		}
	}
	return true
}

// WriteFileIfChanged writes buf into a file.
// Does not overwrite if the file already has the specified content.
// Uses 0666 permission if overwriting is neccessary.
func WriteFileIfChanged(fn string, buf []byte) error {
	if CheckFileHasContent(fn, buf) {
		return nil
	}
	return ioutil.WriteFile(fn, buf, 0666)
}

// LineAndCharacter locates line and pos from offset into a file
func LineAndCharacter(input string, offset int) (line int, character int, err error) {
	lf := rune(0x0A)
	if offset > len(input) || offset < 0 {
		return 0, 0, fmt.Errorf("can't find offset %d within the input", offset)
	}

	for i, b := range input {
		if b == lf {
			line++
			character = 0
		}
		character++
		if i == offset {
			break
		}
	}
	return line, character, nil
}
