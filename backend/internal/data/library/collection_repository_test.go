package library

import (
	"context"
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/ste950/spestube/backend/internal/core/apperr"
	"github.com/ste950/spestube/backend/internal/core/sqlitedb"
	"github.com/ste950/spestube/backend/internal/data/migrations"
	"github.com/ste950/spestube/backend/internal/domain/library"
)

func TestCollectionRepository(t *testing.T) {
	ctx := context.Background()
	db := openDB(t)
	seedChannel(t, db)
	videos := NewVideoRepository(db)
	repo := NewCollectionRepository(db)
	if _, err := videos.AddMissing(ctx, []library.Video{video("a", library.StatusReleased), video("b", library.StatusReleased)}); err != nil {
		t.Fatal(err)
	}

	orto, err := repo.Create(ctx, library.CollectionFields{Name: "Orto", Description: "idee"}, t0)
	if err != nil {
		t.Fatal(err)
	}
	cucina, err := repo.Create(ctx, library.CollectionFields{Name: "Cucina"}, t0)
	if err != nil {
		t.Fatal(err)
	}
	// Il vincolo del database regge anche se il servizio non controllasse il nome.
	if _, err := repo.Create(ctx, library.CollectionFields{Name: "ORTO"}, t0); apperr.KindOf(err) != apperr.KindConflict {
		t.Errorf("nome duplicato: %v", err)
	}
	if found, ok, err := repo.FindByName(ctx, "orto"); err != nil || !ok || found.ID != orto {
		t.Errorf("FindByName = %+v, %v, %v", found, ok, err)
	}

	// Il video b entra dopo il video a: è lui la copertina.
	if err := repo.AddVideo(ctx, orto, "a", t0); err != nil {
		t.Fatal(err)
	}
	if err := repo.AddVideo(ctx, orto, "b", t0.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := repo.AddVideo(ctx, orto, "a", t0.Add(time.Hour)); err != nil {
		t.Errorf("aggiunta ripetuta: %v", err)
	}
	if err := repo.AddVideo(ctx, cucina, "a", t0); err != nil {
		t.Fatal(err)
	}

	got, err := repo.Get(ctx, orto)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Orto" || got.Description != "idee" || got.VideoCount != 2 || got.CoverVideoID != "b" || got.Position != 0 {
		t.Errorf("raccolta = %+v", got)
	}
	list, err := repo.Videos(ctx, orto)
	if err != nil || len(list) != 2 || list[0].ID != "b" || list[1].ID != "a" || !list[1].AddedAt.Equal(t0) {
		t.Errorf("video della raccolta = %+v, %v (l'aggiunta ripetuta non sposta la data)", list, err)
	}
	if v, _ := videos.Get(ctx, "a"); len(v.CollectionIDs) != 2 {
		t.Errorf("raccolte del video a = %v", v.CollectionIDs)
	}

	if err := repo.SetOrder(ctx, []int64{cucina, orto}, t0); err != nil {
		t.Fatal(err)
	}
	all, err := repo.List(ctx)
	if err != nil || len(all) != 2 || all[0].ID != cucina || all[1].Position != 1 {
		t.Errorf("ordine = %+v, %v", all, err)
	}

	if err := repo.Update(ctx, cucina, library.CollectionFields{Name: "Orto"}, t0); apperr.KindOf(err) != apperr.KindConflict {
		t.Errorf("rinomina su un nome usato: %v", err)
	}
	if err := repo.RemoveVideo(ctx, orto, "b"); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.Get(ctx, orto); got.VideoCount != 1 || got.CoverVideoID != "a" {
		t.Errorf("dopo la rimozione = %+v", got)
	}

	if err := repo.Delete(ctx, orto); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(ctx, orto); apperr.KindOf(err) != apperr.KindNotFound {
		t.Errorf("raccolta eliminata: %v", err)
	}
	if v, err := videos.Get(ctx, "b"); err != nil || len(v.CollectionIDs) != 0 {
		t.Errorf("il video deve restare, senza la raccolta eliminata: %+v, %v", v, err)
	}
	if err := repo.Delete(ctx, orto); apperr.KindOf(err) != apperr.KindNotFound {
		t.Errorf("seconda eliminazione: %v", err)
	}
}

// I video che erano in Da vedere quando si aggiorna l'app non si perdono: la migrazione li
// mette in una raccolta con lo stesso nome.
func TestMigrationMovesQueuedVideosIntoACollection(t *testing.T) {
	ctx := context.Background()
	db, err := sqlitedb.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	first, err := fs.ReadFile(migrations.FS, "0001_init.sql")
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlitedb.Migrate(ctx, db, fstest.MapFS{"0001_init.sql": {Data: first}}); err != nil {
		t.Fatal(err)
	}
	seedChannel(t, db)
	if _, err := db.ExecContext(ctx, `
		INSERT INTO videos (id, channel_id, title, published_at, fetched_at, source, status, queue_reason, queued_at)
		VALUES ('coda', ?, 'In coda', 0, 0, 'uulf', 'queued', 'per l''orto', ?),
		       ('oggi', ?, 'In Oggi', 0, 0, 'uulf', 'released', '', NULL)`,
		chanA, t0.Unix(), chanA); err != nil {
		t.Fatal(err)
	}

	if err := sqlitedb.Migrate(ctx, db, migrations.FS); err != nil {
		t.Fatal(err)
	}
	repo := NewCollectionRepository(db)
	cols, err := repo.List(ctx)
	if err != nil || len(cols) != 1 || cols[0].Name != "Da vedere" || cols[0].VideoCount != 1 {
		t.Fatalf("raccolte = %+v, %v", cols, err)
	}
	items, _ := repo.Videos(ctx, cols[0].ID)
	if len(items) != 1 || items[0].ID != "coda" || !items[0].AddedAt.Equal(t0) {
		t.Errorf("video della raccolta = %+v", items)
	}
	v, err := NewVideoRepository(db).Get(ctx, "coda")
	if err != nil || v.Status != library.StatusArchived {
		t.Errorf("il video che era in coda = %+v, %v", v, err)
	}
	if v, _ := NewVideoRepository(db).Get(ctx, "oggi"); v.Status != library.StatusReleased {
		t.Errorf("un video in Oggi non deve cambiare: %+v", v)
	}
}

func TestMigrationWithoutQueuedVideosCreatesNoCollection(t *testing.T) {
	db := openDB(t)
	cols, err := NewCollectionRepository(db).List(context.Background())
	if err != nil || len(cols) != 0 {
		t.Errorf("raccolte = %+v, %v", cols, err)
	}
}
