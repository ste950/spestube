package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ste950/spestube/backend/internal/core/apperr"
)

// YouTubeDataAPIURL è la radice della YouTube Data API v3. Si usa solo se c'è una chiave.
const YouTubeDataAPIURL = "https://www.googleapis.com/youtube/v3"

// ChannelLookup trova un canale a partire dal suo handle.
type ChannelLookup interface {
	ByHandle(ctx context.Context, handle string) (ChannelPayload, error)
}

type ChannelPayload struct {
	ID    string
	Title string
}

// DataAPI parla con la YouTube Data API.
type DataAPI struct {
	client  *http.Client
	baseURL string
	key     string
}

func NewDataAPI(baseURL, key string, timeout time.Duration) *DataAPI {
	return &DataAPI{client: &http.Client{Timeout: timeout}, baseURL: baseURL, key: key}
}

type channelsListResponse struct {
	Items []struct {
		ID      string `json:"id"`
		Snippet struct {
			Title string `json:"title"`
		} `json:"snippet"`
	} `json:"items"`
}

// 1 MiB: la risposta per un canale è di pochi kilobyte.
const maxAPIBytes = 1 << 20

func (a *DataAPI) ByHandle(ctx context.Context, handle string) (ChannelPayload, error) {
	q := url.Values{"part": {"snippet"}, "forHandle": {handle}, "key": {a.key}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/channels?"+q.Encode(), nil)
	if err != nil {
		return ChannelPayload{}, fmt.Errorf("richiesta del canale %s: %w", handle, err)
	}
	resp, err := a.client.Do(req)
	if err != nil {
		// L'errore del client contiene l'URL, e l'URL contiene la chiave: nel log va
		// solo il fatto che la richiesta è fallita.
		return ChannelPayload{}, apiUnavailable(fmt.Errorf("richiesta fallita"))
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusForbidden {
		return ChannelPayload{}, apperr.Unavailable("youtube_api_rejected",
			"la YouTube Data API ha rifiutato la richiesta: controlla la chiave e la quota",
			fmt.Errorf("status %d", resp.StatusCode))
	}
	if resp.StatusCode != http.StatusOK {
		return ChannelPayload{}, apiUnavailable(fmt.Errorf("status %d", resp.StatusCode))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxAPIBytes))
	if err != nil {
		return ChannelPayload{}, apiUnavailable(err)
	}
	var out channelsListResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return ChannelPayload{}, apiUnavailable(err)
	}
	// Un handle che non esiste non è un errore dell'API: la lista torna vuota.
	if len(out.Items) == 0 || out.Items[0].ID == "" {
		return ChannelPayload{}, apperr.NotFound("channel_not_found", "nessun canale con questo handle")
	}
	return ChannelPayload{ID: out.Items[0].ID, Title: out.Items[0].Snippet.Title}, nil
}

func apiUnavailable(err error) error {
	return apperr.Unavailable("youtube_unavailable", "YouTube non risponde", err)
}
