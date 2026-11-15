# Database SQLite

L'apertura del database con i pragma nella DSN, e il motore delle migrazioni. Vedi
[Persistenza](../rules/persistence.rule.md).

```go
// backend/internal/core/sqlitedb/sqlitedb.go

// Package sqlitedb apre il database SQLite con i pragma che il resto del codice dà per
// scontati, e applica le migrazioni.
package sqlitedb

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"net/url"
	"sort"
	"strings"

	_ "modernc.org/sqlite" // registra il driver "sqlite", in Go puro: niente CGO
)

// Open apre il file con i pragma nella DSN, così valgono per ogni connessione del
// pool e non solo per la prima.
func Open(ctx context.Context, path string) (*sql.DB, error) {
	q := url.Values{}
	for _, p := range []string{
		"journal_mode(WAL)",
		"busy_timeout(5000)",
		"foreign_keys(1)",
		"synchronous(NORMAL)",
	} {
		q.Add("_pragma", p)
	}
	// IMMEDIATE: la transazione prende il lock di scrittura subito. Con DEFERRED due
	// transazioni che leggono e poi scrivono possono bloccarsi a vicenda, e
	// busy_timeout non le sblocca.
	q.Set("_txlock", "immediate")

	db, err := sql.Open("sqlite", "file:"+path+"?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("apertura di %s: %w", path, err)
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("apertura di %s: %w", path, err)
	}
	return db, nil
}

// Migrate applica in ordine i file .sql di migrations non ancora registrati, ognuno
// nella sua transazione. Il nome del file è la versione.
func Migrate(ctx context.Context, db *sql.DB, migrations fs.FS) error {
	if _, err := db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY)`); err != nil {
		return fmt.Errorf("tabella delle migrazioni: %w", err)
	}
	names, err := fs.Glob(migrations, "*.sql")
	if err != nil {
		return err
	}
	sort.Strings(names) // 0001_, 0002_…: l'ordine lessicografico è l'ordine di esecuzione

	for _, name := range names {
		version := strings.TrimSuffix(name, ".sql")
		var done int
		if err := db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version).Scan(&done); err != nil {
			return err
		}
		if done > 0 {
			continue
		}
		body, err := fs.ReadFile(migrations, name)
		if err != nil {
			return err
		}
		if err := apply(ctx, db, version, string(body)); err != nil {
			return fmt.Errorf("migrazione %s: %w", name, err)
		}
	}
	return nil
}

func apply(ctx context.Context, db *sql.DB, version, body string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // dopo un Commit riuscito non fa niente

	if _, err := tx.ExecContext(ctx, body); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
		return err
	}
	return tx.Commit()
}
```

Un `*sql.DB` è un pool, sicuro da usare da più goroutine: si apre una volta in `bootstrap/` e
si passa ai costruttori dei repository. Le scritture concorrenti si serializzano da sole sul
lock di SQLite, che grazie a `busy_timeout` aspetta invece di fallire.
