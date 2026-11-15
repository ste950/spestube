# Distribuzione

Il prodotto di questo stack è **una cartella**: `web/dist/`, con un `index.html`, il JavaScript
e il CSS con un hash nel nome, e il contenuto di `public/` copiato così com'è. Si serve con
qualsiasi server di file statici. Niente runtime da installare, niente processo da tenere in
vita.

## 🔨 La build

```bash
npm ci
npx vite build
```

-   La build la fa la **CI**, non la macchina di destinazione e non quella di chi scrive: il
    pacchetto che si pubblica è quello costruito da un lock e da un ambiente noti.
-   Le variabili `VITE_*` entrano in questo momento, e solo in questo — vedi
    [Configurazione](./configuration.rule.md#%EF%B8%8F-una-build-per-ambiente).
-   La **versione** del sito è quella di `package.json`, e la build la scrive dove serve
    (un attributo nel documento, un campo inviato con le segnalazioni di errore) leggendola
    con `define` nella configurazione di Vite. Vedi
    Versionamento e release.

## 🔀 Il fallback su `index.html`

È la regola che si dimentica, e si scopre in produzione. In un'applicazione a pagina singola i
percorsi interni — `/orders/42` — **non sono file**: esistono solo dopo che il router è
partito. Il server riceve la richiesta di `/orders/42`, cerca un file con quel nome, non lo
trova e risponde 404.

Il server va quindi configurato così:

-   un file che esiste (`/assets/index-3f9a.js`, `/favicon.svg`) si serve com'è;
-   **qualsiasi altro percorso** risponde con `index.html`, e da lì il router decide. Un
    percorso che nessuna feature dichiara arriva alla pagina dei percorsi sconosciuti di
    `src/app/`, che è la vera pagina 404 del sito.

Il sintomo di un fallback mancante è inconfondibile: navigando dentro il sito tutto funziona,
ricaricando una pagina interna o aprendo un link condiviso si ottiene l'errore del server. È
il motivo per cui la [verifica nel browser](./verification.rule.md#%EF%B8%8F-verifica-nel-browser)
comprende la ricarica su una rotta profonda.

## 🗄️ La cache

Due regole opposte per due tipi di file, e sbagliarle produce o un sito lento o un sito vecchio:

-   **i file con l'hash nel nome** (`/assets/*`) non cambiano mai: un contenuto diverso ha un
    nome diverso. Si servono con una cache lunga e `immutable`, e un visitatore non li riscarica
    finché il sito non cambia davvero;
-   **`index.html` non si mette in cache** (`no-cache`): è il file che dice quali asset caricare,
    e una copia vecchia nel browser continua a chiedere quelli della versione precedente.

## 🧭 Il percorso di base

Un sito servito alla radice del dominio non ha niente da configurare. Un sito servito sotto un
percorso (`https://<host>/<percorso>/`) lo dichiara **in un posto**, l'opzione `base` di Vite,
e il router lo riceve da lì (`import.meta.env.BASE_URL`) come `basename`. Un percorso scritto a
mano in un link o in un asset si rompe il giorno in cui il sito si sposta.

## 🛡️ Intestazioni

Il server aggiunge le intestazioni di sicurezza che un sito statico non può darsi da solo:
almeno `Content-Security-Policy` (con le sole origini da cui il sito carica davvero codice e
dati), `X-Content-Type-Options: nosniff` e `Referrer-Policy`. La politica dei contenuti è
anche il controllo che rende evidente una dipendenza che carica script da un'origine non
dichiarata.

## 🧭 Esempi

-   [Server statico](../examples/static_hosting.example.md): la configurazione di un server di
    file con fallback, cache e intestazioni.
-   [Vite e TypeScript](../examples/vite_and_tsconfig.example.md): `base`, la versione e gli
    alias nella configurazione della build.
