# Entità di Dominio

Un tipo del dominio è una `struct` semplice, passata per valore, senza tag `json` né `db`: la
forma nel database e quella nel contratto API sono affari di `data/` e di `api/`. Gli insiemi
chiusi sono tipi `string` con le costanti e un `Valid()`; la logica derivata sta nei metodi.

```go
// backend/internal/domain/catalog/item.go

// Package catalog è il soggetto degli item salvati: cosa sono, quando vanno
// aggiornati, e le regole per aggiungerne uno.
package catalog

import "time"

// Status è un insieme chiuso: le costanti qui sotto sono gli unici valori validi.
type Status string

const (
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
)

func (s Status) Valid() bool {
	return s == StatusActive || s == StatusArchived
}

// Item si passa per valore: chi lo riceve ne ha una copia, e nessuno lo modifica
// alle spalle di un altro.
type Item struct {
	ID          int64
	URL         string
	Title       string
	Status      Status
	CreatedAt   time.Time
	RefreshedAt time.Time // zero se la sorgente non ha mai risposto
}

// IsStale riporta se il titolo va riletto dalla sorgente. Un item archiviato non lo è
// mai: nessuno lo guarda, e ogni richiesta alla sorgente ha un costo.
func (i Item) IsStale(now time.Time, maxAge time.Duration) bool {
	if i.Status != StatusActive {
		return false
	}
	return i.RefreshedAt.IsZero() || now.Sub(i.RefreshedAt) > maxAge
}
```

-   **`time.Time` zero** al posto di un puntatore per «mai successo»: `IsZero()` si legge meglio
    di un controllo `nil`, e nessuno deve ricordarsi di dereferenziare.
-   **Il metodo riceve `now`**: la regola si prova a qualsiasi ora senza toccare l'orologio
    ([Test](../rules/testing.rule.md#%EF%B8%8F-tempo-e-casualità)).
