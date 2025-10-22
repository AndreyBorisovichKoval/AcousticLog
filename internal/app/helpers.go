// C:\_Projects_Go\AcousticLog\internal\app\helpers.go

package app

import (
	"time"
)

func parseHHMM(s string) (int, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return 0, err
	}
	return t.Hour()*60 + t.Minute(), nil
}

func (a *App) isDay(now time.Time) bool {
	m := now.Hour()*60 + now.Minute()
	if a.dayStart <= a.dayEnd {
		return m >= a.dayStart && m < a.dayEnd
	}
	return m >= a.dayStart || m < a.dayEnd
}

func (a *App) currentLimit(now time.Time) (string, float64) {
	if a.isDay(now) {
		return "DAY", a.dayLimit
	}
	return "NIGHT", a.nightLimit
}

func nextStopAt(loc *time.Location, hhmm string) (time.Time, error) {
	now := time.Now().In(loc)
	t, err := time.ParseInLocation("15:04", hhmm, loc)
	if err != nil {
		return time.Time{}, err
	}
	stop := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, loc)
	if !now.Before(stop) {
		stop = stop.Add(24 * time.Hour)
	}
	return stop, nil
}
