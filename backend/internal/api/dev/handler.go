// Package dev espone le scorciatoie di sviluppo: forzare un download e un rilascio senza
// aspettare l'ora. Si monta solo se la configurazione lo chiede, perché un rilascio a
// comando è esattamente il feed senza fine che il prodotto esiste per togliere.
package dev

import (
	"net/http"

	"github.com/ste950/spestube/backend/internal/core/apperr"
	"github.com/ste950/spestube/backend/internal/core/httpx"
	"github.com/ste950/spestube/backend/internal/domain/library"
)

type fetchResponse struct {
	// Errors elenca i canali il cui feed non ha risposto. Gli altri sono stati letti.
	Errors []string `json:"errors"`
}

type releaseResponse struct {
	Released  int `json:"released"`
	Discarded int `json:"discarded"`
}

type Handler struct {
	svc *library.Service
}

func NewHandler(svc *library.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/dev/fetch", h.fetch)
	mux.HandleFunc("POST /api/dev/release", h.release)
}

func (h *Handler) fetch(w http.ResponseWriter, r *http.Request) {
	out := fetchResponse{Errors: []string{}}
	if err := h.svc.FetchAll(r.Context()); err != nil {
		// Un canale irraggiungibile non fa fallire il giro: è un esito da riportare.
		kind := apperr.KindOf(err)
		if kind != apperr.KindUnavailable && kind != apperr.KindNotFound {
			httpx.WriteError(w, r, err)
			return
		}
		out.Errors = append(out.Errors, err.Error())
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) release(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.ReleaseNow(r.Context())
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, releaseResponse{Released: res.Released, Discarded: res.Discarded})
}
