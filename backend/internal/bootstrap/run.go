// Package bootstrap è il composition root: costruisce tutto, avvia tutto e ferma tutto.
package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/ste950/spestube/backend/internal/api"
	"github.com/ste950/spestube/backend/internal/api/meta"
	"github.com/ste950/spestube/backend/internal/core/schedule"
	"github.com/ste950/spestube/backend/internal/core/sqlitedb"
	datalibrary "github.com/ste950/spestube/backend/internal/data/library"
	"github.com/ste950/spestube/backend/internal/data/library/sources"
	"github.com/ste950/spestube/backend/internal/data/migrations"
	datasettings "github.com/ste950/spestube/backend/internal/data/settings"
	"github.com/ste950/spestube/backend/internal/domain/library"
	"github.com/ste950/spestube/backend/internal/domain/settings"
	"github.com/ste950/spestube/backend/internal/webui"
)

// Run blocca finché ctx non viene cancellato, poi ferma server e job e chiude il
// database, in quest'ordine.
func Run(ctx context.Context, cfg Config, version string) error {
	if dir := filepath.Dir(cfg.DBPath); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("cartella del database: %w", err)
		}
	}
	db, err := sqlitedb.Open(ctx, cfg.DBPath)
	if err != nil {
		return err
	}
	defer db.Close() // per ultimo: i job potrebbero star finendo una scrittura

	if err := sqlitedb.Migrate(ctx, db, migrations.FS); err != nil {
		return err
	}

	settingsSvc := settings.NewService(
		datasettings.NewRepository(db), settings.Defaults(cfg.ReleaseTime), time.Now)

	// Senza chiave la sorgente dell'API non esiste proprio: il repository degli handle
	// riceve nil e risponde che serve la chiave. L'assegnazione è dentro l'if perché un
	// puntatore nil messo in un'interfaccia non è più un'interfaccia nil.
	var lookup sources.ChannelLookup
	if cfg.YouTubeAPIKey != "" {
		lookup = sources.NewDataAPI(sources.YouTubeDataAPIURL, cfg.YouTubeAPIKey, cfg.FetchTimeout)
	}
	feeds := sources.NewRSSFeeds(sources.YouTubeFeedsURL, cfg.FetchTimeout, "spestube/"+version)

	videos := datalibrary.NewVideoRepository(db)
	librarySvc := library.NewService(library.Deps{
		Channels: datalibrary.NewChannelRepository(db),
		Videos:   videos,
		Feeds:    datalibrary.NewFeedRepository(feeds),
		Handles:  datalibrary.NewHandleRepository(lookup),
		Policy:   policyFromSettings{settingsSvc},
		Now:      time.Now,
		// Il fuso è quello del processo, cioè la variabile TZ: "alle 7" sono le 7 di casa.
		Loc: time.Local,
	})

	collectionsSvc := library.NewCollections(datalibrary.NewCollectionRepository(db), videos, time.Now)

	srv := &http.Server{
		Addr: ":" + strconv.Itoa(cfg.Port),
		Handler: api.Routes(api.Services{
			Library:     librarySvc,
			Collections: collectionsSvc,
			Settings:    settingsSvc,
			Meta:        meta.Info{Version: version, Dev: cfg.Dev, HandleLookup: lookup != nil},
			Ping:        db.PingContext,
			Static:      webui.Handler(),
		}),
		ReadHeaderTimeout: 10 * time.Second,
	}

	var wg sync.WaitGroup
	wg.Go(func() { schedule.Every(ctx, "fetch", cfg.FetchInterval, librarySvc.FetchAll) })
	wg.Go(func() { schedule.Every(ctx, "release", cfg.ReleaseInterval, librarySvc.ReleaseDue) })

	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.ListenAndServe() }()
	slog.Info("in ascolto", "addr", srv.Addr, "version", version, "dev", cfg.Dev,
		"release_time", cfg.ReleaseTime.String(), "api_key", lookup != nil)

	select {
	case <-ctx.Done():
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server: %w", err)
		}
	}

	// 10 s per finire le richieste in corso: è il tempo che Docker concede prima del
	// SIGKILL, e un download di feed in corso si interrompe con il contesto.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Warn("shutdown del server", "err", err)
	}
	wg.Wait()
	return nil
}

// policyFromSettings collega due soggetti che non si conoscono: library chiede orario e
// filtri, settings li possiede. Il collegamento sta qui perché è l'unico posto che può
// importarli entrambi.
type policyFromSettings struct {
	svc *settings.Service
}

func (p policyFromSettings) Policy(ctx context.Context) (library.Policy, error) {
	s, err := p.svc.Get(ctx)
	if err != nil {
		return library.Policy{}, err
	}
	return library.Policy{
		ReleaseAt: s.ReleaseTime,
		Filters: library.Filters{
			ExcludeShorts:    s.ExcludeShorts,
			ExcludeLive:      s.ExcludeLive,
			MinDuration:      time.Duration(s.MinDurationMinutes) * time.Minute,
			ExcludedKeywords: s.ExcludedKeywords,
		},
	}, nil
}
