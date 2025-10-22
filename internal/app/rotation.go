// C:\_Projects_Go\AcousticLog\internal\app\rotation.go

package app

import (
	"fmt"
	"path/filepath"
	"sync/atomic"
	"time"

	iofs "acousticlog/internal/io"
	sysx "acousticlog/internal/sys"
)

func (a *App) rotateIfDateChanged(now time.Time) {
	newDate := now.Format("2006-01-02")
	if newDate == a.currentDate {
		return
	}

	// закрываем старые CSV
	if a.csvWriter != nil {
		a.csvWriter.Flush()
		_ = a.csvFile.Close()
	}
	if a.csvAllWriter != nil {
		a.csvAllWriter.Flush()
		_ = a.csvAllFile.Close()
	}

	root, csvDir, wavDir, err := iofs.EnsureOutDirForDate(newDate)
	if err != nil {
		fmt.Printf("%s[ROTATE ERROR] %v%s\n", sysx.ClrRed, err, sysx.ClrReset)
		return
	}

	file, writer, path, err := iofs.CreateCSV(csvDir, "sound_log", a.csvDelim, iofs.DefaultCSVHeader)
	if err != nil {
		fmt.Printf("%s[ROTATE ERROR] CSV(events): %v%s\n", sysx.ClrRed, err, sysx.ClrReset)
		return
	}
	allFile, allWriter, allPath, err := iofs.CreateCSV(csvDir, "sound_all", a.csvDelim, iofs.DefaultCSVHeader)
	if err != nil {
		_ = file.Close()
		fmt.Printf("%s[ROTATE ERROR] CSV(all): %v%s\n", sysx.ClrRed, err, sysx.ClrReset)
		return
	}

	a.outDirRoot, a.outDirCSV, a.outDirWAV = root, csvDir, wavDir
	a.csvFile, a.csvWriter, a.csvPath = file, writer, path
	a.csvAllFile, a.csvAllWriter, a.csvAllPath = allFile, allWriter, allPath
	a.currentDate = newDate

	// UI
	if !a.quiet && !a.liveNoClear {
		a.printLiveHeader()
	}
	fmt.Printf("%s🔁 Ротация по дате: теперь пишем в %s%s\n", sysx.ClrGreen, filepath.Clean(root), sysx.ClrReset)
}

func (a *App) printStats() {
	fmt.Printf("\n%s📊 Статистика:%s\n", sysx.ClrCyan, sysx.ClrReset)
	fmt.Printf("Обработано буферов: %d\n", atomic.LoadUint64(&a.stats.BuffersProcessed))
	fmt.Printf("События CSV: %d | Полный CSV: %d\n", atomic.LoadUint64(&a.stats.CSVEventsWritten), atomic.LoadUint64(&a.stats.CSVAllWritten))
	fmt.Printf("WAV файлов: %d | Ошибки WAV: %d\n", atomic.LoadUint64(&a.stats.WAVFilesSaved), atomic.LoadUint64(&a.stats.WAVErrors))
	fmt.Printf("Ошибки CSV: %d | Проверок диска: %d\n", atomic.LoadUint64(&a.stats.CSVErrors), atomic.LoadUint64(&a.stats.DiskChecks))
}
