// Package migrations contiene lo schema del database, un file per migrazione.
package migrations

import "embed"

// FS entra nel binario: chi lo avvia ha sempre lo schema che il codice si aspetta.
//
//go:embed *.sql
var FS embed.FS
