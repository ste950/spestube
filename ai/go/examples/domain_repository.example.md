# Contratto Repository

Le interfacce dei repository stanno nel dominio, perché è il dominio a consumarle. Parlano di
entità e di `context`, mai di righe o di `*sql.DB`. Vedi
[Struttura del progetto](../rules/project_structure.rule.md).

```go
// backend/internal/domain/catalog/repository.go

package catalog

import (
	"context"
	"time"
)

// ItemRepository è il contratto dell'archivio degli item. Get restituisce un errore
// apperr.KindNotFound se l'item non c'è; FindByURL restituisce ok=false, perché lì
// l'assenza è la risposta normale e non un errore.
type ItemRepository interface {
	Get(ctx context.Context, id int64) (Item, error)
	FindByURL(ctx context.Context, url string) (item Item, ok bool, err error)
	Add(ctx context.Context, item Item) (Item, error)
	ListActive(ctx context.Context) ([]Item, error)
	SetTitle(ctx context.Context, id int64, title string, refreshedAt time.Time) error
}

// TitleRepository legge il titolo corrente di una pagina, da qualunque parte arrivi.
type TitleRepository interface {
	Title(ctx context.Context, url string) (string, error)
}
```

Convenzione sull'assenza, la stessa ovunque:

| Firma | Assenza |
|---|---|
| `Get(ctx, id) (Item, error)` | errore `apperr.KindNotFound`: chi chiama si aspettava che ci fosse |
| `FindBy…(ctx, …) (Item, bool, error)` | `ok == false`, nessun errore: l'assenza è una risposta normale |

Interfacce **piccole**: un metodo che nessun consumatore usa non si aggiunge «per completezza».
Ogni metodo in più è un metodo in più da scrivere nel fake.
