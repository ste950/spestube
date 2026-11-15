# Commenti nel Codice

Un commento è codice che nessun compilatore controlla: nessuno lo aggiorna quando la riga
sotto cambia, e il giorno in cui mente costa più di quanto sia mai valso. Se ne scrivono
pochi, e solo quelli che nessun nome e nessuna firma potrebbero dire al posto loro.

È lo stesso principio degli altri stack
(Python); cambia la convenzione di Go sui commenti
di documentazione.

## ✍️ Il commento spiega il *perché*, mai il *cosa*

```go
// ❌ dice quello che si vede
// Incrementa il contatore degli errori.
f.ConsecutiveErrors++

// ✅ dice quello che non si vede
// Un 304 non è un successo che azzera il contatore: dice solo che il server è vivo,
// non che il feed è leggibile.
if status != http.StatusNotModified {
	f.ConsecutiveErrors = 0
}
```

Si commenta una **decisione non ovvia**, un **caso limite**, un **vincolo esterno**, un
**numero scelto** (un timeout, una soglia), un `//nolint` o un errore ignorato.

## 📚 I commenti di documentazione

Go ha una convenzione che gli strumenti leggono: il commento sopra un identificatore
**esportato** inizia con il suo nome ed è una frase.

```go
// Match riporta se la regola si applica all'articolo. Una regola senza condizioni
// non si applica a niente: è il default prudente per una regola appena creata.
func (r Rule) Match(a Article) bool {
```

-   Si scrive dove aggiunge qualcosa alla firma — un'assenza che non è un errore, un effetto
    collaterale, una precondizione. Su `func (r *Repo) Get(ctx, id) (Item, error)` un
    «Get restituisce l'item» è rumore.
-   Ogni package ha **una** riga di documentazione (`// Package catalog …`), in un file solo.
-   Gli identificatori non esportati si commentano solo per il perché.

## 🚫 Divieti

-   **Riferimenti a task, issue o PR** nel codice (`// vedi #42`, `// TODO(issue 17)`): il
    numero invecchia, il codice resta. Il perché si scrive per intero, o va nel messaggio di
    commit.
-   **Codice commentato.** È nel controllo di versione.
-   **Commenti di sezione** (`// ===== HANDLERS =====`): se un file ne ha bisogno, sono due
    file.
