// +build darwin freebsd openbsd netbsd dragonfly

package main

import (
	"os"

	"golang.org/x/sys/unix"
)

func enableVT(f *os.File) (ok bool, cleanup func()) {
	_, err := unix.IoctlGetTermios(int(f.Fd()), unix.TIOCGETA)
	return err == nil, func() {}
}
