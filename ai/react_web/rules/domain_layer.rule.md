# Layer: Domain (Dominio)

Il layer `web/src/domain/` contiene la logica di business pura: entità, contratti verso
l'esterno e use case. È **condiviso da tutte le feature** e non sa niente di React, del browser,
della rete e del router.

## 📁 Struttura

Il dominio è organizzato **per soggetto** (concetto di business):

```
domain/
└── <soggetto>/
    ├── entities/        # tipi immutabili di business (es. order.ts → type Order)
    ├── repositories/    # porte: contratti verso l'esterno (es. order-repository.ts)
    └── use-cases/       # (opzionale) regole di business nostre
```

## 🗂️ Soggetti

Un soggetto è un concetto di business, non una pagina. Criteri:

-   Un soggetto ha **almeno una porta propria**: se un gruppo di entità non ha una fonte dati
    dedicata, probabilmente appartiene a un soggetto esistente.
-   Un **read model trasversale** (una vista di lettura che aggrega più soggetti) è un
    soggetto autonomo: ha le proprie entità e il proprio contratto; l'aggregazione la fa
    `data/` (vedi [Repository composti](./data_layer.rule.md#-repository-composti)).
-   Se due soggetti hanno bisogno di importarsi a vicenda le entità, vanno fusi in uno solo.

## 🧱 Entità, porte e use case

Dentro un soggetto convivono tre cose diverse, e confonderle è l'errore più comune di questo
layer. La domanda che le distingue è una sola: **c'è un confine da attraversare?**

| | cos'è | contratto? | implementazione |
| --- | --- | --- | --- |
| **Entità** | dati di business e logica derivata da quei dati | no | è l'entità stessa |
| **Porta** (`repositories/`) | contratto verso l'esterno: rete, storage, servizi di terzi | **sì** | in `data/<soggetto>/` |
| **Use case** | una regola di business nostra, senza I/O | **no** | è lo use case stesso |

-   Le **porte** hanno un contratto perché c'è un confine: dietro può esserci un backend, un
    file locale o un oggetto finto in un test, e chi sta sopra non deve accorgersene.
    Sostituire la fonte significa scrivere un'altra implementazione e cambiare una riga in
    `bootstrap/dependencies.ts`.
-   Gli **use case** non ce l'hanno, e dargliela è un errore: sarebbe la copia esatta
    dell'unica implementazione possibile, cioè della logica stessa. Un calcolo non ha un
    "dietro" da nascondere. Se ti accorgi di stare scrivendo l'`interface` di uno use case,
    quasi sempre significa che quello che hai in mano non è uno use case ma una porta.

Esiste un quarto oggetto con un contratto, ed è il motivo per cui questa tabella si sbaglia
spesso: la **sorgente**. Anche lei attraversa un confine, ma non è una porta di dominio e in
questa tabella non entra — sta tutta in `data/<soggetto>/sources/`, contratto compreso. La
differenza è di chi pone la domanda: una porta dichiara **cosa serve al dominio**, e per
questo è scritta nella sua lingua; una sorgente dichiara **come si parla a una fonte**, e
parla la lingua della fonte. La definizione sta in
[data/ – Data Layer](./data_layer.rule.md#-sorgenti).

## 🧬 Entità

-   Un'entità è un **`type` con campi `readonly`**, non una `class`. Non ha metodi che
    mutano, non ha ciclo di vita, non ha identità di oggetto: è un valore, e in TypeScript un
    valore si descrive con un tipo. Le `class` restano dove servono davvero — non qui.
-   La logica **derivata** da un'entità (un totale, uno stato calcolato, una scadenza
    superata) è una **funzione pura esportata accanto al tipo**, che prende l'entità e
    restituisce il derivato. Metterla in un metodo obbligherebbe a costruire un'istanza per
    ogni oggetto che arriva da una fonte, e ogni mapper diventerebbe un costruttore.
-   Gli insiemi chiusi di valori si dichiarano come **unioni di stringhe letterali**
    (`type OrderStatus = 'draft' | 'paid' | 'shipped'`), non come `enum`. Un'unione di
    letterali esiste solo a compile time, si confronta con `===`, si restringe in uno `switch`
    ed è già serializzabile: l'`enum` di TypeScript genera codice a runtime e, nella variante
    numerica, accetta silenziosamente numeri che non appartengono all'insieme.
-   Uno `switch` su un insieme chiuso finisce con un ramo di default che assegna il valore a
    `never`: quando l'insieme cresce, la compilazione fallisce nel punto in cui manca il caso
    invece di ricadere in un comportamento di ripiego.

## ⚖️ Result e fallimenti

Un fallimento **previsto** — la rete non risponde, la risorsa non esiste, il dato salvato è
illeggibile — non è un'eccezione: è un esito, e viaggia nel tipo di ritorno.

```ts
export type Result<T> = { readonly ok: true; readonly value: T } | { readonly ok: false; readonly failure: Failure };
```

È un'**unione discriminata** su `ok`, ed è l'equivalente naturale del tipo sigillato degli
altri stack. Il restringimento lo fa il controllo di flusso del compilatore: dopo
`if (!result.ok) return result;` TypeScript sa che `result.value` esiste, e prima di quel
controllo sa che non esiste. La garanzia è la stessa di un `fold`, senza chiedere a chi legge
di imparare una funzione nuova.

Le alternative, e perché sono scartate:

-   **Le eccezioni** non compaiono nella firma. Una funzione che lancia ha lo stesso tipo di
    una che non lancia, quindi non esiste nessun punto in cui il compilatore chieda «e se
    fallisse?». È la ragione per cui in questo stack si torna a un valore, mentre nello stack
    Python — dove l'eccezione arriva a **un solo** handler HTTP — la scelta opposta è
    corretta.
-   **Una classe `Result` con `fold`** sembra più vicina allo stack Flutter, ma in TypeScript
    perde proprio ciò che rende utile il pattern: un metodo non restringe il tipo di `this`,
    quindi si torna a un accesso che il compilatore non protegge, e ogni esito va istanziato
    con `new`.

Il `Failure` è a sua volta un'unione discriminata su un campo `kind`
(`'network' | 'notFound' | 'parse' | …`), **non** una gerarchia di sottoclassi di `Error`:
una catena di `instanceof` non è mai esaustiva, mentre uno `switch` su `kind` chiuso da
`never` fallisce in compilazione il giorno in cui nasce un modo nuovo di fallire. Il
vocabolario delle failure è **condiviso** e vive in `core/result/`: un soggetto ne dichiara
una propria solo se ha un modo di fallire che non esiste altrove.

Dove le eccezioni restano: nascono dalle librerie, dentro le sorgenti, e lì si fermano — è
il confine in cui vengono assorbite e diventano `Failure`. Vedi
[Errori](./data_layer.rule.md#errori).

## 🧩 Use Case

Uno use case incapsula una singola operazione di logica di business, e va estratto qui quando
una regola di dominio (un calcolo, un ordinamento, una selezione) rischia di finire in un
componente o in uno store.

Convenzioni:

-   File in `domain/<soggetto>/use-cases/`, nominato `<verbo>-<soggetto>.ts`; il simbolo
    esportato è `<verbo><Soggetto>` (es. `sort-orders.ts` → `sortOrders`).
-   Uno use case è una **funzione esportata**, non una classe con un solo metodo. In
    TypeScript una funzione è già un valore iniettabile e sostituibile: la classe
    aggiungerebbe un `new` e nient'altro. Quando riceve delle porte, è una **funzione che le
    prende e restituisce la funzione finale** (`makeSortOrders(repo)`), così la dipendenza è
    esplicita e il test passa un oggetto letterale.
-   Una sola responsabilità per use case: un caso d'uso, un output.
-   Ritorna **entità di dominio**, mai forme pensate per la UI: una percentuale per una barra
    di avanzamento, un'etichetta già formattata e una data già tradotta restano nella
    presentation.
-   Non importa niente fuori da `@core/*` e dal proprio soggetto.

## 🚧 Regole di Dipendenza

-   ✅ Può importare solo `@core/*` (`Result`, `Failure`, contratto `UseCase`).
-   ❌ Mai `react`, `react-dom`, `react-router`, `@data/*`, `@features/*`, `@shared/*`, e nessuna
    libreria che faccia I/O.

Il divieto su `react` è il più facile da violare senza accorgersene, perché un hook sembra un
posto ragionevole dove tenere una regola di business. Non lo è: una funzione pura si prova
chiamandola, un hook si prova solo rendendo qualcosa.

La tabella completa, e la sua forma eseguibile, sono in
[Struttura del Progetto](./project_structure.rule.md#%EF%B8%8F-regole-di-dipendenza-tra-livelli).
