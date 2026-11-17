// Package channels espone i canali seguiti sotto /api/channels.
package channels

import (
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
	mux.HandleFunc("GET /api/channels", h.list)
	mux.HandleFunc("POST /api/channels", h.add)
	mux.HandleFunc("DELETE /api/channels/{id}", h.remove)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	channels, err := h.svc.ListChannels(r.Context())
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	out := listResponse{Channels: make([]channelResponse, 0, len(channels))}
	for _, ch := range channels {
		out.Channels = append(out.Channels, fromDomain(ch))
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) add(w http.ResponseWriter, r *http.Request) {
	var req addRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	ch, err := h.svc.AddChannel(r.Context(), req.Channel)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, fromDomain(ch))
}

func (h *Handler) remove(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.RemoveChannel(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
