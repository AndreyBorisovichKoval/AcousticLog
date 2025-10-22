// C:\_Projects_Go\AcousticLog\internal\app\config.go

package app

import (
	"errors"
	"flag"
)

type Config struct {
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

	spl := flag.Float64("spl-offset", 90, "")
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

	flag.Parse()
	if *appendL {
		*noClear = true
	}
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

	delim := ';'
	if *csvDelimStr != "" {
		delim = rune((*csvDelimStr)[0])
	}

	return &Config{
		SPLOffset: *spl, DayLimit: *day, NightLimit: *night,
		DayStartHHMM: *dayStart, DayEndHHMM: *dayEnd, StopAtHHMM: *stopAt,
		SampleRate: *sr, BufferMs: *bufms,
		Timezone: *tz, LogAll: *logAll, ImpulseDelta: *impulse,
		DiskWarnMB: *warnMB, DiskStopMB: *stopMB,
		LiveLines: *lines, LiveNoClear: *noClear, LiveWavDepth: *wavDepth,
		AutoMode: autoMode, QuietMode: quiet,
		CSVDelim: delim,
	}, nil
}
