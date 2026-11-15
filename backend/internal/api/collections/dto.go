package collections

import (
	"time"

	"github.com/ste950/spestube/backend/internal/domain/library"
)

type collectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type orderRequest struct {
	IDs []int64 `json:"ids"`
}

type collectionResponse struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	VideoCount   int       `json:"video_count"`
	CoverVideoID *string   `json:"cover_video_id"` // null se la raccolta è vuota
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type listResponse struct {
	Collections []collectionResponse `json:"collections"`
}

// videoResponse ripete la forma del video di /api/videos, più la data di ingresso: le
// risorse non si importano fra loro, e il contratto è lo stesso per scelta.
type videoResponse struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	ChannelID       string    `json:"channel_id"`
	ChannelTitle    string    `json:"channel_title"`
	PublishedAt     time.Time `json:"published_at"`
	DurationSeconds *int64    `json:"duration_seconds"`
	Status          string    `json:"status"`
	CollectionIDs   []int64   `json:"collection_ids"`
	AddedAt         time.Time `json:"added_at"`
}

type detailResponse struct {
	Collection collectionResponse `json:"collection"`
	Videos     []videoResponse    `json:"videos"`
}

func fromDomain(c library.Collection) collectionResponse {
	out := collectionResponse{
		ID: c.ID, Name: c.Name, Description: c.Description, VideoCount: c.VideoCount,
		CreatedAt: c.CreatedAt.UTC(), UpdatedAt: c.UpdatedAt.UTC(),
	}
	if c.CoverVideoID != "" {
		cover := c.CoverVideoID
		out.CoverVideoID = &cover
	}
	return out
}

func videoFromDomain(v library.CollectionVideo) videoResponse {
	out := videoResponse{
		ID: v.ID, Title: v.Title, ChannelID: v.ChannelID, ChannelTitle: v.ChannelTitle,
		PublishedAt: v.PublishedAt.UTC(), Status: string(v.Status),
		CollectionIDs: v.CollectionIDs, AddedAt: v.AddedAt.UTC(),
	}
	if out.CollectionIDs == nil {
		out.CollectionIDs = []int64{}
	}
	if v.Duration > 0 {
		sec := int64(v.Duration / time.Second)
		out.DurationSeconds = &sec
	}
	return out
}
