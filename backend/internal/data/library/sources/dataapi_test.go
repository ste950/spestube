package sources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ste950/spestube/backend/internal/core/apperr"
)

func TestDataAPIByHandle(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.URL.Path != "/channels" || q.Get("key") != "chiave-finta" || q.Get("part") != "snippet" {
			t.Errorf("richiesta inattesa: %s", r.URL)
		}
		switch q.Get("forHandle") {
		case "@esiste":
			w.Write([]byte(`{"items":[{"id":"UCaaaaaaaaaaaaaaaaaaaaaa","snippet":{"title":"Canale Vero"}}]}`))
		case "@manca":
			w.Write([]byte(`{"kind":"youtube#channelListResponse","pageInfo":{"totalResults":0}}`))
		case "@quota":
			w.WriteHeader(http.StatusForbidden)
		}
	}))
	defer srv.Close()
	api := NewDataAPI(srv.URL, "chiave-finta", 5*time.Second)
	ctx := context.Background()

	ch, err := api.ByHandle(ctx, "@esiste")
	if err != nil || ch.ID != "UCaaaaaaaaaaaaaaaaaaaaaa" || ch.Title != "Canale Vero" {
		t.Errorf("ByHandle = %+v, %v", ch, err)
	}
	if _, err := api.ByHandle(ctx, "@manca"); apperr.KindOf(err) != apperr.KindNotFound {
		t.Errorf("handle inesistente: genere %v, atteso NotFound", apperr.KindOf(err))
	}
	if _, err := api.ByHandle(ctx, "@quota"); apperr.KindOf(err) != apperr.KindUnavailable {
		t.Errorf("403: genere %v, atteso Unavailable", apperr.KindOf(err))
	}

	srv.Close()
	_, err = api.ByHandle(ctx, "@esiste")
	if apperr.KindOf(err) != apperr.KindUnavailable {
		t.Errorf("rete assente: genere %v, atteso Unavailable", apperr.KindOf(err))
	}
	if strings.Contains(err.Error(), "chiave-finta") {
		t.Errorf("l'errore contiene la chiave API: %v", err)
	}
}
