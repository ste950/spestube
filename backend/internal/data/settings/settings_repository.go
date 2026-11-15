// Package settings implementa su SQLite l'archivio delle impostazioni.
package settings

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) All(ctx context.Context) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT key, value FROM settings`)
	if err != nil {
		return nil, fmt.Errorf("lettura delle impostazioni: %w", err)
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, rows.Err()
}

// Put salva tutte le coppie in una transazione: un salvataggio a metà lascerebbe
// l'orario nuovo con i filtri vecchi.
func (r *Repository) Put(ctx context.Context, values map[string]string, at time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("salvataggio delle impostazioni: %w", err)
	}
	defer tx.Rollback()

	for k, v := range values {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)
			 ON CONFLICT (key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
			k, v, at.Unix()); err != nil {
			return fmt.Errorf("salvataggio dell'impostazione %s: %w", k, err)
		}
	}
	return tx.Commit()
}
