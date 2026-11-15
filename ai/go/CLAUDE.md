# Framework AI — Backend Go a Binario Unico

Questo file è il punto di ingresso per gli agenti AI: elenca le regole trasversali, tutte le
skill disponibili e gli esempi di codice condivisi, con i loro percorsi e una breve
descrizione. Nessun contenuto sostanziale va scritto qui — solo link.

Il framework descrive un backend Go che si distribuisce come **un binario statico**: `domain/`
puro organizzato per soggetto, `data/` per le implementazioni (SQLite in Go puro, sorgenti
esterne), `api/` come sola presentation su `net/http`, `bootstrap/` come composition root
scritto a mano. Errori come valori con un genere e un codice stabile, goroutine con un
proprietario, libreria standard prima di tutto. È indipendente da qualsiasi progetto: nessuna
regola, skill o esempio deve citare un'applicazione specifica.

## 🔌 Come è agganciato a questo repo

Questa cartella è una copia dello stack, tenuta dentro il repo: il `CLAUDE.md` alla radice la
importa con `@ai/go/CLAUDE.md`. Non esiste una cartella sorella da clonare. Origine e
modalità di aggiornamento sono in [`ai/README.md`](../README.md).

Le regole qui dentro assumono il layout `backend/` con `cmd/<binario>/` e `internal/`: il
modulo Go vive in `backend/` alla radice del repo. Tutto ciò che è specifico di un progetto
(path del modulo, soggetti del dominio, prefisso delle variabili d'ambiente, codici di errore
del contratto API, servizi esterni, vincoli di memoria) resta nel `CLAUDE.md` del progetto.

## 📖 Regole Trasversali

-   [Struttura del Progetto](./rules/project_structure.rule.md): organizzazione di `backend/`, regole di dipendenza tra livelli, poche dipendenze e corrispondenze con lo stack Python.
-   [Verifica delle Modifiche](./rules/verification.rule.md): i check dello stack e dove gira ciascuno — in CI tutti, in locale il solo `gofmt`; i comandi di lavoro e cosa fare quando cambiano le dipendenze.
-   [Commenti nel Codice](./rules/code_comments.rule.md): il perché e non il cosa; la convenzione dei commenti di documentazione; nessun riferimento a task, issue o PR nel codice.

### Vincoli Trasversali

-   [Gestione degli Errori](./rules/error_handling.rule.md): `apperr` con `Kind` e codice stabile, chi crea, chi avvolge con `%w` e chi traduce in HTTP — in un punto solo.
-   [Configurazione](./rules/configuration.rule.md): una `Config` caricata una volta dall'ambiente con un prefisso, file `KEY=VALUE`, validazione all'avvio.
-   [Concorrenza](./rules/concurrency.rule.md): ogni goroutine ha un proprietario, il `context` come segnale di stop, shutdown ordinato, SQLite a scrittore unico.
-   [Persistenza](./rules/persistence.rule.md): `database/sql` con `modernc.org/sqlite`, pragma obbligatori, migrazioni SQL numerate in `embed.FS`, regole sullo schema.
-   [API HTTP](./rules/http_api.rule.md): `ServeMux` della libreria standard, convenzioni di rotte e status, DTO separati dalle entità, middleware.
-   [Test](./rules/testing.rule.md): tre livelli con la sola libreria standard, table-driven, fake al posto dei mock, tempo iniettato.
-   [Distribuzione](./rules/deployment.rule.md): build statica incrociata, servizio systemd con limiti di memoria e utente dedicato, `/healthz`.

## 🧠 Lista delle Skill

-   **`skills/backend_bootstrap/backend_bootstrap.skill.md`**: Skill per creare da zero lo scheletro del backend: modulo, layout, configurazione, database, `/healthz`, shutdown, test di architettura e CI.
-   **`skills/endpoint_creation/endpoint_creation.skill.md`**: Skill per creare una risorsa HTTP — DTO, handler, registrazione e test del contratto.
-   **`skills/repository_creation/repository_creation.skill.md`**: Skill per il flusso completo entità → interfaccia → migrazione → implementazione SQL → fake → cablaggio.
-   **`skills/external_source_integration/external_source_integration.skill.md`**: Skill per isolare un servizio o un formato di terze parti dietro una sorgente di `data/`.
-   **`skills/background_job_creation/background_job_creation.skill.md`**: Skill per creare un lavoro periodico con il ciclo in `core/` e la goroutine posseduta dal composition root.
-   **`skills/test_creation/test_creation.skill.md`**: Skill per scrivere i test ai tre livelli e il test di architettura.

## 📦 Esempi Condivisi

Snippet di codice concreti, riusabili da tutte le skill. Sono estratti da un modulo che compila
e i cui test passano: un progetto nuovo parte da lì.

### Infrastruttura

-   [go.mod](./examples/go_mod.example.md): path del modulo, direttiva `go`, l'unica dipendenza diretta e i comandi dei moduli.
-   [Errore applicativo](./examples/app_error.example.md): `apperr.Error` con `Kind`, codice stabile e causa; `KindOf`.
-   [Errori HTTP e JSON](./examples/http_errors.example.md): `httpx.WriteJSON`, `DecodeJSON` e `WriteError`, l'unico posto che sceglie uno status per un errore.
-   [Config](./examples/config.example.md): caricamento dall'ambiente con prefisso, default, validazione e file `KEY=VALUE` di sviluppo.
-   [Main e shutdown](./examples/main_and_shutdown.example.md): `main`, composition root, avvio di server e job, arresto ordinato.
-   [Database SQLite](./examples/sqlite_db.example.md): apertura con i pragma nella DSN e motore delle migrazioni.
-   [Migrazioni](./examples/migration.example.md): file SQL numerati in `embed.FS`, la prima tabella e come si aggiunge la seconda migrazione.

### Livelli e risorse

-   [Entità di Dominio](./examples/domain_entity.example.md): struct per valore, insieme chiuso con `Valid()`, logica derivata con l'ora iniettata.
-   [Contratto Repository](./examples/domain_repository.example.md): interfacce piccole nel dominio e la convenzione sull'assenza.
-   [Servizio di Dominio](./examples/service.example.md): i casi d'uso come metodi di un `Service`, validazione ed errori del dominio.
-   [Implementazione di Repository](./examples/repository_impl.example.md): mapper riga → entità e repository SQL che traduce `sql.ErrNoRows`.
-   [Handler HTTP](./examples/http_handler.example.md): DTO, handler sottili, `Register` sul `ServeMux`, middleware di recover e log.
-   [Job in Background](./examples/background_worker.example.md): `schedule.Every`, l'avvio nel composition root, intervalli per elemento, orari fissi e parallelismo limitato.
-   [Sorgente Esterna](./examples/external_source.example.md): interfaccia, payload e adattatore HTTP in `sources/`, composti da un repository.

### Test

-   [Test Table-Driven](./examples/table_test.example.md): fake del soggetto, test del servizio a tabella e test del repository su SQLite vero.
-   [Test di un Handler](./examples/handler_test.example.md): `httptest` sulla risorsa vera, status e codici di errore.
-   [Test di Architettura](./examples/architecture_test.example.md): le regole di dipendenza fra livelli lette con `go/parser`, al posto dei `grep` a mano.

### CI, release e distribuzione

-   [Job di CI](./examples/ci_job.example.yml): la verifica su macchina pulita, con gli stessi comandi di [Verifica delle Modifiche](./rules/verification.rule.md). È un `.yml` e non un `.example.md` perché si copia dentro `.github/` così com'è.
-   [Versione](./examples/versioning.example.md): `version.txt` come unico posto del numero, `-ldflags -X` per portarlo nel binario, `release-please` con `release-type: simple`.
-   [Unit systemd](./examples/systemd_unit.example.md): il servizio con `EnvironmentFile`, `StateDirectory`, `MemoryMax` + `GOMEMLIMIT` e l'indurimento di base.

