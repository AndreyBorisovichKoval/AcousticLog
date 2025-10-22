// C:\_Projects_Go\AcousticLog\internal\app\liveui.go

package app

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	sysx "acousticlog/internal/sys"
)

func shortenPath(path string, keep int) string {
	if keep <= 0 || path == "" {
		return path
	}
	parts := strings.Split(path, string(filepath.Separator))
	if len(parts) <= keep {
		return path
	}
	return filepath.Join(parts[len(parts)-keep:]...)
}

func (a *App) printLiveHeader() {
	now := time.Now().In(a.loc)
	fmt.Println(sysx.ClrCyan + "================================================" + sysx.ClrReset)
	fmt.Println("🔊  " + sysx.ClrBold + "AcousticLog — Real-time Noise Monitor" + sysx.ClrReset)
	fmt.Println("🔊" + sysx.ClrCyan + "================================================" + sysx.ClrReset)
	fmt.Printf("📅 Local time: %s %s(TZ=%s)%s\n", now.Format("2006-01-02 15:04:05"), sysx.ClrGray, a.loc, sysx.ClrReset)
	fmt.Printf("⚙️  Аудио: %d Гц, 16-бит, Моно | Буфер: %d мс (%d байт/буфер)\n",
		a.Fmt.NSamplesPerSec, a.bufMs, int(a.Fmt.NAvgBytesPerSec)/1000*a.bufMs)
	fmt.Printf("📁 CSV (events) → %s\n", a.csvPath)
	fmt.Printf("📁 CSV (all)    → %s\n", a.csvAllPath)
	fmt.Printf("⚙️  Порог: день %.1f дБ, ночь %.1f дБ | калибровка %+0.1f дБ | импульс ≥ %.1f дБ\n",
		a.dayLimit, a.nightLimit, a.splOffset, a.impulseDelta)
	fmt.Printf("💾 Контроль диска: предупреждение < %d МБ, останов < %d МБ\n", a.diskWarnMB, a.diskStopMB)
	fmt.Printf("🖥️  Вывод: %s; предел строк: %d | Глубина пути WAV: %d\n",
		map[bool]string{true: "без очистки экрана", false: "с очисткой экрана"}[a.liveNoClear], a.maxLines, a.liveWavDepth)
	fmt.Println(sysx.ClrMagenta + "Timestamp                 Mode   dBFS     dB_SPL  Limit  Status   WAV_File" + sysx.ClrReset)
	fmt.Println(sysx.ClrMagenta + "--------------------------------------------------------------------------------" + sysx.ClrReset)
	a.linesPrinted = 0
}
