package webui

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func get(t *testing.T, h http.Handler, method, path string) (int, string, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	body, _ := io.ReadAll(rec.Result().Body)
	return rec.Code, string(body), rec.Header().Get("Cache-Control")
}

func TestHandler(t *testing.T) {
	h := NewHandler(fstest.MapFS{
		"index.html":           {Data: []byte("<html>app</html>")},
		"sw.js":                {Data: []byte("// service worker")},
		"assets/index-abc1.js": {Data: []byte("console.log(1)")},
	})
	cases := []struct {
		name      string
		path      string
		wantCode  int
		wantBody  string
		wantCache string
	}{
		{"radice", "/", http.StatusOK, "<html>app</html>", "no-cache"},
		{"percorso del router", "/guarda/abc", http.StatusOK, "<html>app</html>", "no-cache"},
		{"file con hash", "/assets/index-abc1.js", http.StatusOK, "console.log(1)", "public, max-age=31536000, immutable"},
		{"service worker sempre ricontrollato", "/sw.js", http.StatusOK, "// service worker", "no-cache"},
		{"file che manca resta un 404", "/assets/vecchio-0000.js", http.StatusNotFound, "", ""},
		{"percorso che esce dalla cartella", "/../../etc/passwd", http.StatusOK, "<html>app</html>", "no-cache"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, body, cache := get(t, h, http.MethodGet, tc.path)
			if code != tc.wantCode || cache != tc.wantCache || (tc.wantBody != "" && body != tc.wantBody) {
				t.Errorf("GET %s = %d %q cache %q", tc.path, code, body, cache)
			}
		})
	}
	if code, _, _ := get(t, h, http.MethodPost, "/"); code != http.StatusMethodNotAllowed {
		t.Errorf("POST / = %d", code)
	}
}

func TestHandlerWithoutABuiltFrontend(t *testing.T) {
	h := NewHandler(fstest.MapFS{".gitkeep": {}})
	for _, path := range []string{"/", "/index.html", "/canali"} {
		if code, _, _ := get(t, h, http.MethodGet, path); code != http.StatusServiceUnavailable {
			t.Errorf("GET %s senza build = %d, atteso 503", path, code)
		}
	}
}

func TestManifestIsServedAsAManifest(t *testing.T) {
	h := NewHandler(fstest.MapFS{
		"index.html":           {Data: []byte("<html>app</html>")},
		"manifest.webmanifest": {Data: []byte(`{"name":"SpesTube"}`)},
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/manifest.webmanifest", nil))
	if got := rec.Header().Get("Content-Type"); got != "application/manifest+json" {
		t.Errorf("Content-Type = %q", got)
	}
}
