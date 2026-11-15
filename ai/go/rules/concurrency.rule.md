# Concorrenza

Una goroutine costa poco da avviare e molto da dimenticare: una che non si ferma tiene vivo il
processo allo shutdown, una che scrive senza sincronizzazione corrompe uno stato in modo che
nessun test ripete. Questa regola fissa chi può avviarle e come finiscono.

## 👤 Ogni goroutine ha un proprietario

-   **Le goroutine di lunga vita** — il server HTTP, uno scheduler, un worker — le avvia solo
    `bootstrap.Run`, e solo lì si aspetta che finiscano. Un package di `domain/` o `data/`
    espone una funzione bloccante `Run(ctx)`, non avvia goroutine per conto suo.
-   **Le goroutine di breve vita** dentro una funzione (fan-out su più feed, più richieste)
    finiscono **prima che la funzione ritorni**: si aspettano con un `sync.WaitGroup`
    (`wg.Go(...)`) e il loro numero è limitato da un semaforo (un canale bufferizzato). Una
    funzione che ritorna lasciando goroutine in volo è un bug anche quando oggi non si vede.

## ⏹️ Il `context` è il segnale di stop

-   `main` crea il contesto radice con `signal.NotifyContext(ctx, SIGINT, SIGTERM)`: è l'unico
    modo in cui il processo viene fermato, e systemd manda `SIGTERM`.
-   Ogni funzione che fa I/O o che può durare prende un `ctx context.Context` come **primo
    parametro**, e lo passa a ogni chiamata sotto (`QueryContext`, `http.NewRequestWithContext`).
-   Un ciclo di lunga vita controlla `ctx.Done()` a ogni giro, dentro un `select`.
-   Nessun `context.Background()` fuori da `main` e dai test: un contesto nuovo a metà catena
    taglia il segnale di stop.

## 🛑 Shutdown ordinato

All'arrivo del segnale, nell'ordine:

1.  il server HTTP smette di accettare connessioni e finisce quelle in corso
    (`srv.Shutdown` con una scadenza);
2.  i job vedono il contesto cancellato e ritornano dal loro `Run`;
3.  `Run` aspetta tutti, poi chiude il database.

Il database si chiude **per ultimo**: chiuderlo prima vuol dire far fallire l'ultima scrittura
di un job che stava finendo bene. Vedi [Main e shutdown](../examples/main_and_shutdown.example.md).

## 🔒 Stato condiviso

-   Il primo strumento è **non condividere**: un worker possiede i suoi dati, e gli altri gli
    parlano con un canale o chiamando il database.
-   Quando lo stato condiviso serve (una cache, un contatore), sta dietro un `sync.Mutex` nella
    stessa struct, con i campi protetti dichiarati subito sotto. Mai un mutex esportato.
-   `go test -race` gira in CI su tutta la suite: è la rete per ciò che questa regola non ha
    previsto ([Verifica](./verification.rule.md)).

## 🗄️ SQLite e la concorrenza

SQLite accetta **uno scrittore alla volta**. Con WAL i lettori non bloccano lo scrittore e
viceversa, e con `busy_timeout` uno scrittore che trova il lock aspetta invece di fallire. Il
dettaglio — pragma e transazioni `IMMEDIATE` — sta in [Persistenza](./persistence.rule.md).
Qui conta la conseguenza: parallelizzare le **scritture** non accelera niente. Si parallelizza
l'I/O di rete (scaricare dieci feed), poi si scrive.

## 🧭 Esempi

-   [Main e shutdown](../examples/main_and_shutdown.example.md)
-   [Job in background](../examples/background_worker.example.md)
