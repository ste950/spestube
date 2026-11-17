package library

import (
	"strings"
	"time"

	"github.com/ste950/spestube/backend/internal/core/clock"
)

// ShortMaxDuration è la durata massima di uno Short su YouTube. Serve solo per i video
// arrivati dal feed del canale, che gli Shorts li contiene.
const ShortMaxDuration = 3 * time.Minute

// Filters sono le regole che un video deve superare per arrivare in Oggi.
type Filters struct {
	ExcludeShorts    bool
	ExcludeLive      bool
	MinDuration      time.Duration // zero: nessuna durata minima
	ExcludedKeywords []string
}

// Policy è ciò che il rilascio deve sapere dalle impostazioni: quando avviene e con
// quali filtri.
type Policy struct {
	ReleaseAt clock.TimeOfDay
	Filters   Filters
}

// Check restituisce il motivo per cui il video va scartato, o DiscardNone se passa.
//
// Una durata sconosciuta non scarta mai: senza chiave API non c'è modo di saperla, e
// nascondere tutti i video sarebbe peggio che lasciarne passare uno corto.
func (f Filters) Check(v Video) DiscardReason {
	title := strings.ToLower(v.Title)
	for _, k := range f.ExcludedKeywords {
		k = strings.ToLower(strings.TrimSpace(k))
		if k != "" && strings.Contains(title, k) {
			return DiscardKeyword
		}
	}
	if f.ExcludeLive && v.IsLive {
		return DiscardLive
	}
	// Un video arrivato dalla playlist dei video lunghi non è uno Short per costruzione,
	// anche se dura un minuto: la durata si guarda solo per il feed del canale.
	if f.ExcludeShorts && v.Source == FeedChannel && v.Duration > 0 && v.Duration <= ShortMaxDuration {
		return DiscardShort
	}
	if f.MinDuration > 0 && v.Duration > 0 && v.Duration < f.MinDuration {
		return DiscardDuration
	}
	return DiscardNone
}
