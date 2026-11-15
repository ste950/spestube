package library_test

import (
	"context"
	"testing"
	"time"

	"github.com/ste950/spestube/backend/internal/domain/library"
	"github.com/ste950/spestube/backend/internal/domain/library/librarytest"
)

type shelf struct {
	svc    *library.Collections
	store  *librarytest.Collections
	videos *librarytest.Videos
	now    time.Time
}

func newShelf(t *testing.T) *shelf {
	t.Helper()
	s := &shelf{videos: librarytest.NewVideos(), now: time.Date(2026, time.November, 15, 10, 0, 0, 0, time.UTC)}
	s.store = librarytest.NewCollections(s.videos)
	s.svc = library.NewCollections(s.store, s.videos, func() time.Time { return s.now })
	return s
}

func (s *shelf) create(t *testing.T, name string) library.Collection {
	t.Helper()
	c, err := s.svc.Create(context.Background(), name, "")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCreateCollection(t *testing.T) {
	s := newShelf(t)
	ctx := context.Background()

	c, err := s.svc.Create(ctx, "  Orto  ", " idee ")
	if err != nil {
		t.Fatal(err)
	}
	if c.Name != "Orto" || c.Description != "idee" || !c.CreatedAt.Equal(s.now) || c.VideoCount != 0 {
		t.Errorf("raccolta = %+v", c)
	}
	// Una nuova raccolta va in fondo: l'ordine scelto finora non cambia.
	second := s.create(t, "Cucina")
	list, _ := s.svc.List(ctx)
	if len(list) != 2 || list[0].ID != c.ID || list[1].ID != second.ID {
		t.Errorf("ordine = %+v", list)
	}

	if _, err := s.svc.Create(ctx, "ORTO", ""); codeOf(err) != "collection_exists" {
		t.Errorf("nome già usato: %v", err)
	}
	if _, err := s.svc.Create(ctx, "", ""); codeOf(err) != "collection_name_required" {
		t.Errorf("nome vuoto: %v", err)
	}
}

func TestUpdateCollection(t *testing.T) {
	s := newShelf(t)
	ctx := context.Background()
	orto := s.create(t, "Orto")
	s.create(t, "Cucina")

	s.now = s.now.Add(time.Hour)
	got, err := s.svc.Update(ctx, orto.ID, "orto", "con le maiuscole cambiate")
	if err != nil {
		t.Fatalf("rinominare una raccolta con il suo stesso nome: %v", err)
	}
	if got.Name != "orto" || got.Description != "con le maiuscole cambiate" || !got.UpdatedAt.Equal(s.now) {
		t.Errorf("raccolta = %+v", got)
	}
	if _, err := s.svc.Update(ctx, orto.ID, "Cucina", ""); codeOf(err) != "collection_exists" {
		t.Errorf("nome di un'altra raccolta: %v", err)
	}
	if _, err := s.svc.Update(ctx, 999, "Altro", ""); codeOf(err) != "collection_not_found" {
		t.Errorf("raccolta inesistente: %v", err)
	}
	if _, err := s.svc.Update(ctx, orto.ID, " ", ""); codeOf(err) != "collection_name_required" {
		t.Errorf("nome vuoto: %v", err)
	}
}

func TestReorderCollections(t *testing.T) {
	s := newShelf(t)
	ctx := context.Background()
	a, b, c := s.create(t, "A"), s.create(t, "B"), s.create(t, "C")

	if err := s.svc.Reorder(ctx, []int64{c.ID, a.ID, b.ID}); err != nil {
		t.Fatal(err)
	}
	list, _ := s.svc.List(ctx)
	if got := list[0].Name + list[1].Name + list[2].Name; got != "CAB" {
		t.Errorf("ordine = %s", got)
	}
	if err := s.svc.Reorder(ctx, []int64{a.ID, b.ID}); codeOf(err) != "collection_order_invalid" {
		t.Errorf("ordine incompleto: %v", err)
	}
}

func TestAddAndRemoveVideos(t *testing.T) {
	s := newShelf(t)
	ctx := context.Background()
	orto := s.create(t, "Orto")
	s.videos.ByID["uscito"] = library.Video{ID: "uscito", Status: library.StatusReleased}
	s.videos.ByID["archiviato"] = library.Video{ID: "archiviato", Status: library.StatusArchived}
	s.videos.ByID["nuovo"] = library.Video{ID: "nuovo", Status: library.StatusNew}
	s.videos.ByID["scartato"] = library.Video{ID: "scartato", Status: library.StatusDiscarded}

	for _, id := range []string{"uscito", "archiviato", "uscito"} {
		if err := s.svc.AddVideo(ctx, orto.ID, id); err != nil {
			t.Fatalf("aggiunta di %s: %v", id, err)
		}
	}
	_, videos, err := s.svc.Get(ctx, orto.ID)
	if err != nil || len(videos) != 2 {
		t.Fatalf("video = %+v, %v (un video aggiunto due volte c'è una volta sola)", videos, err)
	}

	// Un video mai mostrato non si può salvare: conoscerne l'id non basta.
	for _, id := range []string{"nuovo", "scartato", "inesistente"} {
		if err := s.svc.AddVideo(ctx, orto.ID, id); codeOf(err) != "video_not_found" {
			t.Errorf("aggiunta di %s: %v", id, err)
		}
	}
	if err := s.svc.AddVideo(ctx, 999, "uscito"); codeOf(err) != "collection_not_found" {
		t.Errorf("raccolta inesistente: %v", err)
	}

	if err := s.svc.RemoveVideo(ctx, orto.ID, "uscito"); err != nil {
		t.Fatal(err)
	}
	if err := s.svc.RemoveVideo(ctx, orto.ID, "uscito"); err != nil {
		t.Errorf("togliere un video che non c'è più: %v", err)
	}
	_, videos, _ = s.svc.Get(ctx, orto.ID)
	if len(videos) != 1 || videos[0].ID != "archiviato" {
		t.Errorf("video = %+v", videos)
	}
}

func TestDeleteCollectionKeepsTheVideos(t *testing.T) {
	s := newShelf(t)
	ctx := context.Background()
	orto := s.create(t, "Orto")
	s.videos.ByID["v"] = library.Video{ID: "v", Status: library.StatusReleased}
	if err := s.svc.AddVideo(ctx, orto.ID, "v"); err != nil {
		t.Fatal(err)
	}

	if err := s.svc.Delete(ctx, orto.ID); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.videos.ByID["v"]; !ok {
		t.Error("il video è sparito con la raccolta")
	}
	if err := s.svc.Delete(ctx, orto.ID); codeOf(err) != "collection_not_found" {
		t.Errorf("seconda eliminazione: %v", err)
	}
}
