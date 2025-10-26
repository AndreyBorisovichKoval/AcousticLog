package app

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	awin "acousticlog/internal/audio/winmm"
	iofs "acousticlog/internal/io"
	"acousticlog/internal/mathx"
	sysx "acousticlog/internal/sys"
)

const (
	DefaultDiskCheckInterval = 5 * time.Minute
	ShutdownTimeout          = 10 * time.Second
)

// mergeInfo — сводка по часовому мерджу для вывода в конце сессии.
type mergeInfo struct {
	Hour    string
	OutPath string
	Clips   int
	Err     error
}

func Run(cfg *Config) error {
	// дать доступ ParseFlags к реальным os.Args
	orig := os.Args
	_osArgs = func() []string { return orig }
	_setOsArgs = func(v []string) { os.Args = v }

	// sysx.EnableANSI() удалено, т.к. уже вызвано в build.PrintHeader

	// Time / TZ
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return fmt.Errorf("timezone %q: %w", cfg.Timezone, err)
	}

	// Dirs & CSV
	root, csvDir, wavDir, err := iofs.EnsureOutDirForDate(time.Now().In(loc).Format("2006-01-02"))
	if err != nil {
		return err
	}
	csvHeader := iofs.DefaultCSVHeader
	f1, w1, p1, err := iofs.CreateCSV(csvDir, "sound_log", cfg.CSVDelim, csvHeader)
	if err != nil {
		return fmt.Errorf("CSV(events): %w", err)
	}
	f2, w2, p2, err := iofs.CreateCSV(csvDir, "sound_all", cfg.CSVDelim, csvHeader)
	if err != nil {
		return fmt.Errorf("CSV(all): %w", err)
	}
	defer f1.Close()
	defer f2.Close()

	// Audio init
	fmtx := awin.WaveFormatPCM1ch16(cfg.SampleRate)
	h, err := awin.WaveInOpen(awin.WAVE_MAPPER, &fmtx)
	if err != nil {
		return fmt.Errorf("waveInOpen: %w", err)
	}

	bytesPerMs := int(fmtx.NAvgBytesPerSec) / 1000
	size := bytesPerMs * cfg.BufferMs
	if size < 512 {
		size = 512
	}
	bufs := make([]*buffer, 3)
	for i := 0; i < 3; i++ {
		mem := make([]byte, size)
		hdr := WAVEHDR{LpData: &mem[0], DwBufferLength: uint32(len(mem))}
		bufs[i] = &buffer{mem: mem, hdr: hdr}
	}

	app := &App{
		cfg:          cfg,
		Handle:       h,
		Fmt:          fmtx,
		Bufs:         bufs,
		csvFile:      f1,
		csvWriter:    w1,
		csvPath:      p1,
		csvAllFile:   f2,
		csvAllWriter: w2,
		csvAllPath:   p2,
		outDirRoot:   root,
		outDirCSV:    csvDir,
		outDirWAV:    wavDir,
		diskWarnMB:   cfg.DiskWarnMB,
		diskStopMB:   cfg.DiskStopMB,
		loc:          loc,
		splOffset:    cfg.SPLOffset,
		dayLimit:     cfg.DayLimit,
		nightLimit:   cfg.NightLimit,
		bufMs:        cfg.BufferMs,
		quiet:        cfg.QuietMode,
		nearMargin:   3.0,
		logAll:       cfg.LogAll,
		csvDelim:     cfg.CSVDelim,
		impulseDelta: cfg.ImpulseDelta,
		liveNoClear:  cfg.LiveNoClear,
		maxLines:     cfg.LiveLines,
		liveWavDepth: cfg.LiveWavDepth,
		chMainCSV:    make(chan []string, 256),
		chAllCSV:     make(chan []string, 512),
		chWAV:        make(chan wavTask, 256),
		currentDate:  time.Now().In(loc).Format("2006-01-02"),
	}

	// day start/end
	if v, err := parseHHMM(cfg.DayStartHHMM); err == nil {
		app.dayStart = v
	} else {
		return err
	}
	if v, err := parseHHMM(cfg.DayEndHHMM); err == nil {
		app.dayEnd = v
	} else {
		return err
	}

	// Prepare buffers
	for _, b := range app.Bufs {
		if err := awin.WaveInPrepareHeader(app.Handle, &b.hdr); err != nil {
			return fmt.Errorf("prepare: %w", err)
		}
		if err := awin.WaveInAddBuffer(app.Handle, &b.hdr); err != nil {
			return fmt.Errorf("addbuf: %w", err)
		}
	}

	// Start audio
	if err := awin.WaveInStart(app.Handle); err != nil {
		return fmt.Errorf("start: %w", err)
	}

	// Workers
	app.startWorkers()

	// Резюмирование незавершённых мерджей при старте
	ResumePendingMerges(context.Background(), app.cfg, app.outDirWAV)

	// UI header
	if !app.quiet {
		app.printLiveHeader()
	} else {
		fmt.Printf("Мониторинг… CSV(events) → %s | CSV(all) → %s | День %.1f / Ночь %.1f дБ | импульс ≥ %.1f дБ\n",
			app.csvPath, app.csvAllPath, app.dayLimit, app.nightLimit, app.impulseDelta)
	}

	// Disk ticker
	app.updateDiskStatus()
	diskCheckTicker := time.NewTicker(DefaultDiskCheckInterval)
	defer diskCheckTicker.Stop()

	// Часовой тикер + фиксация текущего часа
	lastHour := time.Now().In(app.loc).Format("15")
	hourTicker := time.NewTicker(10 * time.Second)
	defer hourTicker.Stop()

	// Сводка мерджей за сессию (по часам)
	var mergedHours []mergeInfo

	// Auto-stop timer (только если не /auto)
	var timer *time.Timer
	if !cfg.AutoMode {
		if dl, err := nextStopAt(loc, cfg.StopAtHHMM); err == nil {
			timer = new(time.Timer)
			*timer = *time.NewTimer(time.Until(dl))
			defer timer.Stop()
		}
	}

	// Signals
	intCh := make(chan os.Signal, 1)
	signal.Notify(intCh, os.Interrupt, syscall.SIGTERM)

	// Loop
	tick := time.NewTicker(20 * time.Millisecond)
	defer tick.Stop()

loop:
	for {
		select {
		case <-tick.C:
			for _, b := range app.Bufs {
				if (b.hdr.DwFlags & awin.WHDR_DONE) != 0 {
					app.process(b)
					b.hdr.DwFlags &^= awin.WHDR_DONE
					b.hdr.DwBytesRecorded = 0
					_ = awin.WaveInAddBuffer(app.Handle, &b.hdr)
				}
			}

		case <-diskCheckTicker.C:
			app.updateDiskStatus()
			if app.diskFreeMB < app.diskStopMB {
				now := time.Now().In(app.loc)
				record := []string{
					now.Format("2006-01-02 15:04:05.000"),
					"SYSTEM", "0.00", "0.00", "0.0",
					fmt.Sprintf("FATAL_DISK_SPACE_LEFT_%.1fMB", float64(app.diskFreeMB)),
					"NO_WAV",
				}
				_ = iofs.SafeWrite(app.csvWriter, record)
				_ = iofs.SafeWrite(app.csvAllWriter, record)
				fmt.Printf("\n%s[FATAL ERROR] КРИТИЧЕСКИ МАЛО МЕСТА (%.1f МБ). Аварийное завершение...%s\n",
					sysx.ClrRed, float64(app.diskFreeMB), sysx.ClrReset)
				break loop
			}

		// Автосклейка завершившегося часа + сводка
		case <-hourTicker.C:
			now := time.Now().In(app.loc)
			h := now.Format("15")
			if h != lastHour {
				prev := lastHour
				lastHour = h
				if !app.cfg.NoHourlyMerge {
					// Определяем переменные заранее, чтобы они были видны во всём блоке...
					var (
						out string
						n   int
						err error
					)

					// Всегда «только что завершившийся» час...
					// mergeTime := time.Now().In(app.loc).Add(-1 * time.Hour)
					mergeTime := now.Add(-1 * time.Hour)
					dayStr := mergeTime.Format("2006-01-02")

					// Папка именно того дня, к которому относится prev...
					_, _, dayWavDir, errDir := iofs.EnsureOutDirForDate(dayStr)
					if errDir != nil {
						mergedHours = append(mergedHours, mergeInfo{
							Hour: prev, Err: fmt.Errorf("ensure out dir for %s: %w", dayStr, errDir),
						})
					} else {
						out, n, err = StartHourlyMerge(context.Background(), app.cfg, dayWavDir, prev)
						mergedHours = append(mergedHours, mergeInfo{
							Hour: prev, OutPath: out, Clips: n, Err: err,
						})
					}
				}
			}

		case <-intCh:
			fmt.Println("\nCtrl+C — остановка…")
			break loop

		default:
			if timer != nil {
				select {
				case <-timer.C:
					fmt.Println("\nДостигнуто время авто-остановки — завершение…")
					timer = nil
					break loop
				default:
				}
			}
		}
	}

	// Синхронный мердж текущего часа на завершение + сводка
	if !app.cfg.NoHourlyMerge {
		now := time.Now().In(app.loc)
		hh := now.Format("15")
		out, n, err := StartHourlyMerge(context.Background(), app.cfg, app.outDirWAV, hh)
		mergedHours = append(mergedHours, mergeInfo{
			Hour:    hh,
			OutPath: out,
			Clips:   n,
			Err:     err,
		})
	}

	a := app
	a.shutdownWithStats(ShutdownTimeout, mergedHours)
	<-a.shutdownCh
	return nil
}

func (a *App) startWorkers() {
	// Main CSV writer (events)
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for rec := range a.chMainCSV {
			if err := iofs.SafeWrite(a.csvWriter, rec); err != nil {
				fmt.Printf("%s[CSV write error] %v%s\n", sysx.ClrRed, err, sysx.ClrReset)
				atomic.AddUint64(&a.stats.CSVErrors, 1)
			}
		}
	}()

	// All CSV writer
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for rec := range a.chAllCSV {
			if err := iofs.SafeWrite(a.csvAllWriter, rec); err != nil {
				fmt.Printf("%s[CSV flush error] %v%s\n", sysx.ClrRed, err, sysx.ClrReset)
				atomic.AddUint64(&a.stats.CSVErrors, 1)
			}
		}
	}()

	// WAV workers (3)
	for i := 0; i < 3; i++ {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			for task := range a.chWAV {
				path, err := iofs.SaveWAVKind(a.outDirWAV, task.when, task.rate, task.pcm, task.kind)
				if err != nil {
					fmt.Printf("%s[WAV error] %v%s\n", sysx.ClrRed, err, sysx.ClrReset)
					atomic.AddUint64(&a.stats.WAVErrors, 1)
					continue
				}
				atomic.AddUint64(&a.stats.WAVFilesSaved, 1)
				if task.after != nil {
					task.after(path)
				}
			}
		}()
	}
}

func (a *App) updateDiskStatus() {
	freeMB, err := sysx.GetFreeDiskSpaceMB(a.outDirRoot)
	if err != nil {
		log.Printf("%s[DISK ERROR] Ошибка проверки диска %s: %v%s", sysx.ClrRed, a.outDirRoot, err, sysx.ClrReset)
		a.diskFreeMB = 0
		return
	}
	a.diskFreeMB = freeMB
	atomic.AddUint64(&a.stats.DiskChecks, 1)
}

func (a *App) process(b *buffer) {
	if a.isShutting.Load() {
		return
	}
	atomic.AddUint64(&a.stats.BuffersProcessed, 1)

	n := int(b.hdr.DwBytesRecorded)
	if n <= 0 || n > len(b.mem) {
		return
	}
	raw := b.mem[:n]
	samples := mathx.BytesToInt16LE(raw)
	if len(samples) == 0 {
		return
	}
	rms := mathx.CalcRMSInt16(samples)
	if rms <= 0 {
		return
	}

	dbFS := 20 * math.Log10(rms)
	dbSPL := 20*math.Log10(rms) + a.splOffset
	if dbSPL < 0 {
		dbSPL = 0
	}

	now := time.Now().In(a.loc)
	a.rotateIfDateChanged(now)
	mode, lim := a.currentLimit(now)

	exceeded := dbSPL >= lim
	impulse := a.prevInit && (dbSPL-a.prevDbSPL) >= a.impulseDelta

	color := sysx.ClrGray
	status := "OK"
	switch {
	case exceeded:
		color = sysx.ClrRed
		status = "EXCEEDED"
	case impulse:
		color = sysx.ClrCyan
		status = "IMPULSE"
	case dbSPL >= lim-a.nearMargin:
		color = sysx.ClrYellow
		status = "NEAR"
	}
	// Определяем папку события для WAV
	kind := ""
	switch status {
	case "EXCEEDED":
		kind = "EXCEEDED"
	case "IMPULSE":
		kind = "IMPULSE"
	}

	freeMB := a.diskFreeMB
	canSaveWAV := freeMB > a.diskWarnMB

	var wavFilename string
	if exceeded || impulse {
		if canSaveWAV {
			wavFilename = filepath.Join(a.outDirWAV, now.Format("15"), kind, fmt.Sprintf("noise_%s.wav", now.Format("20060102_150405.000")))
		} else {
			wavFilename = fmt.Sprintf("DISK_LOW_SPACE_%.1fMB", float64(freeMB))
		}
	}

	if !a.quiet {
		shortWav := ""
		if wavFilename != "" {
			if strings.HasPrefix(wavFilename, "DISK_LOW_SPACE") {
				shortWav = sysx.ClrYellow + "DISK LOW" + sysx.ClrReset
			} else {
				shortWav = shortenPath(wavFilename, a.liveWavDepth)
			}
		}
		fmt.Printf("%s%-23s  %-5s %7.1f  %6.1f  %5.1f  %-7s %s%s\n",
			color, now.Format("2006-01-02 15:04:05.000"), mode, dbFS, dbSPL, lim, status, shortWav, sysx.ClrReset)

		if !a.liveNoClear {
			a.linesPrinted++
			if a.linesPrinted >= a.maxLines {
				a.printLiveHeader()
			}
		}

		if !canSaveWAV && (exceeded || impulse) {
			fmt.Printf("%s[DISK WARNING] МЕСТО ЗАКАНЧИВАЕТСЯ: %.1f МБ. WAV-файлы НЕ ЗАПИСАНЫ.%s\n",
				sysx.ClrYellow, float64(freeMB), sysx.ClrReset)
		}
	}

	ts := now.Format("2006-01-02 15:04:05.000")
	row := []string{ts, mode,
		strconv.FormatFloat(dbFS, 'f', 2, 64), strconv.FormatFloat(dbSPL, 'f', 2, 64),
		strconv.FormatFloat(lim, 'f', 1, 64), status, wavFilename}

	a.chAllCSV <- row
	atomic.AddUint64(&a.stats.CSVAllWritten, 1)
	if exceeded || impulse || a.logAll {
		a.chMainCSV <- row
		atomic.AddUint64(&a.stats.CSVEventsWritten, 1)
	}

	if (exceeded || impulse) && canSaveWAV {
		select {
		case a.chWAV <- wavTask{when: now, rate: int(a.Fmt.NSamplesPerSec), pcm: append([]byte(nil), raw...), kind: kind}:
		default:
			// дроп без блокировки
		}
	}

	a.prevDbSPL = dbSPL
	a.prevInit = true
}

// shutdownWithStats — завершение + расширенная сводка мерджей по часам (кол-во клипов и размер).
func (a *App) shutdownWithStats(timeout time.Duration, mergedHours []mergeInfo) {
	if a.isShutting.Swap(true) {
		return
	}
	fmt.Printf("\n%s🔄 Завершение работы...%s\n", sysx.ClrYellow, sysx.ClrReset)

	_ = awin.WaveInStop(a.Handle)
	for _, b := range a.Bufs {
		_ = awin.WaveInUnprepareHeader(a.Handle, &b.hdr)
	}
	_ = awin.WaveInClose(a.Handle)

	close(a.chMainCSV)
	close(a.chAllCSV)
	close(a.chWAV)

	done := make(chan struct{})
	go func() { a.wg.Wait(); close(done) }()

	select {
	case <-done:
		fmt.Printf("%s✅ Все воркеры завершили работу%s\n", sysx.ClrGreen, sysx.ClrReset)
	case <-time.After(timeout):
		fmt.Printf("%s⚠️  Завершение по таймауту%s\n", sysx.ClrYellow, sysx.ClrReset)
	}

	if a.csvWriter != nil {
		a.csvWriter.Flush()
	}
	if a.csvAllWriter != nil {
		a.csvAllWriter.Flush()
	}

	// Базовая статистика
	a.printStats()
	fmt.Printf("📄 CSV(events): %s\n📄 CSV(all):    %s\n", a.csvPath, a.csvAllPath)

	// Новая секция: сводка по часовым мерджам (по ходу сессии + финальный)
	if len(mergedHours) > 0 {
		fmt.Printf("\n%s🧩 Hourly merge summary:%s\n", sysx.ClrCyan, sysx.ClrReset)
		totalClips := 0
		var totalBytes int64
		totalOk := 0

		for _, mi := range mergedHours {
			if mi.Err != nil {
				fmt.Printf(" - %s: ERROR: %v\n", mi.Hour, mi.Err)
				continue
			}
			base := "(no output)"
			sizeMB := 0.0
			if mi.OutPath != "" {
				if fi, err := os.Stat(mi.OutPath); err == nil {
					totalBytes += fi.Size()
					sizeMB = float64(fi.Size()) / (1024.0 * 1024.0)
				}
				base = filepath.Base(mi.OutPath)
			}
			fmt.Printf(" - %s: %d clips → %s (%.2f MB)\n", mi.Hour, mi.Clips, base, sizeMB)
			totalClips += mi.Clips
			totalOk++
		}
		fmt.Printf("Всего часов склеено: %d | Всего фрагментов объединено: %d | Суммарный объём: %.2f MB\n",
			totalOk, totalClips, float64(totalBytes)/(1024.0*1024.0))
	} else {
		fmt.Println("\n🧩 Hourly merge summary: нет выполненных склеек за сессию")
	}

	if a.shutdownCh == nil {
		a.shutdownCh = make(chan struct{})
	}
	close(a.shutdownCh)
}
