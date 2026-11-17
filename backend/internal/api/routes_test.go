package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ste950/spestube/backend/internal/api/meta"
	"github.com/ste950/spestube/backend/internal/core/clock"
	"github.com/ste950/spestube/backend/internal/domain/library"
	"github.com/ste950/spestube/backend/internal/domain/library/librarytest"
	"github.com/ste950/spestube/backend/internal/domain/settings"
	"github.com/ste950/spestube/backend/internal/domain/settings/settingstest"
)

const chanA = "UCaaaaaaaaaaaaaaaaaaaaaa"

var now = time.Date(2026, time.November, 15, 9, 0, 0, 0, time.UTC)

// newServer monta le rotte vere su servizi veri, con i fake al posto di database e YouTube.
func newServer(t *testing.T, dev bool) (*httptest.Server, *librarytest.Videos) {
	srv, videos, _ := newServerWithCollections(t, dev)
	return srv, videos
}

func newServerWithCollections(t *testing.T, dev bool) (*httptest.Server, *librarytest.Videos, *librarytest.Collections) {
	t.Helper()
	videos := librarytest.NewVideos()
	feeds := &librarytest.Feeds{
		LongVideos: map[string]library.Feed{chanA: {
			ChannelTitle: "Officina Radio",
			Entries: []library.FeedEntry{
				{VideoID: "v1", Title: "Antenne LoRa", PublishedAt: now.Add(-time.Hour)},
				{VideoID: "v2", Title: "Saldare bene", PublishedAt: now.Add(-48 * time.Hour)},
			},
		}},
		Channel: map[string]library.Feed{},
	}
	clockFn := func() time.Time { return now }
	lib := library.NewService(library.Deps{
		Channels: librarytest.NewChannels(), Videos: videos, Feeds: feeds,
		Handles: &librarytest.Handles{NoKey: true},
		Policy:  librarytest.Policy{ReleaseAt: clock.TimeOfDay{Hour: 7}},
		Now:     clockFn, Loc: time.UTC,
	})
	collections := librarytest.NewCollections(videos)
	srv := httptest.NewServer(Routes(Services{
		Library:     lib,
		Collections: library.NewCollections(collections, videos, clockFn),
		Settings:    settings.NewService(settingstest.NewRepository(), settings.Defaults(clock.TimeOfDay{Hour: 7}), clockFn),
		Meta:        meta.Info{Version: "test", Dev: dev},
		Ping:        func(context.Context) error { return nil },
		Static: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte("frontend"))
		}),
	}))
	t.Cleanup(srv.Close)
	return srv, videos, collections
}

func call(t *testing.T, srv *httptest.Server, method, path, body string) (int, map[string]any) {
	t.Helper()
	req, err := http.NewRequest(method, srv.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	if resp.StatusCode != http.StatusNoContent && strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			t.Fatalf("%s %s: corpo non JSON: %v", method, path, err)
		}
	}
	return resp.StatusCode, out
}

func errorCode(body map[string]any) string {
	e, _ := body["error"].(map[string]any)
	code, _ := e["code"].(string)
	return code
}

func TestAddChannelThenSeeItsVideosAfterRelease(t *testing.T) {
	srv, _ := newServer(t, true)

	status, body := call(t, srv, "POST", "/api/channels", `{"channel":"https://www.youtube.com/channel/`+chanA+`"}`)
	if status != http.StatusCreated || body["title"] != "Officina Radio" || body["feed_mode"] != "uulf" {
		t.Fatalf("aggiunta: %d %v", status, body)
	}

	// Scaricati ma non ancora rilasciati: Oggi è vuoto, ed è una lista, non null.
	status, body = call(t, srv, "GET", "/api/feed/today", "")
	if list, ok := body["videos"].([]any); status != http.StatusOK || !ok || len(list) != 0 {
		t.Fatalf("Oggi prima del rilascio: %d %v", status, body)
	}
	if body["next_release_at"] != "2026-11-16T07:00:00Z" {
		t.Errorf("prossimo rilascio = %v", body["next_release_at"])
	}

	status, body = call(t, srv, "POST", "/api/dev/release", "")
	if status != http.StatusOK || body["released"] != float64(2) {
		t.Fatalf("rilascio forzato: %d %v", status, body)
	}

	status, body = call(t, srv, "GET", "/api/feed/today", "")
	list, _ := body["videos"].([]any)
	if status != http.StatusOK || len(list) != 2 {
		t.Fatalf("Oggi dopo il rilascio: %d %v", status, body)
	}
	first := list[0].(map[string]any)
	if first["channel_id"] != chanA || first["status"] != "released" || first["duration_seconds"] != nil {
		t.Errorf("video = %v", first)
	}
	// Il contratto non espone contatori: non devono comparire nemmeno per sbaglio.
	for _, forbidden := range []string{"views", "view_count", "likes", "like_count"} {
		if _, ok := first[forbidden]; ok {
			t.Errorf("il video espone %q", forbidden)
		}
	}

	status, body = call(t, srv, "GET", "/api/channels", "")
	if list, _ := body["channels"].([]any); status != http.StatusOK || len(list) != 1 {
		t.Errorf("canali: %d %v", status, body)
	}
	if status, _ := call(t, srv, "DELETE", "/api/channels/"+chanA, ""); status != http.StatusNoContent {
		t.Errorf("rimozione: %d", status)
	}
}

func TestChannelErrors(t *testing.T) {
	srv, _ := newServer(t, false)
	cases := []struct {
		name       string
		body       string
		wantStatus int
		wantCode   string
	}{
		{"handle senza chiave API", `{"channel":"@officina"}`, http.StatusBadRequest, "handle_needs_api_key"},
		{"riferimento non riconosciuto", `{"channel":"officina"}`, http.StatusBadRequest, "channel_ref_unsupported"},
		{"canale inesistente", `{"channel":"UCzzzzzzzzzzzzzzzzzzzzzz"}`, http.StatusNotFound, "channel_not_found"},
		{"campo sconosciuto", `{"url":"x"}`, http.StatusBadRequest, "invalid_body"},
		{"corpo non JSON", `canale`, http.StatusBadRequest, "invalid_body"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, body := call(t, srv, "POST", "/api/channels", tc.body)
			if status != tc.wantStatus || errorCode(body) != tc.wantCode {
				t.Errorf("risposta = %d %q, attesa %d %q", status, errorCode(body), tc.wantStatus, tc.wantCode)
			}
		})
	}
	if status, body := call(t, srv, "DELETE", "/api/channels/UCmanca", ""); status != http.StatusNotFound || errorCode(body) != "channel_not_found" {
		t.Errorf("rimozione di un canale assente: %d %v", status, body)
	}
}

func TestVideoActions(t *testing.T) {
	srv, videos := newServer(t, false)
	videos.ByID["uscito"] = library.Video{ID: "uscito", Title: "Uscito", Status: library.StatusReleased, ReleasedAt: now}
	videos.ByID["nuovo"] = library.Video{ID: "nuovo", Title: "Nuovo", Status: library.StatusNew}

	if status, body := call(t, srv, "GET", "/api/videos/nuovo", ""); status != http.StatusNotFound || errorCode(body) != "video_not_found" {
		t.Errorf("un video non rilasciato si può aprire: %d %v", status, body)
	}
	status, body := call(t, srv, "GET", "/api/videos/uscito", "")
	if ids, ok := body["collection_ids"].([]any); status != http.StatusOK || !ok || len(ids) != 0 {
		t.Errorf("video: %d %v", status, body)
	}
	// La coda non esiste più: le sue rotte non rispondono.
	if status, _ := call(t, srv, "GET", "/api/queue", ""); status != http.StatusNotFound {
		t.Errorf("GET /api/queue = %d", status)
	}
	if status, body := call(t, srv, "POST", "/api/videos/uscito/watched", ""); status != http.StatusOK || body["status"] != "watched" {
		t.Errorf("visto: %d %v", status, body)
	}
	if status, body := call(t, srv, "POST", "/api/videos/uscito/discard", ""); status != http.StatusConflict || errorCode(body) != "video_not_discardable" {
		t.Errorf("scarto di un video visto: %d %v", status, body)
	}
}

func TestCollections(t *testing.T) {
	srv, videos, _ := newServerWithCollections(t, false)
	videos.ByID["v1"] = library.Video{ID: "v1", Title: "Antenne LoRa", ChannelTitle: "Officina Radio", Status: library.StatusReleased}
	videos.ByID["nuovo"] = library.Video{ID: "nuovo", Title: "Nuovo", Status: library.StatusNew}

	status, body := call(t, srv, "POST", "/api/collections", `{"name":"  Orto  ","description":"idee"}`)
	if status != http.StatusCreated || body["name"] != "Orto" || body["video_count"] != float64(0) || body["cover_video_id"] != nil {
		t.Fatalf("creazione: %d %v", status, body)
	}
	orto := int(body["id"].(float64))
	_, body = call(t, srv, "POST", "/api/collections", `{"name":"Cucina","description":""}`)
	cucina := int(body["id"].(float64))
	path := func(format string, args ...any) string { return fmt.Sprintf(format, args...) }

	for _, tc := range []struct {
		name, method, path, body string
		status                   int
		code                     string
	}{
		{"nome vuoto", "POST", "/api/collections", `{"name":" ","description":""}`, 400, "collection_name_required"},
		{"nome già usato", "POST", "/api/collections", `{"name":"orto","description":""}`, 409, "collection_exists"},
		{"rinomina su un nome usato", "PUT", path("/api/collections/%d", cucina), `{"name":"ORTO","description":""}`, 409, "collection_exists"},
		{"raccolta inesistente", "GET", "/api/collections/999", "", 404, "collection_not_found"},
		{"id non valido", "GET", "/api/collections/abc", "", 400, "invalid_id"},
		{"video non rilasciato", "PUT", path("/api/collections/%d/videos/nuovo", orto), "", 404, "video_not_found"},
		{"ordine incompleto", "PUT", "/api/collections/order", path(`{"ids":[%d]}`, orto), 409, "collection_order_invalid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status, body := call(t, srv, tc.method, tc.path, tc.body)
			if status != tc.status || errorCode(body) != tc.code {
				t.Errorf("risposta = %d %q, attesa %d %q", status, errorCode(body), tc.status, tc.code)
			}
		})
	}

	// Salvare due volte lo stesso video è lo stesso che salvarlo una volta.
	for range 2 {
		if status, _ := call(t, srv, "PUT", path("/api/collections/%d/videos/v1", orto), ""); status != http.StatusNoContent {
			t.Fatalf("aggiunta del video: %d", status)
		}
	}
	status, body = call(t, srv, "GET", path("/api/collections/%d", orto), "")
	col, _ := body["collection"].(map[string]any)
	list, _ := body["videos"].([]any)
	if status != http.StatusOK || col["video_count"] != float64(1) || col["cover_video_id"] != "v1" || len(list) != 1 {
		t.Fatalf("raccolta: %d %v", status, body)
	}
	if first := list[0].(map[string]any); first["channel_title"] != "Officina Radio" || first["added_at"] == nil {
		t.Errorf("video della raccolta = %v", first)
	}
	_, video := call(t, srv, "GET", "/api/videos/v1", "")
	if ids, _ := video["collection_ids"].([]any); len(ids) != 1 || ids[0] != float64(orto) {
		t.Errorf("collection_ids = %v", video["collection_ids"])
	}

	status, body = call(t, srv, "PUT", path("/api/collections/%d", orto), `{"name":"Orto e giardino","description":""}`)
	if status != http.StatusOK || body["name"] != "Orto e giardino" || body["description"] != "" {
		t.Errorf("modifica: %d %v", status, body)
	}

	if status, _ := call(t, srv, "PUT", "/api/collections/order", path(`{"ids":[%d,%d]}`, cucina, orto)); status != http.StatusNoContent {
		t.Errorf("riordino: %d", status)
	}
	_, body = call(t, srv, "GET", "/api/collections", "")
	cols, _ := body["collections"].([]any)
	if len(cols) != 2 || cols[0].(map[string]any)["name"] != "Cucina" {
		t.Errorf("ordine dopo il riordino = %v", cols)
	}

	if status, _ := call(t, srv, "DELETE", path("/api/collections/%d/videos/v1", orto), ""); status != http.StatusNoContent {
		t.Errorf("rimozione del video: %d", status)
	}
	if status, _ := call(t, srv, "DELETE", path("/api/collections/%d", orto), ""); status != http.StatusNoContent {
		t.Errorf("eliminazione: %d", status)
	}
	if status, body := call(t, srv, "DELETE", path("/api/collections/%d", orto), ""); status != http.StatusNotFound || errorCode(body) != "collection_not_found" {
		t.Errorf("seconda eliminazione: %d %v", status, body)
	}
	// Il video non sparisce con la raccolta.
	if status, _ := call(t, srv, "GET", "/api/videos/v1", ""); status != http.StatusOK {
		t.Errorf("il video è sparito con la raccolta: %d", status)
	}
}

func TestSettings(t *testing.T) {
	srv, _ := newServer(t, false)

	status, body := call(t, srv, "GET", "/api/settings", "")
	if status != http.StatusOK || body["release_time"] != "07:00" || body["exclude_shorts"] != true {
		t.Fatalf("default: %d %v", status, body)
	}
	if kw, ok := body["excluded_keywords"].([]any); !ok || len(kw) != 0 {
		t.Errorf("excluded_keywords = %v, attesa una lista vuota", body["excluded_keywords"])
	}

	put := `{"release_time":"19:30","exclude_shorts":true,"exclude_live":false,"min_duration_minutes":5,"excluded_keywords":["reaction"]}`
	status, body = call(t, srv, "PUT", "/api/settings", put)
	if status != http.StatusOK || body["release_time"] != "19:30" || body["exclude_live"] != false {
		t.Fatalf("salvataggio: %d %v", status, body)
	}
	if _, again := call(t, srv, "GET", "/api/settings", ""); again["release_time"] != "19:30" || again["min_duration_minutes"] != float64(5) {
		t.Errorf("rilette = %v", again)
	}

	bad := strings.Replace(put, `"19:30"`, `"25:00"`, 1)
	if status, body := call(t, srv, "PUT", "/api/settings", bad); status != http.StatusBadRequest || errorCode(body) != "settings_invalid" {
		t.Errorf("orario non valido: %d %v", status, body)
	}
	bad = strings.Replace(put, `"min_duration_minutes":5`, `"min_duration_minutes":-1`, 1)
	if status, body := call(t, srv, "PUT", "/api/settings", bad); status != http.StatusBadRequest || errorCode(body) != "settings_invalid" {
		t.Errorf("durata minima non valida: %d %v", status, body)
	}
	// Le miniature non sono un'impostazione: un client che prova a mandarne una viene
	// fermato, invece di credere di averla salvata.
	bad = strings.Replace(put, `}`, `,"thumbnail_mode":"bw"}`, 1)
	if status, body := call(t, srv, "PUT", "/api/settings", bad); status != http.StatusBadRequest || errorCode(body) != "invalid_body" {
		t.Errorf("campo delle miniature: %d %v", status, body)
	}
	if _, body := call(t, srv, "GET", "/api/settings", ""); body["thumbnail_mode"] != nil {
		t.Errorf("le impostazioni espongono ancora thumbnail_mode: %v", body)
	}
}

func TestDevEndpointsExistOnlyInDevMode(t *testing.T) {
	prod, _ := newServer(t, false)
	for _, path := range []string{"/api/dev/release", "/api/dev/fetch"} {
		if status, body := call(t, prod, "POST", path, ""); status != http.StatusNotFound || errorCode(body) != "not_found" {
			t.Errorf("%s fuori dallo sviluppo: %d %v", path, status, body)
		}
	}
	_, info := call(t, prod, "GET", "/api/meta", "")
	if info["dev"] != false || info["version"] != "test" {
		t.Errorf("meta = %v", info)
	}

	dev, _ := newServer(t, true)
	if status, body := call(t, dev, "POST", "/api/dev/fetch", ""); status != http.StatusOK || body["errors"] == nil {
		t.Errorf("fetch in sviluppo: %d %v", status, body)
	}
}

func TestUnknownPaths(t *testing.T) {
	srv, _ := newServer(t, false)

	if status, body := call(t, srv, "GET", "/api/nope", ""); status != http.StatusNotFound || errorCode(body) != "not_found" {
		t.Errorf("percorso /api/ sconosciuto: %d %v", status, body)
	}
	// Tutto il resto è del frontend.
	resp, err := srv.Client().Get(srv.URL + "/da-vedere")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("percorso del frontend: %d", resp.StatusCode)
	}
	resp, err = srv.Client().Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("healthz: %d", resp.StatusCode)
	}
}
