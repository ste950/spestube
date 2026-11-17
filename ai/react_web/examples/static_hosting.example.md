La configurazione di un server di file statici per `web/dist/`, con le tre cose che le regole di
[Distribuzione](../rules/deployment.rule.md) chiedono: il fallback su `index.html`, due
politiche di cache opposte e le intestazioni di sicurezza. È scritta per nginx perché è il caso
più comune; qualsiasi altro server ha le stesse tre direttive con un altro nome.

```nginx
# web/deploy/nginx.conf

server {
    listen 80;
    root /usr/share/nginx/html;

    # Intestazioni valide per ogni risposta. `always` le aggiunge anche agli errori, che sono
    # pagine come le altre.
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    # Le sole origini da cui il sito carica codice e dati. `connect-src` va allargato all'API
    # del progetto, e nient'altro: una dipendenza che chiama un'origine non dichiarata fallisce
    # qui, e lo si vede.
    add_header Content-Security-Policy "default-src 'self'; connect-src 'self' <origine_api>; img-src 'self' data:; style-src 'self' 'unsafe-inline'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'" always;

    # I file con l'hash nel nome non cambiano mai: un contenuto nuovo ha un nome nuovo.
    location /assets/ {
        add_header Cache-Control "public, max-age=31536000, immutable";
        # Un `add_header` dentro una `location` sostituisce quelli del blocco esterno invece di
        # aggiungersi. Qui si ripete solo `nosniff`, l'unica che conta per uno script o un
        # foglio: la politica dei contenuti vale per i documenti.
        add_header X-Content-Type-Options "nosniff" always;
        try_files $uri =404;
    }

    # Tutto il resto: il file, se esiste; altrimenti `index.html`, e decide il router. È il
    # fallback senza cui una pagina interna ricaricata risponde 404.
    location / {
        add_header Cache-Control "no-cache";
        add_header X-Content-Type-Options "nosniff" always;
        add_header Referrer-Policy "strict-origin-when-cross-origin" always;
        add_header Content-Security-Policy "default-src 'self'; connect-src 'self' <origine_api>; img-src 'self' data:; style-src 'self' 'unsafe-inline'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'" always;
        try_files $uri /index.html;
    }
}
```

Se il sito è servito in un container, l'immagine è il server con dentro la cartella costruita
dalla CI — la build non avviene nell'immagine, perché il pacchetto da pubblicare è quello che
la CI ha già verificato:

```dockerfile
# web/deploy/Dockerfile

FROM nginx:<versione>-alpine
COPY deploy/nginx.conf /etc/nginx/conf.d/default.conf
COPY dist/ /usr/share/nginx/html/
```

Lo stesso con Caddy, per confronto — `try_files` e le due politiche di cache sono le stesse tre
righe:

```text
<host> {
    root * /srv/web
    header /assets/* Cache-Control "public, max-age=31536000, immutable"
    header /index.html Cache-Control "no-cache"
    try_files {path} /index.html
    file_server
}
```

Punti chiave:

-   **`try_files $uri /index.html`** è il fallback: un file che esiste si serve, un percorso
    interno riceve la pagina e il router decide. Un percorso che nessuna feature dichiara
    arriva alla pagina dei percorsi sconosciuti del sito, che risponde comunque 200 — il server
    non sa quali percorsi esistano, lo sa solo il router.
-   **`/assets/` con `immutable`**, `index.html` con `no-cache`: il primo non si riscarica mai,
    il secondo si ricontrolla sempre, e dopo un rilascio il visitatore riceve la versione nuova
    al primo caricamento.
-   **Le intestazioni si ripetono in ogni `location`** che ne aggiunge una sua: è la trappola
    di nginx che lascia una pagina senza `Content-Security-Policy` senza che nessuno se ne
    accorga.
-   **`<origine_api>`** è il segnaposto dell'API del progetto, l'unica origine che il sito
    chiama oltre a sé stesso. Una politica con `connect-src *` non verifica niente.
-   `style-src 'unsafe-inline'` serve al foglio del tema che `main.tsx` inserisce all'avvio
    ([Token di tema](./theme_tokens.example.md)): è l'unico stile non servito come file. Gli
    script restano su `'self'`, senza eccezioni.
-   La ricarica su una rotta profonda è il controllo che dice se questa configurazione è
    stata applicata: è nella [verifica nel browser](../rules/verification.rule.md#%EF%B8%8F-verifica-nel-browser).
