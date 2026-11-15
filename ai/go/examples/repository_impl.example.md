# Implementazione di Repository

L'implementazione SQL di un contratto del dominio, in `internal/data/<soggetto>/`. Non dichiara
l'interfaccia che soddisfa: lo verifica il compilatore nel composition root. Vedi
[Persistenza](../rules/persistence.rule.md).

Il mapper fra riga ed entità, con i tipi delle colonne:

```go
// backend/internal/data/catalog/mapper.go

// Package catalog implementa su SQLite e su HTTP i contratti del soggetto catalog.
package catalog

import (
	"database/sql"
	"fmt"
	"time"

	"example.com/backend/internal/domain/catalog"
)

// itemRow è la riga così come la legge Scan: i tipi sono quelli della colonna, non
// quelli del dominio.
type itemRow struct {
	id          int64
	url         string
	title       string
	status      string
	createdAt   int64
	refreshedAt sql.NullInt64
}

const itemColumns = `id, url, title, status, created_at, refreshed_at`

func (r *itemRow) scanFrom(s interface{ Scan(...any) error }) error {
	return s.Scan(&r.id, &r.url, &r.title, &r.status, &r.createdAt, &r.refreshedAt)
}

// toEntity rifiuta uno stato sconosciuto: letto dal nostro database è un dato
// corrotto, non un valore da far passare con un default.
func toEntity(r itemRow) (catalog.Item, error) {
	status := catalog.Status(r.status)
	if !status.Valid() {
		return catalog.Item{}, fmt.Errorf("item %d: stato sconosciuto %q", r.id, r.status)
	}
	it := catalog.Item{
		ID:        r.id,
		URL:       r.url,
		Title:     r.title,
		Status:    status,
		CreatedAt: time.Unix(r.createdAt, 0).UTC(),
	}
	if r.refreshedAt.Valid {
		it.RefreshedAt = time.Unix(r.refreshedAt.Int64, 0).UTC()
	}
	return it, nil
}
```

Il repository:

```go
// backend/internal/data/catalog/item_repository.go

package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"example.com/backend/internal/core/apperr"
	"example.com/backend/internal/domain/catalog"
)

// ItemRepository soddisfa catalog.ItemRepository senza dichiararlo: lo verifica il
// compilatore nel composition root, dove l'una diventa l'altra.
type ItemRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) Get(ctx context.Context, id int64) (catalog.Item, error) {
	var row itemRow
	err := row.scanFrom(r.db.QueryRowContext(ctx,
		`SELECT `+itemColumns+` FROM items WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return catalog.Item{}, apperr.NotFound("item_not_found", "item inesistente")
	}
	if err != nil {
		return catalog.Item{}, fmt.Errorf("lettura dell'item %d: %w", id, err)
	}
	return toEntity(row)
}

func (r *ItemRepository) FindByURL(ctx context.Context, url string) (catalog.Item, bool, error) {
	var row itemRow
	err := row.scanFrom(r.db.QueryRowContext(ctx,
		`SELECT `+itemColumns+` FROM items WHERE url = ?`, url))
	if errors.Is(err, sql.ErrNoRows) {
		return catalog.Item{}, false, nil
	}
	if err != nil {
		return catalog.Item{}, false, fmt.Errorf("ricerca per url: %w", err)
	}
	it, err := toEntity(row)
	return it, err == nil, err
}

func (r *ItemRepository) Add(ctx context.Context, it catalog.Item) (catalog.Item, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO items (url, title, status, created_at) VALUES (?, ?, ?, ?)`,
		it.URL, it.Title, string(it.Status), it.CreatedAt.Unix())
	if err != nil {
		return catalog.Item{}, fmt.Errorf("inserimento dell'item: %w", err)
	}
	if it.ID, err = res.LastInsertId(); err != nil {
		return catalog.Item{}, err
	}
	return it, nil
}

func (r *ItemRepository) ListActive(ctx context.Context) ([]catalog.Item, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+itemColumns+` FROM items WHERE status = ? ORDER BY created_at DESC`,
		string(catalog.StatusActive))
	if err != nil {
		return nil, fmt.Errorf("elenco degli item: %w", err)
	}
	defer rows.Close()

	var items []catalog.Item
	for rows.Next() {
		var row itemRow
		if err := row.scanFrom(rows); err != nil {
			return nil, err
		}
		it, err := toEntity(row)
		if err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	// Un errore a metà lettura non arriva da Next, che si limita a restituire false.
	return items, rows.Err()
}

func (r *ItemRepository) SetTitle(ctx context.Context, id int64, title string, at time.Time) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE items SET title = ?, refreshed_at = ? WHERE id = ?`, title, at.Unix(), id)
	if err != nil {
		return fmt.Errorf("aggiornamento del titolo dell'item %d: %w", id, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return apperr.NotFound("item_not_found", "item inesistente")
	}
	return nil
}
```

-   **`sql.ErrNoRows` diventa `NotFound`** qui, e da qui non esce nessun errore di
    `database/sql` non avvolto.
-   **`rows.Close()` in `defer` e `rows.Err()` dopo il ciclo**: senza il secondo, una lettura
    interrotta a metà restituisce una lista corta e nessun errore.
-   **Nessuna logica di business**: l'ordinamento per data è un `ORDER BY`, non una decisione.
