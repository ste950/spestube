// Package videos espone i video: il feed di oggi e le azioni su un video.
package videos

import (
	"context"
	"net/http"

	"github.com/ste950/spestube/backend/internal/core/httpx"
	"github.com/ste950/spestube/backend/internal/domain/library"
)

type Handler struct {
	svc *library.Service
}

func NewHandler(svc *library.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/feed/today", h.today)
	mux.HandleFunc("GET /api/videos/{id}", h.get)
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
		Videos:        listFromDomain(today.Videos),
		NextReleaseAt: today.NextReleaseAt.UTC(),
	})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.Video(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, fromDomain(v))
}

// action serve le azioni che rileggono il video da capo: le raccolte che lo contengono
// arrivano dalla lettura, non dal valore che l'azione ha in mano.
func (h *Handler) action(do func(context.Context, string) (library.Video, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := do(r.Context(), r.PathValue("id")); err != nil {
			httpx.WriteError(w, r, err)
			return
		}
		v, err := h.svc.Video(r.Context(), r.PathValue("id"))
		if err != nil {
			httpx.WriteError(w, r, err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, fromDomain(v))
	}
}
