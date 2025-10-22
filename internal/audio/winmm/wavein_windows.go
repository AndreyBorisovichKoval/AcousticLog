// C:\_Projects_Go\AcousticLog\internal\audio\winmm\wavein_windows.go

//go:build windows

package winmm

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	WAVE_MAPPER      = 0xFFFFFFFF
	MMSYSERR_NOERROR = 0
	CALLBACK_NULL    = 0
	WHDR_DONE        = 0x00000001
)

var (
	winmm               = windows.NewLazySystemDLL("winmm.dll")
	procWaveInOpen      = winmm.NewProc("waveInOpen")
	procWaveInClose     = winmm.NewProc("waveInClose")
	procWaveInPrepare   = winmm.NewProc("waveInPrepareHeader")
	procWaveInUnprepare = winmm.NewProc("waveInUnprepareHeader")
	procWaveInAddBuffer = winmm.NewProc("waveInAddBuffer")
	procWaveInStart     = winmm.NewProc("waveInStart")
	procWaveInStop      = winmm.NewProc("waveInStop")
)

func WaveInOpen(deviceID uint32, pwfx *WAVEFORMATEX) (uintptr, error) {
	var h uintptr
	r0, _, _ := procWaveInOpen.Call(uintptr(unsafe.Pointer(&h)), uintptr(deviceID),
		uintptr(unsafe.Pointer(pwfx)), 0, 0, uintptr(CALLBACK_NULL))
	if r0 != MMSYSERR_NOERROR {
		return 0, fmt.Errorf("waveInOpen failed: %d", r0)
	}
	return h, nil
}
func WaveInClose(h uintptr) error {
	r0, _, _ := procWaveInClose.Call(h)
	if r0 != MMSYSERR_NOERROR {
		return fmt.Errorf("waveInClose failed: %d", r0)
	}
	return nil
}
func WaveInPrepareHeader(h uintptr, pwh *WAVEHDR) error {
	r0, _, _ := procWaveInPrepare.Call(h, uintptr(unsafe.Pointer(pwh)), unsafe.Sizeof(*pwh))
	if r0 != MMSYSERR_NOERROR {
		return fmt.Errorf("waveInPrepareHeader failed: %d", r0)
	}
	return nil
}
func WaveInUnprepareHeader(h uintptr, pwh *WAVEHDR) error {
	r0, _, _ := procWaveInUnprepare.Call(h, uintptr(unsafe.Pointer(pwh)), unsafe.Sizeof(*pwh))
	if r0 != MMSYSERR_NOERROR {
		return fmt.Errorf("waveInUnprepareHeader failed: %d", r0)
	}
	return nil
}
func WaveInAddBuffer(h uintptr, pwh *WAVEHDR) error {
	r0, _, _ := procWaveInAddBuffer.Call(h, uintptr(unsafe.Pointer(pwh)), unsafe.Sizeof(*pwh))
	if r0 != MMSYSERR_NOERROR {
		return fmt.Errorf("waveInAddBuffer failed: %d", r0)
	}
	return nil
}
func WaveInStart(h uintptr) error {
	r0, _, _ := procWaveInStart.Call(h)
	if r0 != MMSYSERR_NOERROR {
		return fmt.Errorf("waveInStart failed: %d", r0)
	}
	return nil
}
func WaveInStop(h uintptr) error {
	r0, _, _ := procWaveInStop.Call(h)
	if r0 != MMSYSERR_NOERROR {
		return fmt.Errorf("waveInStop failed: %d", r0)
	}
	return nil
}
