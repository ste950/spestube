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
	ListNewFetchedBefore(ctx context.Context, cutoff time.Time) ([]Video, error)
	Save(ctx context.Context, v Video) error
	ArchiveReleasedBefore(ctx context.Context, cutoff, at time.Time) (int, error)
}

// CollectionRepository è l'archivio delle raccolte. Get restituisce un errore
// apperr.KindNotFound se la raccolta non c'è; FindByName restituisce ok=false.
type CollectionRepository interface {
	// List restituisce le raccolte nell'ordine scelto dall'utente.
	List(ctx context.Context) ([]Collection, error)
	Get(ctx context.Context, id int64) (Collection, error)
	// FindByName confronta i nomi senza guardare le maiuscole.
	FindByName(ctx context.Context, name string) (c Collection, ok bool, err error)
	// Create mette la raccolta in fondo all'ordine e restituisce il suo id.
	Create(ctx context.Context, fields CollectionFields, at time.Time) (int64, error)
	Update(ctx context.Context, id int64, fields CollectionFields, at time.Time) error
	// Delete elimina la raccolta, non i video che conteneva.
	Delete(ctx context.Context, id int64) error
	// SetOrder assegna le posizioni nell'ordine dato, tutte insieme.
	SetOrder(ctx context.Context, ids []int64, at time.Time) error
	// Videos restituisce i video della raccolta, dall'ultimo aggiunto.
	Videos(ctx context.Context, id int64) ([]CollectionVideo, error)
	// AddVideo non fa niente se il video c'è già; RemoveVideo se non c'è.
	AddVideo(ctx context.Context, collectionID int64, videoID string, at time.Time) error
	RemoveVideo(ctx context.Context, collectionID int64, videoID string) error
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
