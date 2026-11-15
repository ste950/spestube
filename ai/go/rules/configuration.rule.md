# Configurazione

## 🎯 Una struttura, caricata una volta

Tutta la configurazione sta in una `Config` di `bootstrap/`, caricata **una volta** all'avvio
e passata a chi ne ha bisogno — come valore, o come i soli campi che servono. Nessun altro
package legge l'ambiente.

```go
// ❌ letta dove serve: invisibile, non testabile, e il default sta in un posto qualunque
timeout := os.Getenv("APP_FETCH_TIMEOUT")

// ✅ un campo della Config, con il suo default e la sua validazione
fetcher := feeds.NewFetcher(client, cfg.FetchTimeout)
```

Vedi [Config](../examples/config.example.md).

## 🌍 Da dove arriva

-   **Variabili d'ambiente con un prefisso del progetto** (`<PREFISSO>_DB_PATH`), dichiarato
    nel `CLAUDE.md` del progetto. È l'interfaccia che systemd, un container e la CI parlano già.
-   **Il «file di configurazione» è un file di ambiente** (`KEY=VALUE`, una per riga): systemd
    lo carica con `EnvironmentFile=`, e in sviluppo lo legge il binario con un flag `-env`.
    Un formato solo, nessuna libreria di parsing, nessuna precedenza fra due sorgenti da
    ricordare.
-   **I segreti** (password, chiavi) arrivano dallo stesso canale ma da un file con permessi
    `0600` che non sta nel repo. Nel repo c'è solo l'esempio, con valori finti —
    Segreti e configurazione.

## ✅ Validazione all'avvio

`Load` restituisce un errore se un valore è malformato o se manca un valore obbligatorio, e il
binario **non parte**. Un errore all'avvio, con il nome della variabile, si corregge in un
minuto; lo stesso errore scoperto alla prima richiesta che usa quel valore si corregge dopo
averlo cercato.

-   Durate come `time.Duration` (`30s`, `5m`), percorsi come stringhe assolute o relative alla
    directory di stato, liste separate da virgola.
-   Ogni campo ha un **default sensato** dove esiste; quelli che non possono averne uno (una
    password) sono obbligatori e lo dicono nel messaggio.
-   La `Config` non contiene logica: è un dato. Le decisioni che dipendono dalla configurazione
    stanno nel codice che la riceve.

## 🧪 Nei test

I test costruiscono la `Config` (o i soli valori) a mano. Un test che chiama `t.Setenv` per
configurare il codice sotto test sta testando il caricamento, non il comportamento: si fa una
volta, nel test di `Load`.

## 🧭 Esempi

-   [Config](../examples/config.example.md)
-   [Unit systemd](../examples/systemd_unit.example.md): `EnvironmentFile=` in produzione.
