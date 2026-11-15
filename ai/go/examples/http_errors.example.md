# Errori HTTP e JSON

`core/httpx` è l'unico posto che sceglie uno status code per un errore, e l'unico che scrive
JSON sulla risposta. Un errore interno si logga qui — dove smette di essere restituito — e al
client arriva un messaggio generico. Vedi [Gestione degli errori](../rules/error_handling.rule.md)
e [API HTTP](../rules/http_api.rule.md).

```go
// backend/internal/core/httpx/httpx.go

// Package httpx raccoglie la lettura e la scrittura JSON e la traduzione degli errori
// in risposte HTTP: l'unico posto che sceglie uno status code per un errore.
package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"example.com/backend/internal/core/apperr"
)

// 1 MiB: abbondante per qualsiasi corpo JSON di questa API, e un client che manda di
// più sta sbagliando qualcosa.
const maxBodyBytes = 1 << 20

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Lo status è già partito: resta solo da registrarlo.
		slog.Error("scrittura della risposta", "err", err)
	}
}

// DecodeJSON rifiuta i campi sconosciuti: un refuso nel nome di un campo diventa un 400
// invece di un valore ignorato in silenzio.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return apperr.Invalid("invalid_body", fmt.Sprintf("corpo della richiesta non valido: %v", err))
	}
	if dec.Decode(&struct{}{}) != io.EOF {
		return apperr.Invalid("invalid_body", "il corpo contiene più di un valore JSON")
	}
	return nil
}

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var statusByKind = map[apperr.Kind]int{
	apperr.KindNotFound:     http.StatusNotFound,
	apperr.KindInvalid:      http.StatusBadRequest,
	apperr.KindConflict:     http.StatusConflict,
	apperr.KindUnauthorized: http.StatusUnauthorized,
	apperr.KindUnavailable:  http.StatusBadGateway,
}

// WriteError è il punto in cui un errore smette di essere restituito: per questo è
// anche il punto in cui, se è interno, lo si scrive nel log.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Kind == apperr.KindInternal {
		slog.Error("errore interno", "method", r.Method, "path", r.URL.Path, "err", err)
		WriteJSON(w, http.StatusInternalServerError,
			errorBody{errorDetail{Code: "internal_error", Message: "errore interno"}})
		return
	}
	if ae.Kind == apperr.KindUnavailable {
		slog.Warn("dipendenza non disponibile", "path", r.URL.Path, "err", err)
	}
	WriteJSON(w, statusByKind[ae.Kind], errorBody{errorDetail{Code: ae.Code, Message: ae.Msg}})
}
```

La risposta di errore ha sempre la stessa forma, che il client può leggere senza sapere
quale endpoint ha chiamato:

```json
{ "error": { "code": "item_not_found", "message": "item inesistente" } }
```
