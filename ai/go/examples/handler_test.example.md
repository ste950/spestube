# Test di un Handler

Il contratto HTTP si prova con `httptest`, montando la risorsa vera sopra il servizio vero, con
i soli repository finti. Si verificano status, corpo e **codice di errore**: è ciò che il
client legge. Vedi [Test](../rules/testing.rule.md) e [API HTTP](../rules/http_api.rule.md).

```go
// backend/internal/api/items/handler_test.go

package items_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"example.com/backend/internal/api/items"
	"example.com/backend/internal/domain/catalog"
	"example.com/backend/internal/domain/catalog/catalogtest"
)

// newServer monta la risorsa vera su un servizio vero, con i soli repository finti:
// si prova il contratto HTTP, non un handler isolato dal resto.
func newServer(t *testing.T, seed ...catalog.Item) *httptest.Server {
	t.Helper()
	svc := catalog.NewService(catalogtest.NewItems(seed...), catalogtest.Titles{},
		func() time.Time { return time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC) })
	mux := http.NewServeMux()
	items.NewHandler(svc).Register(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestAddReturnsCreated(t *testing.T) {
	srv := newServer(t)

	resp, err := http.Post(srv.URL+"/api/items", "application/json",
		strings.NewReader(`{"url":"https://example.org/x"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, atteso 201", resp.StatusCode)
	}
	var body struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.ID == 0 || body.Title != "example.org" {
		t.Errorf("corpo = %+v", body)
	}
}

func TestErrorsCarryStableCodes(t *testing.T) {
	srv := newServer(t)

	tests := []struct {
		name, method, path, body string
		wantStatus               int
		wantCode                 string
	}{
		{"item assente", "GET", "/api/items/42", "", 404, "item_not_found"},
		{"id non numerico", "GET", "/api/items/abc", "", 400, "invalid_id"},
		{"campo sconosciuto", "POST", "/api/items", `{"link":"https://x.org"}`, 400, "invalid_body"},
		{"url non valido", "POST", "/api/items", `{"url":"nope"}`, 400, "invalid_url"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(tc.method, srv.URL+tc.path, strings.NewReader(tc.body))
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			var body struct {
				Error struct{ Code string } `json:"error"`
			}
			json.NewDecoder(resp.Body).Decode(&body)
			if resp.StatusCode != tc.wantStatus || body.Error.Code != tc.wantCode {
				t.Errorf("= %d %q, atteso %d %q", resp.StatusCode, body.Error.Code, tc.wantStatus, tc.wantCode)
			}
		})
	}
}
```
