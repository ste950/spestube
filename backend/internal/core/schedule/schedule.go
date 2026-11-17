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
