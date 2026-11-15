# API HTTP

`internal/api/` traduce fra HTTP e dominio, e basta. Un handler legge la richiesta, chiama un
servizio, scrive la risposta: se contiene un `if` che decide qualcosa di business, quel `if`
appartiene al dominio.

## 🧭 Routing con la libreria standard

Il `http.ServeMux` della libreria standard riconosce metodo e parametri di percorso:

```go
mux.HandleFunc("GET /api/items/{id}", h.get)
mux.HandleFunc("POST /api/items", h.create)
// nell'handler: r.PathValue("id")
```

Non serve un router esterno. `api/routes.go` monta ogni risorsa e applica i middleware;
ogni risorsa espone un `Register(mux *http.ServeMux)` e non sa niente delle altre.

## 📐 Convenzioni

-   **Prefisso `/api/`** per tutto ciò che è JSON. Il resto del percorso è libero per altro
    (file statici, `/healthz`).
-   **Nomi di risorsa al plurale**, in kebab-case: `/api/items`, `/api/read-later`.
-   **Status code espliciti**: `200` lettura, `201` creazione con il corpo creato, `204` nessun
    corpo. Gli errori li sceglie `httpx.WriteError` dal `Kind`, non l'handler.
-   **Liste avvolte in un oggetto** (`{"items": [...], "next": "…"}`), mai un array nudo: un
    campo in più domani non rompe il client.
-   **Paginazione a cursore** per le liste che crescono senza limite: il cursore è opaco per il
    client.
-   **Tempi in RFC 3339 UTC** nel JSON.

## 📦 DTO

Il corpo di richiesta e di risposta è una struct di `api/<risorsa>/dto.go` con i tag `json`,
costruita da una funzione `fromDomain(...)`. Le entità del dominio **non** hanno tag `json` e
non si serializzano direttamente: il formato del contratto API e la forma del dominio cambiano
per ragioni diverse.

La lettura del corpo passa da `httpx.DecodeJSON`, che limita la dimensione e rifiuta i campi
sconosciuti: un client che manda `"titel"` riceve un `400` invece di vedere ignorato il suo
campo. Vedi [Handler HTTP](../examples/http_handler.example.md).

## 🧱 Middleware

In `api/routes.go`, nell'ordine dall'esterno: recupero dai `panic` (diventa un `500` e una
riga di log), log della richiesta, autenticazione dove serve. Un middleware è una
`func(http.Handler) http.Handler`, senza librerie.

## 🚫 Divieti

-   `database/sql` o un tipo di `data/` in un handler.
-   `json.NewEncoder(w).Encode(...)` sparso: la scrittura passa da `httpx.WriteJSON`, che imposta
    l'header e lo status nell'ordine giusto.
-   Un handler che importa un'altra risorsa di `api/`.
-   Leggere la configurazione dentro un handler: arriva al costruttore.

## 🧭 Esempi

-   [Handler HTTP](../examples/http_handler.example.md)
-   [Errori HTTP](../examples/http_errors.example.md)
-   [Test di un handler](../examples/handler_test.example.md)
