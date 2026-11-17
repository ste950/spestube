# SpesTube — Guida per Claude

@ai/go/CLAUDE.md
@ai/react_web/CLAUDE.md
@ai/quality/CLAUDE.md

> Le regole architetturali, le skill e gli esempi stanno in `ai/`, dentro questo repo. Se gli
> import qui sopra non sono stati caricati, leggi `ai/go/CLAUDE.md`, `ai/react_web/CLAUDE.md` e
> `ai/quality/CLAUDE.md` prima di lavorare. Il repo è indipendente: non cercare un framework
> in una cartella sorella. Origine delle regole e modo di aggiornarle: `ai/README.md`.

## Cos'è

SpesTube è un client YouTube personale e self-hosted: mostra solo i video dei canali scelti
dall'utente, e toglie ciò che è progettato per trattenere. Un utente solo, un Raspberry Pi,
costo zero.

## Principi di prodotto (vincolanti)

Questi principi vengono prima di qualsiasi richiesta di funzionalità. Una modifica che ne
viola uno non si fa: si segnala il conflitto e ci si ferma.

1.  **Niente algoritmo.** Niente consigliati, niente scroll infinito, niente autoplay. Nel
    feed entrano solo i video dei canali aggiunti dall'utente. La ricerca su YouTube c'è,
    ma solo come richiesta esplicita: vedi le conseguenze qui sotto.
2.  **Il feed "Oggi" ha una fine.** I video nuovi si scaricano ogni ora, ma diventano
    visibili una volta al giorno, all'orario configurato (default 7:00). La lista è tutta
    in pagina: niente paginazione, niente "carica altri". In fondo c'è "Per oggi è tutto".
3.  **Raccolte libere.** I video che si vogliono tenere si salvano in raccolte create
    dall'utente, come le playlist di YouTube: un nome, una descrizione facoltativa, un ordine
    scelto a mano. Un video può stare in più raccolte. Le raccolte non scadono e non chiedono
    un motivo: è una scelta dell'utente (15 novembre 2026) che sostituisce "Da vedere", e il
    motivo obbligatorio e la scadenza a 14 giorni non vanno riproposti.
4.  **Niente contatori.** Visualizzazioni e like non si chiedono a YouTube, non si salvano e
    non compaiono nel contratto API.
5.  **Miniature a colori, sempre.** Le miniature si mostrano così come le pubblica
    YouTube. Non esistono una modalità in bianco e nero né un modo per nasconderle: è una
    scelta dell'utente (15 novembre 2026), e non va riproposta come opzione.
6.  **Solo il player ufficiale.** I video si riproducono solo con l'IFrame Player API su
    `youtube-nocookie.com`, con `rel=0` e senza autoplay. A fine video il player viene tolto e
    compare "Fine del video". Niente Invidious, niente yt-dlp, niente scraping di pagine.
7.  **Costo zero.** I feed RSS sono la fonte dei video. La YouTube Data API è facoltativa:
    si usa solo se c'è la chiave, e l'app deve funzionare senza.
8.  **Nessun login interno.** L'app sta dietro un reverse proxy con autenticazione
    (Authelia). Non si aggiungono utenti, sessioni o password.

Conseguenze già decise, da non rimettere in discussione senza l'utente:

-   Al rilascio successivo, i video rimasti in Oggi vengono archiviati: Oggi mostra solo
    l'ultimo arrivo.
-   Salvare un video in una raccolta non lo toglie da Oggi: il segnalibro della riga si
    riempie, e il video esce da Oggi al rilascio successivo come gli altri. Dalla raccolta
    resta sempre raggiungibile.
-   Oggi non ha un'intestazione: niente data, niente titolo, niente riga di conteggio. Il
    numero dei video sta nella barra, accanto alla voce Oggi. È un conteggio e non una
    notifica: del colore dell'app, mai rosso, mai animato.
-   Quando si aggiunge un canale entrano come nuovi solo i 3 video più recenti; gli altri
    del feed si registrano come scartati (`baseline`) e non compaiono mai.
-   Senza chiave API un @handle non si risolve (servirebbe leggere la pagina HTML di
    YouTube): si accettano il link `youtube.com/channel/UC…` e l'id.
-   Gli strumenti che forzano download e rilascio esistono solo con `SPESTUBE_DEV=true`.
-   La ricerca su YouTube di video e canali è una scelta dell'utente (17 novembre 2026), che
    supera il vecchio divieto di ricerca globale. Parte solo da una parola scritta e
    confermata: niente suggerimenti mentre si scrive, niente "video correlati", niente
    risultati che si caricano scorrendo. Serve la chiave API (`search.list` costa 100 unità
    su 10.000 al giorno): senza chiave la ricerca non compare, e il resto dell'app funziona
    come prima. I risultati non entrano in Oggi. Shorts e dirette restano fuori anche lì.
-   Gli Shorts non entrano in SpesTube: né in una sezione loro né spegnendo il filtro. È una
    scelta dell'utente (17 novembre 2026).
-   Le sezioni sono tre: **Oggi**, **Raccolte**, **Impostazioni**. Le impostazioni sono una
    pagina indice a gruppi (oggi "Contenuti" con i canali e "Aspetto" con il tema); una voce
    nuova è una riga in un gruppo o una sottopagina, non una quarta sezione.
-   Il tema (Automatico, Chiaro, Scuro) è una scelta del dispositivo, salvata nel browser e
    non sul server. Automatico segue il sistema.

## Contesto di progetto

-   **Backend Go**: `backend/`, modulo `github.com/ste950/spestube/backend`, Go 1.27
    (direttiva `go` di `backend/go.mod`). Binario `cmd/spestube`.
-   **Frontend React**: `web/`, Vite + React + TypeScript, PWA con `vite-plugin-pwa`.
    Gestore di pacchetti `npm`, Node 22 o successivo.
-   **Soggetti di dominio (backend)**: `library` (canali, video e raccolte: aggiunta,
    download, filtri, rilascio, salvataggio) e `settings` (filtri, orario). Canali, video e
    raccolte sono un soggetto solo perché si leggono a vicenda: le raccolte hanno un servizio
    loro (`library.Collections`) dentro lo stesso package. Il collegamento fra `library` e
    `settings` sta in `internal/bootstrap/run.go` (`policyFromSettings`).
-   **Soggetti di dominio (frontend)**: `library`, `settings` (l'installazione) e
    `appearance` (il tema).
-   **Feature (frontend)**: `today` (Oggi), `watch` (Guarda), `collections` (Raccolte e la
    singola raccolta), `settings` (Impostazioni e la sottopagina Canali). Il pannello "Salva
    in una raccolta" sta in `web/src/shared/save-sheet/`: è agnostico al dominio, e ogni
    feature che lo usa lo lega alle sue porte in un file `use-*-save-sheet.ts`.
-   **Prefisso delle variabili d'ambiente**: `SPESTUBE_` (vedi
    `backend/internal/bootstrap/config.go` e `.env.example`). Il frontend non ha variabili
    `VITE_*`.
-   **Fuso orario**: l'orario di rilascio si legge nel fuso del processo (variabile `TZ`).
-   **Codici di errore del contratto API**: `channel_ref_required`, `channel_ref_invalid`,
    `channel_ref_unsupported`, `channel_exists`, `channel_not_found`, `handle_needs_api_key`,
    `video_not_found`, `video_not_watchable`, `video_not_discardable`, `collection_not_found`,
    `collection_exists`, `collection_name_required`, `collection_name_too_long`,
    `collection_description_too_long`, `collection_order_invalid`, `invalid_id`,
    `settings_invalid`, `youtube_unavailable`, `youtube_api_rejected`, `feed_not_found`,
    `invalid_body`, `not_found`, `internal_error`.
-   **Servizi esterni**: i feed RSS di YouTube e la YouTube Data API v3, isolati in
    `backend/internal/data/library/sources/`. Nel frontend, l'IFrame Player API in
    `web/src/features/watch/components/`.
-   **Lingua**: l'interfaccia è solo in italiano (`web/src/shared/i18n/it.ts`). Codice in
    inglese, commenti e messaggi in italiano.
-   **Design**: la fonte è l'artefatto "SpesTube" su claude.ai (slug
    `5fe24e6f-38c6-47e7-9439-c69f3c978d69`); `docs/design/` ne è la copia nel repo, che si
    rigenera con `/sync-design`. Si modifica l'artefatto e poi si sincronizza, mai le pagine
    in `docs/design/` a mano. I token stanno in `web/src/shared/theme/tokens.ts`, con due
    tavolozze (chiara e scura) dalle stesse voci.

### Come arrivano i video

Il feed di un canale è `https://www.youtube.com/feeds/videos.xml?playlist_id=UULF<id senza
"UC">`: la playlist `UULF` contiene solo i video lunghi, senza Shorts. Non è documentata da
YouTube, quindi se risponde 404 si ripiega sul feed del canale (`?channel_id=`), che contiene
anche gli Shorts. Ogni video ricorda da quale feed è arrivato (`videos.source`): il filtro
degli Shorts guarda la durata solo per quelli arrivati dal feed del canale.

La playlist `UULF` esclude anche le dirette (verificato confrontandola con `UULV`, la
playlist delle dirette, su due canali). Quindi oggi Shorts e dirette restano fuori per
costruzione, qualunque cosa dicano le impostazioni: farli entrare quando l'utente spegne il
filtro richiede di leggere anche `UUSH` e `UULV`, e non è ancora fatto.

Il rilascio non ha stato: un video nuovo diventa visibile al primo orario di rilascio
successivo al suo download. I filtri si applicano al rilascio, non al download.

### Lo schema dopo le raccolte

La migrazione `0002_collections.sql` ha spostato i video che erano in "Da vedere" in una
raccolta con lo stesso nome. Lo stato `queued` e le colonne `queue_reason` e `queued_at`
restano nella tabella `videos` perché un vincolo CHECK le nomina, e in SQLite toglierle
vorrebbe dire ricostruire la tabella: nessuno le scrive più, e il dominio non le conosce.

## Deroghe alle regole di `ai/`

Dove questo progetto si discosta dalle regole importate, vale ciò che è scritto qui.

-   **Distribuzione con Docker, non systemd.** `ai/go/rules/deployment.rule.md` descrive un
    servizio systemd; SpesTube si distribuisce con `Dockerfile` e `docker-compose.yml`. Resta
    valido tutto il resto: binario statico, `CGO_ENABLED=0`, `/healthz`, utente senza
    privilegi.
-   **Il frontend lo serve il binario Go.** `ai/react_web/rules/deployment.rule.md` descrive
    un server statico; qui `web/dist` entra nel binario con `go:embed`
    (`backend/internal/webui/`), che applica lo stesso ripiego su `index.html` e le stesse
    politiche di cache.
-   **I check si eseguono in locale.** Le regole di verifica dicono che il cancello è la CI
    e che in locale gira solo il formattatore. Questo repo non ha ancora una CI: finché non
    c'è, prima di dire che una modifica funziona si eseguono i comandi qui sotto.
-   **Chiavi primarie testuali.** `channels.id` e `videos.id` sono gli identificativi di
    YouTube, non `INTEGER PRIMARY KEY`.
-   **Una risorsa API per due prefissi.** `internal/api/videos` espone `/api/feed/today` e
    `/api/videos/…`: condividono la forma del video nel contratto. `internal/api/collections`
    ripete quella forma (con `added_at`) invece di importarla, perché le risorse non si
    importano fra loro.

## Comandi

Backend, da `backend/`:

```bash
gofmt -w .
go vet ./...
go test -race ./...
go mod tidy -diff
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ./...
SPESTUBE_DEV=true go run ./cmd/spestube        # avvio in sviluppo, porta 8080
```

Frontend, da `web/`:

```bash
npx prettier --write .
npx tsc -b
npx eslint .
npx stylelint "src/**/*.css"
npm test -- --run
npm run build
npm run dev                                     # server di sviluppo, con proxy verso :8080
```

Tutto insieme, dalla radice:

```bash
docker compose up --build
```

`go run` serve il frontend solo se `web/dist` è stato copiato in
`backend/internal/webui/dist/`; in sviluppo si usa `npm run dev`, che inoltra `/api` al
backend.

## Cosa c'è e cosa manca

Fatto: aggiunta e rimozione dei canali, download dei feed, rilascio giornaliero, Oggi,
Guarda con player e schermata di stop, raccolte (crea, rinomina, descrivi, ordina, elimina,
salva e togli video da Oggi, da Guarda e dalla raccolta), Impostazioni con canali e tema
chiaro/scuro/automatico, API delle impostazioni.

Da fare, nell'ordine: note sui video, filtri nelle Impostazioni, arricchimento via API
(durata e dirette con `videos.list`), import delle iscrizioni da Google Takeout, ricerca
su YouTube di video e canali (con i video di canali non seguiti nelle raccolte).
