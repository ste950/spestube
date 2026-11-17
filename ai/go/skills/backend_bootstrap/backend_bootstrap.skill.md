---
description: Skill per creare da zero lo scheletro del backend Go — modulo, layout a livelli, configurazione, database, server con /healthz, shutdown ordinato, test di architettura e job di CI.
goal: Partire con un binario che compila per il bersaglio, si avvia, risponde su /healthz e si ferma bene, con i confini fra livelli già difesi da un test.
---
# Skill: Bootstrap del Backend

## Obiettivo

Uno scheletro su cui ogni task successivo aggiunge un pezzo senza dover decidere di nuovo dove
va: il layout, il cablaggio, la configurazione e la verifica ci sono già.

## Quando usarla

-   Il repo non ha ancora `backend/go.mod`.
-   Il backend esiste ma non ha il layout di
    [Struttura del progetto](../../rules/project_structure.rule.md): prima di aggiungere
    funzionalità, si riallinea.

## Procedura

1.  **Il modulo**
    -   `go mod init <path del modulo>` in `backend/`, con il path dichiarato nel `CLAUDE.md`
        del progetto. Direttiva `go` alla versione stabile corrente.
    -   [Vedi Esempio](../../examples/go_mod.example.md)

2.  **Il layout vuoto**
    -   `cmd/<binario>/`, `internal/{bootstrap,core,domain,data,api}/`, `architecture/`.
    -   Una cartella vuota non si committa: nasce con il primo file.

3.  **Il nucleo di `core/`**
    -   `apperr` ([esempio](../../examples/app_error.example.md)), `httpx`
        ([esempio](../../examples/http_errors.example.md)), `sqlitedb`
        ([esempio](../../examples/sqlite_db.example.md)).

4.  **Lo schema iniziale**
    -   `internal/data/migrations/` con `migrations.go` e `0001_init.sql`.
    -   [Vedi Esempio](../../examples/migration.example.md)

5.  **Configurazione e composition root**
    -   `bootstrap/config.go` con il prefisso del progetto
        ([esempio](../../examples/config.example.md)), `bootstrap/run.go` e
        `cmd/<binario>/main.go` ([esempio](../../examples/main_and_shutdown.example.md)).
    -   `api/routes.go` con `GET /healthz` e i middleware di recover e log
        ([esempio](../../examples/http_handler.example.md)).

6.  **Il test di architettura**
    -   `architecture/architecture_test.go`, così il primo package che sbaglia livello è già
        rosso. [Vedi Esempio](../../examples/architecture_test.example.md)

7.  **CI e versione**
    -   Il job di [CI](../../examples/ci_job.example.yml) — come workflow o come job di quello
        esistente — e `backend/version.txt`
        ([esempio](../../examples/versioning.example.md)).

8.  **La distribuzione, se il progetto la chiede già**
    -   `deploy/<binario>.service` e l'ambiente d'esempio
        ([esempio](../../examples/systemd_unit.example.md)).

## Vincoli

-   Nessuna dipendenza oltre al driver SQLite: il resto è libreria standard.
-   Nessun soggetto di dominio inventato per «riempire» lo scheletro: il primo soggetto arriva
    con il primo task che ne ha bisogno.
-   `.gitignore` del repo: il database locale (`*.db`, `*.db-wal`, `*.db-shm`), `dist/` e i
    file d'ambiente reali.

## Verifica

-   In locale: `gofmt -w .`. Il resto lo fa la CI
    ([Verifica delle modifiche](../../rules/verification.rule.md)).
-   Comando di lavoro per vedere che parte: `go run ./cmd/<binario>` e
    `curl localhost:<porta>/healthz` → `dev`.
