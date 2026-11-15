# Verifica delle Modifiche

## 🎯 Il cancello non è su questa macchina

I check di questo stack girano **nella CI**, sul [job di verifica](../examples/ci_job.example.yml).
In locale se ne esegue uno solo, il formattatore, e non è un cancello: il perché — e perché
l'agente apre la pull request in draft invece di dichiarare che la verifica passa — sta in
Dove gira la verifica.

## 📋 I check dello stack

Dalla cartella `backend/`:

| Famiglia | Comando | Dove gira |
|---|---|---|
| Formattazione | `test -z "$(gofmt -l .)"` | CI — **e in locale** nella forma che riscrive, `gofmt -w .` |
| Moduli allineati | `go mod tidy -diff` | solo CI |
| Analisi statica | `go vet ./...` | solo CI |
| Test (architettura compresa) | `go test -race ./...` | solo CI |
| Build per il bersaglio | `CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ./...` | solo CI |

-   **`go mod tidy -diff`** fallisce se `go.mod` o `go.sum` non corrispondono agli import: una
    dipendenza aggiunta al codice e non al modulo, o rimasta nel modulo dopo l'ultimo uso.
-   **`-race`** costa poco su una suite piccola e trova la classe di errori che nessun altro
    check vede: due goroutine sullo stesso campo senza sincronizzazione. Vedi
    [Concorrenza](./concurrency.rule.md).
-   **La build incrociata** è la «build di verifica»: compila per l'architettura e con il
    `CGO_ENABLED` della [distribuzione](./deployment.rule.md). Una dipendenza che chiede CGO
    passa i test sul runner e fallisce solo qui — che è il punto.

Nessuno di questi si lascia indietro: un avviso di `go vet` non passa «perché il test passa».
Solo, a dirlo è il runner.

## ✅ L'unico comando locale

```bash
gofmt -w .
```

Sta in locale perché costa un istante, non compila niente e la sua violazione è certa. Niente
`go vet`, niente `go test ./...` «per sicurezza».

## 📦 Quando cambiano le dipendenze

```bash
go get <modulo>@<versione>
go mod tidy
```

`go.mod` e `go.sum` si committano insieme. Una dipendenza nuova si giustifica nel corpo della
PR: lo stack ne vuole [poche](./project_structure.rule.md#-poche-dipendenze).

## ⚙️ Comandi di lavoro

Servono a scrivere il codice, non a dichiararlo verificato.

```bash
go run ./cmd/<binario>                       # avvio in sviluppo
go test ./internal/domain/<soggetto>/ -run TestNome -v   # il test che si sta scrivendo
go doc <package>.<Simbolo>                   # la firma vera, invece di ricordarla
```

## 🚧 Verifica dei confini fra layer

Il compilatore non conosce le regole di dipendenza di questo framework: la garanzia è il
[test di architettura](../examples/architecture_test.example.md), che gira dentro
`go test ./...` come qualsiasi altro test. Mentre si lavora resta comodo un `grep`:

```bash
grep -rn '"database/sql"\|"net/http"' internal/domain    # deve essere vuoto
```

## 🔗 Correlate

-   Dove gira la verifica: perché il cancello è uno
    solo, e cosa resta in locale.
-   Quando la CI è rossa: cosa si fa di un check
    rosso — si ripara la causa, mai il check.
