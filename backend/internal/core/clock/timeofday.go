// Package clock rappresenta un orario del giorno e calcola quando cade rispetto a un
// istante. Non sa a cosa serve l'orario.
package clock

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TimeOfDay è un orario senza data, al minuto.
type TimeOfDay struct {
	Hour   int
	Minute int
}

// ParseTimeOfDay accetta "7:00" e "07:00".
func ParseTimeOfDay(s string) (TimeOfDay, error) {
	h, m, ok := strings.Cut(strings.TrimSpace(s), ":")
	if !ok {
		return TimeOfDay{}, fmt.Errorf("orario non valido %q (es. 07:00)", s)
	}
	hour, errH := strconv.Atoi(h)
	minute, errM := strconv.Atoi(m)
	if errH != nil || errM != nil || hour < 0 || hour > 23 || minute < 0 || minute > 59 || len(m) != 2 {
		return TimeOfDay{}, fmt.Errorf("orario non valido %q (es. 07:00)", s)
	}
	return TimeOfDay{Hour: hour, Minute: minute}, nil
}

func (t TimeOfDay) String() string {
	return fmt.Sprintf("%02d:%02d", t.Hour, t.Minute)
}

// Last è l'occorrenza più recente dell'orario che non supera now, nel fuso loc.
func (t TimeOfDay) Last(now time.Time, loc *time.Location) time.Time {
	n := now.In(loc)
	at := t.on(n, 0)
	if at.After(n) {
		at = t.on(n, -1)
	}
	return at
}

// Next è la prima occorrenza dell'orario strettamente dopo now, nel fuso loc.
func (t TimeOfDay) Next(now time.Time, loc *time.Location) time.Time {
	n := now.In(loc)
	at := t.on(n, 0)
	if !at.After(n) {
		at = t.on(n, 1)
	}
	return at
}

// on usa time.Date con il giorno spostato invece di sommare 24 ore: nei giorni del
// cambio dell'ora un giorno non dura 24 ore, e l'orario deve restare quello dell'orologio.
func (t TimeOfDay) on(day time.Time, offsetDays int) time.Time {
	return time.Date(day.Year(), day.Month(), day.Day()+offsetDays, t.Hour, t.Minute, 0, 0, day.Location())
}
