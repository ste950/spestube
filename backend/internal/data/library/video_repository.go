package library

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ste950/spestube/backend/internal/core/apperr"
	"github.com/ste950/spestube/backend/internal/domain/library"
)

type VideoRepository struct {
	db *sql.DB
}

func NewVideoRepository(db *sql.DB) *VideoRepository {
	return &VideoRepository{db: db}
}

func (r *VideoRepository) Get(ctx context.Context, id string) (library.Video, error) {
	var row videoRow
	err := row.scanFrom(r.db.QueryRowContext(ctx,
		`SELECT `+videoColumns+videoFrom+`WHERE v.id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return library.Video{}, apperr.NotFound("video_not_found", "video inesistente")
	}
	if err != nil {
		return library.Video{}, fmt.Errorf("lettura del video %s: %w", id, err)
	}
	return row.toEntity()
}

func (r *VideoRepository) AddMissing(ctx context.Context, videos []library.Video) (int, error) {
	if len(videos) == 0 {
		return 0, nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("inserimento dei video: %w", err)
	}
	defer tx.Rollback()

	added := 0
	for _, v := range videos {
		res, err := tx.ExecContext(ctx,
			`INSERT INTO videos (id, channel_id, title, published_at, fetched_at, source,
			                     duration_seconds, is_live, status, discard_reason)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT (id) DO NOTHING`,
			v.ID, v.ChannelID, v.Title, v.PublishedAt.Unix(), v.FetchedAt.Unix(), string(v.Source),
			durationOrNull(v.Duration), v.IsLive, string(v.Status), string(v.Discard))
		if err != nil {
			return 0, fmt.Errorf("inserimento del video %s: %w", v.ID, err)
		}
		n, _ := res.RowsAffected()
		added += int(n)
	}
	return added, tx.Commit()
}

func (r *VideoRepository) ListReleased(ctx context.Context) ([]library.Video, error) {
	return r.list(ctx, `WHERE v.status = ? ORDER BY v.published_at DESC`, string(library.StatusReleased))
}

// ListQueued ordina per ingresso in coda: in cima c'è il video più vicino alla scadenza.
func (r *VideoRepository) ListQueued(ctx context.Context) ([]library.Video, error) {
	return r.list(ctx, `WHERE v.status = ? ORDER BY v.queued_at ASC`, string(library.StatusQueued))
}

func (r *VideoRepository) ListNewFetchedBefore(ctx context.Context, cutoff time.Time) ([]library.Video, error) {
	return r.list(ctx, `WHERE v.status = ? AND v.fetched_at < ? ORDER BY v.published_at DESC`,
		string(library.StatusNew), cutoff.Unix())
}

func (r *VideoRepository) list(ctx context.Context, where string, args ...any) ([]library.Video, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+videoColumns+videoFrom+where, args...)
	if err != nil {
		return nil, fmt.Errorf("elenco dei video: %w", err)
	}
	defer rows.Close()

	var out []library.Video
	for rows.Next() {
		var row videoRow
		if err := row.scanFrom(rows); err != nil {
			return nil, err
		}
		v, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// Save scrive lo stato del video e ciò che lo accompagna. Titolo, canale e date di
// pubblicazione non cambiano dopo l'inserimento.
func (r *VideoRepository) Save(ctx context.Context, v library.Video) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE videos SET status = ?, discard_reason = ?, released_at = ?, queue_reason = ?,
		                   queued_at = ?, watched_at = ?, archived_at = ?
		 WHERE id = ?`,
		string(v.Status), string(v.Discard), timeOrNull(v.ReleasedAt), v.QueueReason,
		timeOrNull(v.QueuedAt), timeOrNull(v.WatchedAt), timeOrNull(v.ArchivedAt), v.ID)
	if err != nil {
		return fmt.Errorf("salvataggio del video %s: %w", v.ID, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return apperr.NotFound("video_not_found", "video inesistente")
	}
	return nil
}

func (r *VideoRepository) ArchiveReleasedBefore(ctx context.Context, cutoff, at time.Time) (int, error) {
	return r.archive(ctx, `status = ? AND released_at < ?`, library.StatusReleased, cutoff, at)
}

// ArchiveQueuedBefore archivia i video entrati in coda fino a cutoff compreso: allo
// scoccare della scadenza il video è scaduto.
func (r *VideoRepository) ArchiveQueuedBefore(ctx context.Context, cutoff, at time.Time) (int, error) {
	return r.archive(ctx, `status = ? AND queued_at <= ?`, library.StatusQueued, cutoff, at)
}

func (r *VideoRepository) archive(ctx context.Context, where string, status library.Status, cutoff, at time.Time) (int, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE videos SET status = ?, archived_at = ? WHERE `+where,
		string(library.StatusArchived), at.Unix(), string(status), cutoff.Unix())
	if err != nil {
		return 0, fmt.Errorf("archiviazione dei video: %w", err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}
