package clock

import (
	"testing"
	"time"
)

func TestParseTimeOfDay(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    TimeOfDay
		wantErr bool
	}{
		{"senza zero iniziale", "7:00", TimeOfDay{7, 0}, false},
		{"con zero iniziale", "07:30", TimeOfDay{7, 30}, false},
		{"fine giornata", "23:59", TimeOfDay{23, 59}, false},
		{"ora fuori intervallo", "24:00", TimeOfDay{}, true},
		{"minuti fuori intervallo", "07:60", TimeOfDay{}, true},
		{"minuti a una cifra", "07:5", TimeOfDay{}, true},
		{"senza due punti", "0700", TimeOfDay{}, true},
		{"vuoto", "", TimeOfDay{}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseTimeOfDay(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ParseTimeOfDay(%q) errore = %v, atteso errore = %v", tc.in, err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("ParseTimeOfDay(%q) = %v, atteso %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestLastAndNext(t *testing.T) {
	rome, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		t.Fatal(err)
	}
	seven := TimeOfDay{7, 0}
	at := func(day, hour, minute int) time.Time {
		return time.Date(2026, time.November, day, hour, minute, 0, 0, rome)
	}
	cases := []struct {
		name     string
		now      time.Time
		wantLast time.Time
		wantNext time.Time
	}{
		{"prima dell'orario", at(15, 6, 59), at(14, 7, 0), at(15, 7, 0)},
		{"all'orario esatto", at(15, 7, 0), at(15, 7, 0), at(16, 7, 0)},
		{"dopo l'orario", at(15, 13, 0), at(15, 7, 0), at(16, 7, 0)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := seven.Last(tc.now, rome); !got.Equal(tc.wantLast) {
				t.Errorf("Last = %v, atteso %v", got, tc.wantLast)
			}
			if got := seven.Next(tc.now, rome); !got.Equal(tc.wantNext) {
				t.Errorf("Next = %v, atteso %v", got, tc.wantNext)
			}
		})
	}
}

func TestLastUsesTheLocationNotUTC(t *testing.T) {
	rome, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		t.Fatal(err)
	}
	// Le 06:30 UTC sono le 07:30 a Roma: l'orario delle 7 è già passato, oggi.
	now := time.Date(2026, time.November, 15, 6, 30, 0, 0, time.UTC)
	want := time.Date(2026, time.November, 15, 7, 0, 0, 0, rome)
	if got := (TimeOfDay{7, 0}).Last(now, rome); !got.Equal(want) {
		t.Errorf("Last = %v, atteso %v", got, want)
	}
}

func TestNextAcrossDaylightSavingChange(t *testing.T) {
	rome, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		t.Fatal(err)
	}
	// Nella notte fra il 28 e il 29 marzo 2026 l'orologio salta un'ora: fra le 7 di
	// sabato e le 7 di domenica passano 23 ore, non 24.
	saturday := time.Date(2026, time.March, 28, 7, 0, 0, 0, rome)
	next := (TimeOfDay{7, 0}).Next(saturday, rome)
	if next.Hour() != 7 || next.Day() != 29 {
		t.Errorf("Next = %v, attese le 7 del 29", next)
	}
	if d := next.Sub(saturday); d != 23*time.Hour {
		t.Errorf("distanza = %v, attese 23h", d)
	}
}
