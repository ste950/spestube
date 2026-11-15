package settings_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/ste950/spestube/backend/internal/core/apperr"
	"github.com/ste950/spestube/backend/internal/core/clock"
	"github.com/ste950/spestube/backend/internal/domain/settings"
	"github.com/ste950/spestube/backend/internal/domain/settings/settingstest"
)

func newService(repo *settingstest.Repository) *settings.Service {
	return settings.NewService(repo, settings.Defaults(clock.TimeOfDay{Hour: 7}), time.Now)
}

func TestDefaultsWhenNothingIsStored(t *testing.T) {
	got, err := newService(settingstest.NewRepository()).Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := settings.Settings{
		ReleaseTime: clock.TimeOfDay{Hour: 7}, ExcludeShorts: true, ExcludeLive: true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("default = %+v, attesi %+v", got, want)
	}
}

func TestUpdateThenGet(t *testing.T) {
	ctx := context.Background()
	svc := newService(settingstest.NewRepository())
	in := settings.Settings{
		ReleaseTime: clock.TimeOfDay{Hour: 19, Minute: 30}, ExcludeShorts: false, ExcludeLive: false,
		MinDurationMinutes: 10, ExcludedKeywords: []string{" reaction ", "", "Drama", "REACTION"},
	}
	if _, err := svc.Update(ctx, in); err != nil {
		t.Fatal(err)
	}
	got, err := svc.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := in
	want.ExcludedKeywords = []string{"reaction", "Drama"} // ripulite: senza spazi, vuote e doppioni
	if !reflect.DeepEqual(got, want) {
		t.Errorf("lette = %+v, attese %+v", got, want)
	}
}

func TestUpdateRejectsInvalidValues(t *testing.T) {
	valid := settings.Defaults(clock.TimeOfDay{Hour: 7})
	cases := []struct {
		name   string
		change func(*settings.Settings)
	}{
		{"durata minima negativa", func(s *settings.Settings) { s.MinDurationMinutes = -1 }},
		{"durata minima assurda", func(s *settings.Settings) { s.MinDurationMinutes = 601 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := settingstest.NewRepository()
			in := valid
			tc.change(&in)
			if _, err := newService(repo).Update(context.Background(), in); apperr.KindOf(err) != apperr.KindInvalid {
				t.Errorf("errore = %v, atteso Invalid", err)
			}
			if len(repo.Values) != 0 {
				t.Error("un salvataggio rifiutato ha scritto qualcosa")
			}
		})
	}
}

func TestGetIgnoresUnreadableStoredValues(t *testing.T) {
	repo := settingstest.NewRepository()
	repo.Values = map[string]string{
		"release_time": "le sette", "min_duration_minutes": "tanti",
	}
	got, err := newService(repo).Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.ReleaseTime != (clock.TimeOfDay{Hour: 7}) || got.MinDurationMinutes != 0 {
		t.Errorf("valori illeggibili non ignorati: %+v", got)
	}
}
