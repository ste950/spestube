# Distribuzione

Il prodotto di questo stack è **un file**: un binario statico che contiene il server, i job, le
migrazioni e — se il progetto lo vuole — i file statici del client. Si copia sulla macchina di
destinazione e lo avvia systemd. Niente runtime da installare, niente container obbligatorio.

## 🔨 La build

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
  go build -trimpath -ldflags "-s -w -X main.version=<versione>" -o dist/<binario> ./cmd/<binario>
```

-   **`CGO_ENABLED=0`** è ciò che rende il binario statico e la build incrociata banale. Ogni
    dipendenza deve rispettarlo: è il motivo per cui il driver SQLite è in Go puro
    ([Persistenza](./persistence.rule.md)), e la CI lo verifica a ogni PR
    ([Verifica](./verification.rule.md)).
-   **`-trimpath`** toglie dal binario i percorsi della macchina che l'ha compilato.
-   **`-s -w`** tolgono simboli e informazioni di debug: un terzo del peso in meno.
-   **La versione** entra con `-X`. Vedi [Versione](../examples/versioning.example.md).

La build non avviene sulla macchina di destinazione se questa è piccola: la fa la CI, o la
macchina di sviluppo, e il binario si copia.

## 🐧 Il servizio

Un'unità systemd per binario, versionata nel repo (di norma `deploy/<binario>.service`). Vedi
[Unit systemd](../examples/systemd_unit.example.md). Vincoli:

-   **Limiti di memoria espliciti**: `MemoryMax=` nell'unità e `GOMEMLIMIT` nell'ambiente,
    un po' sotto. Il primo è il tetto che il kernel fa rispettare; il secondo dice al garbage
    collector di Go di lavorare di più prima di arrivarci, invece di farsi uccidere.
-   **Utente dedicato senza privilegi** (`DynamicUser=yes`) e i dati in `StateDirectory=`:
    systemd crea la cartella con i permessi giusti, il binario la riceve in
    `$STATE_DIRECTORY`.
-   **Configurazione con `EnvironmentFile=`**, fuori dal repo
    ([Configurazione](./configuration.rule.md)).
-   **Arresto con `SIGTERM`**, che il binario gestisce ([Concorrenza](./concurrency.rule.md)),
    e `Restart=on-failure`.
-   **Indurimento di base**: `NoNewPrivileges`, `ProtectSystem=strict`, `ProtectHome`,
    `PrivateTmp`.

## 🩺 Salute

Il binario espone `GET /healthz` senza autenticazione: `200` e la versione se il database
risponde. È ciò che si interroga dopo un aggiornamento per sapere che il binario giusto è
partito.

## 🧭 Esempi

-   [Unit systemd](../examples/systemd_unit.example.md)
-   [Versione](../examples/versioning.example.md)
