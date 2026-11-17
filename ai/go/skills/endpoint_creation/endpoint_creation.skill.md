---
description: Skill per creare una risorsa HTTP — handler, DTO, registrazione sul ServeMux e test del contratto con httptest.
goal: Esporre un caso d'uso del dominio via HTTP con handler sottili, errori nel formato uniforme e un test che fissa status e codici.
---
# Skill: Creazione di un Endpoint

## Obiettivo

Un endpoint che traduce fra HTTP e dominio senza contenere decisioni, e il cui contratto —
status, corpo, codici di errore — è fissato da un test.

## Quando usarla

-   Un caso d'uso del dominio deve diventare raggiungibile dal client.
-   Un handler esistente contiene logica: la si sposta nel servizio, e l'handler si riscrive
    seguendo questi passi.

## Procedura

1.  **Verifica il contesto**
    -   Leggi [API HTTP](../../rules/http_api.rule.md) e
        [Gestione degli errori](../../rules/error_handling.rule.md).
    -   Il caso d'uso esiste nel `Service` del soggetto? Se no, prima quello
        ([Servizio di dominio](../../examples/service.example.md)).

2.  **I DTO**
    -   In `internal/api/<risorsa>/dto.go`: struct di richiesta e di risposta con i tag `json`, e
        `fromDomain`. Liste avvolte in un oggetto, tempi in UTC.

3.  **Gli handler**
    -   In `internal/api/<risorsa>/handler.go`: lettura con `httpx.DecodeJSON` e `r.PathValue`,
        chiamata al servizio, `httpx.WriteJSON` o `httpx.WriteError`. Nient'altro.
    -   [Vedi Esempio](../../examples/http_handler.example.md)

4.  **La registrazione**
    -   Il `Register(mux)` della risorsa con i pattern `"<METODO> /api/<risorsa>/…"`, chiamato da
        `internal/api/routes.go`. Se la risorsa ha bisogno di un servizio nuovo, lo si aggiunge a
        `api.Services` e lo si riempie nel composition root.

5.  **Il test**
    -   In `internal/api/<risorsa>/handler_test.go`: server `httptest` sulla risorsa vera e i
        fake del soggetto; il caso buono e ogni codice di errore che l'endpoint può produrre.
    -   [Vedi Esempio](../../examples/handler_test.example.md)

6.  **Il contratto**
    -   Un codice di errore nuovo si aggiunge all'elenco del `CLAUDE.md` del progetto, e
        l'endpoint al documento che descrive l'API (di norma il PRD).

## Vincoli

-   Nessun `database/sql`, nessun tipo di `data/` in `api/` — lo fa fallire il
    [test di architettura](../../examples/architecture_test.example.md).
-   Nessuno status code scelto a mano per un errore: lo sceglie `httpx.WriteError` dal `Kind`.
-   Una risorsa non importa un'altra risorsa.

## Verifica

-   In locale `gofmt -w .`; i check in CI.
-   Comando di lavoro: `go test ./internal/api/<risorsa>/ -v`.
