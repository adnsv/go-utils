package ansi

import "os"

// OutputState allows restoring terminal to its original state
type OutputState struct {
	supported   bool
	restoreProc func()
}

func (s *OutputState) Supported() bool {
	return s.supported
}

func (s *OutputState) Restore() {
	if s.restoreProc != nil {
		s.restoreProc()
		s.restoreProc = nil
	}
}

// SetupOutput prepares/validates an output to support virtual terminal escape sequences
//
// Windows hosts:
// - configures terminal for UTF-8 output (65001 codepage)
// - enables virtual terminal processing mode
//
// Other hosts (linux, bsd):
// - checks for termios support
//
// Returns:
// - ok: indicates if the output supports virtual terminal escape sequences (will fail for regular files)
// - cleanup: to be used for restoring the output to its original state
//
func SetupOutput(output *os.File) *OutputState {
	ok, proc := implSetupOutput(output)
	return &OutputState{supported: ok, restoreProc: proc}
}

// SetupStdout configures os.Stdout to support virtual terminal escape sequences
func SetupStdout() *OutputState {
	return SetupOutput(os.Stdout)
}
