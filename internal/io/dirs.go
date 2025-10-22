// C:\_Projects_Go\AcousticLog\internal\io\dirs.go

package io

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func EnsureOutDir() (root, csvDir, wavDir string, err error) {
	return EnsureOutDirForDate(time.Now().Format("2006-01-02"))
}

func EnsureOutDirForDate(dateStr string) (root, csvDir, wavDir string, err error) {
	base := `C:\DataSound_Temp`
	if st, e := os.Stat(`D:\`); e == nil && st.IsDir() {
		base = `D:\DataSound_Temp`
	}
	root = filepath.Join(base, dateStr)
	csvDir = filepath.Join(root, "CSV")
	wavDir = filepath.Join(root, "WAV")
	if err = os.MkdirAll(csvDir, 0o755); err != nil {
		return "", "", "", fmt.Errorf("mkdir %s: %w", csvDir, err)
	}
	if err = os.MkdirAll(wavDir, 0o755); err != nil {
		return "", "", "", fmt.Errorf("mkdir %s: %w", wavDir, err)
	}
	return
}
