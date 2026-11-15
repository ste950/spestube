# Struttura del Progetto

## 📁 Struttura delle Directory Principali

-   **`backend/`**: il modulo Go del backend, con il suo `go.mod`. Tutti i comandi dello stack
    si lanciano da qui. Il path del modulo lo sceglie il progetto (negli esempi è
    `example.com/backend`) e lo dichiara nel suo `CLAUDE.md`.
-   **`CLAUDE.md`**: punto di ingresso per gli agenti AI. Importa questo framework e dichiara
    le specificità del progetto (soggetti del dominio, prefisso delle variabili d'ambiente,
    codici di errore, servizi esterni). Regole, skill ed esempi non vivono nel repo del
    progetto.

## 📂 Struttura di `backend/`

```
backend/
├── go.mod                         # path del modulo, versione di Go, dipendenze dirette
├── go.sum                         # versionato: rende riproducibile la build
│
├── cmd/
│   └── <binario>/
│       └── main.go                # solo: legge la config, chiama bootstrap.Run, esce
│
├── internal/
│   ├── bootstrap/                 # 🆕 COMPOSITION ROOT
│   │   ├── config.go              # Config caricata una volta dall'ambiente
│   │   ├── app.go                 # costruisce repository, servizi, handler e job
│   │   └── run.go                 # avvia server e job, gestisce lo shutdown
│   │
│   ├── core/                      # 🔧 INFRASTRUTTURA AGNOSTICA AL DOMINIO
│   │   ├── apperr/                # errore applicativo con Kind e codice stabile
│   │   ├── httpx/                 # JSON in/out, scrittura uniforme degli errori
│   │   └── sqlitedb/              # apertura del database, pragma, migrazioni
│   │
│   ├── domain/                    # 🧠 DOMAIN LAYER (per soggetto)
│   │   └── <soggetto>/
│   │       ├── <entità>.go        # tipi del dominio, logica derivata come metodi
│   │       ├── repository.go      # interfacce dei soli repository
│   │       └── service.go         # i casi d'uso, come metodi di un Service
│   │
│   ├── data/                      # 🔌 DATA LAYER (per soggetto)
│   │   ├── migrations/            # NNNN_<nome>.sql + embed.FS: lo schema è uno solo
│   │   └── <soggetto>/
│   │       ├── <nome>_repository.go   # implementazione SQL dell'interfaccia di domain/
│   │       ├── mapper.go          # riga o payload → entità
│   │       └── sources/           # fonti esterne: interfaccia, payload e adattatore
│   │
│   └── api/                       # 🧩 PRESENTATION (una cartella per risorsa)
│       ├── routes.go              # monta le risorse sul ServeMux, middleware
│       └── <risorsa>/
│           ├── handler.go         # gli handler HTTP
│           └── dto.go             # corpi di richiesta e risposta
│
└── architecture/
    └── architecture_test.go       # le regole di dipendenza qui sotto, eseguibili
```

I test unitari stanno **accanto al codice** (`service_test.go` vicino a `service.go`), come
vuole la toolchain. Fa eccezione solo `architecture/`, che non testa un package ma il modulo.

### Perché `internal/`

Tutto il codice applicativo sta sotto `internal/`: il compilatore impedisce a qualsiasi modulo
esterno di importarlo. Un backend non è una libreria, e `internal/` lo dice in modo che nessuno
debba ricordarselo. Fuori da `internal/` resta solo `cmd/`.

### Soggetti e risorse sono cose diverse

-   Un **soggetto** è un concetto di business (es. `catalog`, `accounts`). Vive in
    `domain/<soggetto>/` (tipi, interfacce, servizio) e `data/<soggetto>/` (implementazioni).
-   Una **risorsa** è un gruppo di endpoint sotto un prefisso di rotta. Vive in
    `api/<risorsa>/` e non contiene logica di dominio.
-   Spesso hanno lo stesso nome, ma non è un obbligo: una risorsa può leggere da più
    soggetti, e un soggetto può non avere nessun endpoint proprio — un job in background, per
    esempio, è un consumatore del dominio quanto un handler.

### Dettagli sui Livelli

#### `internal/bootstrap/` — Composition root
L'unico package che può importare tutto. Costruisce le implementazioni di `data/`, le passa ai
servizi di `domain/` come interfacce, consegna i servizi agli handler e ai job, e possiede il
ciclo di vita del processo. Non esiste un container di dependency injection: il cablaggio è
codice Go scritto a mano, e il compilatore controlla che ogni implementazione soddisfi
l'interfaccia che riceve. Vedi [Main e shutdown](../examples/main_and_shutdown.example.md).

#### `internal/core/` — Infrastruttura riutilizzabile
Errore applicativo, helper HTTP, apertura del database. Nulla che parli di un concetto di
business: se un file di `core/` nomina un soggetto, sta nel posto sbagliato.

#### `internal/domain/` e `internal/data/` — Layer di dominio
Il dominio è organizzato per soggetto ed è puro: niente `database/sql`, niente `net/http`,
niente driver. Le **interfacce dei repository** stanno qui, perché è il dominio a consumarle —
è l'idioma Go «l'interfaccia la definisce chi la usa». Vedi
[Contratto repository](../examples/domain_repository.example.md) e
[Persistenza](./persistence.rule.md).

Le **sorgenti esterne** sono l'eccezione dichiarata: la loro interfaccia sta in
`data/<soggetto>/sources/`, perché a consumarle è il repository e non il dominio. Vedi
[Sorgente esterna](../examples/external_source.example.md).

#### `internal/api/` — Presentation
Solo traduzione fra HTTP e dominio. Vedi [API HTTP](./http_api.rule.md).

## ⚙️ Regole di Dipendenza tra Livelli

-   `api/<risorsa>` → `domain/`, `core/`. Mai un'altra risorsa di `api/`, mai `data/`, mai
    `bootstrap/`.
-   `data/<soggetto>` → `domain/`, `core/`, e `data/` del **proprio** soggetto. Mai `api/`, mai
    `bootstrap/`.
-   `domain/<soggetto>` → solo `core/apperr` e la libreria standard **senza I/O** (`time`,
    `strings`, `errors`, `context`…). Mai `database/sql`, `net/http`, un driver, `data/`,
    `api/`. Import fra soggetti diversi: da evitare — se serve, è segnale che i due soggetti
    vanno fusi, o che il collegamento va fatto nel composition root.
-   `core/` → nessuno dei precedenti.
-   `bootstrap/` → tutto.
-   `cmd/<binario>` → solo `bootstrap/`.

Go vieta già i cicli di import: un ciclo fra due package non compila. Non vieta però una
dipendenza nella direzione sbagliata che non chiude un ciclo — `domain` che importa
`database/sql` compila benissimo. Per questo le regole qui sopra sono un
[test di architettura](../examples/architecture_test.example.md), non una raccomandazione.

## 📦 Poche dipendenze

La libreria standard copre il server HTTP (con il routing per metodo e parametro del
`ServeMux`), il JSON, il logging strutturato (`log/slog`), i test (`testing`, `httptest`) e
l'embedding dei file. Una dipendenza esterna entra quando fa una cosa che la libreria standard
non fa (un driver SQLite, un parser di formati) — non per comodità sintattica. Niente
framework web, niente librerie di asserzioni, niente generatori di mock, niente container di DI.

## 🧭 Corrispondenze con gli altri stack

| Python | Go |
| --- | --- |
| `Protocol` | `interface`, soddisfatta implicitamente |
| `Depends` + `bootstrap/dependencies.py` | costruttori chiamati a mano in `bootstrap/app.go` |
| use case come classe invocabile | metodo di un `Service` del soggetto |
| `@dataclass(frozen=True)` | `struct` passata per valore |
| `StrEnum` | tipo `string` con costanti e un metodo `Valid()` |
| eccezioni tipizzate + un handler | `error` con `Kind` + un solo `httpx.WriteError` |
| `async def` + threadpool | goroutine con `context.Context` e un proprietario |
| `pydantic-settings` | `Config` caricata dall'ambiente in `bootstrap/` |
| `tests/` separata | `_test.go` accanto al codice |

## 🧭 Esempi

-   [go.mod](../examples/go_mod.example.md)
-   [Main e shutdown](../examples/main_and_shutdown.example.md)
-   [Test di architettura](../examples/architecture_test.example.md)
