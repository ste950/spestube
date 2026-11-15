package settings

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/ste950/spestube/backend/internal/core/sqlitedb"
	"github.com/ste950/spestube/backend/internal/data/migrations"
)

func TestPutThenAll(t *testing.T) {
	ctx := context.Background()
	db, err := sqlitedb.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := sqlitedb.Migrate(ctx, db, migrations.FS); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)

	if got, err := repo.All(ctx); err != nil || len(got) != 0 {
		t.Fatalf("database nuovo: %v, %v", got, err)
	}
	now := time.Unix(1_800_000_000, 0)
	if err := repo.Put(ctx, map[string]string{"release_time": "07:00", "exclude_live": "true"}, now); err != nil {
		t.Fatal(err)
	}
	// Una seconda scrittura sovrascrive la chiave e lascia stare le altre.
	if err := repo.Put(ctx, map[string]string{"release_time": "19:00"}, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	got, err := repo.All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got["release_time"] != "19:00" || got["exclude_live"] != "true" {
		t.Errorf("impostazioni = %v", got)
	}
}
