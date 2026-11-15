# Handler HTTP

Una risorsa di `internal/api/`: handler sottili, DTO con i tag `json`, un `Register` che monta
le rotte. Vedi [API HTTP](../rules/http_api.rule.md).

I DTO:

```go
// backend/internal/api/items/dto.go

package items

import (
	"time"

	"example.com/backend/internal/domain/catalog"
)

type addRequest struct {
	URL string `json:"url"`
}

type itemResponse struct {
	ID          int64      `json:"id"`
	URL         string     `json:"url"`
	Title       string     `json:"title"`
	CreatedAt   time.Time  `json:"created_at"`
	RefreshedAt *time.Time `json:"refreshed_at"` // null finché la sorgente non ha risposto
}

// listResponse avvolge la lista in un oggetto: un campo in più domani (un cursore)
// non rompe i client di oggi.
type listResponse struct {
	Items []itemResponse `json:"items"`
}

func fromDomain(it catalog.Item) itemResponse {
	out := itemResponse{ID: it.ID, URL: it.URL, Title: it.Title, CreatedAt: it.CreatedAt.UTC()}
	if !it.RefreshedAt.IsZero() {
		t := it.RefreshedAt.UTC()
		out.RefreshedAt = &t
	}
	return out
}
```

Gli handler:

```go
// backend/internal/api/items/handler.go

// Package items espone il soggetto catalog sotto /api/items.
package items

import (
	"net/http"
	"strconv"

	"example.com/backend/internal/core/apperr"
	"example.com/backend/internal/core/httpx"
	"example.com/backend/internal/domain/catalog"
)

type Handler struct {
	svc *catalog.Service
}

func NewHandler(svc *catalog.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/items", h.list)
	mux.HandleFunc("POST /api/items", h.add)
	mux.HandleFunc("GET /api/items/{id}", h.get)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.List(r.Context())
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	out := listResponse{Items: make([]itemResponse, 0, len(items))} // [] e non null
	for _, it := range items {
		out.Items = append(out.Items, fromDomain(it))
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) add(w http.ResponseWriter, r *http.Request) {
	var req addRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	it, err := h.svc.Add(r.Context(), req.URL)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, fromDomain(it))
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httpx.WriteError(w, r, apperr.Invalid("invalid_id", "id non valido"))
		return
	}
	it, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, fromDomain(it))
}
```

Il montaggio delle risorse e i middleware:

```go
// backend/internal/api/routes.go

// Package api monta le risorse sul router e applica i middleware.
package api

import (
	"log/slog"
	"net/http"
	"time"

	"example.com/backend/internal/api/items"
	"example.com/backend/internal/domain/catalog"
)

// Services è ciò di cui le risorse hanno bisogno: lo riempie il composition root.
type Services struct {
	Catalog *catalog.Service
	Version string
}

func Routes(s Services) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(s.Version))
	})
	items.NewHandler(s.Catalog).Register(mux)
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
```
