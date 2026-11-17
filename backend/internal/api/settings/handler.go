// Package settings espone le impostazioni sotto /api/settings.
package settings

import (
	"net/http"

	"github.com/ste950/spestube/backend/internal/core/apperr"
	"github.com/ste950/spestube/backend/internal/core/clock"
	"github.com/ste950/spestube/backend/internal/core/httpx"
	"github.com/ste950/spestube/backend/internal/domain/settings"
)

// settingsBody è sia la risposta sia il corpo del PUT: le impostazioni si salvano
// tutte insieme, così il client non deve sapere quali campi sono cambiati.
type settingsBody struct {
	ReleaseTime        string   `json:"release_time"`
	ExcludeShorts      bool     `json:"exclude_shorts"`
	ExcludeLive        bool     `json:"exclude_live"`
	MinDurationMinutes int      `json:"min_duration_minutes"`
	ExcludedKeywords   []string `json:"excluded_keywords"`
}

func fromDomain(s settings.Settings) settingsBody {
	keywords := s.ExcludedKeywords
	if keywords == nil {
		keywords = []string{} // [] e non null
	}
	return settingsBody{
		ReleaseTime: s.ReleaseTime.String(), ExcludeShorts: s.ExcludeShorts, ExcludeLive: s.ExcludeLive,
		MinDurationMinutes: s.MinDurationMinutes, ExcludedKeywords: keywords,
	}
}

type Handler struct {
	svc *settings.Service
}

func NewHandler(svc *settings.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/settings", h.get)
	mux.HandleFunc("PUT /api/settings", h.put)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	s, err := h.svc.Get(r.Context())
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, fromDomain(s))
}

func (h *Handler) put(w http.ResponseWriter, r *http.Request) {
	var body settingsBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	releaseTime, err := clock.ParseTimeOfDay(body.ReleaseTime)
	if err != nil {
		httpx.WriteError(w, r, apperr.Invalid("settings_invalid", "orario di arrivo non valido (es. 07:00)"))
		return
	}
	saved, err := h.svc.Update(r.Context(), settings.Settings{
		ReleaseTime: releaseTime, ExcludeShorts: body.ExcludeShorts, ExcludeLive: body.ExcludeLive,
		MinDurationMinutes: body.MinDurationMinutes, ExcludedKeywords: body.ExcludedKeywords,
	})
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, fromDomain(saved))
}
