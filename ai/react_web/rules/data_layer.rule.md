# Layer: Data (Dati)

Il layer `web/src/data/` si occupa della comunicazione con l'esterno: API remote, storage
del browser (`localStorage`, IndexedDB), SDK di terzi. È **condiviso da tutte le feature**,
organizzato per soggetto in modo speculare a `domain/`, ed è il solo posto in cui l'applicazione sa *come* i dati sono
fatti davvero.

## 📁 Struttura

```
data/
└── <soggetto>/
    ├── sources/         # contratti e implementazioni delle fonti esterne
    ├── repositories/    # implementazioni delle porte di domain/<soggetto>/
    ├── models/          # (se serve) le forme che le fonti restituiscono
    └── mappers/         # (se serve) modello → entità
```

## ⚙️ Responsabilità

-   Implementare le porte di `domain/<soggetto>/repositories/`.
-   Attraversare i confini tecnici del soggetto — rete, storage, SDK di terzi — dentro le
    **sorgenti** di `sources/`, e da nessun'altra parte: è lì che sta la strategia di
    trasporto (il retry, la paginazione della fonte, il formato che arriva). Il repository le
    **compone**, e decide cosa chiedere a quale; non le sostituisce e non parla al loro posto.
-   Convertire ogni rappresentazione esterna in entità di dominio, passando dai `mappers/`.
-   Assorbire quello che le librerie sollevano e restituirlo come `Failure` dentro un
    `Result`.

## 🌐 Sorgenti

Una **sorgente** è l'oggetto che sa parlare con **una** fonte esterna: l'API HTTP del
backend, `localStorage`, IndexedDB, l'SDK di un terzo. Vive in
`data/<soggetto>/sources/`, dove stanno insieme il suo contratto e la sua implementazione. Il
modulo dell'implementazione è **l'unico** dell'applicazione che importa la libreria di quel
confine.

Esiste perché il repository ha già un mestiere — soddisfare la porta del dominio: comporre,
decidere, mappare, tradurre gli errori — e il modo di parlare con la fonte è un mestiere
diverso, che cambia per motivi diversi. Tenerli nello stesso file significa rimettere le mani
sulla logica di composizione ogni volta che cambia un endpoint, e non avere nessun posto dove
mettere una cache locale accanto a una fonte remota.

### La sorgente merita un contratto esplicito?

Sì, e la domanda è legittima: in TypeScript un `interface` è **strutturale**, quindi un
oggetto conforme lo soddisfa senza dichiararlo, e un modulo è già un singleton — sembra che
si possa esportare direttamente la funzione che chiama la rete e chiuderla lì.

Il motivo per cui il contratto si scrive comunque è uno solo, ed è bene che resti quello:
**poter sostituire la fonte in un test del repository**. Il test di un repository deve poter
provare la composizione — cosa si chiede alla fonte remota, quando si ripiega sulla copia
locale, cosa succede quando la rete fallisce — senza rete e senza browser. Con un
contratto dichiarato, il finto è un oggetto letterale annotato con l'`interface`: il
compilatore verifica che abbia tutti i metodi con le firme giuste, e quando il contratto
cambia sono i test a smettere di compilare. Senza contratto, resta l'intercettazione del
modulo da parte del test runner: funziona, ma sposta la verità dal tipo a una stringa con un
percorso dentro il file di test, che nessuno ricontrolla quando il file si sposta.

Non si scrive, invece, «per astrarre»: un contratto con una sola implementazione e nessun
test che lo finge è una riga in più che non difende niente. La regola è pratica — c'è una
sorgente, quindi c'è un repository che la compone, quindi c'è un test che deve fingerla.

Proprio perché la conformità è strutturale, l'implementazione **non dichiara** di soddisfare
il contratto: è la **factory** ad annotare il tipo di ritorno
(`export function createHttpOrderSource(...): OrderSource`), ed è lì che TypeScript verifica.
Una factory, e non una `class`: quello che si costruisce è una chiusura sulle sue dipendenze
— l'URL di base, il client, la chiave di storage — e non un oggetto con un ciclo di vita.

### Perché il contratto sta in `data/` e non in `domain/`

La porta del dominio **è il repository**, ed è l'unica che il dominio deve conoscere: dietro
di essa può esserci HTTP, un file o un finto, e chi sta sopra non se ne accorge. Dichiarare
in `domain/` anche il contratto della sorgente significherebbe scrivere nel dominio che quel
soggetto ha una fonte remota — esattamente ciò che la porta serve a nascondere.

Tenerlo in `data/` lo rende anche irraggiungibile invece che vietato: `domain/` e `features/`
non possono importare `@data/*`, quindi nessuno può farsi dare la sorgente e scavalcare il
repository, nemmeno volendo.

### Perimetro

Una sorgente per ogni **confine tecnico** attraversato dal soggetto. Due confini sono due
sorgenti anche quando servono lo stesso dato: la fonte remota e la copia locale non sono la
stessa cosa, e distinguerle è ciò che rende la cache una responsabilità dichiarata del
repository invece di un `if` sepolto nel codice di rete.

### Cosa attraversa il confine

La sorgente parla la lingua della fonte e restituisce `Promise<Result<Modello>>`: i
**modelli** di `models/`, imbustati come ogni altro esito di questa applicazione.

Un modello è un tipo **nostro**, con i campi e i tipi che la fonte ha davvero, ed è ciò che
permette di sostituire la libreria senza toccare niente sopra la sorgente. Non è `any`, non è
`unknown` lasciato tale e non è `Record<string, unknown>`: un contratto che restituisce forme
indistinte non dice niente, due sorgenti diverse hanno la stessa firma e il mapper non ha da
dove partire.

La divisione del lavoro è netta: la sorgente **valida e deserializza** (legge quello che è
arrivato e lo porta dentro il tipo dichiarato, fallendo con una `Failure` di parsing se non
corrisponde), il mapper **interpreta** (default, insiemi chiusi di valori, regole di
dominio). Le entità le costruisce il repository chiamando i `mappers/`, così la conversione
resta in un punto solo: una sorgente che restituisce entità toglie al repository il suo
mestiere e obbliga a duplicare il mapping a ogni fonte in più.

Qui il confine ha un dettaglio che gli altri stack non hanno: quello che arriva da `fetch` è
`any`, e `any` non è un errore per il compilatore — si propaga in silenzio fino a dove
qualcuno lo usa. La sorgente è il punto in cui va fermato: l'esito grezzo si tratta come
`unknown` e si converte nel modello solo dopo averne controllato la forma. Il controllo è
codice, non una `as`: un `as` è un'affermazione che nessuno verifica, ed è il modo più rapido
per far arrivare `undefined` dentro un'entità.

### Errori

La sorgente **restituisce** i suoi fallimenti, non li solleva: è il confine in cui quello che
la fonte fa davvero — la rete che cade, un JSON illeggibile, un SDK che esplode — viene
assorbito e diventa una `Failure` dentro un `Result`. È qui che vive il `try/catch`, ed è
l'unico posto di questo layer dove ha diritto di esistere
([Result e fallimenti](./domain_layer.rule.md#%EF%B8%8F-result-e-fallimenti)).

Il `catch` di TypeScript riceve `unknown`, ed è corretto che sia così: quello che una
libreria lancia non è necessariamente un `Error`. Non si legge `err.message` senza prima
controllare cosa si è preso in mano, e il messaggio della libreria non è il messaggio che
arriva al dominio.

Le failure sono quelle **condivise** di `core/result/` — rete, cache, parsing, non trovato:
un vocabolario solo, che il repository non deve tradurre. Una sorgente ne dichiara una
propria, accanto al suo contratto, soltanto se il soggetto ha un modo di fallire che non
esiste altrove.

### Chi può chiamare una sorgente

Solo i **repository** dello stesso soggetto. Mai uno use case, mai una feature, mai uno
store: chi sta sopra il repository non sa nemmeno che le sorgenti esistano. Vale anche fra
soggetti: la sorgente di un soggetto la chiama il **suo** repository, non quello di un altro
— un soggetto che ha bisogno dei dati di un altro passa dalla porta di dominio di quell'altro
(vedi [Repository composti](#-repository-composti)).

Il confine è verificabile in due modi, e vanno usati entrambi: nessun file fuori da
`src/data/<soggetto>/repositories/` importa da `../sources/`, e nessun file fuori da
`src/data/<soggetto>/sources/` nomina `fetch`, un client HTTP o una libreria di storage. Il
secondo è quello che conta davvero, perché è il divieto che chiude la porta anche a chi
riscrive la chiamata di rete direttamente dentro uno store.

### Naming

-   Il **contratto** prende il nome del confine che nasconde: `order-source.ts` →
    `OrderSource`, `order-cache.ts` → `OrderCache`.
-   L'**implementazione** ci mette davanti la strategia di trasporto:
    `http-order-source.ts` → `createHttpOrderSource`, `local-order-cache.ts`.
-   Il **repository** prende il nome della sua strategia di composizione
    (`cached-order-repository.ts` → `createCachedOrderRepository`), perché è quello che lo
    distingue da un'altra implementazione della stessa porta.

## 💉 Iniezione delle dipendenze

**Questo stack non ha un service locator, e la differenza rispetto agli altri è voluta.**

In TypeScript un modulo è già un singleton: viene valutato una volta sola e ogni `import`
successivo riceve lo stesso valore. Un contenitore come `GetIt` risolverebbe un problema che
qui non esiste — avere una sola istanza — al prezzo di una registrazione per tipo e di una
risoluzione che nessun compilatore controlla: `getIt<OrderRepository>()` fallisce a runtime se
qualcuno ha dimenticato di registrare, e la dimenticanza non si vede nel diff.

Al suo posto c'è un **composition root** in `src/bootstrap/dependencies.ts`: un modulo che
costruisce le implementazioni chiamando le factory di `data/` e le esporta **annotate con i
contratti di `domain/`**.

```ts
export const orderRepository: OrderRepository = createCachedOrderRepository(
  createHttpOrderSource(apiClient),
  createLocalOrderCache(),
);
```

Sono tre righe che fanno tutto il lavoro di un contenitore: l'annotazione è il punto in cui
TypeScript verifica la conformità strutturale, il tipo esportato è il contratto e non
l'implementazione (quindi chi consuma non può usare un metodo che la porta non dichiara), e
una dipendenza mancante è un errore di compilazione, non una sorpresa all'avvio.

Le regole che restano, e sono quelle che rendono la scelta verificabile:

-   **Le feature importano solo `@bootstrap/dependencies`**, mai un modulo di `@data/*`. È
    l'unica riga di `bootstrap/` che possono toccare, ed è la ragione per cui la tabella dei
    divieti la nomina esplicitamente.
-   **Ogni cosa costruita altrove riceve le sue dipendenze come argomenti**: un repository
    prende le sorgenti dalla factory, uno use case prende le porte da `make*`, uno store le
    prende quando viene creato. Nessun modulo di `data/`, `domain/` o `features/` importa
    `@bootstrap/dependencies` per servirsi da sé — fuori dal composition root lo conosce solo
    il modulo di due righe che lega lo store di ciascuna feature alle sue porte.
-   Un test non ha bisogno di sostituire niente a livello globale: costruisce la stessa
    catena con gli oggetti finti che gli servono.

Un `Context` di React **non** è il posto della dependency injection di dominio: obbligherebbe
a rendere qualcosa per ottenere un repository, e legherebbe la vita degli oggetti condivisi
all'albero dei componenti. Resta quello che è — un canale per ciò che è davvero di
presentazione, come il tema.

## 🧬 Repository composti

Un repository compone due cose diverse, e vanno tenute distinte:

-   Le **sorgenti del proprio soggetto** — la remota e la locale, per esempio. È il caso
    normale, ed è di questo che parla [Sorgenti](#-sorgenti): il soggetto è uno, cambia solo
    da dove arriva il dato.
-   I **repository di altri soggetti**, quando il soggetto è un read model che non ha dati
    propri. Lì si aggregano domini, non si sceglie un trasporto.

Un'implementazione può fare entrambe le cose, ma le riceve da due posti diversi: le sorgenti
sono contratti di `data/`, i repository altrui sono porte di `domain/`. Il coupling fra
soggetti resta confinato in quel file, ed è così che una pagina mostra dati "di un altro
soggetto": la feature parla con la **propria** porta di dominio, la composizione vive qui, e
le feature non si importano mai a vicenda.

## 🚧 Regole di Dipendenza

-   ✅ Può importare `@domain/*` (anche di altri soggetti: è così che si compongono i read
    model) e `@core/*`.
-   ✅ Può importare `@data/<proprio soggetto>/*`: è così che un repository riceve la sua
    sorgente e i suoi mapper.
-   ❌ Mai `@features/*`, `@shared/*`, `@bootstrap/*`, `@app/*`.
-   ❌ Mai `react`, `react-dom`, `react-router`: questo layer non rende niente. Le API del
    browser che conservano dati (`localStorage`, `sessionStorage`, IndexedDB) sono un'altra
    cosa e stanno **dentro le sorgenti**.

## 🚧 Errori ricorrenti

-   **Una sorgente che restituisce entità**: toglie al repository il suo mestiere e obbliga a
    riscrivere la conversione a ogni fonte in più.
-   **`as` al posto di un controllo di forma**: fa passare il compilatore e rimanda il
    problema a dove il campo viene letto, lontano dal confine che avrebbe dovuto fermarlo.
-   **`catch` che diventa una lista vuota**: un errore imprevisto mascherato da assenza di
    dati è un bug che nessuno vedrà mai.
-   **Logica di business nel repository**: ordinamenti, punteggi e selezioni sono use case.
-   **Chiamate di rete dentro uno store o un componente**: è il divieto che il test di
    architettura cerca per primo.
