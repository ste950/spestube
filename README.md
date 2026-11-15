# SpesTube

Self-hosted YouTube client without the algorithm: only the channels you choose, no Shorts,
no autoplay, a feed that ends. Go + React PWA.

**YouTube, solo quello che hai scelto di guardare.**

SpesTube è un client personale e self-hosted per YouTube. Toglie ciò che è progettato per
trattenerti (consigliati, autoplay, Shorts, scroll infinito) e lascia il valore vero: i
canali che segui, il player, le cose che impari.

> Client indipendente, non affiliato a YouTube. I video restano su YouTube e si guardano
> con il suo player ufficiale incorporato.

## Perché si chiama così

*Spes* in latino è la speranza. C'è una storia che racconta di due lupi che combattono
dentro ognuno di noi, uno fatto di paura e uno di speranza. Quale dei due vince? Quello che
nutri. SpesTube serve a scegliere cosa dare da mangiare alla testa.

## Cosa cambia rispetto a YouTube

| Su YouTube | In SpesTube |
| --- | --- |
| Consigliati dall'algoritmo | Solo i canali che scegli tu |
| Autoplay e "prossimo video" | Alla fine il video si ferma: "Fine del video" |
| Scroll infinito, feed che si rinnova | Il feed finisce: i video arrivano una volta al giorno |
| Shorts | Esclusi |
| Visualizzazioni e like | Non ci sono |
| Ricerca su tutto YouTube | Non c'è |

## Come funziona

1. Aggiungi i canali che vuoi seguire.
2. Ogni ora SpesTube legge i loro feed RSS, gratuiti e senza quota.
3. Una volta al giorno, all'orario che scegli (di default le 7:00), i video nuovi compaiono
   in **Oggi**. In fondo alla lista c'è scritto "Per oggi è tutto".
4. Guardi un video in **Guarda**, con il player ufficiale su `youtube-nocookie.com`. Quando
   finisce, si ferma.
5. Ciò che resta in Oggi viene archiviato al rilascio del giorno dopo.

Quando aggiungi un canale entrano solo i suoi 3 video più recenti: gli arretrati restano
fuori.

## Avvio con Docker

Servono Docker e Docker Compose.

```bash
git clone https://github.com/ste950/spestube.git
```

```bash
cd spestube
```

```bash
cp .env.example .env
```

```bash
docker compose up -d --build
```

Poi apri <http://localhost:8080>. Il database sta nel volume `spestube-data` e sopravvive
agli aggiornamenti dell'immagine.

### Il primo canale

Vai in **Canali** e incolla il link del canale nella forma
`https://www.youtube.com/channel/UC…`, oppure il suo id. L'id si trova sulla pagina del
canale: *Condividi canale*, poi *Copia ID canale*.

I video del canale arrivano in Oggi al prossimo orario di rilascio. Per vederli subito
durante una prova, imposta `SPESTUBE_DEV=true` nel file `.env`, riavvia con
`docker compose up -d` e premi **Rilascia ora** in Oggi. Nell'uso normale lascia
`SPESTUBE_DEV=false`: un bottone che fa arrivare video a comando è il contrario di ciò che
SpesTube vuole essere.

### Configurazione

| Variabile | Default | A cosa serve |
| --- | --- | --- |
| `SPESTUBE_PORT` | `8080` | La porta su cui apri l'app |
| `SPESTUBE_DB_PATH` | `data/spestube.db` | Dove sta il database (con Docker è sempre `/data/spestube.db`) |
| `SPESTUBE_RELEASE_TIME` | `07:00` | L'orario in cui arrivano i video nuovi |
| `SPESTUBE_YT_API_KEY` | vuota | Chiave della YouTube Data API v3, facoltativa |
| `SPESTUBE_DEV` | `false` | Attiva gli strumenti di sviluppo |
| `TZ` | `Europe/Rome` | Il fuso orario dell'orario di arrivo (solo Docker Compose) |
| `SPESTUBE_BIND` | `127.0.0.1` | L'indirizzo su cui ascolta la porta (solo Docker Compose) |

**La chiave API è facoltativa.** Senza chiave SpesTube funziona con i soli feed RSS. Con la
chiave puoi aggiungere un canale anche dal suo `@handle`.

### Accesso e sicurezza

SpesTube **non ha un login**. È pensata per stare dietro un reverse proxy che si occupa
dell'autenticazione, per esempio Authelia. Per questo la porta, di default, ascolta solo su
`127.0.0.1`: la raggiunge il proxy che gira sulla stessa macchina, non il resto della rete.

Se vuoi aprirla alla rete di casa senza proxy, imposta `SPESTUBE_BIND=0.0.0.0`: chiunque
sia sulla rete potrà usarla.

### Raspberry Pi e altre architetture

`docker compose up --build` costruisce l'immagine per la macchina su cui lo lanci, quindi
sul Raspberry Pi produce un'immagine `linux/arm64`. Per costruire entrambe le architetture
da un'altra macchina:

```bash
docker buildx build --platform linux/arm64,linux/amd64 -t spestube:latest .
```

### Installarla su iPhone

Apri SpesTube in Safari, tocca *Condividi* e poi *Aggiungi alla schermata Home*. Serve un
indirizzo HTTPS, che è quello che ti dà il reverse proxy.

## Sviluppo

Servono Go 1.27 e Node 22 o successivo.

Il backend, da `backend/`:

```bash
SPESTUBE_DEV=true go run ./cmd/spestube
```

Il frontend, da `web/`, in un altro terminale:

```bash
npm install
```

```bash
npm run dev
```

Il server di sviluppo del frontend inoltra le chiamate `/api` al backend sulla porta 8080.

I test:

```bash
cd backend && go test -race ./...
```

```bash
cd web && npm test -- --run
```

### Com'è fatto

- `backend/`: Go con la sola libreria standard e SQLite in Go puro (`modernc.org/sqlite`,
  senza cgo). Il frontend compilato entra nel binario con `go:embed`: il risultato è un
  file solo.
- `web/`: Vite, React e TypeScript, installabile come PWA.
- `docs/design/`: le schermate di riferimento.
- `ai/` e `CLAUDE.md`: le regole architetturali e i principi di prodotto per chi lavora
  al codice, persone e agenti AI.

## Stato del progetto

Funzionano: canali, download dei feed, rilascio giornaliero, Oggi, Guarda.

In arrivo: ingresso in "Da vedere" con un motivo, note sui video, pannello dei filtri,
durata dei video tramite API, import delle iscrizioni da Google Takeout.

## Limiti noti

- Il feed RSS di YouTube dà solo gli ultimi 15 video di un canale.
- Senza chiave API la durata dei video non è nota, quindi il filtro sulla durata minima
  non ha effetto.
- L'esclusione degli Shorts si appoggia a una playlist di YouTube non documentata
  (`UULF…`). Se smette di funzionare, SpesTube ripiega sul feed del canale, che contiene
  anche gli Shorts.

## Licenza

[AGPL-3.0](LICENSE).
