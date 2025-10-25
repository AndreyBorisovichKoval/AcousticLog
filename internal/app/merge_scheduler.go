package app

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	iomerge "acousticlog/internal/io"
)

func StartHourlyMergeAsync(ctx context.Context, cfg *Config, dayWavDir, hour string) {
	if cfg != nil && cfg.NoHourlyMerge { return }
	go func() {
		day := filepath.Base(filepath.Dir(dayWavDir)) // YYYY-MM-DD
		outName := fmt.Sprintf("merged_exceeded_%s_%s.wav", day, hour)
		opts := iomerge.MergeOptions{
			OutDir:   filepath.Join(dayWavDir, cfg.HourlyMergeOut),
			OutName:  outName,
			LockName: fmt.Sprintf("_merge_%s.lock", hour),
		}
		ctx2, cancel := context.WithTimeout(ctx, 10*time.Minute); defer cancel()
		if _, err := iomerge.MergeHour(ctx2, dayWavDir, hour, opts); err != nil {
			fmt.Println("[merge]", hour, "error:", err)
		}
	}()
}
