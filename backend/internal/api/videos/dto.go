package videos

import (
	"time"

	"github.com/ste950/spestube/backend/internal/domain/library"
)

type queueRequest struct {
	Reason string `json:"reason"`
}

// videoResponse non ha contatori di visualizzazioni né di like: non esistono nel
// dominio, e non devono comparire nel contratto.
type videoResponse struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	ChannelID       string     `json:"channel_id"`
	ChannelTitle    string     `json:"channel_title"`
	PublishedAt     time.Time  `json:"published_at"`
	DurationSeconds *int64     `json:"duration_seconds"` // null se sconosciuta
	Status          string     `json:"status"`
	QueueReason     string     `json:"queue_reason"`
	QueuedAt        *time.Time `json:"queued_at"`
	QueueExpiresAt  *time.Time `json:"queue_expires_at"`
	QueueDaysLeft   int        `json:"queue_days_left"`
}

type todayResponse struct {
	Videos        []videoResponse `json:"videos"`
	NextReleaseAt time.Time       `json:"next_release_at"`
}

type queueResponse struct {
	Videos []videoResponse `json:"videos"`
}

func fromDomain(v library.Video, now time.Time) videoResponse {
	out := videoResponse{
		ID: v.ID, Title: v.Title, ChannelID: v.ChannelID, ChannelTitle: v.ChannelTitle,
		PublishedAt: v.PublishedAt.UTC(), Status: string(v.Status),
		QueueReason: v.QueueReason, QueueDaysLeft: v.QueueDaysLeft(now),
	}
	if v.Duration > 0 {
		sec := int64(v.Duration / time.Second)
		out.DurationSeconds = &sec
	}
	if v.Status == library.StatusQueued {
		queued, expires := v.QueuedAt.UTC(), v.QueueExpiresAt().UTC()
		out.QueuedAt, out.QueueExpiresAt = &queued, &expires
	}
	return out
}

func listFromDomain(videos []library.Video, now time.Time) []videoResponse {
	out := make([]videoResponse, 0, len(videos)) // [] e non null
	for _, v := range videos {
		out = append(out, fromDomain(v, now))
	}
	return out
}
