// Package videos espone i video: il feed di oggi, la coda e le azioni su un video.
// Le tre rotte stanno in una risorsa sola perché condividono la forma del video nel
// contratto, e una risorsa non può importarne un'altra.
package videos

import (
	"context"
	"net/http"
	"time"

	"github.com/ste950/spestube/backend/internal/core/httpx"
	"github.com/ste950/spestube/backend/internal/domain/library"
)

type Handler struct {
	svc *library.Service
	now func() time.Time
}

func NewHandler(svc *library.Service, now func() time.Time) *Handler {
	return &Handler{svc: svc, now: now}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/feed/today", h.today)
	mux.HandleFunc("GET /api/queue", h.queue)
	mux.HandleFunc("GET /api/videos/{id}", h.get)
	mux.HandleFunc("POST /api/videos/{id}/queue", h.enqueue)
	mux.HandleFunc("POST /api/videos/{id}/watched", h.action(func(ctx context.Context, id string) (library.Video, error) {
		return h.svc.MarkWatched(ctx, id)
	}))
	mux.HandleFunc("POST /api/videos/{id}/discard", h.action(func(ctx context.Context, id string) (library.Video, error) {
		return h.svc.Discard(ctx, id)
	}))
}

func (h *Handler) today(w http.ResponseWriter, r *http.Request) {
	today, err := h.svc.Today(r.Context())
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, todayResponse{
		Videos:        listFromDomain(today.Videos, h.now()),
		NextReleaseAt: today.NextReleaseAt.UTC(),
	})
}

func (h *Handler) queue(w http.ResponseWriter, r *http.Request) {
	videos, err := h.svc.Queue(r.Context())
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, queueResponse{Videos: listFromDomain(videos, h.now())})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.Video(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, fromDomain(v, h.now()))
}

func (h *Handler) enqueue(w http.ResponseWriter, r *http.Request) {
	var req queueRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	v, err := h.svc.Enqueue(r.Context(), r.PathValue("id"), req.Reason)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, fromDomain(v, h.now()))
}

func (h *Handler) action(do func(context.Context, string) (library.Video, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v, err := do(r.Context(), r.PathValue("id"))
		if err != nil {
			httpx.WriteError(w, r, err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, fromDomain(v, h.now()))
	}
}
