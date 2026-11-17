package sources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ste950/spestube/backend/internal/core/apperr"
)

func fixture(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/uulf_feed.xml")
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestParseFeed(t *testing.T) {
	feed, err := ParseFeed(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if feed.Author != "Veritasium" {
		t.Errorf("Author = %q, atteso Veritasium", feed.Author)
	}
	// Il titolo del feed di una playlist non è il nome del canale.
	if feed.Title != "Videos" {
		t.Errorf("Title = %q, atteso Videos", feed.Title)
	}
	if len(feed.Entries) != 3 {
		t.Fatalf("voci = %d, attese 3", len(feed.Entries))
	}
	first := feed.Entries[0]
	if first.VideoID != "KfODG5REyss" {
		t.Errorf("VideoID = %q", first.VideoID)
	}
	if want := "I finally saw quantum mechanics with my own eyes... (It's spookier than I expected)"; first.Title != want {
		t.Errorf("Title = %q, atteso %q", first.Title, want)
	}
	if want := time.Date(2026, time.November, 4, 15, 52, 59, 0, time.UTC); !first.Published.Equal(want) {
		t.Errorf("Published = %v, atteso %v", first.Published, want)
	}
}

func TestParseFeedSkipsMalformedEntries(t *testing.T) {
	body := []byte(`<?xml version="1.0"?>
<feed xmlns:yt="http://www.youtube.com/xml/schemas/2015" xmlns="http://www.w3.org/2005/Atom">
 <title>Canale</title>
 <entry><yt:videoId>buono</yt:videoId><title>Uno</title><published>2026-01-02T03:04:05+00:00</published></entry>
 <entry><title>Senza id</title><published>2026-01-02T03:04:05+00:00</published></entry>
 <entry><yt:videoId>senzadata</yt:videoId><title>Data rotta</title><published>ieri</published></entry>
</feed>`)
	feed, err := ParseFeed(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(feed.Entries) != 1 || feed.Entries[0].VideoID != "buono" {
		t.Errorf("voci = %+v, attesa solo quella valida", feed.Entries)
	}
}

func TestParseFeedEmptyAndBroken(t *testing.T) {
	empty, err := ParseFeed([]byte(`<feed xmlns="http://www.w3.org/2005/Atom"><title>Vuoto</title></feed>`))
	if err != nil || len(empty.Entries) != 0 {
		t.Errorf("feed vuoto: voci = %d, err = %v", len(empty.Entries), err)
	}
	if _, err := ParseFeed([]byte(`<html>non è un feed`)); err == nil {
		t.Error("un documento rotto deve dare errore")
	}
}

func TestRSSFeedsFetch(t *testing.T) {
	body := fixture(t)
	var gotQuery, gotAgent string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery, gotAgent = r.URL.RawQuery, r.Header.Get("User-Agent")
		switch r.URL.Query().Get("playlist_id") {
		case "UULFesiste":
			w.Write(body)
		case "UULFrotto":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	feeds := NewRSSFeeds(srv.URL, 5*time.Second, "spestube/test")
	ctx := context.Background()

	feed, err := feeds.Playlist(ctx, "UULFesiste")
	if err != nil {
		t.Fatal(err)
	}
	if len(feed.Entries) != 3 {
		t.Errorf("voci = %d, attese 3", len(feed.Entries))
	}
	if gotQuery != "playlist_id=UULFesiste" || gotAgent != "spestube/test" {
		t.Errorf("richiesta: query %q, user agent %q", gotQuery, gotAgent)
	}

	if _, err := feeds.Playlist(ctx, "UULFmanca"); apperr.KindOf(err) != apperr.KindNotFound {
		t.Errorf("404: genere %v, atteso NotFound (%v)", apperr.KindOf(err), err)
	}
	if _, err := feeds.Playlist(ctx, "UULFrotto"); apperr.KindOf(err) != apperr.KindUnavailable {
		t.Errorf("500: genere %v, atteso Unavailable (%v)", apperr.KindOf(err), err)
	}
	if _, err := feeds.Channel(ctx, "UCqualcosa"); apperr.KindOf(err) != apperr.KindNotFound || gotQuery != "channel_id=UCqualcosa" {
		t.Errorf("feed del canale: query %q, err %v", gotQuery, err)
	}

	srv.Close()
	if _, err := feeds.Playlist(ctx, "UULFesiste"); apperr.KindOf(err) != apperr.KindUnavailable {
		t.Errorf("rete assente: genere %v, atteso Unavailable", apperr.KindOf(err))
	}
}
