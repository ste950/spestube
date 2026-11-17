// Package library è il soggetto dei canali seguiti e dei loro video: come si aggiunge un
// canale, quali video entrano, quando diventano visibili e quando escono di scena.
package library

import (
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ste950/spestube/backend/internal/core/apperr"
)

// FeedMode dice da quale feed di YouTube arrivano i video di un canale.
type FeedMode string

const (
	// FeedLongVideos è la playlist "UULF…": contiene solo i video lunghi, senza Shorts.
	FeedLongVideos FeedMode = "uulf"
	// FeedChannel è il feed del canale: contiene tutto, Shorts compresi. Si usa solo se
	// la playlist dei video lunghi non risponde.
	FeedChannel FeedMode = "channel"
)

func (m FeedMode) Valid() bool {
	return m == FeedLongVideos || m == FeedChannel
}

type Channel struct {
	ID             string
	Title          string
	Handle         string // vuoto se il canale è stato aggiunto dal suo id
	FeedMode       FeedMode
	AddedAt        time.Time
	LastFetchedAt  time.Time // zero se il feed non è mai stato letto
	LastFetchError string
}

// ChannelRef è ciò che l'utente ha indicato per aggiungere un canale: l'id, oppure
// l'handle da risolvere. Esattamente uno dei due campi è valorizzato.
type ChannelRef struct {
	ID     string
	Handle string // con la chiocciola
}

var (
	channelIDRe = regexp.MustCompile(`^UC[A-Za-z0-9_-]{22}$`)
	handleRe    = regexp.MustCompile(`^@[\p{L}\p{N}._-]{3,30}$`)
)

// ParseChannelRef riconosce un id ("UC…"), un handle ("@nome") e i link di YouTube che
// contengono l'uno o l'altro. I vecchi link /c/nome e /user/nome non si accettano:
// risolverli richiederebbe leggere la pagina HTML di YouTube.
func ParseChannelRef(input string) (ChannelRef, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return ChannelRef{}, apperr.Invalid("channel_ref_required", "indica un canale: link, @handle o id")
	}
	if channelIDRe.MatchString(s) {
		return ChannelRef{ID: s}, nil
	}
	if strings.HasPrefix(s, "@") {
		return handleRef(s)
	}

	raw := s
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ChannelRef{}, unsupportedRef()
	}
	host := strings.ToLower(u.Hostname())
	host = strings.TrimPrefix(strings.TrimPrefix(host, "www."), "m.")
	if host != "youtube.com" {
		return ChannelRef{}, unsupportedRef()
	}
	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	switch {
	case len(segments) >= 2 && segments[0] == "channel" && channelIDRe.MatchString(segments[1]):
		return ChannelRef{ID: segments[1]}, nil
	case strings.HasPrefix(segments[0], "@"):
		return handleRef(segments[0])
	}
	return ChannelRef{}, unsupportedRef()
}

func handleRef(s string) (ChannelRef, error) {
	if !handleRe.MatchString(s) {
		return ChannelRef{}, apperr.Invalid("channel_ref_invalid", "l'handle non è valido")
	}
	return ChannelRef{Handle: s}, nil
}

func unsupportedRef() error {
	return apperr.Invalid("channel_ref_unsupported",
		"non riconosco il canale: usa un link youtube.com/channel/UC…, un @handle o l'id del canale")
}
