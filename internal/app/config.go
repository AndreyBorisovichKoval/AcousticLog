// C:\_Projects_Go\AcousticLog\internal\app\config.go

package app

import (
	"errors"
	"flag"
)

type Config struct {
	NoHourlyMerge  bool
	HourlyMergeOut string

	// thresholds & logic
	SPLOffset    float64
	DayLimit     float64
	NightLimit   float64
	DayStartHHMM string
	DayEndHHMM   string
	ImpulseDelta float64
	LogAll       bool

	// audio
	SampleRate int
	BufferMs   int

	// runtime
	Timezone   string
	StopAtHHMM string
	AutoMode   bool
	QuietMode  bool

	// live UI
	LiveLines    int
	LiveNoClear  bool
	LiveWavDepth int

	// disk
	DiskWarnMB uint64
	DiskStopMB uint64

	// csv
	CSVDelim rune
}

func ParseFlags() (*Config, error) {
	// подсос реальных os.Args через замыкание (см. Run)
	autoMode := hasToken("/auto")
	if autoMode {
		stripToken("/auto")
	}
	if hasToken("/run") {
		stripToken("/run")
	}
	quiet := hasToken("/quiet")
	if quiet {
		stripToken("/quiet")
	}

	// --- флаги
	spl := flag.Float64("spl-offset", 114, "")
	day := flag.Float64("day-limit", 55, "")
	night := flag.Float64("night-limit", 45, "")
	dayStart := flag.String("day-start", "07:00", "")
	dayEnd := flag.String("day-end", "23:00", "")
	stopAt := flag.String("stop-at", "02:00", "")

	sr := flag.Int("samplerate", 16000, "")
	bufms := flag.Int("duration", 200, "")
	tz := flag.String("tz", "Asia/Dushanbe", "")

	logAll := flag.Bool("log-all", false, "")
	impulse := flag.Float64("impulse-delta", 15, "")

	warnMB := flag.Uint64("disk-warn-mb", 100, "")
	stopMB := flag.Uint64("disk-stop-mb", 50, "")

	lines := flag.Int("live-lines", 70, "")
	noClear := flag.Bool("live-no-clear", false, "")
	appendL := flag.Bool("append-live", false, "")
	csvDelimStr := flag.String("csv-delim", ";", "")
	wavDepth := flag.Int("live-wav-depth", 3, "")

	consolePage := flag.Bool("console-page", false, "")
	consolePageSize := flag.Int("console-page-size", 70, "")

	noHourly := flag.Bool("no-hourly-merge", false, "")
	hourlyOut := flag.String("hourly-merge-out", "_Merged_Exceeded", "")

	flag.Parse()

	// --- приведение поведения консоли
	// 1) Если включена постраничность — принудительно noClear=false и назначаем размер страницы.
	if *consolePage {
		*noClear = false
		if consolePageSize != nil && *consolePageSize > 0 {
			*lines = *consolePageSize
		} else {
			*lines = 70
		}
	}
	// 2) Если включён append-live — всегда не очищаем экран (append-режим подразумевает накопление).
	if *appendL {
		*noClear = true
	}

	// --- валидации
	if *day <= 0 || *night <= 0 || *day < *night {
		return nil, errors.New("некорректные пороги: day>0, night>0, day>=night")
	}
	if *impulse <= 0 {
		return nil, errors.New("impulse-delta должен быть > 0")
	}
	if *wavDepth < 2 {
		*wavDepth = 2
	} else if *wavDepth > 4 {
		*wavDepth = 4
	}

	// --- csv delimiter: корректно берём первую руну (а не первый байт)
	delim := ';'
	if *csvDelimStr != "" {
		rs := []rune(*csvDelimStr)
		if len(rs) > 0 {
			delim = rs[0]
		}
	}

	return &Config{
		// thresholds & logic
		SPLOffset:    *spl,
		DayLimit:     *day,
		NightLimit:   *night,
		DayStartHHMM: *dayStart,
		DayEndHHMM:   *dayEnd,
		ImpulseDelta: *impulse,
		LogAll:       *logAll,

		// audio
		SampleRate: *sr,
		BufferMs:   *bufms,

		// runtime
		Timezone:   *tz,
		StopAtHHMM: *stopAt,
		AutoMode:   autoMode,
		QuietMode:  quiet,

		// live UI
		LiveLines:    *lines,
		LiveNoClear:  *noClear,
		LiveWavDepth: *wavDepth,

		// disk
		DiskWarnMB: *warnMB,
		DiskStopMB: *stopMB,

		// csv
		CSVDelim: delim,

		// hourly merge
		NoHourlyMerge:  *noHourly,
		HourlyMergeOut: *hourlyOut,
	}, nil
}

// // C:\_Projects_Go\AcousticLog\internal\app\config.go

// package app

// import (
// 	"errors"
// 	"flag"
// )

// type Config struct {
// 	NoHourlyMerge  bool
// 	HourlyMergeOut string
// 	// thresholds & logic
// 	SPLOffset    float64
// 	DayLimit     float64
// 	NightLimit   float64
// 	DayStartHHMM string
// 	DayEndHHMM   string
// 	ImpulseDelta float64
// 	LogAll       bool

// 	// audio
// 	SampleRate int
// 	BufferMs   int

// 	// runtime
// 	Timezone   string
// 	StopAtHHMM string
// 	AutoMode   bool
// 	QuietMode  bool

// 	// live UI
// 	LiveLines    int
// 	LiveNoClear  bool
// 	LiveWavDepth int

// 	// disk
// 	DiskWarnMB uint64
// 	DiskStopMB uint64

// 	// csv
// 	CSVDelim rune
// }

// func ParseFlags() (*Config, error) {
// 	// подсос реальных os.Args через замыкание (см. Run)
// 	autoMode := hasToken("/auto")
// 	if autoMode {
// 		stripToken("/auto")
// 	}
// 	if hasToken("/run") {
// 		stripToken("/run")
// 	}
// 	quiet := hasToken("/quiet")
// 	if quiet {
// 		stripToken("/quiet")
// 	}

// 	spl := flag.Float64("spl-offset", 114, "")
// 	day := flag.Float64("day-limit", 55, "")
// 	night := flag.Float64("night-limit", 45, "")
// 	dayStart := flag.String("day-start", "07:00", "")
// 	dayEnd := flag.String("day-end", "23:00", "")
// 	stopAt := flag.String("stop-at", "02:00", "")
// 	sr := flag.Int("samplerate", 16000, "")
// 	bufms := flag.Int("duration", 200, "")
// 	tz := flag.String("tz", "Asia/Dushanbe", "")
// 	logAll := flag.Bool("log-all", false, "")
// 	impulse := flag.Float64("impulse-delta", 15, "")
// 	warnMB := flag.Uint64("disk-warn-mb", 100, "")
// 	stopMB := flag.Uint64("disk-stop-mb", 50, "")
// 	lines := flag.Int("live-lines", 70, "")
// 	noClear := flag.Bool("live-no-clear", false, "")
// 	appendL := flag.Bool("append-live", false, "")
// 	csvDelimStr := flag.String("csv-delim", ";", "")
// 	wavDepth := flag.Int("live-wav-depth", 3, "")
// 	consolePage := flag.Bool("console-page", false, "")
// 	consolePageSize := flag.Int("console-page-size", 70, "")
// 	noHourly := flag.Bool("no-hourly-merge", false, "")
// 	hourlyOut := flag.String("hourly-merge-out", "_Merged_Exceeded", "")

// 	flag.Parse()
// 	// map console-page to live behavior
// 	if *consolePage {
// 		*noClear = false
// 		if consolePageSize != nil && *consolePageSize > 0 {
// 			*lines = *consolePageSize
// 		} else {
// 			*lines = 70
// 		}
// 	} else {
// 		*noClear = true
// 	}
// 	if *appendL {
// 		*noClear = true
// 	}
// 	if *appendL {
// 		*noClear = true
// 	}
// 	if *day <= 0 || *night <= 0 || *day < *night {
// 		return nil, errors.New("некорректные пороги: day>0, night>0, day>=night")
// 	}
// 	if *impulse <= 0 {
// 		return nil, errors.New("impulse-delta должен быть > 0")
// 	}
// 	if *wavDepth < 2 {
// 		*wavDepth = 2
// 	} else if *wavDepth > 4 {
// 		*wavDepth = 4
// 	}

// 	delim := ';'
// 	if *csvDelimStr != "" {
// 		delim = rune((*csvDelimStr)[0])
// 	}

// 	return &Config{
// 		SPLOffset: *spl, DayLimit: *day, NightLimit: *night,
// 		DayStartHHMM: *dayStart, DayEndHHMM: *dayEnd, StopAtHHMM: *stopAt,
// 		SampleRate: *sr, BufferMs: *bufms,
// 		Timezone: *tz, LogAll: *logAll, ImpulseDelta: *impulse,
// 		DiskWarnMB: *warnMB, DiskStopMB: *stopMB,
// 		LiveLines: *lines, LiveNoClear: *noClear, LiveWavDepth: *wavDepth,
// 		AutoMode: autoMode, QuietMode: quiet,
// 		CSVDelim:       delim,
// 		NoHourlyMerge:  *noHourly,
// 		HourlyMergeOut: *hourlyOut,
// 	}, nil
// }
