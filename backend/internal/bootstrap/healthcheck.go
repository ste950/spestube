package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Healthcheck interroga /healthz del processo in ascolto sulla porta configurata, e
// restituisce un errore se non risponde 200.
func Healthcheck(ctx context.Context, cfg Config) error {
	// Tre secondi: /healthz fa solo un ping al database, e un servizio che ci mette di
	// più non sta bene.
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	url := "http://127.0.0.1:" + strconv.Itoa(cfg.Port) + "/healthz"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("il servizio non risponde: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("il servizio risponde %d", resp.StatusCode)
	}
	return nil
}
