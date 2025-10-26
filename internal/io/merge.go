package io

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type MergeOptions struct {
	OutDir   string
	OutName  string // .wav
	LockName string
}

var (
	ErrNoClips     = errors.New("no exceeded clips to merge")
	ErrFmtMismatch = errors.New("wav format mismatch between clips")
)

func FindExceededClips(hourDir string) ([]string, error) {
	src := filepath.Join(hourDir, "EXCEEDED")
	entries, err := os.ReadDir(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		low := strings.ToLower(e.Name())
		if strings.HasSuffix(low, ".wav") && strings.HasPrefix(low, "noise_") {
			files = append(files, filepath.Join(src, e.Name()))
		}
	}
	sort.Strings(files)
	return files, nil
}

type wavInfo struct {
	fmtChunk []byte
	dataSize uint32
	dataOff  int64
}

func readWAVInfo(f *os.File) (wavInfo, error) {
	var info wavInfo
	hdr := make([]byte, 12)
	if _, err := io.ReadFull(f, hdr); err != nil {
		return info, err
	}
	if string(hdr[0:4]) != "RIFF" || string(hdr[8:12]) != "WAVE" {
		return info, errors.New("not a RIFF/WAVE file")
	}
	for {
		var ch [8]byte
		if _, err := io.ReadFull(f, ch[:]); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return info, err
		}
		tag := string(ch[0:4])
		size := binary.LittleEndian.Uint32(ch[4:8])
		switch tag {
		case "fmt ":
			buf := make([]byte, 8+size)
			copy(buf[0:8], ch[:])
			if _, err := io.ReadFull(f, buf[8:]); err != nil {
				return info, err
			}
			info.fmtChunk = buf
		case "data":
			info.dataSize = size
			pos, _ := f.Seek(0, io.SeekCurrent)
			info.dataOff = pos
			if _, err := f.Seek(int64(size), io.SeekCurrent); err != nil {
				return info, err
			}
		default:
			if _, err := f.Seek(int64(size), io.SeekCurrent); err != nil {
				return info, err
			}
		}
		if size%2 == 1 {
			if _, err := f.Seek(1, io.SeekCurrent); err != nil {
				return info, err
			}
		}
	}
	if len(info.fmtChunk) == 0 {
		return info, errors.New("missing fmt chunk")
	}
	return info, nil
}

func sameFmt(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	return bytes.Equal(a[8:], b[8:])
}

func writeWAV(out *os.File, fmtChunk []byte, totalData uint32, concat func(w io.Writer) error) error {
	riffSize := uint32(4+8) + uint32(len(fmtChunk)) + totalData
	if _, err := out.Write([]byte("RIFF")); err != nil {
		return err
	}
	if err := binary.Write(out, binary.LittleEndian, riffSize); err != nil {
		return err
	}
	if _, err := out.Write([]byte("WAVE")); err != nil {
		return err
	}
	if _, err := out.Write(fmtChunk[:8]); err != nil {
		return err
	}
	if _, err := out.Write(fmtChunk[8:]); err != nil {
		return err
	}
	if (len(fmtChunk)-8)%2 == 1 {
		if _, err := out.Write([]byte{0}); err != nil {
			return err
		}
	}
	if _, err := out.Write([]byte("data")); err != nil {
		return err
	}
	if err := binary.Write(out, binary.LittleEndian, totalData); err != nil {
		return err
	}
	if err := concat(out); err != nil {
		return err
	}
	if totalData%2 == 1 {
		if _, err := out.Write([]byte{0}); err != nil {
			return err
		}
	}
	return nil
}

func MergeHour(ctx context.Context, dayWavDir, hour string, opts MergeOptions) (string, error) {
	outDir := opts.OutDir
	if outDir == "" {
		outDir = filepath.Join(dayWavDir, "_Merged_Exceeded")
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", err
	}

	lock := filepath.Join(outDir, opts.LockName)
	if _, err := os.Stat(lock); err == nil {
		return "", errors.New("merge already in progress")
	}
	_ = os.WriteFile(lock, []byte("1"), 0644)
	defer os.Remove(lock)

	hourDir := filepath.Join(dayWavDir, hour)
	clips, err := FindExceededClips(hourDir)
	if err != nil {
		return "", err
	}
	if len(clips) == 0 {
		return "", ErrNoClips
	}

	outWav := filepath.Join(outDir, opts.OutName)
	if _, err := os.Stat(outWav); err == nil {
		_ = os.Remove(outWav)
	}

	// Read format from first and validate all, sum size
	f0, err := os.Open(clips[0])
	if err != nil {
		return "", err
	}
	info0, err := readWAVInfo(f0)
	f0.Close()
	if err != nil {
		return "", err
	}

	var total uint32 = info0.dataSize
	infos := make([]wavInfo, len(clips))
	infos[0] = info0
	for i := 1; i < len(clips); i++ {
		f, err := os.Open(clips[i])
		if err != nil {
			return "", err
		}
		info, err := readWAVInfo(f)
		f.Close()
		if err != nil {
			return "", err
		}
		if !sameFmt(info0.fmtChunk, info.fmtChunk) {
			return "", ErrFmtMismatch
		}
		infos[i] = info
		total += info.dataSize
	}

	tmp := outWav + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return "", err
	}
	defer func() { out.Close(); os.Remove(tmp) }()

	buf := make([]byte, 64*1024)
	concat := func(w io.Writer) error {
		for i, p := range clips {
			f, err := os.Open(p)
			if err != nil {
				return err
			}
			info := infos[i]
			if _, err := f.Seek(info.dataOff, io.SeekStart); err != nil {
				f.Close()
				return err
			}
			left := int64(info.dataSize)
			for left > 0 {
				ch := int64(len(buf))
				if left < ch {
					ch = left
				}
				n, er := f.Read(buf[:ch])
				if n > 0 {
					if _, ew := w.Write(buf[:n]); ew != nil {
						f.Close()
						return ew
					}
					left -= int64(n)
				}
				if er == io.EOF && left == 0 {
					break
				}
				if er != nil && er != io.EOF {
					f.Close()
					return er
				}
			}
			f.Close()
		}
		return nil
	}

	if err := writeWAV(out, info0.fmtChunk, total, concat); err != nil {
		return "", err
	}
	if err := out.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tmp, outWav); err != nil {
		return "", err
	}
	return outWav, nil
}
