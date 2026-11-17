package library

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"

	"github.com/ste950/spestube/backend/internal/core/apperr"
	"github.com/ste950/spestube/backend/internal/domain/library"
)

type CollectionRepository struct {
	db *sql.DB
}

func NewCollectionRepository(db *sql.DB) *CollectionRepository {
	return &CollectionRepository{db: db}
}

func (r *CollectionRepository) List(ctx context.Context) ([]library.Collection, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+collectionColumns+` FROM collections c ORDER BY c.position, c.id`)
	if err != nil {
		return nil, fmt.Errorf("elenco delle raccolte: %w", err)
	}
	defer rows.Close()

	var out []library.Collection
	for rows.Next() {
		var row collectionRow
		if err := row.scanFrom(rows); err != nil {
			return nil, err
		}
		out = append(out, row.toEntity())
	}
	return out, rows.Err()
}

func (r *CollectionRepository) Get(ctx context.Context, id int64) (library.Collection, error) {
	var row collectionRow
	err := row.scanFrom(r.db.QueryRowContext(ctx,
		`SELECT `+collectionColumns+` FROM collections c WHERE c.id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return library.Collection{}, collectionNotFound()
	}
	if err != nil {
		return library.Collection{}, fmt.Errorf("lettura della raccolta %d: %w", id, err)
	}
	return row.toEntity(), nil
}

func (r *CollectionRepository) FindByName(ctx context.Context, name string) (library.Collection, bool, error) {
	var row collectionRow
	// La colonna è COLLATE NOCASE: il confronto ignora le maiuscole da sé.
	err := row.scanFrom(r.db.QueryRowContext(ctx,
		`SELECT `+collectionColumns+` FROM collections c WHERE c.name = ?`, name))
	if errors.Is(err, sql.ErrNoRows) {
		return library.Collection{}, false, nil
	}
	if err != nil {
		return library.Collection{}, false, fmt.Errorf("ricerca della raccolta per nome: %w", err)
	}
	return row.toEntity(), true, nil
}

func (r *CollectionRepository) Create(ctx context.Context, f library.CollectionFields, at time.Time) (int64, error) {
	// La posizione si calcola nella stessa istruzione: due creazioni insieme non possono
	// prendere lo stesso posto.
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO collections (name, description, position, created_at, updated_at)
		 VALUES (?, ?, (SELECT COALESCE(MAX(position), -1) + 1 FROM collections), ?, ?)`,
		f.Name, f.Description, at.Unix(), at.Unix())
	if err != nil {
		return 0, translateNameConflict(err, "creazione della raccolta")
	}
	return res.LastInsertId()
}

func (r *CollectionRepository) Update(ctx context.Context, id int64, f library.CollectionFields, at time.Time) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE collections SET name = ?, description = ?, updated_at = ? WHERE id = ?`,
		f.Name, f.Description, at.Unix(), id)
	if err != nil {
		return translateNameConflict(err, fmt.Sprintf("modifica della raccolta %d", id))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return collectionNotFound()
	}
	return nil
}

func (r *CollectionRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM collections WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("eliminazione della raccolta %d: %w", id, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return collectionNotFound()
	}
	return nil
}

func (r *CollectionRepository) SetOrder(ctx context.Context, ids []int64, at time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ordine delle raccolte: %w", err)
	}
	defer tx.Rollback()

	for position, id := range ids {
		if _, err := tx.ExecContext(ctx,
			`UPDATE collections SET position = ?, updated_at = ? WHERE id = ?`,
			position, at.Unix(), id); err != nil {
			return fmt.Errorf("posizione della raccolta %d: %w", id, err)
		}
	}
	return tx.Commit()
}

func (r *CollectionRepository) Videos(ctx context.Context, id int64) ([]library.CollectionVideo, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+videoColumns+`, i.added_at`+videoFrom+
			`JOIN collection_items i ON i.video_id = v.id
			 WHERE i.collection_id = ?
			 ORDER BY i.added_at DESC, v.id`, id)
	if err != nil {
		return nil, fmt.Errorf("video della raccolta %d: %w", id, err)
	}
	defer rows.Close()

	var out []library.CollectionVideo
	for rows.Next() {
		var row videoRow
		var addedAt int64
		if err := row.scanFrom(appendScan{rows, &addedAt}); err != nil {
			return nil, err
		}
		v, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		out = append(out, library.CollectionVideo{Video: v, AddedAt: unix(addedAt)})
	}
	return out, rows.Err()
}

func (r *CollectionRepository) AddVideo(ctx context.Context, collectionID int64, videoID string, at time.Time) error {
	if _, err := r.db.ExecContext(ctx,
		`INSERT INTO collection_items (collection_id, video_id, added_at) VALUES (?, ?, ?)
		 ON CONFLICT (collection_id, video_id) DO NOTHING`,
		collectionID, videoID, at.Unix()); err != nil {
		return fmt.Errorf("aggiunta del video %s alla raccolta %d: %w", videoID, collectionID, err)
	}
	return nil
}

func (r *CollectionRepository) RemoveVideo(ctx context.Context, collectionID int64, videoID string) error {
	if _, err := r.db.ExecContext(ctx,
		`DELETE FROM collection_items WHERE collection_id = ? AND video_id = ?`,
		collectionID, videoID); err != nil {
		return fmt.Errorf("rimozione del video %s dalla raccolta %d: %w", videoID, collectionID, err)
	}
	return nil
}

// appendScan aggiunge una destinazione in coda a quelle di scanFrom: la riga di un video in
// una raccolta è la riga del video più la data di ingresso.
type appendScan struct {
	s     scanner
	extra any
}

func (a appendScan) Scan(dest ...any) error {
	return a.s.Scan(append(dest, a.extra)...)
}

// translateNameConflict riconosce il vincolo di unicità sul nome: è il caso di due
// salvataggi con lo stesso nome partiti insieme, che il controllo del servizio non vede.
func translateNameConflict(err error, what string) error {
	var se *sqlite.Error
	if errors.As(err, &se) && se.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
		return apperr.Conflict("collection_exists", "hai già una raccolta con questo nome")
	}
	return fmt.Errorf("%s: %w", what, err)
}

func collectionNotFound() error {
	return apperr.NotFound("collection_not_found", "raccolta inesistente")
}
