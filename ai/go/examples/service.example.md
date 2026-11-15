# Servizio di Dominio

I casi d'uso di un soggetto sono i metodi del suo `Service`. Il servizio riceve le interfacce
dei repository e l'orologio nel costruttore; non sa da dove arrivano i dati né chi lo chiama —
un handler HTTP o un job periodico.

```go
// backend/internal/domain/catalog/service.go

package catalog

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"example.com/backend/internal/core/apperr"
)

// Un giorno: il titolo di una pagina cambia di rado, e chi lo legge tollera di vederlo
// vecchio di qualche ora.
const titleMaxAge = 24 * time.Hour

// Service raccoglie i casi d'uso del soggetto. Riceve le interfacce, non le
// implementazioni: le sceglie il composition root.
type Service struct {
	items  ItemRepository
	titles TitleRepository
	now    func() time.Time
}

func NewService(items ItemRepository, titles TitleRepository, now func() time.Time) *Service {
	return &Service{items: items, titles: titles, now: now}
}

func (s *Service) Get(ctx context.Context, id int64) (Item, error) {
	return s.items.Get(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]Item, error) {
	return s.items.ListActive(ctx)
}

// Add salva un URL nuovo. Il titolo provvisorio è l'host: quello vero arriva al primo
// RefreshStale, perché aspettare la sorgente qui renderebbe l'aggiunta lenta quanto
// la pagina più lenta.
func (s *Service) Add(ctx context.Context, rawURL string) (Item, error) {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Item{}, apperr.Invalid("invalid_url", "l'indirizzo deve essere un URL http o https")
	}
	if _, ok, err := s.items.FindByURL(ctx, u.String()); err != nil {
		return Item{}, err
	} else if ok {
		return Item{}, apperr.Conflict("item_exists", "l'indirizzo è già salvato")
	}
	return s.items.Add(ctx, Item{
		URL:       u.String(),
		Title:     u.Host,
		Status:    StatusActive,
		CreatedAt: s.now(),
	})
}

// RefreshStale rilegge i titoli scaduti. Una sorgente irraggiungibile non ferma il
// giro: si riprova al prossimo, e l'errore torna a chi ha chiamato per il log.
func (s *Service) RefreshStale(ctx context.Context) error {
	items, err := s.items.ListActive(ctx)
	if err != nil {
		return err
	}
	now := s.now()
	var errs []error
	for _, it := range items {
		if !it.IsStale(now, titleMaxAge) {
			continue
		}
		title, err := s.titles.Title(ctx, it.URL)
		if err != nil {
			if apperr.KindOf(err) == apperr.KindUnavailable {
				errs = append(errs, fmt.Errorf("item %d: %w", it.ID, err))
				continue
			}
			return err
		}
		if err := s.items.SetTitle(ctx, it.ID, title, now); err != nil {
			return err
		}
	}
	return errors.Join(errs...)
}
```

-   **La validazione sta qui**, non nell'handler: un job o un comando da riga di comando che
    chiama `Add` ottiene le stesse garanzie.
-   **`errors.Join`** raccoglie gli errori di un giro senza fermarlo: chi chiama decide se
    loggarli, e `apperr.KindOf` trova ancora il genere del primo.
