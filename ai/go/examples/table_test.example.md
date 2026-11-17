# Test Table-Driven

I test del dominio: una tabella di casi con un nome ciascuno, il servizio vero, i fake al posto
dei repository. Vedi [Test](../rules/testing.rule.md).

I fake del soggetto, condivisi fra i test del dominio e quelli delle API:

```go
// backend/internal/domain/catalog/catalogtest/fake.go

// Package catalogtest contiene i fake del soggetto catalog, usati dai test del dominio
// e da quelli delle API.
package catalogtest

import (
	"context"
	"sort"
	"sync"
	"time"

	"example.com/backend/internal/core/apperr"
	"example.com/backend/internal/domain/catalog"
)

// Il compilatore verifica qui che il fake soddisfi ancora il contratto: se
// l'interfaccia cambia, si rompe questa riga e non un test a caso.
var _ catalog.ItemRepository = (*Items)(nil)

// Items è un archivio in memoria. Il mutex c'è perché i test delle API passano dal
// server HTTP, che serve ogni richiesta in una goroutine.
type Items struct {
	mu     sync.Mutex
	byID   map[int64]catalog.Item
	nextID int64
}

func NewItems(seed ...catalog.Item) *Items {
	f := &Items{byID: map[int64]catalog.Item{}}
	for _, it := range seed {
		f.nextID++
		it.ID = f.nextID
		f.byID[it.ID] = it
	}
	return f
}

func (f *Items) Get(_ context.Context, id int64) (catalog.Item, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	it, ok := f.byID[id]
	if !ok {
		return catalog.Item{}, apperr.NotFound("item_not_found", "item inesistente")
	}
	return it, nil
}

func (f *Items) FindByURL(_ context.Context, url string) (catalog.Item, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, it := range f.byID {
		if it.URL == url {
			return it, true, nil
		}
	}
	return catalog.Item{}, false, nil
}

func (f *Items) Add(_ context.Context, it catalog.Item) (catalog.Item, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	it.ID = f.nextID
	f.byID[it.ID] = it
	return it, nil
}

func (f *Items) ListActive(_ context.Context) ([]catalog.Item, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []catalog.Item
	for _, it := range f.byID {
		if it.Status == catalog.StatusActive {
			out = append(out, it)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (f *Items) SetTitle(_ context.Context, id int64, title string, at time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	it, ok := f.byID[id]
	if !ok {
		return apperr.NotFound("item_not_found", "item inesistente")
	}
	it.Title, it.RefreshedAt = title, at
	f.byID[id] = it
	return nil
}

// Titles risponde da una mappa; un URL assente simula una sorgente irraggiungibile.
type Titles map[string]string

func (t Titles) Title(_ context.Context, url string) (string, error) {
	if title, ok := t[url]; ok {
		return title, nil
	}
	return "", apperr.Unavailable("source_unavailable", "sorgente irraggiungibile", nil)
}
```

I test del servizio:

```go
// backend/internal/domain/catalog/service_test.go

package catalog_test

import (
	"context"
	"testing"
	"time"

	"example.com/backend/internal/core/apperr"
	"example.com/backend/internal/domain/catalog"
	"example.com/backend/internal/domain/catalog/catalogtest"
)

var now = time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)

func fixedNow() time.Time { return now }

func TestAdd(t *testing.T) {
	existing := catalog.Item{URL: "https://example.org/a", Status: catalog.StatusActive}

	tests := []struct {
		name     string
		url      string
		wantKind apperr.Kind // KindInternal = nessun errore atteso
		wantErr  bool
	}{
		{name: "url valido", url: "https://example.org/b"},
		{name: "schema non http", url: "ftp://example.org/b", wantErr: true, wantKind: apperr.KindInvalid},
		{name: "senza host", url: "https://", wantErr: true, wantKind: apperr.KindInvalid},
		{name: "già salvato", url: "https://example.org/a", wantErr: true, wantKind: apperr.KindConflict},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := catalog.NewService(catalogtest.NewItems(existing), catalogtest.Titles{}, fixedNow)

			got, err := svc.Add(context.Background(), tc.url)

			if tc.wantErr {
				if k := apperr.KindOf(err); err == nil || k != tc.wantKind {
					t.Fatalf("errore = %v (kind %d), atteso kind %d", err, k, tc.wantKind)
				}
				return
			}
			if err != nil {
				t.Fatalf("errore inatteso: %v", err)
			}
			if got.Title != "example.org" || !got.CreatedAt.Equal(now) {
				t.Errorf("item = %+v, atteso titolo provvisorio example.org creato alle %v", got, now)
			}
		})
	}
}

func TestRefreshStaleSkipsUnavailableSources(t *testing.T) {
	items := catalogtest.NewItems(
		catalog.Item{URL: "https://up.example", Status: catalog.StatusActive},
		catalog.Item{URL: "https://down.example", Status: catalog.StatusActive},
	)
	titles := catalogtest.Titles{"https://up.example": "Su"}
	svc := catalog.NewService(items, titles, fixedNow)

	err := svc.RefreshStale(context.Background())

	if apperr.KindOf(err) != apperr.KindUnavailable {
		t.Fatalf("errore = %v, atteso l'errore della sorgente giù", err)
	}
	up, _ := items.Get(context.Background(), 1)
	if up.Title != "Su" || !up.RefreshedAt.Equal(now) {
		t.Errorf("item raggiungibile = %+v, atteso aggiornato nonostante l'altro", up)
	}
}
```

I test del repository, su un database vero in una cartella temporanea:

```go
// backend/internal/data/catalog/item_repository_test.go

package catalog_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"example.com/backend/internal/core/apperr"
	"example.com/backend/internal/core/sqlitedb"
	datacatalog "example.com/backend/internal/data/catalog"
	"example.com/backend/internal/data/migrations"
	"example.com/backend/internal/domain/catalog"
)

// newRepo apre un database vero su un file temporaneo, con le migrazioni vere: un
// database in memoria si comporterebbe diversamente proprio su WAL e lock.
func newRepo(t *testing.T) *datacatalog.ItemRepository {
	t.Helper()
	ctx := context.Background()
	db, err := sqlitedb.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := sqlitedb.Migrate(ctx, db, migrations.FS); err != nil {
		t.Fatal(err)
	}
	return datacatalog.NewItemRepository(db)
}

func TestItemRepository(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	created := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	added, err := repo.Add(ctx, catalog.Item{
		URL: "https://example.org", Title: "x", Status: catalog.StatusActive, CreatedAt: created,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("lettura", func(t *testing.T) {
		got, err := repo.Get(ctx, added.ID)
		if err != nil || got.URL != added.URL || !got.CreatedAt.Equal(created) {
			t.Fatalf("Get = %+v, %v", got, err)
		}
	})
	t.Run("assenza", func(t *testing.T) {
		if _, err := repo.Get(ctx, 999); apperr.KindOf(err) != apperr.KindNotFound {
			t.Fatalf("errore = %v, atteso NotFound", err)
		}
		if _, ok, err := repo.FindByURL(ctx, "https://altro.org"); ok || err != nil {
			t.Fatalf("FindByURL = %v, %v; attesa assenza senza errore", ok, err)
		}
	})
	t.Run("url unico", func(t *testing.T) {
		if _, err := repo.Add(ctx, added); err == nil {
			t.Fatal("secondo inserimento dello stesso url accettato")
		}
	})
}
```
