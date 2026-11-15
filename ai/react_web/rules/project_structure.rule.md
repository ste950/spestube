# Struttura del Progetto

## 📁 Struttura delle Directory Principali

-   **`web/`**: Contiene il sito, un'applicazione React a pagina singola costruita con Vite.
    Tutto il codice sta in `web/src/`, i test in `web/tests/`.
-   **`CLAUDE.md`**: Punto di ingresso per gli agenti AI. Importa questo framework e dichiara
    le specificità del progetto (nome del sito, alias di import, servizi esterni, dove viene
    pubblicato). Regole, skill ed esempi non vivono nel repo del progetto: sono referenziati
    dal framework condiviso.

## 🌐 Che cosa copre questo stack

Un sito che si costruisce in **file statici** — un `index.html`, il JavaScript e il CSS con
l'hash nel nome — e che nel browser diventa un'applicazione: il router cambia pagina senza
ricaricare, i dati arrivano da un'API. Non c'è un server che renda le pagine.

È la scelta che tiene questo stack vicino agli altri due che disegnano interfacce: niente
codice che gira sul server accanto a quello che gira nel browser, quindi niente confine da
regolare fra i due, e la stessa architettura di `flutter/` e
`react_native/` vale così com'è. La pubblicazione è una
cartella di file (vedi [Distribuzione](./deployment.rule.md)).

Il limite è dichiarato, non nascosto: il primo contenuto compare dopo che il JavaScript è
partito. Per un'area riservata, un pannello, uno strumento interno è irrilevante; per un sito
pubblico che vive di motori di ricerca e di anteprime nei social non lo è, e quel progetto ha
bisogno di un rendering lato server che questo stack non descrive. Se ne accorge chi scrive il
`CLAUDE.md` del progetto, non chi scrive la terza pagina.

## 📂 Struttura di `web/`

```
web/
├── package.json
├── index.html                     # l'unica pagina HTML: Vite parte da qui
├── vite.config.ts                 # build, alias, CSS Modules e configurazione dei test
├── tsconfig.json                  # solo riferimenti ai due progetti qui sotto
├── tsconfig.app.json              # strict + alias, per src/ e tests/
├── tsconfig.node.json             # per i file di configurazione che girano in Node
├── eslint.config.js               # lint e confini fra livelli
├── stylelint.config.js            # i fogli di stile non scrivono valori a mano
├── architecture/
│   └── boundaries.json            # la tabella dei divieti, letta dal linter e dal test
├── public/                        # file serviti così come sono (favicon, robots.txt)
│
├── src/
│   ├── main.tsx                   # ingresso: tema, provider e router, nient'altro
│   │
│   ├── app/                       # 🧭 COMPOSIZIONE DELL'APP
│   │   ├── router.tsx             # compone le rotte delle feature
│   │   ├── layouts/               # i layout che più rotte condividono
│   │   └── not-found-page.tsx     # la pagina dei percorsi che nessuna feature dichiara
│   │
│   ├── bootstrap/                 # 🆕 COMPOSITION ROOT
│   │   ├── dependencies.ts        # costruisce le implementazioni, le espone come contratti
│   │   └── config.ts              # l'unico modulo che legge import.meta.env
│   │
│   ├── core/                      # 🔧 INFRASTRUTTURA AGNOSTICA AL DOMINIO
│   │   ├── result/                # Result, Failure e i loro costruttori
│   │   └── use-cases/             # contratto base UseCase<In, Out>
│   │
│   ├── domain/                    # 🧠 DOMAIN LAYER (per soggetto)
│   │   └── <soggetto>/
│   │       ├── entities/          # tipi immutabili di business
│   │       ├── repositories/      # porte: contratti verso l'esterno
│   │       └── use-cases/
│   │
│   ├── data/                      # 🔌 DATA LAYER (per soggetto)
│   │   └── <soggetto>/
│   │       ├── sources/           # contratti e implementazioni delle fonti esterne
│   │       ├── repositories/      # implementazioni delle porte di domain/
│   │       ├── models/            # forme che arrivano dalle fonti
│   │       └── mappers/           # modello → entità
│   │
│   ├── features/                  # 🧩 FEATURE (solo presentation)
│   │   └── <nome_feature>/
│   │       ├── routes.tsx         # i percorsi della feature e le pagine che montano
│   │       ├── pages/             # le pagine
│   │       ├── state/             # store della feature
│   │       └── components/        # componenti specifici della feature
│   │
│   └── shared/                    # 🎨 COMPONENTI, TEMA E TRADUZIONI CONDIVISI
│       ├── components/
│       ├── theme/                 # token, foglio globale (vedi Convenzioni UI)
│       └── i18n/
│
└── tests/
    ├── architecture/              # i test che difendono i confini fra livelli
    ├── domain/                    # test degli use case (nessun I/O)
    ├── data/                      # test dei repository con sorgenti finte
    └── features/                  # test degli store e delle pagine
```

## 🧭 Le rotte stanno nella feature

È il punto in cui questo stack si stacca da React Native e torna a Flutter, e conviene dire
perché. In `react_native/` la rotta nasce
dalla posizione di un file, quindi le rotte non possono appartenere a una feature e vivono in
un albero separato. Qui il router è **codice**: un elenco di oggetti che qualcuno compone. Il
vincolo che là imponeva la divisione non c'è, e tenere i percorsi dentro la feature costa meno
di separarli.

La divisione è questa:

-   **`src/features/<f>/routes.tsx`** dichiara i percorsi della feature e, per ciascuno, il
    piccolo componente che legge i parametri, li converte e monta la pagina. È l'unico file
    della feature che `app/` può importare.
-   **`src/app/router.tsx`** compone le rotte delle feature sotto i layout comuni, aggiunge la
    pagina dei percorsi sconosciuti e crea il router. Non conosce una pagina, uno store o un
    componente di nessuna feature: conosce solo i loro elenchi di rotte.
-   **`src/features/<f>/pages/`** contiene le pagine, che non sanno da quale URL sono state
    raggiunte.

Due alternative scartate. Le **rotte da file** — un plugin che genera il router dalla
posizione dei file, come fa Expo Router — riportano qui il vincolo di React Native senza
nessuno dei motivi che là lo rendono inevitabile, e sciolgono il confine della feature
nell'albero delle cartelle delle rotte. Un **router centrale** che importa tutte le pagine di
tutte le feature funziona, ma fa di `app/` il posto in cui si cambia ogni volta che una feature
aggiunge un percorso, e ogni import di una pagina da parte di `app/` è un pezzo di feature che
esce dalla feature.

La forma completa — compresa la ragione per cui niente `loader` — è in
[Struttura di una Feature](./feature_overview.rule.md#-la-feature-e-le-sue-rotte).

## 🗂️ Feature e soggetti sono cose diverse

-   Una **feature** è una porzione di UI (una pagina o un gruppo di pagine con il suo stato).
    Vive in `src/features/` e non contiene logica di dominio.
-   Un **soggetto** è un concetto di business (es. `orders`, `accounts`) condiviso da tutte le
    feature. Vive in `src/domain/<soggetto>/` (entità e porte) e `src/data/<soggetto>/`
    (implementazioni).
-   Spesso feature e soggetto hanno lo stesso nome, ma non è un obbligo: una feature può
    consumare più soggetti, e un read model trasversale può essere un soggetto senza una
    feature omonima.

## 🧱 Dettagli sui Livelli

#### `src/main.tsx` — Ingresso
Monta il foglio del tema, i provider di presentazione e il router. Non è un livello e non
contiene niente che si possa provare: se cresce oltre la ventina di righe, sta facendo il
lavoro di `app/` o di `bootstrap/`.

#### `src/app/` — Composizione dell'app
Router, layout comuni, pagina dei percorsi sconosciuti. Vedi
[Le rotte stanno nella feature](#-le-rotte-stanno-nella-feature).

#### `src/bootstrap/` — Composition Root
L'unico modulo che può importare tutto. Costruisce le implementazioni di `data/` e le espone
**annotate con i contratti di `domain/`**: è lì che TypeScript verifica la conformità, ed è
l'unico punto da cambiare per sostituire una fonte. Vedi
[Iniezione delle dipendenze](./data_layer.rule.md#-iniezione-delle-dipendenze) e
[Configurazione](./configuration.rule.md).

#### `src/core/` — Infrastruttura Riutilizzabile
`Result`, il vocabolario delle `Failure`, il contratto base `UseCase`. Niente che parli di un
concetto di business e niente che parli di React.

#### `src/domain/` e `src/data/` — Layer Condivisi
Il dominio è unico e organizzato per soggetto. Vedi
[domain/ – Domain Layer](./domain_layer.rule.md) e
[data/ – Data Layer](./data_layer.rule.md).

#### `src/features/` — Feature
Solo presentation, rotte comprese. Layout e regole di dipendenza in
[Struttura di una Feature](./feature_overview.rule.md).

#### `src/shared/` — Componenti, Tema e Traduzioni
Componenti generici, token di tema e cataloghi, agnostici al dominio. Vedi
[Convenzioni UI](./ui_conventions.rule.md).

## 🔗 Import assoluti con alias

Gli import fra livelli usano **alias assoluti** (`@app/*`, `@bootstrap/*`, `@core/*`,
`@domain/*`, `@data/*`, `@features/*`, `@shared/*`); i percorsi relativi restano ammessi
**solo dentro la stessa cartella**.

Non è una preferenza estetica. `../../../domain/orders/entities/order` e
`@domain/orders/entities/order` dicono la stessa cosa a chi legge, ma non a chi verifica: nel
primo caso il livello di destinazione dipende da dove si trova il file di partenza, e una
regola di dipendenza diventa un calcolo di percorsi invece di un confronto di prefissi. Con
gli alias, la tabella qui sotto è direttamente eseguibile, e spostare un file non riscrive
gli import di nessun altro.

Gli alias li leggono tre strumenti — `tsc`, Vite e Vitest — e vanno dichiarati **una volta
sola**: due elenchi tenuti a mano divergono, e il sintomo è un import che compila e poi non si
risolve nella build. Come si fa è in
[Vite e TypeScript](../examples/vite_and_tsconfig.example.md).

## ⚙️ Regole di Dipendenza tra Livelli

Da leggere come una tabella di **prefissi di import vietati**, perché è esattamente la forma
in cui un test di architettura la consuma:

| livello | può importare | non può importare, mai |
| --- | --- | --- |
| `src/app` | `@features/*/routes`, `@shared/*`, `@bootstrap/config`, `react-router` | `@domain/*`, `@data/*`, `@core/*`, il resto di `@bootstrap/*`, qualsiasi altro file di una feature |
| `src/features/<f>` | `@domain/*`, `@core/*`, `@shared/*`, `@bootstrap/dependencies`, la **propria** feature, `react-router` | un'**altra** feature, `@data/*`, `@app/*`, il resto di `@bootstrap/*` |
| `src/data/<s>` | `@domain/*` (anche di altri soggetti), `@core/*`, il **proprio** `@data/<s>/*` | `@features/*`, `@shared/*`, `@bootstrap/*`, `@app/*`, `react`, `react-dom`, `react-router` |
| `src/domain/<s>` | solo `@core/*` | `@data/*`, `@features/*`, `@shared/*`, `@bootstrap/*`, `@app/*`, `react`, `react-dom`, `react-router`, qualsiasi libreria di I/O |
| `src/core` | niente dei precedenti | `@domain/*`, `@data/*`, `@features/*`, `@shared/*`, `@app/*` |
| `src/shared` | `@core/*` | `@domain/*`, `@data/*`, `@features/*`, `@app/*`, `@bootstrap/*` |
| `src/bootstrap` | tutto (composition root) | — |

Tre vincoli che non sono divieti di import fra livelli ma pattern proibiti dentro un file, e
valgono con la stessa forza:

-   **Il codice che parla con l'esterno sta solo nelle sorgenti.** `fetch`, un client HTTP,
    `localStorage`, IndexedDB, l'SDK di un terzo: compaiono in `src/data/<soggetto>/sources/`
    e da nessun'altra parte. Vedi [Sorgenti](./data_layer.rule.md#-sorgenti).
-   **`import.meta.env` si legge solo in `src/bootstrap/config.ts`.** Vedi
    [Configurazione](./configuration.rule.md).
-   **Nessun valore di stile scritto a mano nei fogli dei componenti.** Vedi
    [Convenzioni UI](./ui_conventions.rule.md#-tema-e-token).

Il compilatore non conosce nessuna di queste regole: `tsc` verifica i tipi, non la direzione
delle frecce. La garanzia è il **linter**, che gli import li legge già, insieme a stylelint per
i fogli di stile: la tabella è una loro configurazione, committata nel progetto e eseguita
dalla CI, e un import vietato fallisce sulla riga in cui è scritto — vedi
[Confini fra livelli](../examples/architecture_test.example.md). Il criterio e la forma
generale, valida per tutti gli stack, sono in
[Test di Architettura](../../quality/rules/architecture_tests.rule.md) e
[Tabella degli import vietati](../../quality/examples/forbidden_imports_table.example.md); come
là, una riga nuova della tabella nasce nello stesso commit della regola che difende.
Mentre si lavora resta comoda la verifica al volo, da `web/`:

```bash
grep -rn "@data/" src/features src/domain src/app
grep -rn "from 'react" src/domain src/data
```

## 🧭 Corrispondenze con gli altri stack

Questo stack è il gemello web di `react_native/`: stesso
linguaggio, stessi `domain/`, `data/`, `core/` e `bootstrap/`, e le regole di quei livelli sono
la stessa regola scritta due volte — quando una cambia, l'altra va riletta. Le righe che
contano sono quelle in cui le tre colonne **non** coincidono.

| Flutter | React Native + Expo | React web |
| --- | --- | --- |
| `features/<f>/pages/` | `src/features/<f>/screens/`, montate da un file di `app/` | `src/features/<f>/pages/` |
| rotte dichiarate nel file della feature | nessun equivalente: le rotte nascono dalla posizione dei file in `app/` | `src/features/<f>/routes.tsx`, composte da `src/app/router.tsx` |
| `Bloc` | uno store leggero per feature, in `state/` | identico a React Native |
| `Result<T>` sealed + `fold` | unione discriminata con restringimento su `ok` | identico a React Native |
| `GetIt` / `service_locator.dart` | composition root in `bootstrap/dependencies.ts` | identico a React Native |
| `shared/` + `AppThemeTokens` | token in un oggetto TypeScript, letti con `useAppTheme()` | token in un oggetto TypeScript, scritti come [variabili CSS](./ui_conventions.rule.md#-tema-e-token) e letti dai CSS Modules |
| verifica sul dispositivo | verifica sul dispositivo | [verifica nel browser](./verification.rule.md#%EF%B8%8F-verifica-nel-browser) |
| pacchetto per gli store | build nativa su un servizio di build | una cartella di file statici, vedi [Distribuzione](./deployment.rule.md) |
