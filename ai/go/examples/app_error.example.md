# Errore applicativo

Il tipo unico che attraversa i livelli: un `Kind` che decide lo status HTTP, un `Code` stabile
che fa parte del contratto API, il messaggio e la causa. Vedi
[Gestione degli errori](../rules/error_handling.rule.md).

```go
// backend/internal/core/apperr/apperr.go

// Package apperr definisce l'errore applicativo: un genere che decide lo status HTTP e
// un codice stabile che fa parte del contratto API.
package apperr

import "errors"

// Kind è il genere dell'errore. Lo legge solo chi traduce l'errore verso l'esterno.
type Kind int

const (
	KindInternal Kind = iota // il default: tutto ciò che non è stato classificato
	KindNotFound
	KindInvalid
	KindConflict
	KindUnauthorized
	KindUnavailable
)

// Error è l'errore che attraversa i livelli. Code è il codice del contratto API
// (es. "item_not_found") e non si rinomina: il client lo confronta.
type Error struct {
	Kind Kind
	Code string
	Msg  string
	Err  error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Msg + ": " + e.Err.Error()
	}
	return e.Msg
}

func (e *Error) Unwrap() error { return e.Err }

func NotFound(code, msg string) *Error { return &Error{Kind: KindNotFound, Code: code, Msg: msg} }
func Invalid(code, msg string) *Error  { return &Error{Kind: KindInvalid, Code: code, Msg: msg} }
func Conflict(code, msg string) *Error { return &Error{Kind: KindConflict, Code: code, Msg: msg} }
func Unauthorized(code, msg string) *Error {
	return &Error{Kind: KindUnauthorized, Code: code, Msg: msg}
}

// Unavailable avvolge l'errore di una dipendenza esterna (rete, servizio di terzi):
// la causa resta raggiungibile per il log, ma al client arriva solo il codice.
func Unavailable(code, msg string, err error) *Error {
	return &Error{Kind: KindUnavailable, Code: code, Msg: msg, Err: err}
}

// KindOf trova il genere anche attraverso più livelli di fmt.Errorf("...: %w", err).
func KindOf(err error) Kind {
	var e *Error
	if errors.As(err, &e) {
		return e.Kind
	}
	return KindInternal
}
```

Uso:

```go
// nel dominio: un errore con un significato per l'utente
return Item{}, apperr.Conflict("item_exists", "l'indirizzo è già salvato")

// a metà strada: contesto in più, catena intatta
return fmt.Errorf("aggiornamento del feed %d: %w", id, err)

// in chi decide: il genere, attraverso tutti gli avvolgimenti
if apperr.KindOf(err) == apperr.KindUnavailable { /* si riprova al prossimo giro */ }
```
