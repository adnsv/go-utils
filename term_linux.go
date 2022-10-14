//go:build linux || aix || zos
// +build linux aix zos

package ansi

import (
	"os"

	"golang.org/x/sys/unix"
)

func implSetupOutput(f *os.File) (ok bool, cleanup func()) {
	_, err := unix.IoctlGetTermios(int(f.Fd()), unix.TCGETS)
	return err == nil, func() {}
}
