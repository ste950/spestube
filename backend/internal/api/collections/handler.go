// Package collections espone le raccolte sotto /api/collections.
package collections

import (
	"net/http"
	"strconv"

	"github.com/ste950/spestube/backend/internal/core/apperr"
	"github.com/ste950/spestube/backend/internal/core/httpx"
	"github.com/ste950/spestube/backend/internal/domain/library"
)

type Handler struct {
	svc *library.Collections
}

func NewHandler(svc *library.Collections) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/collections", h.list)
	mux.HandleFunc("POST /api/collections", h.create)
	// `order` è più specifico di `{id}`: il ServeMux sceglie questa rotta per quel percorso.
	mux.HandleFunc("PUT /api/collections/order", h.reorder)
	mux.HandleFunc("GET /api/collections/{id}", h.get)
	mux.HandleFunc("PUT /api/collections/{id}", h.update)
	mux.HandleFunc("DELETE /api/collections/{id}", h.delete)
	mux.HandleFunc("PUT /api/collections/{id}/videos/{videoId}", h.addVideo)
	mux.HandleFunc("DELETE /api/collections/{id}/videos/{videoId}", h.removeVideo)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	cols, err := h.svc.List(r.Context())
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	out := listResponse{Collections: make([]collectionResponse, 0, len(cols))}
	for _, c := range cols {
		out.Collections = append(out.Collections, fromDomain(c))
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req collectionRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	c, err := h.svc.Create(r.Context(), req.Name, req.Description)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, fromDomain(c))
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := collectionID(w, r)
	if !ok {
		return
	}
	c, videos, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	out := detailResponse{Collection: fromDomain(c), Videos: make([]videoResponse, 0, len(videos))}
	for _, v := range videos {
		out.Videos = append(out.Videos, videoFromDomain(v))
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := collectionID(w, r)
	if !ok {
		return
	}
	var req collectionRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	c, err := h.svc.Update(r.Context(), id, req.Name, req.Description)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, fromDomain(c))
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := collectionID(w, r)
	if !ok {
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) reorder(w http.ResponseWriter, r *http.Request) {
	var req orderRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	if err := h.svc.Reorder(r.Context(), req.IDs); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) addVideo(w http.ResponseWriter, r *http.Request) {
	id, ok := collectionID(w, r)
	if !ok {
		return
	}
	if err := h.svc.AddVideo(r.Context(), id, r.PathValue("videoId")); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) removeVideo(w http.ResponseWriter, r *http.Request) {
	id, ok := collectionID(w, r)
	if !ok {
		return
	}
	if err := h.svc.RemoveVideo(r.Context(), id, r.PathValue("videoId")); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func collectionID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httpx.WriteError(w, r, apperr.Invalid("invalid_id", "id della raccolta non valido"))
		return 0, false
	}
	return id, true
}
