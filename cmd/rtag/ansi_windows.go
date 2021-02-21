package main

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32                         = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode               = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode               = kernel32.NewProc("SetConsoleMode")
	cEnableVirtualTerminalProcessing = uint32(0x4)
)

func enableVT(f *os.File) (ok bool, cleanup func()) {
	fd := f.Fd()

	var mode uint32
	r, _, _ := procGetConsoleMode.Call(fd, uintptr(unsafe.Pointer(&mode)))

	if r == 0 {
		return false, func() {}
	}

	if mode&cEnableVirtualTerminalProcessing != 0 {
		// already has VT enabled
		return true, func() {}
	}

	r, _, _ = procSetConsoleMode.Call(fd, uintptr(mode|cEnableVirtualTerminalProcessing))
	if r == 0 {
		return false, func() {}
	}

	return true, func() {
		procSetConsoleMode.Call(fd, uintptr(mode))
	}
}
