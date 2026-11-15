package bootstrap

import (
	"strings"
	"testing"

	"github.com/ste950/spestube/backend/internal/core/clock"
)

func env(values map[string]string) func(string) string {
	return func(k string) string { return values[k] }
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(env(nil))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 8080 || cfg.DBPath != "data/spestube.db" || cfg.ReleaseTime != (clock.TimeOfDay{Hour: 7}) ||
		cfg.YouTubeAPIKey != "" || cfg.Dev {
		t.Errorf("default = %+v", cfg)
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	cfg, err := Load(env(map[string]string{
		"SPESTUBE_PORT": "9090", "SPESTUBE_DB_PATH": "/data/spestube.db", "SPESTUBE_RELEASE_TIME": "19:30",
		"SPESTUBE_YT_API_KEY": "  chiave  ", "SPESTUBE_DEV": "true",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 9090 || cfg.DBPath != "/data/spestube.db" || cfg.ReleaseTime != (clock.TimeOfDay{Hour: 19, Minute: 30}) ||
		cfg.YouTubeAPIKey != "chiave" || !cfg.Dev {
		t.Errorf("config = %+v", cfg)
	}
}

func TestLoadRejectsMalformedValues(t *testing.T) {
	cases := map[string]string{
		"SPESTUBE_PORT":         "ottanta",
		"SPESTUBE_RELEASE_TIME": "25:00",
		"SPESTUBE_DEV":          "forse",
	}
	for name, value := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := Load(env(map[string]string{name: value}))
			// Il messaggio deve dire quale variabile correggere.
			if err == nil || !strings.Contains(err.Error(), name) {
				t.Errorf("errore = %v, atteso un errore che nomina %s", err, name)
			}
		})
	}
}
