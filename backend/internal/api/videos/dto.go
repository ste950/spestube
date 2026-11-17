package videos

import (
	"time"

	"github.com/ste950/spestube/backend/internal/domain/library"
)

// videoResponse non ha contatori di visualizzazioni né di like: non esistono nel
// dominio, e non devono comparire nel contratto.
type videoResponse struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	ChannelID       string    `json:"channel_id"`
	ChannelTitle    string    `json:"channel_title"`
	PublishedAt     time.Time `json:"published_at"`
	DurationSeconds *int64    `json:"duration_seconds"` // null se sconosciuta
	Status          string    `json:"status"`
	CollectionIDs   []int64   `json:"collection_ids"` // le raccolte che lo contengono
}

type todayResponse struct {
	Videos        []videoResponse `json:"videos"`
	NextReleaseAt time.Time       `json:"next_release_at"`
}

func fromDomain(v library.Video) videoResponse {
	out := videoResponse{
		ID: v.ID, Title: v.Title, ChannelID: v.ChannelID, ChannelTitle: v.ChannelTitle,
		PublishedAt: v.PublishedAt.UTC(), Status: string(v.Status),
		CollectionIDs: v.CollectionIDs,
	}
	if out.CollectionIDs == nil {
		out.CollectionIDs = []int64{} // [] e non null
	}
	if v.Duration > 0 {
		sec := int64(v.Duration / time.Second)
		out.DurationSeconds = &sec
	}
	return out
}

func listFromDomain(videos []library.Video) []videoResponse {
	out := make([]videoResponse, 0, len(videos)) // [] e non null
	for _, v := range videos {
		out = append(out, fromDomain(v))
	}
	return out
}
