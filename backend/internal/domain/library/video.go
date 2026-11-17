package library

import (
	"time"

	"github.com/ste950/spestube/backend/internal/core/apperr"
)

// Status è il punto del percorso in cui si trova un video.
type Status string

const (
	StatusNew       Status = "new"       // scaricato, non ancora visibile
	StatusReleased  Status = "released"  // visibile in Oggi
	StatusWatched   Status = "watched"   // visto
	StatusDiscarded Status = "discarded" // scartato, da un filtro o dall'utente
	StatusArchived  Status = "archived"  // uscito da Oggi senza essere visto
)

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusReleased, StatusWatched, StatusDiscarded, StatusArchived:
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
	WatchedAt    time.Time
	ArchivedAt   time.Time
	// CollectionIDs sono le raccolte che contengono il video, in sola lettura. L'appartenenza
	// a una raccolta non dipende dallo stato: un video salvato resta salvato anche quando
	// esce da Oggi.
	CollectionIDs []int64
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

func (v Video) MarkWatched(now time.Time) (Video, error) {
	switch v.Status {
	case StatusReleased, StatusArchived:
		v.Status = StatusWatched
		v.WatchedAt = now
		return v, nil
	case StatusWatched:
		return v, nil
	}
	return Video{}, apperr.Conflict("video_not_watchable", "questo video non è disponibile")
}

func (v Video) DiscardByUser() (Video, error) {
	if v.Status != StatusReleased {
		return Video{}, apperr.Conflict("video_not_discardable", "questo video non si può scartare")
	}
	v.Status = StatusDiscarded
	v.Discard = DiscardUser
	return v, nil
}

// Playable riporta se il video si può aprire in Guarda e salvare in una raccolta. Un video
// nuovo o scartato non è mai stato mostrato, e conoscerne l'id non deve bastare a vederlo
// prima del rilascio.
func (v Video) Playable() bool {
	switch v.Status {
	case StatusReleased, StatusWatched, StatusArchived:
		return true
	}
	return false
}
