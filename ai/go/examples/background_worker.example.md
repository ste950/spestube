# Job in Background

Un lavoro periodico si divide in due: **cosa** fare è un metodo del servizio di dominio
(`RefreshStale`), **quando** farlo è un ciclo agnostico in `core/schedule`. Il composition root
li unisce e possiede la goroutine. Vedi [Concorrenza](../rules/concurrency.rule.md).

```go
// backend/internal/core/schedule/schedule.go

// Package schedule ripete un lavoro a intervalli finché il contesto non viene
// cancellato. Non sa niente del lavoro che ripete.
package schedule

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"time"
)

// Every esegue job subito e poi ogni interval, finché ctx è vivo. Un errore si scrive
// nel log e non ferma il ciclo: il giro dopo è il nuovo tentativo.
//
// Ogni attesa ha fino a un decimo di variazione casuale, così più cicli avviati
// insieme non colpiscono la stessa risorsa nello stesso istante per sempre.
func Every(ctx context.Context, name string, interval time.Duration, job func(context.Context) error) {
	for {
		if err := job(ctx); err != nil && ctx.Err() == nil {
			slog.Warn("lavoro periodico fallito", "job", name, "err", err)
		}
		wait := interval + rand.N(interval/10+1)
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
	}
}
```

Nel composition root ([Main e shutdown](./main_and_shutdown.example.md)):

```go
var wg sync.WaitGroup
wg.Go(func() {
	schedule.Every(ctx, "refresh-titles", cfg.RefreshInterval, catalogSvc.RefreshStale)
})
// … allo shutdown, dopo srv.Shutdown:
wg.Wait()
```

Varianti frequenti:

-   **Intervallo per elemento** (ogni feed il suo): il ciclo gira a passo fisso e corto (es. un
    minuto) e il servizio sceglie dal database gli elementi la cui prossima scadenza è passata.
    Un ciclo per elemento sarebbero N goroutine da tenere allineate con il database.
-   **Orari fissi** (un riepilogo alle 8 e alle 20): il servizio riceve `now` e decide se
    l'orario è passato dall'ultimo invio, registrato nel database. Il ciclo resta lo stesso, e
    un riavvio del processo non fa saltare né raddoppiare un invio.
-   **Lavoro in parallelo dentro un giro** (scaricare più feed): semaforo e `WaitGroup` dentro
    il metodo del servizio, che ritorna solo quando tutte le goroutine sono finite.

```go
sem := make(chan struct{}, 4) // al più 4 download insieme: la rete del dispositivo è piccola
var wg sync.WaitGroup
for _, f := range due {
	sem <- struct{}{}
	wg.Go(func() {
		defer func() { <-sem }()
		s.fetchOne(ctx, f)
	})
}
wg.Wait()
```
