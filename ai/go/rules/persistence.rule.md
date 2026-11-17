# Persistenza

Il default dello stack è **SQLite** attraverso `database/sql`, con un driver in Go puro. Un
backend a binario unico che gira su una macchina piccola non ha bisogno di un server di
database, e un file solo si copia, si fa il backup e si ispeziona con `sqlite3`.

## 🧩 Driver e apertura

-   Driver **`modernc.org/sqlite`** (nome `"sqlite"`): è SQLite tradotto in Go, quindi niente
    CGO e la build incrociata per ARM è un `GOARCH=arm64`. FTS5 e JSON sono compilati dentro.
-   Niente ORM e niente generatori di query: SQL scritto a mano nei repository, con i
    parametri `?`. Le query sono poche e stanno vicino a chi le usa.
-   Il database si apre **una volta**, in `bootstrap/`, con `core/sqlitedb.Open`, e il
    `*sql.DB` si passa ai costruttori dei repository. Vedi
    [Database SQLite](../examples/sqlite_db.example.md).

## ⚙️ Pragma obbligatori

Si impostano nella DSN, così valgono per **ogni** connessione del pool e non solo per la prima:

| Pragma | Valore | Perché |
|---|---|---|
| `journal_mode` | `WAL` | lettori e scrittore non si bloccano a vicenda |
| `busy_timeout` | `5000` | chi trova il lock aspetta fino a 5 s invece di fallire con `SQLITE_BUSY` |
| `foreign_keys` | `1` | in SQLite sono spente di default, e un vincolo dichiarato ma spento mente |
| `synchronous` | `NORMAL` | con WAL è sicuro contro i crash del processo e molto più veloce di `FULL` |

E `_txlock=immediate`: le transazioni prendono il lock di scrittura **all'inizio**. Con il
default (`DEFERRED`) due transazioni che prima leggono e poi scrivono possono finire in uno
stallo che `busy_timeout` non risolve.

## 🧾 Migrazioni

-   Lo schema vive in `data/migrations/` come file SQL numerati (`0001_init.sql`,
    `0002_<cosa>.sql`), inclusi nel binario con `embed.FS`. Vedi
    [Migrazioni](../examples/migration.example.md).
-   All'avvio si applicano quelle non ancora registrate in `schema_migrations`, ognuna nella
    sua transazione. Il binario che parte ha sempre lo schema che si aspetta.
-   **Una migrazione pubblicata non si modifica**: si scrive la successiva. Il file è già
    stato eseguito su un database che non si può rieseguire.
-   Lo schema è uno solo per tutto il modulo, anche se le tabelle appartengono a soggetti
    diversi: l'ordine delle migrazioni è globale, e spezzarlo per soggetto significa
    inventarsi un ordinamento fra cartelle.

## 🗂️ Schema

-   Chiavi primarie `INTEGER PRIMARY KEY`; tempi come `INTEGER` in secondi Unix UTC (o
    millisecondi, dichiarandolo): si confrontano e si ordinano senza parsing.
-   Un `UNIQUE` per ogni invariante di unicità del dominio — è il database a garantirla, non un
    controllo «se non esiste, inserisci» che due scrittori possono superare insieme.
-   Un indice per ogni colonna su cui si filtra o si ordina in una query frequente.
-   Gli insiemi chiusi di valori sono `TEXT` con un `CHECK (col IN (...))`.

## 🚫 Divieti

-   `*sql.Rows` che esce da un repository, o un `rows` non chiuso (`defer rows.Close()`, e
    `rows.Err()` dopo il ciclo).
-   Query costruite concatenando valori: solo parametri.
-   Leggere tutto e filtrare in Go ciò che una `WHERE` sa filtrare.
-   Un secondo `sql.Open` sullo stesso file da un altro punto del codice.

## 🧭 Esempi

-   [Database SQLite](../examples/sqlite_db.example.md)
-   [Migrazioni](../examples/migration.example.md)
-   [Implementazione di repository](../examples/repository_impl.example.md)
