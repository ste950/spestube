# Configurazione

## 🌍 Tutto quello che il sito sa, lo sa anche chi lo visita

Un sito statico non ha un processo che legge l'ambiente all'avvio. La configurazione entra
**nella build**: Vite sostituisce ogni `import.meta.env.VITE_<NOME>` con il suo valore e lo
scrive dentro il JavaScript che ogni browser scarica. Da quel momento è un testo leggibile con
gli strumenti di sviluppo da chiunque apra la pagina.

La conseguenza va presa alla lettera: **nel sito non esistono segreti.** Una chiave di API, un
token di servizio, una password messi in una variabile `VITE_*` sono pubblicati, con la stessa
irreversibilità di un segreto committato — vedi
Segreti e configurazione, dove il criterio
«è già nelle mani di chi usa il prodotto?» qui ha sempre risposta sì.

Ciò che richiede un segreto si fa su un server: il backend del progetto, o un servizio che
espone solo quello che il sito può sapere. Le chiavi «pubbliche» che alcuni servizi
distribuiscono apposta per il browser sono un'altra cosa — sono configurazione, e lo dice il
servizio, non chi la mette nella build.

## 🔤 Il prefisso

Vite espone al codice solo le variabili che iniziano con `VITE_`, e questo è un vantaggio:
tutto il resto dell'ambiente della build — compresi i segreti che la CI usa per pubblicare —
resta fuori dal pacchetto per costruzione. Non si allarga il prefisso (`envPrefix`) per
comodità: è il filtro che impedisce a un segreto della CI di finire nel sito.

Il prefisso del progetto si aggiunge dopo: `VITE_<PREFISSO>_API_BASE_URL`. Quello che dice
Il prefisso delle variabili d'ambiente
vale anche qui, ed è il `CLAUDE.md` del progetto a dichiararlo.

## 🎯 Un modulo solo legge l'ambiente

`import.meta.env` compare in **un file**: `src/bootstrap/config.ts`. Legge le variabili, le
controlla e le espone come costanti tipizzate; il composition root le passa alle sorgenti, che
le **ricevono** e non le leggono.

-   **Si controlla all'avvio.** Una variabile mancante o malformata produce un errore che dice
    quale variabile e cosa ci si aspettava, al caricamento del modulo, invece di un `undefined`
    che arriva a una `fetch` e diventa una richiesta verso `undefined/orders`.
-   **Vale anche per ciò che non è una variabile del progetto**: `import.meta.env.BASE_URL`,
    che Vite ricava dall'opzione `base`, passa da `config.ts` come il resto, e il router lo
    riceve da lì.
-   **I tipi non mentono.** Tutto quello che arriva da `import.meta.env` è `string | undefined`;
    `config.ts` lo converte una volta (un URL, un numero, un insieme chiuso di valori) e chi
    usa la configurazione riceve il tipo giusto.
-   **Le variabili si dichiarano in `src/vite-env.d.ts`**, così un nome sbagliato in
    `config.ts` è un errore di `tsc` e non un `undefined` silenzioso.
-   **Un file `.env.example`** nella cartella `web/` elenca tutte le variabili con un valore
    d'esempio e un commento su cosa sono. I `.env` con i valori veri restano locali —
    il `.gitignore` di un progetto Vite lo prevede già per `.env.local` e
    `.env.*.local`.

La forma completa è in [Configurazione all'avvio](../examples/config.example.md).

## 🧪 Nei test

I test non leggono l'ambiente: costruiscono sorgenti e repository passando i valori che
servono. È la stessa ragione per cui le sorgenti ricevono l'URL invece di leggerlo, ed è ciò che
rende `config.ts` l'unico modulo da cui un test non dipende mai.

## 🏗️ Una build per ambiente

Poiché la configurazione è dentro il pacchetto, **una build vale per un ambiente**: quella di
prova e quella di produzione sono due build, con due insiemi di variabili. La pipeline lo
dichiara nel workflow di release, e non si prova a riscrivere un valore dentro i file già
costruiti: è un `sed` su JavaScript minimizzato, e la prima volta che il valore cambia
lunghezza o compare due volte, rompe il sito.

Un progetto che deve davvero cambiare configurazione **senza ricostruire** — la stessa
build promossa da un ambiente all'altro — la legge a runtime da un file servito accanto al
sito (`/config.json`), caricato da `main.tsx` prima del primo render. È una scelta del
progetto, che la dichiara nel suo `CLAUDE.md`; le regole di questo file valgono uguali: un
modulo solo la legge, la controlla e la tipizza, e niente di quello che contiene è un segreto.
