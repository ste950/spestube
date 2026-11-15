# Sorgente Esterna

Una fonte di terze parti — un sito, un'API, un servizio di notifica — si attraversa dentro
`internal/data/<soggetto>/sources/`: interfaccia, payload e adattatore. Il repository la
riceve nel costruttore e la compone; il dominio non la vede mai.

```go
// backend/internal/data/catalog/sources/pages.go

// Package sources isola le fonti esterne del soggetto: il formato che non si
// controlla e la rete che può mancare non escono da qui.
package sources

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"example.com/backend/internal/core/apperr"
)

// PageSource è il contratto che il repository consuma. Sta qui, nel data layer,
// perché è il repository e non il dominio a usarlo.
type PageSource interface {
	Fetch(ctx context.Context, url string) (Page, error)
}

// Page è il payload: la forma del dato come arriva dalla fonte, non un'entità.
type Page struct {
	Title string
}

// HTTPPages legge le pagine sul web.
type HTTPPages struct {
	client    *http.Client
	userAgent string
}

// Il timeout è sul client e non solo sul contesto: copre anche il caso in cui chi
// chiama abbia passato un contesto senza scadenza.
func NewHTTPPages(timeout time.Duration, userAgent string) *HTTPPages {
	return &HTTPPages{client: &http.Client{Timeout: timeout}, userAgent: userAgent}
}

var titleRe = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)

// 512 KiB: il <title> sta nell'head, e leggere oltre è solo traffico.
const maxPageBytes = 512 << 10

func (p *HTTPPages) Fetch(ctx context.Context, url string) (Page, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Page{}, apperr.Invalid("invalid_url", "indirizzo non valido")
	}
	req.Header.Set("User-Agent", p.userAgent)

	resp, err := p.client.Do(req)
	if err != nil {
		return Page{}, apperr.Unavailable("source_unavailable", "pagina irraggiungibile", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Page{}, apperr.Unavailable("source_unavailable",
			"pagina irraggiungibile", fmt.Errorf("status %d", resp.StatusCode))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxPageBytes))
	if err != nil {
		return Page{}, apperr.Unavailable("source_unavailable", "lettura della pagina", err)
	}
	m := titleRe.FindSubmatch(body)
	if m == nil {
		return Page{}, nil // una pagina senza titolo non è un errore della fonte
	}
	return Page{Title: strings.TrimSpace(html.UnescapeString(string(m[1])))}, nil
}
```

Il repository che la compone e soddisfa il contratto del dominio:

```go
// backend/internal/data/catalog/title_repository.go

package catalog

import (
	"context"

	"example.com/backend/internal/data/catalog/sources"
)

// TitleRepository soddisfa catalog.TitleRepository componendo una sorgente: è lui a
// decidere cosa fare quando la fonte risponde male, non il dominio.
type TitleRepository struct {
	pages sources.PageSource
}

func NewTitleRepository(pages sources.PageSource) *TitleRepository {
	return &TitleRepository{pages: pages}
}

func (r *TitleRepository) Title(ctx context.Context, url string) (string, error) {
	page, err := r.pages.Fetch(ctx, url)
	if err != nil {
		return "", err
	}
	if page.Title == "" {
		return url, nil // meglio l'indirizzo di un titolo vuoto nella lista
	}
	return page.Title, nil
}
```

-   **Timeout sul client**, limite sulla lettura del corpo, `User-Agent` esplicito: una fonte
    lenta o enorme non deve poter bloccare un job.
-   **Gli errori di rete diventano `Unavailable`**, quelli di input `Invalid`: il dominio decide
    cosa fare guardando il genere, non il testo.
-   **Si prova contro un `httptest.Server`** che risponde con una fixture registrata dalla fonte
    vera, mai contro la rete ([Test](../rules/testing.rule.md)).
