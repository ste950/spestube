package library

import (
	"strings"
	"time"

	"github.com/ste950/spestube/backend/internal/core/apperr"
)

// Status è il punto del percorso in cui si trova un video.
type Status string

const (
	StatusNew       Status = "new"       // scaricato, non ancora visibile
	StatusReleased  Status = "released"  // visibile in Oggi
	StatusQueued    Status = "queued"    // in Da vedere, con un motivo
	StatusWatched   Status = "watched"   // visto
	StatusDiscarded Status = "discarded" // scartato, da un filtro o dall'utente
	StatusArchived  Status = "archived"  // uscito di scena da solo
)

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusReleased, StatusQueued, StatusWatched, StatusDiscarded, StatusArchived:
		return true
	}
	return false
}

// DiscardReason dice perché un video è stato scartato. Vuoto se non lo è.
type DiscardReason string

const (
	DiscardNone     DiscardReason = ""
	DiscardUser     DiscardReason = "user"
	DiscardBaseline DiscardReason = "baseline" // già pubblicato quando il canale è stato aggiunto
	DiscardShort    DiscardReason = "short"
	DiscardLive     DiscardReason = "live"
	DiscardDuration DiscardReason = "duration"
	DiscardKeyword  DiscardReason = "keyword"
)

func (r DiscardReason) Valid() bool {
	switch r {
	case DiscardNone, DiscardUser, DiscardBaseline, DiscardShort, DiscardLive, DiscardDuration, DiscardKeyword:
		return true
	}
	return false
}

// QueueLifetime è quanto un video resta in Da vedere prima di essere archiviato da solo.
// Due settimane: abbastanza per un weekend libero, troppo poco per accumulare una lista
// che non si guarderà mai.
const QueueLifetime = 14 * 24 * time.Hour

type Video struct {
	ID           string
	ChannelID    string
	ChannelTitle string // in sola lettura: lo valorizza l'archivio leggendo il canale
	Title        string
	PublishedAt  time.Time
	FetchedAt    time.Time
	Source       FeedMode      // il feed da cui è arrivato
	Duration     time.Duration // zero se sconosciuta (senza chiave API)
	IsLive       bool
	Status       Status
	Discard      DiscardReason
	ReleasedAt   time.Time
	QueueReason  string
	QueuedAt     time.Time
	WatchedAt    time.Time
	ArchivedAt   time.Time
}

// Release decide la sorte di un video nuovo al momento del rilascio: visibile in Oggi,
// oppure scartato dal primo filtro che non supera.
func (v Video) Release(now time.Time, f Filters) Video {
	if reason := f.Check(v); reason != DiscardNone {
		v.Status = StatusDiscarded
		v.Discard = reason
		return v
	}
	v.Status = StatusReleased
	v.ReleasedAt = now
	return v
}

// Enqueue mette il video in Da vedere. Il motivo è obbligatorio: è l'attrito voluto che
// distingue "mi serve" da "lo salvo, non si sa mai".
func (v Video) Enqueue(reason string, now time.Time) (Video, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return Video{}, apperr.Invalid("queue_reason_required", "per mettere un video in Da vedere serve un motivo")
	}
	if v.Status != StatusReleased && v.Status != StatusArchived {
		return Video{}, apperr.Conflict("video_not_queueable", "questo video non può entrare in Da vedere")
	}
	v.Status = StatusQueued
	v.QueueReason = reason
	v.QueuedAt = now
	v.ArchivedAt = time.Time{}
	return v, nil
}

func (v Video) MarkWatched(now time.Time) (Video, error) {
	switch v.Status {
	case StatusReleased, StatusQueued, StatusArchived:
		v.Status = StatusWatched
		v.WatchedAt = now
		return v, nil
	case StatusWatched:
		return v, nil
	}
	return Video{}, apperr.Conflict("video_not_watchable", "questo video non è disponibile")
}

func (v Video) DiscardByUser() (Video, error) {
	if v.Status != StatusReleased && v.Status != StatusQueued {
		return Video{}, apperr.Conflict("video_not_discardable", "questo video non si può scartare")
	}
	v.Status = StatusDiscarded
	v.Discard = DiscardUser
	return v, nil
}

// Playable riporta se il video si può aprire in Guarda. Un video nuovo o scartato non è
// mai stato mostrato, e conoscerne l'id non deve bastare a vederlo prima del rilascio.
func (v Video) Playable() bool {
	switch v.Status {
	case StatusReleased, StatusQueued, StatusWatched, StatusArchived:
		return true
	}
	return false
}

// QueueExpiresAt è l'istante in cui il video in coda viene archiviato. Zero se non è in coda.
func (v Video) QueueExpiresAt() time.Time {
	if v.Status != StatusQueued {
		return time.Time{}
	}
	return v.QueuedAt.Add(QueueLifetime)
}

// QueueDaysLeft conta i giorni che restano, arrotondando per eccesso: a 30 ore dalla
// scadenza ne restano 2, a 3 ore ne resta 1.
func (v Video) QueueDaysLeft(now time.Time) int {
	left := v.QueueExpiresAt().Sub(now)
	if v.Status != StatusQueued || left <= 0 {
		return 0
	}
	const day = 24 * time.Hour
	return int((left + day - 1) / day)
}
