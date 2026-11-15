// Package settings è il soggetto delle impostazioni: i filtri e l'orario in cui arrivano
// i video.
package settings

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/ste950/spestube/backend/internal/core/apperr"
	"github.com/ste950/spestube/backend/internal/core/clock"
)

// Dieci ore: nessun video ragionevole è più lungo, e una durata minima più alta è quasi
// certamente un errore di battitura che nasconderebbe tutto.
const maxMinDurationMinutes = 600

type Settings struct {
	ReleaseTime        clock.TimeOfDay
	ExcludeShorts      bool
	ExcludeLive        bool
	MinDurationMinutes int
	ExcludedKeywords   []string
}

// Defaults sono le impostazioni di un'installazione nuova. L'orario di rilascio arriva
// dalla configurazione del processo.
func Defaults(releaseTime clock.TimeOfDay) Settings {
	return Settings{
		ReleaseTime:   releaseTime,
		ExcludeShorts: true,
		ExcludeLive:   true,
	}
}

func (s Settings) validate() error {
	if s.MinDurationMinutes < 0 || s.MinDurationMinutes > maxMinDurationMinutes {
		return apperr.Invalid("settings_invalid", "durata minima non valida")
	}
	return nil
}

// Repository conserva le impostazioni come coppie chiave-valore: aggiungerne una non
// richiede una migrazione.
type Repository interface {
	All(ctx context.Context) (map[string]string, error)
	Put(ctx context.Context, values map[string]string, at time.Time) error
}

const (
	keyReleaseTime   = "release_time"
	keyExcludeShorts = "exclude_shorts"
	keyExcludeLive   = "exclude_live"
	keyMinDuration   = "min_duration_minutes"
	keyKeywords      = "excluded_keywords"
)

type Service struct {
	repo     Repository
	defaults Settings
	now      func() time.Time
}

func NewService(repo Repository, defaults Settings, now func() time.Time) *Service {
	return &Service{repo: repo, defaults: defaults, now: now}
}

// Get restituisce i default con sopra ciò che è stato salvato. Un valore salvato che
// non si legge più si ignora: meglio il default di un'app che non parte.
func (s *Service) Get(ctx context.Context) (Settings, error) {
	stored, err := s.repo.All(ctx)
	if err != nil {
		return Settings{}, err
	}
	out := s.defaults
	if v, ok := stored[keyReleaseTime]; ok {
		if t, err := clock.ParseTimeOfDay(v); err == nil {
			out.ReleaseTime = t
		}
	}
	if v, ok := stored[keyExcludeShorts]; ok {
		out.ExcludeShorts = v == "true"
	}
	if v, ok := stored[keyExcludeLive]; ok {
		out.ExcludeLive = v == "true"
	}
	if v, ok := stored[keyMinDuration]; ok {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 && n <= maxMinDurationMinutes {
			out.MinDurationMinutes = n
		}
	}
	if v, ok := stored[keyKeywords]; ok {
		out.ExcludedKeywords = splitKeywords(v)
	}
	return out, nil
}

func (s *Service) Update(ctx context.Context, in Settings) (Settings, error) {
	if err := in.validate(); err != nil {
		return Settings{}, err
	}
	in.ExcludedKeywords = cleanKeywords(in.ExcludedKeywords)
	values := map[string]string{
		keyReleaseTime:   in.ReleaseTime.String(),
		keyExcludeShorts: strconv.FormatBool(in.ExcludeShorts),
		keyExcludeLive:   strconv.FormatBool(in.ExcludeLive),
		keyMinDuration:   strconv.Itoa(in.MinDurationMinutes),
		keyKeywords:      strings.Join(in.ExcludedKeywords, "\n"),
	}
	if err := s.repo.Put(ctx, values, s.now()); err != nil {
		return Settings{}, err
	}
	return in, nil
}

// Le parole chiave si salvano una per riga: una virgola può far parte di una parola
// chiave, un a capo no.
func splitKeywords(v string) []string {
	return cleanKeywords(strings.Split(v, "\n"))
}

func cleanKeywords(in []string) []string {
	out := make([]string, 0, len(in))
	seen := map[string]bool{}
	for _, k := range in {
		k = strings.TrimSpace(k)
		if lower := strings.ToLower(k); k != "" && !seen[lower] {
			seen[lower] = true
			out = append(out, k)
		}
	}
	return out
}
