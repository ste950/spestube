package library

import (
	"context"
	"strings"

	"github.com/ste950/spestube/backend/internal/core/apperr"
	"github.com/ste950/spestube/backend/internal/data/library/sources"
	"github.com/ste950/spestube/backend/internal/domain/library"
)

// FeedRepository soddisfa library.FeedRepository componendo la sorgente dei feed: è qui
// che un FeedMode diventa un indirizzo di YouTube.
type FeedRepository struct {
	feeds sources.FeedSource
}

func NewFeedRepository(feeds sources.FeedSource) *FeedRepository {
	return &FeedRepository{feeds: feeds}
}

func (r *FeedRepository) Recent(ctx context.Context, channelID string, mode library.FeedMode) (library.Feed, error) {
	var (
		payload sources.FeedPayload
		err     error
	)
	if mode == library.FeedLongVideos {
		payload, err = r.feeds.Playlist(ctx, LongVideosPlaylistID(channelID))
	} else {
		payload, err = r.feeds.Channel(ctx, channelID)
	}
	if err != nil {
		return library.Feed{}, err
	}

	// Nel feed di una playlist il titolo è "Videos": il nome del canale è l'autore.
	feed := library.Feed{ChannelTitle: payload.Author}
	if feed.ChannelTitle == "" {
		feed.ChannelTitle = payload.Title
	}
	for _, e := range payload.Entries {
		feed.Entries = append(feed.Entries, library.FeedEntry{
			VideoID: e.VideoID, Title: e.Title, PublishedAt: e.Published,
		})
	}
	return feed, nil
}

// LongVideosPlaylistID costruisce l'id della playlist che YouTube tiene per ogni canale
// con i soli video lunghi: l'id del canale con "UULF" al posto di "UC". Non è
// documentata, ed è il motivo per cui esiste il ripiego sul feed del canale.
func LongVideosPlaylistID(channelID string) string {
	return "UULF" + strings.TrimPrefix(channelID, "UC")
}

// HandleRepository soddisfa library.HandleRepository con la YouTube Data API.
type HandleRepository struct {
	lookup sources.ChannelLookup
}

// NewHandleRepository accetta lookup nil: è il caso di un'installazione senza chiave API.
func NewHandleRepository(lookup sources.ChannelLookup) *HandleRepository {
	return &HandleRepository{lookup: lookup}
}

func (r *HandleRepository) ChannelID(ctx context.Context, handle string) (string, error) {
	// Senza chiave un handle si potrebbe risolvere solo leggendo la pagina HTML di
	// YouTube, cioè con lo scraping che questo progetto non fa.
	if r.lookup == nil {
		return "", apperr.Invalid("handle_needs_api_key",
			"per aggiungere un canale dall'@handle serve la chiave della YouTube Data API: "+
				"senza chiave usa il link youtube.com/channel/UC… o l'id del canale")
	}
	ch, err := r.lookup.ByHandle(ctx, handle)
	if err != nil {
		return "", err
	}
	return ch.ID, nil
}
