// Package library implementa su SQLite e su YouTube i contratti del soggetto library.
package library

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ste950/spestube/backend/internal/domain/library"
)

type scanner interface{ Scan(...any) error }

type channelRow struct {
	id             string
	title          string
	handle         string
	feedMode       string
	addedAt        int64
	lastFetchedAt  sql.NullInt64
	lastFetchError string
}

const channelColumns = `id, title, handle, feed_mode, added_at, last_fetched_at, last_fetch_error`

func (r *channelRow) scanFrom(s scanner) error {
	return s.Scan(&r.id, &r.title, &r.handle, &r.feedMode, &r.addedAt, &r.lastFetchedAt, &r.lastFetchError)
}

func (r channelRow) toEntity() (library.Channel, error) {
	mode := library.FeedMode(r.feedMode)
	if !mode.Valid() {
		return library.Channel{}, fmt.Errorf("canale %s: feed sconosciuto %q", r.id, r.feedMode)
	}
	return library.Channel{
		ID: r.id, Title: r.title, Handle: r.handle, FeedMode: mode,
		AddedAt:        unix(r.addedAt),
		LastFetchedAt:  nullUnix(r.lastFetchedAt),
		LastFetchError: r.lastFetchError,
	}, nil
}

type videoRow struct {
	id           string
	channelID    string
	channelTitle string
	title        string
	publishedAt  int64
	fetchedAt    int64
	source       string
	duration     sql.NullInt64
	isLive       bool
	status       string
	discard      string
	releasedAt   sql.NullInt64
	queueReason  string
	queuedAt     sql.NullInt64
	watchedAt    sql.NullInt64
	archivedAt   sql.NullInt64
}

// Le colonne sono qualificate perché ogni lettura fa la join con channels per il nome.
const videoColumns = `v.id, v.channel_id, c.title, v.title, v.published_at, v.fetched_at, v.source,
	v.duration_seconds, v.is_live, v.status, v.discard_reason, v.released_at, v.queue_reason,
	v.queued_at, v.watched_at, v.archived_at`

const videoFrom = ` FROM videos v JOIN channels c ON c.id = v.channel_id `

func (r *videoRow) scanFrom(s scanner) error {
	return s.Scan(&r.id, &r.channelID, &r.channelTitle, &r.title, &r.publishedAt, &r.fetchedAt, &r.source,
		&r.duration, &r.isLive, &r.status, &r.discard, &r.releasedAt, &r.queueReason,
		&r.queuedAt, &r.watchedAt, &r.archivedAt)
}

// toEntity rifiuta i valori fuori dagli insiemi chiusi: letti dal nostro database sono
// un dato corrotto, non qualcosa da far passare con un default.
func (r videoRow) toEntity() (library.Video, error) {
	status := library.Status(r.status)
	source := library.FeedMode(r.source)
	discard := library.DiscardReason(r.discard)
	if !status.Valid() || !source.Valid() || !discard.Valid() {
		return library.Video{}, fmt.Errorf("video %s: stato %q, feed %q o scarto %q sconosciuto",
			r.id, r.status, r.source, r.discard)
	}
	v := library.Video{
		ID: r.id, ChannelID: r.channelID, ChannelTitle: r.channelTitle, Title: r.title,
		PublishedAt: unix(r.publishedAt), FetchedAt: unix(r.fetchedAt), Source: source,
		IsLive: r.isLive, Status: status, Discard: discard,
		ReleasedAt: nullUnix(r.releasedAt), QueueReason: r.queueReason, QueuedAt: nullUnix(r.queuedAt),
		WatchedAt: nullUnix(r.watchedAt), ArchivedAt: nullUnix(r.archivedAt),
	}
	if r.duration.Valid {
		v.Duration = time.Duration(r.duration.Int64) * time.Second
	}
	return v, nil
}

func unix(sec int64) time.Time { return time.Unix(sec, 0).UTC() }

func nullUnix(n sql.NullInt64) time.Time {
	if !n.Valid {
		return time.Time{}
	}
	return unix(n.Int64)
}

// timeOrNull scrive NULL per l'istante zero: nel database "mai successo" è un'assenza,
// non il primo gennaio dell'anno 1.
func timeOrNull(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.Unix()
}

func durationOrNull(d time.Duration) any {
	if d <= 0 {
		return nil
	}
	return int64(d / time.Second)
}
