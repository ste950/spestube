// Package meta dice al client con quale installazione sta parlando: la versione e ciò
// che la configurazione rende disponibile.
package meta

import (
	"net/http"

	"github.com/ste950/spestube/backend/internal/core/httpx"
)

// Info sono i fatti dell'installazione che cambiano ciò che il client mostra.
type Info struct {
	Version string
	// Dev: gli endpoint di sviluppo sono attivi.
	Dev bool
	// HandleLookup: c'è la chiave API, quindi un canale si può aggiungere dal suo @handle.
	HandleLookup bool
}

type infoResponse struct {
	Version      string `json:"version"`
	Dev          bool   `json:"dev"`
	HandleLookup bool   `json:"handle_lookup"`
}

type Handler struct {
	info Info
}

func NewHandler(info Info) *Handler {
	return &Handler{info: info}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/meta", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, infoResponse{
			Version: h.info.Version, Dev: h.info.Dev, HandleLookup: h.info.HandleLookup,
		})
	})
}
