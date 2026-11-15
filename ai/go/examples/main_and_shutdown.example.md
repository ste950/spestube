# Main e shutdown

`main` fa tre cose: legge la configurazione, crea il contesto che i segnali cancellano, chiama
`bootstrap.Run`. `Run` è il composition root: costruisce le implementazioni, le consegna al
dominio come interfacce, avvia server e job e li ferma nell'ordine giusto. Vedi
[Struttura del progetto](../rules/project_structure.rule.md) e
[Concorrenza](../rules/concurrency.rule.md).

```go
// backend/cmd/app/main.go

package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"example.com/backend/internal/bootstrap"
)

// version la imposta la build con -ldflags "-X main.version=…".
var version = "dev"

func main() {
	envFile := flag.String("env", "", "file KEY=VALUE da caricare (sviluppo)")
	flag.Parse()

	if err := run(*envFile); err != nil {
		fmt.Fprintln(os.Stderr, "errore:", err)
		os.Exit(1)
	}
}

func run(envFile string) error {
	if envFile != "" {
		if err := bootstrap.LoadEnvFile(envFile); err != nil {
			return err
		}
	}
	cfg, err := bootstrap.Load(os.Getenv)
	if err != nil {
		return err
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		return fmt.Errorf("%sLOG_LEVEL: %w", bootstrap.Prefix, err)
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	// SIGTERM è come systemd ferma il servizio: è il segnale di stop di tutto il resto.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return bootstrap.Run(ctx, cfg, version)
}
```

```go
// backend/internal/bootstrap/run.go

// Package bootstrap è il composition root: costruisce tutto, avvia tutto e ferma tutto.
package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"example.com/backend/internal/api"
	"example.com/backend/internal/core/schedule"
	"example.com/backend/internal/core/sqlitedb"
	datacatalog "example.com/backend/internal/data/catalog"
	"example.com/backend/internal/data/catalog/sources"
	"example.com/backend/internal/data/migrations"
	"example.com/backend/internal/domain/catalog"
)

// Run blocca finché ctx non viene cancellato, poi ferma server e job e chiude il
// database, in quest'ordine.
func Run(ctx context.Context, cfg Config, version string) error {
	db, err := sqlitedb.Open(ctx, cfg.DBPath)
	if err != nil {
		return err
	}
	defer db.Close() // per ultimo: i job potrebbero star finendo una scrittura

	if err := sqlitedb.Migrate(ctx, db, migrations.FS); err != nil {
		return err
	}

	// Il cablaggio: implementazioni di data/ consegnate al dominio come interfacce.
	// Se un'implementazione smette di soddisfare il contratto, non compila qui.
	pages := sources.NewHTTPPages(cfg.FetchTimeout, "app/"+version)
	catalogSvc := catalog.NewService(
		datacatalog.NewItemRepository(db),
		datacatalog.NewTitleRepository(pages),
		time.Now,
	)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           api.Routes(api.Services{Catalog: catalogSvc, Version: version}),
		ReadHeaderTimeout: 10 * time.Second,
	}

	var wg sync.WaitGroup
	wg.Go(func() {
		schedule.Every(ctx, "refresh-titles", cfg.RefreshInterval, catalogSvc.RefreshStale)
	})

	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.ListenAndServe() }()
	slog.Info("in ascolto", "addr", cfg.Addr, "version", version)

	select {
	case <-ctx.Done():
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server: %w", err)
		}
	}

	// 10 s per finire le richieste in corso: systemd aspetta 90 s prima del SIGKILL,
	// quindi c'è margine anche per i job.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Warn("shutdown del server", "err", err)
	}
	wg.Wait()
	return nil
}
```

Note:

-   **Il cablaggio è codice**: `catalog.NewService(datacatalog.NewItemRepository(db), …)`. Se
    `ItemRepository` smette di soddisfare l'interfaccia del dominio, non compila questa riga —
    è il controllo che altrove fa un container di DI a runtime.
-   **`wg.Go`** (Go 1.25+) avvia la goroutine e la registra nel `WaitGroup` in una chiamata: non
    si può dimenticare l'`Add`.
-   **L'ordine dello shutdown** è nel codice e non nei `defer` sparsi: server, poi job, poi
    database (il `defer db.Close()` è il primo registrato, quindi l'ultimo eseguito).
