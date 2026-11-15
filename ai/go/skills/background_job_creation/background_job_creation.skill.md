---
description: Skill per creare un lavoro periodico o in background — il lavoro come metodo del servizio, il ciclo in core/schedule, la goroutine posseduta dal composition root.
goal: Aggiungere un job che si ferma con il processo, non perde né raddoppia lavoro a un riavvio e si prova senza aspettare l'orologio.
---
# Skill: Creazione di un Job in Background

## Obiettivo

Un lavoro che gira da solo — un aggiornamento periodico, un invio a orari fissi, una pulizia —
senza goroutine orfane e senza stato che vive solo in memoria.

## Quando usarla

-   Qualcosa deve succedere senza una richiesta HTTP che lo scateni.
-   Un handler sta per avviare una goroutine che sopravvive alla richiesta.

## Procedura

1.  **Verifica il contesto**
    -   Leggi [Concorrenza](../../rules/concurrency.rule.md).

2.  **Il lavoro, nel dominio**
    -   Un metodo del `Service` con firma `func(ctx context.Context) error`, che fa **un giro**
        e ritorna. Riceve l'ora dall'orologio iniettato.
    -   Lo stato che deve sopravvivere a un riavvio (l'ultimo invio, la prossima scadenza) sta
        nel database, non in un campo del servizio.
    -   [Vedi Esempio](../../examples/service.example.md)

3.  **Il ciclo**
    -   `schedule.Every` di `core/` se il progetto non ce l'ha ancora.
    -   [Vedi Esempio](../../examples/background_worker.example.md)

4.  **L'avvio**
    -   In `bootstrap/run.go`, `wg.Go(func() { schedule.Every(ctx, "<nome>", <intervallo>,
        svc.<Metodo>) })`, con l'intervallo dalla `Config`.

5.  **Il parallelismo, se serve**
    -   Dentro il metodo, con semaforo e `WaitGroup`, e ritorno solo a goroutine finite.

6.  **Il test**
    -   Si chiama il metodo direttamente, con l'orologio fisso e i fake: un giro, verifiche
        sullo stato dei fake. Il ciclo non si testa con il tempo vero.

## Vincoli

-   Nessun `go func()` fuori da `bootstrap/` per un lavoro di lunga vita.
-   Nessun `time.Sleep` nel lavoro: le attese stanno nel ciclo, che ascolta il contesto.
-   Un errore di un giro si logga e il ciclo continua; un errore all'avvio ferma il processo.

## Verifica

-   `go test -race` in CI copre l'accesso concorrente allo stato.
-   In locale `gofmt -w .`.
