package channels

import (
	"time"

	"github.com/ste950/spestube/backend/internal/domain/library"
)

type addRequest struct {
	// Channel è ciò che l'utente ha scritto: un link, un @handle o un id.
	Channel string `json:"channel"`
}

type channelResponse struct {
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	Handle         string     `json:"handle"`
	FeedMode       string     `json:"feed_mode"`
	AddedAt        time.Time  `json:"added_at"`
	LastFetchedAt  *time.Time `json:"last_fetched_at"` // null finché il feed non è mai stato letto
	LastFetchError string     `json:"last_fetch_error"`
}

type listResponse struct {
	Channels []channelResponse `json:"channels"`
}

func fromDomain(ch library.Channel) channelResponse {
	out := channelResponse{
		ID: ch.ID, Title: ch.Title, Handle: ch.Handle, FeedMode: string(ch.FeedMode),
		AddedAt: ch.AddedAt.UTC(), LastFetchError: ch.LastFetchError,
	}
	if !ch.LastFetchedAt.IsZero() {
		t := ch.LastFetchedAt.UTC()
		out.LastFetchedAt = &t
	}
	return out
}
