// C:\_Projects_Go\AcousticLog\internal\app\types.go

package app

import (
	"encoding/csv"
	"os"
	"sync"
	"sync/atomic"
	"time"

	awin "acousticlog/internal/audio/winmm"
)

type AppStats struct {
	BuffersProcessed uint64
	CSVEventsWritten uint64
	CSVAllWritten    uint64
	WAVFilesSaved    uint64
	WAVErrors        uint64
	CSVErrors        uint64
	DiskChecks       uint64
}

type App struct {
	cfg *Config

	// audio
	Handle uintptr
	Fmt    WAVEFORMATEX
	Bufs   []*buffer

	// CSV
	csvFile      *os.File
	csvWriter    *csv.Writer
	csvPath      string
	csvAllFile   *os.File
	csvAllWriter *csv.Writer
	csvAllPath   string

	// dirs
	outDirRoot string
	outDirCSV  string
	outDirWAV  string

	// disk
	diskFreeMB     uint64
	diskWarnMB     uint64
	diskStopMB     uint64
	diskCheckMutex sync.Mutex

	// time & limits
	loc        *time.Location
	splOffset  float64
	dayLimit   float64
	nightLimit float64
	dayStart   int
	dayEnd     int

	// state
	stopCh       chan struct{}
	shutdownCh   chan struct{}
	isShutting   atomic.Bool
	prevDbSPL    float64
	prevInit     bool
	quiet        bool
	linesPrinted int
	nearMargin   float64
	logAll       bool
	csvDelim     rune
	impulseDelta float64

	// live UI
	bufMs        int
	liveNoClear  bool
	maxLines     int
	liveWavDepth int

	// pipelines
	chMainCSV chan []string
	chAllCSV  chan []string
	chWAV     chan wavTask
	wg        sync.WaitGroup

	// rotation
	currentDate string

	// stats
	stats AppStats
}

// Алиасы ровно на типы пакета winmm (совместимость без кастов)
type WAVEFORMATEX = awin.WAVEFORMATEX
type WAVEHDR = awin.WAVEHDR

type buffer struct {
	mem []byte
	hdr WAVEHDR
}

type wavTask struct {
	when time.Time
	rate int
	pcm  []byte
}
