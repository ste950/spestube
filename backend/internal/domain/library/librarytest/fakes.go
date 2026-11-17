// Package librarytest contiene i fake dei contratti del soggetto library: implementazioni
// vere e minime, in memoria, per i test del dominio e delle API.
package librarytest

import (
	"context"
	"sort"
	"strings"
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

// Collections è l'archivio delle raccolte in memoria. Le posizioni seguono l'ordine della
// lista; i video li legge da Videos, come farebbe la join del database.
type Collections struct {
	Items []library.Collection
	Store *Videos // dove vivono i video, per la loro forma e le loro raccolte
	// Members[id raccolta] sono i video, nell'ordine di ingresso.
	Members map[int64][]library.CollectionVideo
	nextID  int64
}

func NewCollections(videos *Videos, items ...library.Collection) *Collections {
	c := &Collections{Store: videos, Members: map[int64][]library.CollectionVideo{}}
	for _, it := range items {
		c.Items = append(c.Items, it)
		c.nextID = max(c.nextID, it.ID)
	}
	return c
}

func (c *Collections) List(context.Context) ([]library.Collection, error) {
	out := make([]library.Collection, 0, len(c.Items))
	for _, it := range c.Items {
		out = append(out, c.withCounts(it))
	}
	return out, nil
}

func (c *Collections) Get(_ context.Context, id int64) (library.Collection, error) {
	for _, it := range c.Items {
		if it.ID == id {
			return c.withCounts(it), nil
		}
	}
	return library.Collection{}, apperr.NotFound("collection_not_found", "raccolta inesistente")
}

func (c *Collections) FindByName(_ context.Context, name string) (library.Collection, bool, error) {
	for _, it := range c.Items {
		if strings.EqualFold(it.Name, name) {
			return c.withCounts(it), true, nil
		}
	}
	return library.Collection{}, false, nil
}

func (c *Collections) Create(_ context.Context, f library.CollectionFields, at time.Time) (int64, error) {
	c.nextID++
	c.Items = append(c.Items, library.Collection{
		ID: c.nextID, Name: f.Name, Description: f.Description, Position: len(c.Items),
		CreatedAt: at, UpdatedAt: at,
	})
	return c.nextID, nil
}

func (c *Collections) Update(_ context.Context, id int64, f library.CollectionFields, at time.Time) error {
	for i, it := range c.Items {
		if it.ID == id {
			c.Items[i].Name, c.Items[i].Description, c.Items[i].UpdatedAt = f.Name, f.Description, at
			return nil
		}
	}
	return apperr.NotFound("collection_not_found", "raccolta inesistente")
}

func (c *Collections) Delete(_ context.Context, id int64) error {
	for i, it := range c.Items {
		if it.ID == id {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			delete(c.Members, id)
			return nil
		}
	}
	return apperr.NotFound("collection_not_found", "raccolta inesistente")
}

func (c *Collections) SetOrder(_ context.Context, ids []int64, at time.Time) error {
	byID := map[int64]library.Collection{}
	for _, it := range c.Items {
		byID[it.ID] = it
	}
	ordered := make([]library.Collection, 0, len(ids))
	for pos, id := range ids {
		it := byID[id]
		it.Position, it.UpdatedAt = pos, at
		ordered = append(ordered, it)
	}
	c.Items = ordered
	return nil
}

func (c *Collections) Videos(_ context.Context, id int64) ([]library.CollectionVideo, error) {
	members := c.Members[id]
	out := make([]library.CollectionVideo, 0, len(members))
	for i := len(members) - 1; i >= 0; i-- {
		out = append(out, members[i])
	}
	return out, nil
}

func (c *Collections) AddVideo(_ context.Context, collectionID int64, videoID string, at time.Time) error {
	for _, m := range c.Members[collectionID] {
		if m.ID == videoID {
			return nil
		}
	}
	v := c.Store.ByID[videoID]
	c.Members[collectionID] = append(c.Members[collectionID], library.CollectionVideo{Video: v, AddedAt: at})
	v.CollectionIDs = append(v.CollectionIDs, collectionID)
	c.Store.ByID[videoID] = v
	return nil
}

func (c *Collections) RemoveVideo(_ context.Context, collectionID int64, videoID string) error {
	members := c.Members[collectionID]
	for i, m := range members {
		if m.ID == videoID {
			c.Members[collectionID] = append(members[:i], members[i+1:]...)
		}
	}
	v, ok := c.Store.ByID[videoID]
	if ok {
		kept := v.CollectionIDs[:0]
		for _, id := range v.CollectionIDs {
			if id != collectionID {
				kept = append(kept, id)
			}
		}
		v.CollectionIDs = kept
		c.Store.ByID[videoID] = v
	}
	return nil
}

func (c *Collections) withCounts(it library.Collection) library.Collection {
	members := c.Members[it.ID]
	it.VideoCount = len(members)
	it.CoverVideoID = ""
	if len(members) > 0 {
		it.CoverVideoID = members[len(members)-1].ID
	}
	return it
}
