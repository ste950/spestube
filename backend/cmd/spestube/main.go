package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	// Il database dei fusi orari entra nel binario: l'immagine Docker non ne ha uno, e
	// senza di esso TZ=Europe/Rome verrebbe ignorata in silenzio.
	_ "time/tzdata"

	"github.com/ste950/spestube/backend/internal/bootstrap"
)

// version la imposta la build con -ldflags "-X main.version=…".
var version = "dev"

func main() {
	envFile := flag.String("env", "", "file KEY=VALUE da caricare (sviluppo)")
	healthcheck := flag.Bool("healthcheck", false, "interroga /healthz del processo in ascolto ed esce")
	flag.Parse()

	if err := run(*envFile, *healthcheck); err != nil {
		fmt.Fprintln(os.Stderr, "errore:", err)
		os.Exit(1)
	}
}

func run(envFile string, healthcheck bool) error {
	if envFile != "" {
		if err := bootstrap.LoadEnvFile(envFile); err != nil {
			return err
		}
	}
	cfg, err := bootstrap.Load(os.Getenv)
	if err != nil {
		return err
	}

	// L'immagine Docker non ha né shell né curl: a chiedere al servizio se sta bene è lo
	// stesso binario, lanciato una seconda volta dal controllo di salute del container.
	if healthcheck {
		return bootstrap.Healthcheck(context.Background(), cfg)
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		return fmt.Errorf("%sLOG_LEVEL: %w", bootstrap.Prefix, err)
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	// SIGTERM è come Docker ferma il container: è il segnale di stop di tutto il resto.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return bootstrap.Run(ctx, cfg, version)
}
