// C:\_Projects_Go\AcousticLog\internal\io\csvlog.go

package io

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

var DefaultCSVHeader = []string{"Timestamp", "Mode", "dBFS", "dB_SPL", "Limit", "Status", "WAV_File"}

func CreateCSV(dir, prefix string, delim rune, header []string) (*os.File, *csv.Writer, string, error) {
	name := fmt.Sprintf("%s_%s.csv", prefix, time.Now().Format("20060102_150405"))
	path := dir + string(os.PathSeparator) + name
	f, err := os.Create(path)
	if err != nil {
		return nil, nil, "", err
	}
	if _, err := f.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		f.Close()
		return nil, nil, "", err
	}
	w := csv.NewWriter(f)
	w.Comma = delim
	w.UseCRLF = true
	if err := w.Write(header); err != nil {
		f.Close()
		return nil, nil, "", err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		f.Close()
		return nil, nil, "", err
	}
	return f, w, path, nil
}

func SafeWrite(w *csv.Writer, rec []string) error {
	if err := w.Write(rec); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}
