// C:\_Projects_Go\AcousticLog\internal\io\wavsave.go

package io

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func SaveWAV(wavRoot string, base time.Time, rate int, pcm []byte) (string, error) {
	hourDir := filepath.Join(wavRoot, base.Format("15"))
	if err := os.MkdirAll(hourDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir hour: %w", err)
	}
	filename := fmt.Sprintf("noise_%s.wav", base.Format("20060102_150405.000"))
	path := filepath.Join(hourDir, filename)

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	dataSize := uint32(len(pcm))
	riffSize := 36 + dataSize
	f.Write([]byte("RIFF"))
	_ = binary.Write(f, binary.LittleEndian, riffSize)
	f.Write([]byte("WAVE"))
	f.Write([]byte("fmt "))
	_ = binary.Write(f, binary.LittleEndian, uint32(16))
	_ = binary.Write(f, binary.LittleEndian, uint16(1))
	_ = binary.Write(f, binary.LittleEndian, uint16(1))
	_ = binary.Write(f, binary.LittleEndian, uint32(rate))
	_ = binary.Write(f, binary.LittleEndian, uint32(rate*2))
	_ = binary.Write(f, binary.LittleEndian, uint16(2))
	_ = binary.Write(f, binary.LittleEndian, uint16(16))
	f.Write([]byte("data"))
	_ = binary.Write(f, binary.LittleEndian, dataSize)
	_, _ = f.Write(pcm)
	return path, nil
}
