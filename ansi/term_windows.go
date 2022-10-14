package ansi

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32                         = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode               = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode               = kernel32.NewProc("SetConsoleMode")
	procSetConsoleOutputCP           = kernel32.NewProc("SetConsoleOutputCP")
	procGetConsoleOutputCP           = kernel32.NewProc("GetConsoleOutputCP")
	cEnableVirtualTerminalProcessing = uint32(0x4)
	//cDisableNewlineAutoReturn        = uint32(0x0008)
)

type mode = uint32

func getMode(fd uintptr) (mode, error) {
	var m mode
	ok, _, e := procGetConsoleMode.Call(fd, uintptr(unsafe.Pointer(&m)))
	if ok != 0 {
		return m, nil
	}
	return m, e
}

func setMode(fd uintptr, m mode) error {
	ok, _, e := procSetConsoleMode.Call(fd, uintptr(m))
	if ok != 0 {
		return nil
	}
	return e
}
func setOutputCP(cp uint32) error {
	ok, _, e := procSetConsoleOutputCP.Call(uintptr(cp))
	if ok != 0 {
		return nil
	}
	return e
}

func getOutputCP() (uint32, error) {
	cp, _, e := procGetConsoleOutputCP.Call()
	if cp == 0 {
		return 0, e
	}
	return uint32(cp), nil
}

func implSetupOutput(f *os.File) (ok bool, cleanup func()) {
	fd := f.Fd()

	mode, err := getMode(fd)
	if err != nil {
		return false, func() {}
	}
	codepage, err := getOutputCP()
	if err != nil {
		return false, func() {}
	}
	modeOk := mode&cEnableVirtualTerminalProcessing != 0 // already has VT enabled
	codepageOk := codepage == 65001
	if modeOk && codepageOk {
		return true, func() {}
	}

	err = setMode(fd, mode|cEnableVirtualTerminalProcessing)
	if err != nil {
		return false, func() {}
	}
	err = setOutputCP(65001)
	if err != nil {
		// restore original immediately
		setMode(fd, mode)
		return false, func() {}
	}

	// return function that restores original console mode and output codepage
	return true, func() {
		setOutputCP(codepage)
		setMode(fd, mode)
	}
}
