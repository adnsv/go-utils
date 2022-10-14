//go:build darwin || freebsd || openbsd || netbsd || dragonfly
// +build darwin freebsd openbsd netbsd dragonfly

package ansi

import (
	"os"

	"golang.org/x/sys/unix"
)

func implSetupOutput(f *os.File) (ok bool, cleanup func()) {
	_, err := unix.IoctlGetTermios(int(f.Fd()), unix.TIOCGETA)
	return err == nil, func() {}
}
