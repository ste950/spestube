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

type ChannelRepository struct {
	db *sql.DB
}

func NewChannelRepository(db *sql.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

func (r *ChannelRepository) Get(ctx context.Context, id string) (library.Channel, error) {
	ch, ok, err := r.Find(ctx, id)
	if err != nil {
		return library.Channel{}, err
	}
	if !ok {
		return library.Channel{}, channelNotFound()
	}
	return ch, nil
}

func (r *ChannelRepository) Find(ctx context.Context, id string) (library.Channel, bool, error) {
	var row channelRow
	err := row.scanFrom(r.db.QueryRowContext(ctx,
		`SELECT `+channelColumns+` FROM channels WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return library.Channel{}, false, nil
	}
	if err != nil {
		return library.Channel{}, false, fmt.Errorf("lettura del canale %s: %w", id, err)
	}
	ch, err := row.toEntity()
	return ch, err == nil, err
}

func (r *ChannelRepository) Add(ctx context.Context, ch library.Channel) error {
	// Il conflitto sulla chiave si traduce qui: due aggiunte dello stesso canale partite
	// insieme superano entrambe il controllo del servizio, e solo il database le distingue.
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO channels (id, title, handle, feed_mode, added_at) VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT (id) DO NOTHING`,
		ch.ID, ch.Title, ch.Handle, string(ch.FeedMode), ch.AddedAt.Unix())
	if err != nil {
		return fmt.Errorf("inserimento del canale %s: %w", ch.ID, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return apperr.Conflict("channel_exists", "segui già questo canale")
	}
	return nil
}

func (r *ChannelRepository) List(ctx context.Context) ([]library.Channel, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+channelColumns+` FROM channels ORDER BY title COLLATE NOCASE`)
	if err != nil {
		return nil, fmt.Errorf("elenco dei canali: %w", err)
	}
	defer rows.Close()

	var out []library.Channel
	for rows.Next() {
		var row channelRow
		if err := row.scanFrom(rows); err != nil {
			return nil, err
		}
		ch, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		out = append(out, ch)
	}
	return out, rows.Err()
}

func (r *ChannelRepository) Remove(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM channels WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("rimozione del canale %s: %w", id, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return channelNotFound()
	}
	return nil
}

func (r *ChannelRepository) MarkFetched(ctx context.Context, id string, mode library.FeedMode, at time.Time, fetchErr string) error {
	// Dopo un errore last_fetched_at non avanza: resta l'istante dell'ultima lettura
	// riuscita, ed è ciò che dice se il canale ha mai avuto un primo download.
	query := `UPDATE channels SET feed_mode = ?, last_fetched_at = ?, last_fetch_error = '' WHERE id = ?`
	args := []any{string(mode), at.Unix(), id}
	if fetchErr != "" {
		query = `UPDATE channels SET last_fetch_error = ? WHERE id = ?`
		args = []any{fetchErr, id}
	}
	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("esito del download del canale %s: %w", id, err)
	}
	return nil
}

func channelNotFound() error {
	return apperr.NotFound("channel_not_found", "canale inesistente")
}
