package api

import (
	"context"
	"encoding/json"
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
	srv := httptest.NewServer(Routes(Services{
		Library:  lib,
		Settings: settings.NewService(settingstest.NewRepository(), settings.Defaults(clock.TimeOfDay{Hour: 7}), clockFn),
		Meta:     meta.Info{Version: "test", Dev: dev},
		Now:      clockFn,
		Ping:     func(context.Context) error { return nil },
		Static: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte("frontend"))
		}),
	}))
	t.Cleanup(srv.Close)
	return srv, videos
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
	if status, body := call(t, srv, "POST", "/api/videos/uscito/queue", `{"reason":"  "}`); status != http.StatusBadRequest || errorCode(body) != "queue_reason_required" {
		t.Errorf("coda senza motivo: %d %v", status, body)
	}
	status, body := call(t, srv, "POST", "/api/videos/uscito/queue", `{"reason":"per l'orto"}`)
	if status != http.StatusOK || body["status"] != "queued" || body["queue_days_left"] != float64(14) || body["queue_expires_at"] != "2026-11-29T09:00:00Z" {
		t.Errorf("in coda: %d %v", status, body)
	}
	status, body = call(t, srv, "GET", "/api/queue", "")
	if list, _ := body["videos"].([]any); status != http.StatusOK || len(list) != 1 {
		t.Errorf("coda: %d %v", status, body)
	}
	if status, body := call(t, srv, "POST", "/api/videos/uscito/watched", ""); status != http.StatusOK || body["status"] != "watched" {
		t.Errorf("visto: %d %v", status, body)
	}
	if status, body := call(t, srv, "POST", "/api/videos/uscito/discard", ""); status != http.StatusConflict || errorCode(body) != "video_not_discardable" {
		t.Errorf("scarto di un video visto: %d %v", status, body)
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
