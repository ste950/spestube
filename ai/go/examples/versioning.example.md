# Versione

La regola trasversale
Versionamento e Release dice che il numero si
alza una volta sola per release, in un commit dedicato, e che deve essere **derivabile** dai
commit. Non dice dove vive: quello lo dichiara lo stack. Go non ha un manifest con un campo
versione — `go.mod` non ne ha — quindi il posto è un file di una riga.

## 📍 Un solo posto: `backend/version.txt`

```
0.1.0
```

È il formato che `release-please` gestisce con `release-type: simple`, ed è l'unico punto del
repo in cui il numero compare. Il `<comando-che-stampa-la-versione>` del
workflow di release è `cat backend/version.txt`.

## 🏃 Come entra nel binario

Con `-ldflags -X` alla build, su una variabile di `main`:

```go
// backend/cmd/<binario>/main.go

// version la imposta la build con -ldflags "-X main.version=…".
var version = "dev"
```

```bash
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath \
  -ldflags "-s -w -X main.version=$(cat version.txt)" \
  -o dist/<binario> ./cmd/<binario>
```

-   **`dev`** è il valore di un `go run` o di un `go build` senza flag: un binario che dice
    `dev` in `/healthz` non viene da una release, ed è giusto che si veda.
-   **Non un `const` in un file `.go`**: sarebbe un secondo numero da alzare insieme a
    `version.txt`, e i due divergerebbero alla prima release fatta a mano.
-   `main` passa la versione a `bootstrap.Run`, che la espone su `/healthz` e la usa nello
    `User-Agent` delle richieste in uscita.

## 🤖 Chi lo alza

Nella PR di un task il numero **non si tocca mai**. Con l'automazione lo alza `release-please`:

```json
{
  "packages": {
    "backend": { "release-type": "simple", "component": "backend" }
  }
}
```

`simple` aggiorna `version.txt` e `CHANGELOG.md` nella cartella del package, e il tag della
release resta quello della regola trasversale.
