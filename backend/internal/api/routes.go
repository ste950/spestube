// Package api monta le risorse sul router e applica i middleware.
package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/ste950/spestube/backend/internal/api/channels"
	"github.com/ste950/spestube/backend/internal/api/collections"
	"github.com/ste950/spestube/backend/internal/api/dev"
	"github.com/ste950/spestube/backend/internal/api/meta"
	apisettings "github.com/ste950/spestube/backend/internal/api/settings"
	"github.com/ste950/spestube/backend/internal/api/videos"
	"github.com/ste950/spestube/backend/internal/core/apperr"
	"github.com/ste950/spestube/backend/internal/core/httpx"
	"github.com/ste950/spestube/backend/internal/domain/library"
	"github.com/ste950/spestube/backend/internal/domain/settings"
)

// Services è ciò di cui le risorse hanno bisogno: lo riempie il composition root.
type Services struct {
	Library     *library.Service
	Collections *library.Collections
	Settings    *settings.Service
	Meta        meta.Info
	// Ping risponde nil se il database è raggiungibile: è ciò che /healthz verifica.
	Ping func(context.Context) error
	// Static serve il frontend compilato a tutto ciò che non è /api/ né /healthz.
	Static http.Handler
}

func Routes(s Services) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := s.Ping(r.Context()); err != nil {
			http.Error(w, "database non raggiungibile", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte(s.Meta.Version))
	})

	channels.NewHandler(s.Library).Register(mux)
	videos.NewHandler(s.Library).Register(mux)
	collections.NewHandler(s.Collections).Register(mux)
	apisettings.NewHandler(s.Settings).Register(mux)
	meta.NewHandler(s.Meta).Register(mux)
	if s.Meta.Dev {
		dev.NewHandler(s.Library).Register(mux)
	}

	// Un percorso /api/ che nessuna risorsa dichiara deve rispondere con l'errore JSON
	// del contratto, non con la pagina del frontend.
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteError(w, r, apperr.NotFound("not_found", "risorsa inesistente"))
	})
	mux.Handle("/", s.Static)

	return recoverer(logRequests(mux))
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Debug("richiesta", "method", r.Method, "path", r.URL.Path, "dur", time.Since(start))
	})
}

// recoverer trasforma un panic — un bug, non un errore previsto — in un 500 e una
// riga di log, invece di far cadere la connessione senza risposta.
func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				slog.Error("panic", "path", r.URL.Path, "value", v)
				http.Error(w, `{"error":{"code":"internal_error","message":"errore interno"}}`,
					http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
