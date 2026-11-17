// Package sources isola YouTube: il formato dei feed, le risposte dell'API e la rete che
// può mancare non escono da qui.
package sources

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ste950/spestube/backend/internal/core/apperr"
)

// YouTubeFeedsURL è l'indirizzo dei feed RSS pubblici: gratuiti e senza quota.
const YouTubeFeedsURL = "https://www.youtube.com/feeds/videos.xml"

// FeedSource legge un feed di YouTube: quello di una playlist o quello di un canale.
type FeedSource interface {
	Playlist(ctx context.Context, playlistID string) (FeedPayload, error)
	Channel(ctx context.Context, channelID string) (FeedPayload, error)
}

// FeedPayload è il feed come arriva: non un'entità.
type FeedPayload struct {
	Author  string
	Title   string
	Entries []FeedEntryPayload
}

type FeedEntryPayload struct {
	VideoID   string
	Title     string
	Published time.Time
}

// RSSFeeds legge i feed sul web.
type RSSFeeds struct {
	client    *http.Client
	baseURL   string
	userAgent string
}

func NewRSSFeeds(baseURL string, timeout time.Duration, userAgent string) *RSSFeeds {
	return &RSSFeeds{client: &http.Client{Timeout: timeout}, baseURL: baseURL, userAgent: userAgent}
}

func (f *RSSFeeds) Playlist(ctx context.Context, playlistID string) (FeedPayload, error) {
	return f.fetch(ctx, "playlist_id", playlistID)
}

func (f *RSSFeeds) Channel(ctx context.Context, channelID string) (FeedPayload, error) {
	return f.fetch(ctx, "channel_id", channelID)
}

// 2 MiB: un feed vero di quindici video pesa meno di 100 KiB, descrizioni comprese.
const maxFeedBytes = 2 << 20

func (f *RSSFeeds) fetch(ctx context.Context, param, id string) (FeedPayload, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.baseURL+"?"+url.Values{param: {id}}.Encode(), nil)
	if err != nil {
		return FeedPayload{}, fmt.Errorf("richiesta del feed %s: %w", id, err)
	}
	req.Header.Set("User-Agent", f.userAgent)

	resp, err := f.client.Do(req)
	if err != nil {
		return FeedPayload{}, feedUnavailable(err)
	}
	defer resp.Body.Close()

	// YouTube risponde 404 sia per un canale inesistente sia per una playlist che il
	// canale non ha: distinguere i due casi spetta a chi sa quale feed ha chiesto.
	if resp.StatusCode == http.StatusNotFound {
		return FeedPayload{}, apperr.NotFound("feed_not_found", "feed inesistente")
	}
	if resp.StatusCode != http.StatusOK {
		return FeedPayload{}, feedUnavailable(fmt.Errorf("status %d", resp.StatusCode))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxFeedBytes))
	if err != nil {
		return FeedPayload{}, feedUnavailable(err)
	}
	payload, err := ParseFeed(body)
	if err != nil {
		return FeedPayload{}, feedUnavailable(err)
	}
	return payload, nil
}

func feedUnavailable(err error) error {
	return apperr.Unavailable("youtube_unavailable", "YouTube non risponde", err)
}

type atomFeed struct {
	Title  string `xml:"title"`
	Author struct {
		Name string `xml:"name"`
	} `xml:"author"`
	Entries []struct {
		VideoID   string `xml:"http://www.youtube.com/xml/schemas/2015 videoId"`
		Title     string `xml:"title"`
		Published string `xml:"published"`
	} `xml:"entry"`
}

// ParseFeed legge il documento Atom di un feed di YouTube. Una voce senza id o con una
// data illeggibile si salta: un video malformato non deve nascondere gli altri.
func ParseFeed(body []byte) (FeedPayload, error) {
	var doc atomFeed
	if err := xml.Unmarshal(body, &doc); err != nil {
		return FeedPayload{}, fmt.Errorf("feed non leggibile: %w", err)
	}
	out := FeedPayload{Author: strings.TrimSpace(doc.Author.Name), Title: strings.TrimSpace(doc.Title)}
	for _, e := range doc.Entries {
		published, err := time.Parse(time.RFC3339, strings.TrimSpace(e.Published))
		if err != nil || strings.TrimSpace(e.VideoID) == "" {
			continue
		}
		out.Entries = append(out.Entries, FeedEntryPayload{
			VideoID:   strings.TrimSpace(e.VideoID),
			Title:     strings.TrimSpace(e.Title),
			Published: published.UTC(),
		})
	}
	return out, nil
}
