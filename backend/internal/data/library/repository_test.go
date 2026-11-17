package library

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/ste950/spestube/backend/internal/core/apperr"
	"github.com/ste950/spestube/backend/internal/core/sqlitedb"
	"github.com/ste950/spestube/backend/internal/data/migrations"
	"github.com/ste950/spestube/backend/internal/domain/library"
)

const chanA = "UCaaaaaaaaaaaaaaaaaaaaaa"

var t0 = time.Date(2026, time.November, 15, 10, 0, 0, 0, time.UTC)

func openDB(t *testing.T) *sql.DB {
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
	return db
}

func seedChannel(t *testing.T, db *sql.DB) *ChannelRepository {
	t.Helper()
	repo := NewChannelRepository(db)
	err := repo.Add(context.Background(), library.Channel{
		ID: chanA, Title: "Officina Radio", Handle: "@officina", FeedMode: library.FeedLongVideos, AddedAt: t0,
	})
	if err != nil {
		t.Fatal(err)
	}
	return repo
}

func video(id string, status library.Status) library.Video {
	return library.Video{
		ID: id, ChannelID: chanA, Title: "Titolo " + id, PublishedAt: t0.Add(-time.Hour),
		FetchedAt: t0, Source: library.FeedLongVideos, Status: status,
	}
}

func TestChannelRepository(t *testing.T) {
	ctx := context.Background()
	db := openDB(t)
	repo := seedChannel(t, db)

	got, err := repo.Get(ctx, chanA)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Officina Radio" || got.Handle != "@officina" || !got.AddedAt.Equal(t0) || !got.LastFetchedAt.IsZero() {
		t.Errorf("canale = %+v", got)
	}

	if err := repo.Add(ctx, library.Channel{ID: chanA, Title: "Doppio", FeedMode: library.FeedLongVideos, AddedAt: t0}); apperr.KindOf(err) != apperr.KindConflict {
		t.Errorf("aggiunta doppia: %v, atteso Conflict", err)
	}
	if _, ok, err := repo.Find(ctx, "UCmanca"); ok || err != nil {
		t.Errorf("Find di un canale assente: ok=%v err=%v", ok, err)
	}
	if _, err := repo.Get(ctx, "UCmanca"); apperr.KindOf(err) != apperr.KindNotFound {
		t.Errorf("Get di un canale assente: %v", err)
	}

	// Un errore si registra senza toccare la data dell'ultima lettura riuscita.
	if err := repo.MarkFetched(ctx, chanA, library.FeedChannel, t0, ""); err != nil {
		t.Fatal(err)
	}
	if err := repo.MarkFetched(ctx, chanA, library.FeedLongVideos, t0.Add(time.Hour), "YouTube non risponde"); err != nil {
		t.Fatal(err)
	}
	got, _ = repo.Get(ctx, chanA)
	if got.FeedMode != library.FeedChannel || !got.LastFetchedAt.Equal(t0) || got.LastFetchError != "YouTube non risponde" {
		t.Errorf("dopo un errore: %+v", got)
	}

	list, err := repo.List(ctx)
	if err != nil || len(list) != 1 {
		t.Errorf("List = %d canali, %v", len(list), err)
	}
}

func TestRemoveChannelRemovesItsVideos(t *testing.T) {
	ctx := context.Background()
	db := openDB(t)
	channels := seedChannel(t, db)
	videos := NewVideoRepository(db)
	if _, err := videos.AddMissing(ctx, []library.Video{video("a", library.StatusNew)}); err != nil {
		t.Fatal(err)
	}

	if err := channels.Remove(ctx, chanA); err != nil {
		t.Fatal(err)
	}
	if _, err := videos.Get(ctx, "a"); apperr.KindOf(err) != apperr.KindNotFound {
		t.Errorf("il video del canale rimosso c'è ancora: %v", err)
	}
	if err := channels.Remove(ctx, chanA); apperr.KindOf(err) != apperr.KindNotFound {
		t.Errorf("rimozione di un canale assente: %v", err)
	}
}

func TestAddMissingLeavesExistingVideosAlone(t *testing.T) {
	ctx := context.Background()
	db := openDB(t)
	seedChannel(t, db)
	repo := NewVideoRepository(db)

	first := video("a", library.StatusNew)
	first.Duration = 18 * time.Minute
	if n, err := repo.AddMissing(ctx, []library.Video{first}); err != nil || n != 1 {
		t.Fatalf("primo inserimento: %d, %v", n, err)
	}
	released := first.Release(t0, library.Filters{})
	if err := repo.Save(ctx, released); err != nil {
		t.Fatal(err)
	}

	// Il feed riporta lo stesso video al giro dopo: non deve tornare nuovo.
	again := video("a", library.StatusNew)
	again.Title = "Titolo cambiato"
	n, err := repo.AddMissing(ctx, []library.Video{again, video("b", library.StatusNew)})
	if err != nil || n != 1 {
		t.Fatalf("secondo inserimento: %d, %v", n, err)
	}
	got, err := repo.Get(ctx, "a")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != library.StatusReleased || got.Title != "Titolo a" || got.Duration != 18*time.Minute || got.ChannelTitle != "Officina Radio" {
		t.Errorf("video = %+v", got)
	}
	if n, err := repo.AddMissing(ctx, nil); n != 0 || err != nil {
		t.Errorf("lista vuota: %d, %v", n, err)
	}
}

func TestVideoListsAndArchiving(t *testing.T) {
	ctx := context.Background()
	db := openDB(t)
	seedChannel(t, db)
	repo := NewVideoRepository(db)

	early, late := video("presto", library.StatusNew), video("tardi", library.StatusNew)
	early.FetchedAt, late.FetchedAt = t0.Add(-time.Hour), t0.Add(time.Hour)
	if _, err := repo.AddMissing(ctx, []library.Video{early, late, video("oggi", library.StatusNew), video("ieri", library.StatusNew)}); err != nil {
		t.Fatal(err)
	}
	save := func(v library.Video) {
		t.Helper()
		if err := repo.Save(ctx, v); err != nil {
			t.Fatal(err)
		}
	}
	oggi, ieri := video("oggi", library.StatusReleased), video("ieri", library.StatusReleased)
	oggi.ReleasedAt, ieri.ReleasedAt = t0, t0.Add(-24*time.Hour)
	save(oggi)
	save(ieri)

	pending, err := repo.ListNewFetchedBefore(ctx, t0)
	if err != nil || len(pending) != 1 || pending[0].ID != "presto" {
		t.Errorf("nuovi scaricati prima del limite = %+v, %v", pending, err)
	}

	if n, err := repo.ArchiveReleasedBefore(ctx, t0, t0.Add(time.Minute)); err != nil || n != 1 {
		t.Errorf("archiviati da Oggi = %d, %v", n, err)
	}
	released, err := repo.ListReleased(ctx)
	if err != nil || len(released) != 1 || released[0].ID != "oggi" {
		t.Errorf("rilasciati = %+v, %v", released, err)
	}
	archived, _ := repo.Get(ctx, "ieri")
	if archived.Status != library.StatusArchived || !archived.ArchivedAt.Equal(t0.Add(time.Minute)) {
		t.Errorf("archiviato = %+v", archived)
	}

}

func TestSaveOfAMissingVideo(t *testing.T) {
	db := openDB(t)
	seedChannel(t, db)
	if err := NewVideoRepository(db).Save(context.Background(), video("manca", library.StatusReleased)); apperr.KindOf(err) != apperr.KindNotFound {
		t.Errorf("salvataggio di un video assente: %v", err)
	}
}

func TestLongVideosPlaylistID(t *testing.T) {
	if got := LongVideosPlaylistID("UCHnyfMqiRRG1u-2MsSQLbXA"); got != "UULFHnyfMqiRRG1u-2MsSQLbXA" {
		t.Errorf("playlist = %q", got)
	}
}

func TestHandleRepositoryWithoutAPIKey(t *testing.T) {
	_, err := NewHandleRepository(nil).ChannelID(context.Background(), "@officina")
	if apperr.KindOf(err) != apperr.KindInvalid {
		t.Errorf("senza chiave: %v, atteso Invalid", err)
	}
}
