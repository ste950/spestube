package library

import (
	"context"
	"time"

	"github.com/ste950/spestube/backend/internal/core/apperr"
)

// Collections raccoglie i casi d'uso delle raccolte. È un servizio a parte perché non ha
// bisogno dei feed né delle impostazioni: solo dei video, per sapere se si possono salvare.
type Collections struct {
	collections CollectionRepository
	videos      VideoRepository
	now         func() time.Time
}

func NewCollections(collections CollectionRepository, videos VideoRepository, now func() time.Time) *Collections {
	return &Collections{collections: collections, videos: videos, now: now}
}

func (c *Collections) List(ctx context.Context) ([]Collection, error) {
	return c.collections.List(ctx)
}

// Get restituisce la raccolta e i suoi video.
func (c *Collections) Get(ctx context.Context, id int64) (Collection, []CollectionVideo, error) {
	col, err := c.collections.Get(ctx, id)
	if err != nil {
		return Collection{}, nil, err
	}
	videos, err := c.collections.Videos(ctx, id)
	if err != nil {
		return Collection{}, nil, err
	}
	return col, videos, nil
}

func (c *Collections) Create(ctx context.Context, name, description string) (Collection, error) {
	fields, err := NewCollectionFields(name, description)
	if err != nil {
		return Collection{}, err
	}
	if err := c.ensureNameFree(ctx, fields.Name, 0); err != nil {
		return Collection{}, err
	}
	id, err := c.collections.Create(ctx, fields, c.now())
	if err != nil {
		return Collection{}, err
	}
	return c.collections.Get(ctx, id)
}

func (c *Collections) Update(ctx context.Context, id int64, name, description string) (Collection, error) {
	fields, err := NewCollectionFields(name, description)
	if err != nil {
		return Collection{}, err
	}
	if _, err := c.collections.Get(ctx, id); err != nil {
		return Collection{}, err
	}
	if err := c.ensureNameFree(ctx, fields.Name, id); err != nil {
		return Collection{}, err
	}
	if err := c.collections.Update(ctx, id, fields, c.now()); err != nil {
		return Collection{}, err
	}
	return c.collections.Get(ctx, id)
}

func (c *Collections) Delete(ctx context.Context, id int64) error {
	return c.collections.Delete(ctx, id)
}

// Reorder applica l'ordine scelto dall'utente: la lista completa degli id, dal primo
// all'ultimo.
func (c *Collections) Reorder(ctx context.Context, ids []int64) error {
	current, err := c.collections.List(ctx)
	if err != nil {
		return err
	}
	currentIDs := make([]int64, len(current))
	for i, col := range current {
		currentIDs[i] = col.ID
	}
	if err := CheckOrder(currentIDs, ids); err != nil {
		return err
	}
	return c.collections.SetOrder(ctx, ids, c.now())
}

// AddVideo salva un video in una raccolta. Salvarlo due volte non è un errore: il risultato
// è lo stesso, ed è ciò che si aspetta chi spunta una casella già spuntata.
func (c *Collections) AddVideo(ctx context.Context, collectionID int64, videoID string) error {
	if _, err := c.collections.Get(ctx, collectionID); err != nil {
		return err
	}
	v, err := c.videos.Get(ctx, videoID)
	if err != nil {
		return err
	}
	if !v.Playable() {
		return apperr.NotFound("video_not_found", "video inesistente")
	}
	return c.collections.AddVideo(ctx, collectionID, videoID, c.now())
}

func (c *Collections) RemoveVideo(ctx context.Context, collectionID int64, videoID string) error {
	if _, err := c.collections.Get(ctx, collectionID); err != nil {
		return err
	}
	return c.collections.RemoveVideo(ctx, collectionID, videoID)
}

// ensureNameFree rifiuta un nome già usato da un'altra raccolta. Il database ha lo stesso
// vincolo: questo controllo esiste per dare un messaggio chiaro nel caso normale.
func (c *Collections) ensureNameFree(ctx context.Context, name string, self int64) error {
	other, ok, err := c.collections.FindByName(ctx, name)
	if err != nil {
		return err
	}
	if ok && other.ID != self {
		return apperr.Conflict("collection_exists", "hai già una raccolta con questo nome")
	}
	return nil
}
