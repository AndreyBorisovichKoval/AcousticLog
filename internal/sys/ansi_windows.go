// internal/sys/ansi_windows.go

//go:build windows

package sys

import (
	"os/exec"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	ClrReset   = "\x1b[0m"
	ClrBold    = "\x1b[1m"
	ClrRed     = "\x1b[31m"
	ClrGreen   = "\x1b[32m"
	ClrYellow  = "\x1b[33m"
	ClrBlue    = "\x1b[34m"
	ClrMagenta = "\x1b[35m"
	ClrCyan    = "\x1b[36m"
	ClrGray    = "\x1b[90m"
)

func EnableANSI() {
	if runtime.GOOS != "windows" {
		return
	}
	const ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x0004
	h, err := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	if err != nil {
		return
	}
	var mode uint32
	k32 := windows.NewLazySystemDLL("kernel32.dll")
	get := k32.NewProc("GetConsoleMode")
	set := k32.NewProc("SetConsoleMode")
	if r1, _, _ := get.Call(uintptr(h), uintptr(unsafe.Pointer(&mode))); r1 == 0 {
		return
	}
	mode |= ENABLE_VIRTUAL_TERMINAL_PROCESSING
	_, _, _ = set.Call(uintptr(h), uintptr(mode))
}

func ClearConsole() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	_ = cmd.Run()
}
