package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	iomerge "acousticlog/internal/io"
)

// StartHourlyMerge — синхронная склейка для указанного часа.
// Возвращает полный путь итогового WAV, количество склеенных фрагментов (по каталогу EXCEEDED) и ошибку.
func StartHourlyMerge(ctx context.Context, cfg *Config, dayWavDir, hour string) (string, int, error) {
	if cfg != nil && cfg.NoHourlyMerge {
		return "", 0, nil
	}

	// Считаем, сколько исходных клипов (WAV) есть в каталоге EXCEEDED за этот час.
	// Это и будет числом «clips» в сводке.
	clips, _ := countClipsEXCEEDED(dayWavDir, hour)

	day := filepath.Base(filepath.Dir(dayWavDir)) // YYYY-MM-DD
	outName := fmt.Sprintf("merged_exceeded_%s_%s.wav", day, hour)

	opts := iomerge.MergeOptions{
		OutDir:   filepath.Join(dayWavDir, cfg.HourlyMergeOut), // ...\WAV\_Merged_Exceeded
		OutName:  outName,
		LockName: fmt.Sprintf("_merge_%s.lock", hour), // _merge_19.lock
	}

	ctx2, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	out, err := iomerge.MergeHour(ctx2, dayWavDir, hour, opts)
	if err != nil {
		fmt.Println("[merge]", hour, "error:", err)
		return out, clips, err
	}
	fmt.Println("[merge] hour", hour, "completed:", filepath.Join(opts.OutDir, opts.OutName))
	return out, clips, nil
}

// ResumePendingMerges — достраивает «застрявшие» часы по lock-файлам (синхронно).
func ResumePendingMerges(ctx context.Context, cfg *Config, dayWavDir string) {
	if cfg != nil && cfg.NoHourlyMerge {
		return
	}
	outDir := filepath.Join(dayWavDir, cfg.HourlyMergeOut)

	entries, err := os.ReadDir(outDir)
	if err != nil {
		// Папка может ещё не существовать
		return
	}

	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "_merge_") && strings.HasSuffix(name, ".lock") {
			hour := strings.TrimSuffix(strings.TrimPrefix(name, "_merge_"), ".lock")
			_, _, _ = StartHourlyMerge(ctx, cfg, dayWavDir, hour)
		}
	}
}

// countClipsEXCEEDED — сколько файлов *.wav в каталоге EXCEEDED за указанный час.
// Если каталога нет — возвращает 0 без ошибки (это нормальный случай).
func countClipsEXCEEDED(dayWavDir, hour string) (int, error) {
	dir := filepath.Join(dayWavDir, hour, "EXCEEDED")
	ents, err := os.ReadDir(dir)
	if err != nil {
		return 0, nil
	}
	n := 0
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		if strings.HasSuffix(name, ".wav") {
			n++
		}
	}
	return n, nil
}
