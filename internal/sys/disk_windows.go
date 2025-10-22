// C:\_Projects_Go\AcousticLog\internal\sys\disk_windows.go

//go:build windows

package sys

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Returns MB available to the current user (not total free on volume)
func GetFreeDiskSpaceMB(path string) (uint64, error) {
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	proc := kernel32.NewProc("GetDiskFreeSpaceExW")
	var free, total, totalFree uint64
	rootPath := filepath.VolumeName(path) + string(filepath.Separator)
	r1, _, err := proc.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(rootPath))),
		uintptr(unsafe.Pointer(&free)),
		uintptr(unsafe.Pointer(&total)),
		uintptr(unsafe.Pointer(&totalFree)),
	)
	if r1 == 0 {
		if err != nil && err != syscall.Errno(0) {
			return 0, fmt.Errorf("GetDiskFreeSpaceExW %s: %w", rootPath, err)
		}
		return 0, fmt.Errorf("GetDiskFreeSpaceExW %s failed", rootPath)
	}
	return free / (1024 * 1024), nil
}
