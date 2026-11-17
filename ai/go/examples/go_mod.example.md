# go.mod

Il modulo del backend vive in `backend/`. Il path del modulo lo sceglie il progetto: negli
esempi del framework è `example.com/backend`, e ogni import interno parte da lì
(`example.com/backend/internal/domain/catalog`).

```
// backend/go.mod

module example.com/backend

go 1.27

require modernc.org/sqlite v1.60.1

require (
	// dipendenze indirette: le scrive `go mod tidy`, non si toccano a mano
)
```

-   **La direttiva `go`** è la versione minima del linguaggio, ed è anche la fonte da cui la CI
    sceglie la toolchain (`go-version-file: backend/go.mod`): la versione si scrive in un
    posto solo.
-   **Una sola dipendenza diretta** nel caso tipico: il driver SQLite. Tutto il resto — server
    HTTP, JSON, log, test — è libreria standard
    ([Struttura del progetto](../rules/project_structure.rule.md#-poche-dipendenze)).
-   **`go.sum` si committa.** È ciò che rende la build riproducibile e verificabile.

Comandi:

```bash
go mod init example.com/backend     # una volta, alla nascita
go get modernc.org/sqlite@latest    # aggiunge o aggiorna una dipendenza
go mod tidy                         # allinea go.mod e go.sum agli import reali
```
