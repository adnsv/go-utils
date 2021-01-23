package runner

import (
	"errors"
	"os/exec"
	"strings"
)

// Error Codes
var (
	ErrEmptyOutput = errors.New("empty output received")
)

// TrimmedOutput runs an executable and returns its output as a string. Returns
// an error if the executable failes. Returns ErrEmptyOutput if there is no
// output.
func TrimmedOutput(name string, arg ...string) (string, error) {
	out, err := exec.Command(name, arg...).Output()
	if err != nil {
		return string(out), err
	}
	ret := strings.TrimSpace(string(out))
	if len(ret) == 0 {
		return "", ErrEmptyOutput
	}
	return ret, nil
}

// WDTrimmedOutput works similarly to TrimmedOutput, but acceprs work dir for
// command execution as a first parameter.
func WDTrimmedOutput(wd string, name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	cmd.Dir = wd
	out, err := cmd.Output()
	if err != nil {
		return string(out), err
	}
	ret := strings.TrimSpace(string(out))
	if len(ret) == 0 {
		return "", ErrEmptyOutput
	}
	return ret, nil
}
