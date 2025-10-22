// C:\_Projects_Go\AcousticLog\internal\build\meta.go

package build

import (
	"fmt"
	"time"

	sysx "acousticlog/internal/sys"
)

const (
	AppVersion = "v1.03"
	Author     = "Andrey Koval (57)"
	BuildDate  = "2025-10-19"
)

var BuildTime = "unknown" // -ldflags "-X acousticlog/internal/build.BuildTime=YYYY-MM-DD HH:MM:SS"

func PrintHeader(tz string) {
	// Включаем ANSI и очищаем консоль здесь, чтобы шапка была цветной
	sysx.EnableANSI()
	sysx.ClearConsole()

	clrReset := sysx.ClrReset
	clrBold := sysx.ClrBold
	clrCyan := sysx.ClrCyan
	clrYellow := sysx.ClrYellow
	clrGray := sysx.ClrGray

	loc, _ := time.LoadLocation(tz)
	if loc == nil {
		loc = time.Local
	}
	now := time.Now().In(loc)

	// Добавлены цвета к выводу:
	fmt.Println(clrCyan + "================================================" + clrReset)
	fmt.Printf("🔊  %s%sAcousticLog — Real-time Noise Monitor %s%s\n", clrBold, clrCyan, AppVersion, clrReset)
	fmt.Printf("👤  Разработчик: %s%s%s\n", clrCyan, Author, clrReset)
	fmt.Println(clrCyan + "================================================" + clrReset)
	fmt.Printf("📅 Local time: %s %s(TZ=%s)%s\n", now.Format("2006-01-02 15:04:05"), clrGray, loc, clrReset)
	fmt.Printf("🧩 Build: %s%s %s(BuildTime=%s)%s\n", clrYellow, BuildDate, clrGray, BuildTime, clrReset)
	fmt.Println(sysx.ClrBlue + "ℹ️  Ctrl+C — остановка и выход из программы" + clrReset)
}
