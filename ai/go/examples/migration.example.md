# Migrazioni

Lo schema vive in `internal/data/migrations/`, un file SQL per migrazione, incluso nel binario.
Il motore che le applica è in [Database SQLite](./sqlite_db.example.md); le regole sullo schema
in [Persistenza](../rules/persistence.rule.md).

```go
// backend/internal/data/migrations/migrations.go

// Package migrations contiene lo schema del database, un file per migrazione.
package migrations

import "embed"

// FS entra nel binario: chi lo avvia ha sempre lo schema che il codice si aspetta.
//
//go:embed *.sql
var FS embed.FS
```

```sql
-- backend/internal/data/migrations/0001_init.sql

CREATE TABLE items (
    id          INTEGER PRIMARY KEY,
    url         TEXT    NOT NULL UNIQUE,
    title       TEXT    NOT NULL,
    status      TEXT    NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'archived')),
    created_at  INTEGER NOT NULL,              -- secondi Unix, UTC
    refreshed_at INTEGER                       -- NULL finché la sorgente non risponde
);

CREATE INDEX items_status_created ON items (status, created_at DESC);
```

Una migrazione successiva **non riscrive** la `0001`: aggiunge.

```sql
-- backend/internal/data/migrations/0002_items_note.sql

ALTER TABLE items ADD COLUMN note TEXT NOT NULL DEFAULT '';
```

SQLite ha un `ALTER TABLE` limitato: per cambiare il tipo o un vincolo di una colonna si crea
la tabella nuova, si copiano le righe, si elimina la vecchia e si rinomina — tutto nello
stesso file, che il motore esegue in una transazione.
