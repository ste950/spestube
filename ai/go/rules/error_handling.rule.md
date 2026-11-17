# Gestione degli Errori

In Go un errore è un valore restituito, non un'eccezione: ogni livello lo riceve, decide se
aggiungere contesto e lo passa avanti. Questa regola fissa **cosa** viaggia — un errore con un
genere e un codice stabile — e **dove** diventa una risposta HTTP: in un posto solo.

## 🧱 L'errore applicativo

`core/apperr` definisce un tipo solo, con tre informazioni:

| Campo | A cosa serve |
|---|---|
| `Kind` | il **genere** (`NotFound`, `Invalid`, `Conflict`, `Unauthorized`, `Unavailable`, `Internal`): decide lo status HTTP e se si scrive nel log |
| `Code` | il **codice stabile** del contratto API (`item_not_found`, `invalid_url`): il client lo legge, non lo si rinomina |
| `Msg` | il messaggio per chi legge, in chiaro |

Più l'errore originale, se c'è, raggiungibile con `errors.Unwrap`. Vedi
[Errore applicativo](../examples/app_error.example.md).

L'elenco dei codici è parte del **contratto API** del progetto: lo dichiara il suo
`CLAUDE.md`, e un codice nuovo si aggiunge lì nella stessa PR che lo introduce.

## ➡️ Chi crea, chi avvolge, chi traduce

-   **Il dominio crea** gli errori che hanno un significato per l'utente:
    `apperr.NotFound("item_not_found", "…")`, `apperr.Invalid("invalid_url", "…")`.
-   **Il data layer traduce** gli errori della sua tecnologia: `sql.ErrNoRows` diventa
    `NotFound` nel repository, un timeout HTTP di una sorgente diventa `Unavailable`. Un
    errore di `database/sql` non esce mai da `data/` così com'è.
-   **Chi passa avanti avvolge** con `fmt.Errorf("aggiornamento del feed %d: %w", id, err)`:
    `%w`, sempre — `%v` spezza la catena, e `errors.As` non trova più il `Kind`.
-   **Un solo punto traduce in HTTP**: `httpx.WriteError`. Gli handler non scelgono status
    code per gli errori. Vedi [Errori HTTP](../examples/http_errors.example.md).

Un errore che non è un `*apperr.Error` è per definizione `Internal`: si scrive nel log con
tutto il contesto e al client arriva un messaggio generico, mai il testo originale.

## 🚫 Divieti

-   **`panic` per un errore previsto.** `panic` è per gli invarianti violati — un bug — e il
    middleware di recover lo trasforma in un 500. Un input sbagliato, una riga assente, una
    rete giù sono errori.
-   **Ignorare un errore** con `_ =` senza un commento che dica perché è sicuro.
-   **Loggare e restituire.** Lo stesso errore finisce nel log due volte e nessuno sa quale
    sia l'origine. Si logga **dove si smette di restituire**: nel writer HTTP, nel ciclo di un
    job.
-   **Confrontare i messaggi** (`err.Error() == "…"`). Si usano `errors.Is` sugli errori
    sentinella e `apperr.KindOf` / `errors.As` sul tipo.
-   **`http.Error` dentro un handler** per un errore di dominio: produce un corpo in testo
    che il client non sa leggere.

## 🧭 Esempi

-   [Errore applicativo](../examples/app_error.example.md)
-   [Errori HTTP](../examples/http_errors.example.md)
-   [Implementazione di repository](../examples/repository_impl.example.md): `sql.ErrNoRows`
    tradotto in `NotFound`.
