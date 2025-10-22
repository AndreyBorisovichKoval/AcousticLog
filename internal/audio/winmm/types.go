// C:\_Projects_Go\AcousticLog\internal\audio\winmm\types.go

package winmm

type WAVEFORMATEX struct {
	WFormatTag      uint16
	NChannels       uint16
	NSamplesPerSec  uint32
	NAvgBytesPerSec uint32
	NBlockAlign     uint16
	WBitsPerSample  uint16
	CbSize          uint16
}

type WAVEHDR struct {
	LpData          *byte
	DwBufferLength  uint32
	DwBytesRecorded uint32
	DwUser          uintptr
	DwFlags         uint32
	DwLoops         uint32
	LpNext          uintptr
	Reserved        uintptr
}

type Buffer struct {
	Mem []byte
	Hdr WAVEHDR
}

func WaveFormatPCM1ch16(sampleRate int) WAVEFORMATEX {
	return WAVEFORMATEX{
		WFormatTag: 1, NChannels: 1, WBitsPerSample: 16,
		NSamplesPerSec:  uint32(sampleRate),
		NBlockAlign:     2,
		NAvgBytesPerSec: uint32(sampleRate * 2),
	}
}
