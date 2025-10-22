// C:\_Projects_Go\AcousticLog\internal\io\wavsave.go

package io

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SaveWAV — обратная совместимость: пишет как EXCEEDED.
func SaveWAV(wavRoot string, base time.Time, rate int, pcm []byte) (string, error) {
	return SaveWAVKind(wavRoot, base, rate, pcm, EventKindExceeded)
}

// SaveWAVKind — сохраняет WAV в ...\WAV\<HH>\<Kind>\noise_YYYYMMDD_HHMMSS.mmm.wav
func SaveWAVKind(wavRoot string, base time.Time, rate int, pcm []byte, kind string) (string, error) {
	hourDir := filepath.Join(wavRoot, base.Format("15"), normalizeEventKind(kind))
	if err := os.MkdirAll(hourDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir hour/kind: %w", err)
	}
	filename := fmt.Sprintf("noise_%s.wav", base.Format("20060102_150405.000"))
	path := filepath.Join(hourDir, filename)

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create wav: %w", err)
	}
	defer f.Close()

	dataSize := uint32(len(pcm))
	overallSize := 36 + dataSize

	f.Write([]byte("RIFF"))
	_ = binary.Write(f, binary.LittleEndian, overallSize)
	f.Write([]byte("WAVE"))

	f.Write([]byte("fmt "))
	_ = binary.Write(f, binary.LittleEndian, uint32(16))
	_ = binary.Write(f, binary.LittleEndian, uint16(1))       // PCM
	_ = binary.Write(f, binary.LittleEndian, uint16(1))       // mono
	_ = binary.Write(f, binary.LittleEndian, uint32(rate))    // sample rate
	_ = binary.Write(f, binary.LittleEndian, uint32(rate*2))  // byte rate
	_ = binary.Write(f, binary.LittleEndian, uint16(2))       // block align
	_ = binary.Write(f, binary.LittleEndian, uint16(16))      // bits per sample

	f.Write([]byte("data"))
	_ = binary.Write(f, binary.LittleEndian, dataSize)
	_, _ = f.Write(pcm)
	return path, nil
}
