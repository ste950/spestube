package library

import (
	"context"
	"time"
)

// ChannelRepository è l'archivio dei canali seguiti. Get restituisce un errore
// apperr.KindNotFound se il canale non c'è; Find restituisce ok=false.
type ChannelRepository interface {
	Get(ctx context.Context, id string) (Channel, error)
	Find(ctx context.Context, id string) (ch Channel, ok bool, err error)
	Add(ctx context.Context, ch Channel) error
	List(ctx context.Context) ([]Channel, error)
	// Remove elimina il canale e, con lui, tutti i suoi video.
	Remove(ctx context.Context, id string) error
	// MarkFetched registra l'esito dell'ultima lettura del feed. fetchErr vuoto: riuscita.
	MarkFetched(ctx context.Context, id string, mode FeedMode, at time.Time, fetchErr string) error
}

// VideoRepository è l'archivio dei video.
type VideoRepository interface {
	Get(ctx context.Context, id string) (Video, error)
	// AddMissing inserisce i video che non ci sono ancora e lascia intatti gli altri.
	// Restituisce quanti ne ha inseriti.
	AddMissing(ctx context.Context, videos []Video) (int, error)
	ListReleased(ctx context.Context) ([]Video, error)
	ListQueued(ctx context.Context) ([]Video, error)
	ListNewFetchedBefore(ctx context.Context, cutoff time.Time) ([]Video, error)
	Save(ctx context.Context, v Video) error
	ArchiveReleasedBefore(ctx context.Context, cutoff, at time.Time) (int, error)
	ArchiveQueuedBefore(ctx context.Context, cutoff, at time.Time) (int, error)
}

// Feed è ciò che YouTube dice di un canale: il nome e i video più recenti.
type Feed struct {
	ChannelTitle string
	Entries      []FeedEntry
}

type FeedEntry struct {
	VideoID     string
	Title       string
	PublishedAt time.Time
}

// FeedRepository legge i video recenti di un canale. Restituisce un errore
// apperr.KindNotFound se quel feed non esiste, apperr.KindUnavailable se YouTube non
// risponde.
type FeedRepository interface {
	Recent(ctx context.Context, channelID string, mode FeedMode) (Feed, error)
}

// HandleRepository traduce un handle ("@nome") nell'id del canale.
type HandleRepository interface {
	ChannelID(ctx context.Context, handle string) (string, error)
}

// PolicySource fornisce orario di rilascio e filtri correnti: cambiano dalle
// impostazioni mentre il processo gira, quindi si leggono a ogni giro.
type PolicySource interface {
	Policy(ctx context.Context) (Policy, error)
}
