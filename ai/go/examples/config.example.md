# Config

La configurazione si carica una volta, dall'ambiente, con i default e la validazione in un
posto solo. `Load` riceve la funzione che legge l'ambiente invece di chiamare `os.Getenv`: il
suo test passa una mappa. Vedi [Configurazione](../rules/configuration.rule.md).

```go
// backend/internal/bootstrap/config.go

package bootstrap

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// Prefix è il prefisso delle variabili d'ambiente del progetto.
const Prefix = "APP_"

// Config è un dato: nessun metodo che decide qualcosa, solo valori validati.
type Config struct {
	Addr            string
	DBPath          string
	RefreshInterval time.Duration
	FetchTimeout    time.Duration
	LogLevel        string
}

// Load legge l'ambiente con i default, e fallisce con il nome della variabile se un
// valore è malformato: il binario non parte con una configurazione a metà.
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		Addr:            "127.0.0.1:8080",
		DBPath:          "app.db",
		RefreshInterval: 30 * time.Minute,
		FetchTimeout:    15 * time.Second,
		LogLevel:        "info",
	}
	// $STATE_DIRECTORY lo imposta systemd con StateDirectory=: il default del database
	// segue la cartella che il servizio possiede.
	if dir := getenv("STATE_DIRECTORY"); dir != "" {
		cfg.DBPath = dir + "/app.db"
	}

	str(getenv, "ADDR", &cfg.Addr)
	str(getenv, "DB_PATH", &cfg.DBPath)
	str(getenv, "LOG_LEVEL", &cfg.LogLevel)
	if err := dur(getenv, "REFRESH_INTERVAL", &cfg.RefreshInterval); err != nil {
		return Config{}, err
	}
	if err := dur(getenv, "FETCH_TIMEOUT", &cfg.FetchTimeout); err != nil {
		return Config{}, err
	}
	if cfg.RefreshInterval < time.Minute {
		return Config{}, fmt.Errorf("%sREFRESH_INTERVAL: almeno 1m, trovato %s", Prefix, cfg.RefreshInterval)
	}
	return cfg, nil
}

func str(getenv func(string) string, key string, dst *string) {
	if v := getenv(Prefix + key); v != "" {
		*dst = v
	}
}

func dur(getenv func(string) string, key string, dst *time.Duration) error {
	v := getenv(Prefix + key)
	if v == "" {
		return nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fmt.Errorf("%s%s: durata non valida %q (es. 30s, 5m)", Prefix, key, v)
	}
	*dst = d
	return nil
}

// LoadEnvFile carica un file KEY=VALUE nell'ambiente del processo, senza sovrascrivere
// ciò che è già impostato. In produzione il file lo legge systemd (EnvironmentFile=):
// questo serve allo sviluppo, con lo stesso formato.
func LoadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("%s:%d: manca '='", path, n)
		}
		if _, set := os.LookupEnv(k); !set {
			os.Setenv(strings.TrimSpace(k), strings.Trim(strings.TrimSpace(v), `"`))
		}
	}
	return sc.Err()
}
```

Il file d'esempio, versionato nel repo con valori finti (di norma `deploy/<binario>.env.example`):

```bash
# Copiare in /etc/<binario>/<binario>.env (permessi 0600) e compilare.
APP_ADDR=127.0.0.1:8080
APP_REFRESH_INTERVAL=30m
APP_FETCH_TIMEOUT=15s
APP_LOG_LEVEL=info
# APP_DB_PATH si omette in produzione: il default segue $STATE_DIRECTORY.
```

In sviluppo: `go run ./cmd/app -env dev.env`. In produzione lo stesso file lo carica systemd
([Unit systemd](./systemd_unit.example.md)).
