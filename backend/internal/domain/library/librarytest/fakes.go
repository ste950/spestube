// Package librarytest contiene i fake dei contratti del soggetto library: implementazioni
// vere e minime, in memoria, per i test del dominio e delle API.
package librarytest

import (
	"context"
	"sort"
	"time"

	"github.com/ste950/spestube/backend/internal/core/apperr"
	"github.com/ste950/spestube/backend/internal/domain/library"
)

type Channels struct {
	ByID map[string]library.Channel
}

func NewChannels(channels ...library.Channel) *Channels {
	c := &Channels{ByID: map[string]library.Channel{}}
	for _, ch := range channels {
		c.ByID[ch.ID] = ch
	}
	return c
}

func (c *Channels) Get(_ context.Context, id string) (library.Channel, error) {
	ch, ok := c.ByID[id]
	if !ok {
		return library.Channel{}, apperr.NotFound("channel_not_found", "canale inesistente")
	}
	return ch, nil
}

func (c *Channels) Find(_ context.Context, id string) (library.Channel, bool, error) {
	ch, ok := c.ByID[id]
	return ch, ok, nil
}

func (c *Channels) Add(_ context.Context, ch library.Channel) error {
	if _, ok := c.ByID[ch.ID]; ok {
		return apperr.Conflict("channel_exists", "segui già questo canale")
	}
	c.ByID[ch.ID] = ch
	return nil
}

func (c *Channels) List(context.Context) ([]library.Channel, error) {
	out := make([]library.Channel, 0, len(c.ByID))
	for _, ch := range c.ByID {
		out = append(out, ch)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (c *Channels) Remove(_ context.Context, id string) error {
	if _, ok := c.ByID[id]; !ok {
		return apperr.NotFound("channel_not_found", "canale inesistente")
	}
	delete(c.ByID, id)
	return nil
}

func (c *Channels) MarkFetched(_ context.Context, id string, mode library.FeedMode, at time.Time, fetchErr string) error {
	ch := c.ByID[id]
	ch.LastFetchError = fetchErr
	if fetchErr == "" {
		ch.FeedMode, ch.LastFetchedAt = mode, at
	}
	c.ByID[id] = ch
	return nil
}

type Videos struct {
	ByID map[string]library.Video
}

func NewVideos(videos ...library.Video) *Videos {
	v := &Videos{ByID: map[string]library.Video{}}
	for _, x := range videos {
		v.ByID[x.ID] = x
	}
	return v
}

func (v *Videos) Get(_ context.Context, id string) (library.Video, error) {
	x, ok := v.ByID[id]
	if !ok {
		return library.Video{}, apperr.NotFound("video_not_found", "video inesistente")
	}
	return x, nil
}

func (v *Videos) AddMissing(_ context.Context, videos []library.Video) (int, error) {
	added := 0
	for _, x := range videos {
		if _, ok := v.ByID[x.ID]; !ok {
			v.ByID[x.ID] = x
			added++
		}
	}
	return added, nil
}

func (v *Videos) ListReleased(context.Context) ([]library.Video, error) {
	return v.filter(func(x library.Video) bool { return x.Status == library.StatusReleased }), nil
}

func (v *Videos) ListQueued(context.Context) ([]library.Video, error) {
	return v.filter(func(x library.Video) bool { return x.Status == library.StatusQueued }), nil
}

func (v *Videos) ListNewFetchedBefore(_ context.Context, cutoff time.Time) ([]library.Video, error) {
	return v.filter(func(x library.Video) bool {
		return x.Status == library.StatusNew && x.FetchedAt.Before(cutoff)
	}), nil
}

func (v *Videos) Save(_ context.Context, x library.Video) error {
	if _, ok := v.ByID[x.ID]; !ok {
		return apperr.NotFound("video_not_found", "video inesistente")
	}
	v.ByID[x.ID] = x
	return nil
}

func (v *Videos) ArchiveReleasedBefore(_ context.Context, cutoff, at time.Time) (int, error) {
	return v.archive(at, func(x library.Video) bool {
		return x.Status == library.StatusReleased && x.ReleasedAt.Before(cutoff)
	}), nil
}

func (v *Videos) ArchiveQueuedBefore(_ context.Context, cutoff, at time.Time) (int, error) {
	return v.archive(at, func(x library.Video) bool {
		return x.Status == library.StatusQueued && !x.QueuedAt.After(cutoff)
	}), nil
}

func (v *Videos) archive(at time.Time, match func(library.Video) bool) int {
	n := 0
	for id, x := range v.ByID {
		if match(x) {
			x.Status, x.ArchivedAt = library.StatusArchived, at
			v.ByID[id] = x
			n++
		}
	}
	return n
}

func (v *Videos) filter(match func(library.Video) bool) []library.Video {
	var out []library.Video
	for _, x := range v.ByID {
		if match(x) {
			out = append(out, x)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Feeds risponde con ciò che c'è nelle due mappe, per canale. Un canale assente da una
// mappa non ha quel feed; Down fa rispondere come se YouTube fosse irraggiungibile.
type Feeds struct {
	LongVideos map[string]library.Feed
	Channel    map[string]library.Feed
	Down       map[string]bool
}

func (f *Feeds) Recent(_ context.Context, channelID string, mode library.FeedMode) (library.Feed, error) {
	if f.Down[channelID] {
		return library.Feed{}, apperr.Unavailable("youtube_unavailable", "YouTube non risponde", nil)
	}
	src := f.LongVideos
	if mode == library.FeedChannel {
		src = f.Channel
	}
	feed, ok := src[channelID]
	if !ok {
		return library.Feed{}, apperr.NotFound("feed_not_found", "feed inesistente")
	}
	return feed, nil
}

// Handles risolve gli handle della mappa. Con NoKey risponde come un'installazione
// senza chiave API.
type Handles struct {
	IDs   map[string]string
	NoKey bool
}

func (h *Handles) ChannelID(_ context.Context, handle string) (string, error) {
	if h.NoKey {
		return "", apperr.Invalid("handle_needs_api_key", "serve la chiave API")
	}
	id, ok := h.IDs[handle]
	if !ok {
		return "", apperr.NotFound("channel_not_found", "nessun canale con questo handle")
	}
	return id, nil
}

// Policy è una library.PolicySource fissa.
type Policy library.Policy

func (p Policy) Policy(context.Context) (library.Policy, error) {
	return library.Policy(p), nil
}
