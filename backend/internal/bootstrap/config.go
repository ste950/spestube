package bootstrap

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ste950/spestube/backend/internal/core/clock"
)

// Prefix è il prefisso delle variabili d'ambiente del progetto.
const Prefix = "SPESTUBE_"

// Config è un dato: nessun metodo che decide qualcosa, solo valori validati.
type Config struct {
	Port   int
	DBPath string
	// ReleaseTime è l'orario di arrivo dei video per un'installazione nuova: dopo, vale
	// quello salvato nelle impostazioni.
	ReleaseTime clock.TimeOfDay
	// YouTubeAPIKey è facoltativa: senza, l'app funziona con i soli feed RSS.
	YouTubeAPIKey string
	// Dev monta gli endpoint che forzano download e rilascio.
	Dev             bool
	LogLevel        string
	FetchInterval   time.Duration
	FetchTimeout    time.Duration
	ReleaseInterval time.Duration
}

// Load legge l'ambiente con i default, e fallisce con il nome della variabile se un
// valore è malformato: il binario non parte con una configurazione a metà.
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		Port:          8080,
		DBPath:        "data/spestube.db",
		ReleaseTime:   clock.TimeOfDay{Hour: 7},
		LogLevel:      "info",
		FetchInterval: time.Hour,
		// YouTube risponde in meno di un secondo: quindici bastano anche alla rete di casa
		// in una brutta giornata, senza tenere fermo un giro di download.
		FetchTimeout: 15 * time.Second,
		// Il rilascio si controlla ogni minuto: è il ritardo massimo fra l'orario scelto
		// e la comparsa dei video.
		ReleaseInterval: time.Minute,
	}

	if v := getenv(Prefix + "PORT"); v != "" {
		port, err := strconv.Atoi(v)
		if err != nil || port < 1 || port > 65535 {
			return Config{}, fmt.Errorf("%sPORT: porta non valida %q", Prefix, v)
		}
		cfg.Port = port
	}
	if v := getenv(Prefix + "DB_PATH"); v != "" {
		cfg.DBPath = v
	}
	if v := getenv(Prefix + "RELEASE_TIME"); v != "" {
		t, err := clock.ParseTimeOfDay(v)
		if err != nil {
			return Config{}, fmt.Errorf("%sRELEASE_TIME: %w", Prefix, err)
		}
		cfg.ReleaseTime = t
	}
	if v := getenv(Prefix + "DEV"); v != "" {
		dev, err := strconv.ParseBool(v)
		if err != nil {
			return Config{}, fmt.Errorf("%sDEV: valore non valido %q (true o false)", Prefix, v)
		}
		cfg.Dev = dev
	}
	if v := getenv(Prefix + "LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	cfg.YouTubeAPIKey = strings.TrimSpace(getenv(Prefix + "YT_API_KEY"))
	return cfg, nil
}

// LoadEnvFile carica un file KEY=VALUE nell'ambiente del processo, senza sovrascrivere
// ciò che è già impostato. In produzione il file lo legge Docker Compose: questo serve
// allo sviluppo, con lo stesso formato.
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
