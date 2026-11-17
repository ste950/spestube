# Struttura di una Feature

Una feature (`web/src/features/<nome_feature>/`) contiene **solo il presentation layer** di una
funzionalità: i suoi percorsi, le pagine, lo stato di presentazione e i componenti. Entità,
porte e use case vivono nei layer condivisi `domain/` e `data/` (vedi
[domain/ – Domain Layer](./domain_layer.rule.md) e [data/ – Data Layer](./data_layer.rule.md)).

## 📁 Layout della Directory

```
<nome_feature>/
├── routes.tsx        # i percorsi della feature e i componenti che montano le pagine
├── pages/            # le pagine
├── state/            # lo store della feature (e i suoi selettori)
└── components/       # componenti specifici della feature
```

Non esiste una cartella `presentation/`: pages/state/components stanno alla radice della
feature, perché la feature *è* il presentation layer.

## 🧭 La feature e le sue rotte

La feature **possiede i propri percorsi**, come in Flutter: `routes.tsx` esporta un elenco di
rotte di React Router (`ordersRoutes: RouteObject[]`) e `src/app/router.tsx` lo compone con
quelli delle altre feature. Il perché di questa scelta, e delle alternative scartate, è in
[Struttura del Progetto](./project_structure.rule.md#-le-rotte-stanno-nella-feature).

`routes.tsx` è l'**indirizzario** della feature, e dentro c'è solo questo:

-   **i percorsi**, relativi: la feature dichiara `orders` e `orders/:orderId`, e non sa sotto
    quale layout `app/` li monterà;
-   **un componente di montaggio per rotta**, di poche righe: legge i parametri, li converte e
    rende la pagina. Niente chiamate a repository o use case, niente stato, niente layout;
-   **il caricamento pigro delle pagine**: ogni pagina si importa con `lazy`, così il codice di
    una feature si scarica quando qualcuno ci arriva e non all'apertura del sito. È gratis
    proprio perché il confine della feature coincide con il confine del pacchetto scaricato.

La conseguenza pratica è che i **parametri di rotta sono stringhe**, e la conversione (un
identificativo numerico, una data, un valore di un insieme chiuso) avviene nel componente di
montaggio — mai lasciata scivolare verso il dominio, che non deve sapere che esiste un URL.
Lo stesso vale per i parametri di ricerca (`?stato=spedito`): li legge il montaggio, la pagina
riceve valori già tipizzati.

L'inverso vale per la navigazione: una pagina che deve portare altrove usa un `<Link>` o
`useNavigate()`, e il percorso di destinazione è l'unica cosa che sa dell'altra feature. Un
percorso non è un import: le feature restano indipendenti. Un percorso sbagliato non fallisce
in compilazione — finisce sulla pagina dei percorsi sconosciuti — ed è il prezzo
dell'indipendenza; il test di una pagina che naviga verifica il percorso che produce.

Un **link è un link**: si naviga con `<Link>` e non con un `onClick` che chiama
`navigate()`, perché solo il primo è un `<a href>` che il browser sa aprire in una nuova
scheda, copiare e annunciare come collegamento. `useNavigate()` resta per la navigazione che
segue un'azione — un salvataggio riuscito, un'uscita.

### Niente `loader` né `action`

React Router sa caricare i dati prima di rendere una rotta (`loader`) e ricevere i form
(`action`). Questo stack **non li usa**, e la ragione è la stessa per cui lo stato sta in uno
store: un `loader` mette la chiamata al repository dentro la tabella delle rotte, cioè nel
posto che deve restare un indirizzario, e la sua logica si prova solo attraverso il router.
Il dato passa dallo store della feature, che riceve le porte quando viene creato e si prova
chiamando una funzione. Il router resta quello che è: dato un URL, quale pagina.

## 🧠 Stato della feature

Lo stato di presentazione di una feature vive in uno **store leggero** (Zustand, o qualsiasi
store con la stessa forma) dichiarato in `state/`. È il ruolo che negli altri stack ha il
Bloc, ed è la stessa scelta di React Native, per gli stessi motivi.

### Perché uno store e non contesto più hook

La domanda è quale delle due forme rende la logica **verificabile fuori dai componenti**, e
la risposta non è simmetrica.

Uno store è un oggetto normale, creato da una funzione normale. Un test lo costruisce,
chiama un'azione, legge lo stato risultante e asserisce — senza rendere niente, senza un
albero di componenti, senza aspettare un effetto. Le transizioni di stato, che sono la parte
che si rompe, si provano come si prova una funzione.

Con `Context` più hook, lo stato vive **dentro** l'albero dei componenti: `useState` esiste
solo durante un render, `useEffect` parte solo dopo, e per provare una transizione bisogna
montare qualcosa e aspettare che React abbia finito. La logica è raggiungibile solo
attraverso il rendering, e il test finisce per verificare React insieme a quello che voleva
verificare davvero.

C'è anche una conseguenza di struttura: con il contesto, «dove sta lo stato» dipende da dove
è montato il provider, quindi da `app/`. Con uno store, lo stato appartiene alla feature come
la sua cartella, e non c'è un albero da consultare per sapere chi lo vede.

Scartata anche una **libreria di cache delle richieste** al posto dello store: risolve bene la
sincronizzazione con un'API, ma lo fa chiamando la rete dal componente, cioè sopra la porta di
dominio invece che dietro. Se un progetto ne ha bisogno, il posto è dentro un repository di
`data/`, dove il dominio non la vede.

`Context` resta quello che è e serve: un canale per ciò che è davvero di presentazione e
attraversa l'albero — le traduzioni in testa. Non è il posto dello stato di una feature, e non
è il posto delle dipendenze di dominio.

### Come è fatto

-   **Uno store per feature**, in `state/`, e non uno per pagina: due pagine della stessa
    feature che mostrano lo stesso dato lo leggono dallo stesso posto.
-   Lo store **riceve le porte di dominio quando viene creato**, non se le importa da dentro.
    È la stessa regola degli use case, ed è ciò che permette a un test di passargli un
    repository finto.
-   Nello stato ci va **il risultato**, non il `Result`: l'azione chiama il repository,
    restringe l'esito su `ok` e scrive nello stato o i dati o il fallimento tradotto in
    qualcosa che la UI sappia mostrare. Lo `switch` sul `kind` della `Failure` è l'ultimo
    punto in cui quel vocabolario esiste.
-   I componenti leggono attraverso **selettori** (`useOrdersStore((s) => s.orders)`), non
    prendendo lo stato intero: un componente che si sottoscrive a tutto si ridisegna a ogni
    cambiamento, anche per campi che non mostra.
-   **Niente logica di dominio nello store.** Un ordinamento, un calcolo, una selezione sono
    use case: se lo store cresce, è quasi sempre perché ne contiene uno.
-   **L'URL non si copia nello store.** Un filtro che sta nei parametri di ricerca è già
    stato, e il browser lo conserva, lo condivide e lo ripristina con il tasto indietro: lo
    store lo riceve dalla pagina come argomento di un'azione, non ne tiene un secondo
    esemplare che prima o poi diverge.

`state/` può contenere anche view model di presentazione — forme derivate dalle entità che
esistono solo per la UI (un'etichetta già formattata, una frazione per una barra di
avanzamento). Quelle non sono dominio, e qui stanno bene.

## ⚙️ Responsabilità

-   Dichiarare i propri percorsi e convertirne i parametri.
-   Rendere la UI a partire dallo stato esposto dallo store.
-   Gestire l'input dell'utente e la navigazione.
-   Tradurre entità di dominio in ciò che si mostra, e fallimenti in ciò che si dice.

## 🚧 Regole di Dipendenza

Una feature può importare:

-   ✅ `@domain/*` (entità, porte, use case)
-   ✅ `@core/*` e `@shared/*`
-   ✅ file della **stessa** feature
-   ✅ `@bootstrap/dependencies` — l'unico modulo di `bootstrap/` che le è concesso
-   ✅ `react-router`, per dichiarare le rotte e navigare

Una feature **non** può importare:

-   ❌ un'**altra feature**, per nessun motivo: le feature sono indipendenti. Se due feature
    devono condividere dati, la condivisione passa dal dominio; se devono condividere un
    componente, quello va in `@shared/*`; se devono condividere stato, lo stato è di dominio e
    il posto giusto è dietro una porta.
-   ❌ `@data/*`: le implementazioni arrivano da `@bootstrap/dependencies` attraverso le porte
    di `domain/`.
-   ❌ il resto di `bootstrap/`: la composizione dell'app non è affare delle feature.
-   ❌ `@app/*`: la feature non sa sotto quale layout è montata.

E una sola cosa della feature è visibile da fuori: `routes.tsx`, che `app/` importa. Pagine,
store e componenti non escono dalla cartella.

L'indipendenza fra feature è una delle regole che un test di architettura deve difendere
sempre — vedi [Test di Architettura](../../quality/rules/architecture_tests.rule.md). Mentre
si lavora su una singola feature, la verifica al volo da `web/`:

```bash
grep -rn "@features/" src/features/<mia_feature> | grep -v "@features/<mia_feature>"
grep -rn "@data/" src/features/<mia_feature>
```
