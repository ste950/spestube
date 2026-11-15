package library

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ste950/spestube/backend/internal/core/apperr"
)

// Limiti scelti perché il nome stia su una riga del telefono e la descrizione in un
// paragrafo: oltre, non è più un'etichetta ma un documento.
const (
	CollectionNameMax        = 80
	CollectionDescriptionMax = 500
)

// Collection è una raccolta di video scelta dall'utente, come le playlist di YouTube. Non
// scade e non chiede motivi: l'ordine e i nomi li decide chi la usa.
type Collection struct {
	ID          int64
	Name        string
	Description string
	Position    int // l'ordine delle raccolte, deciso dall'utente
	CreatedAt   time.Time
	UpdatedAt   time.Time
	// In sola lettura, li calcola l'archivio.
	VideoCount   int
	CoverVideoID string // il video aggiunto per ultimo, vuoto se la raccolta è vuota
}

// CollectionVideo è un video dentro una raccolta, con il momento in cui ci è entrato.
type CollectionVideo struct {
	Video
	AddedAt time.Time
}

// CollectionFields sono i campi che l'utente scrive, ripuliti e validati.
type CollectionFields struct {
	Name        string
	Description string
}

func NewCollectionFields(name, description string) (CollectionFields, error) {
	name = strings.Join(strings.Fields(name), " ")
	description = strings.TrimSpace(description)
	switch {
	case name == "":
		return CollectionFields{}, apperr.Invalid("collection_name_required", "dai un nome alla raccolta")
	case utf8.RuneCountInString(name) > CollectionNameMax:
		return CollectionFields{}, apperr.Invalid("collection_name_too_long", "il nome è troppo lungo")
	case utf8.RuneCountInString(description) > CollectionDescriptionMax:
		return CollectionFields{}, apperr.Invalid("collection_description_too_long", "la descrizione è troppo lunga")
	}
	return CollectionFields{Name: name, Description: description}, nil
}

// CheckOrder verifica che wanted sia un riordino di current: gli stessi id, ognuno una
// volta sola. Un ordine che ne dimentica uno o ne inventa un altro è un client che ha una
// lista vecchia, e applicarlo a metà lascerebbe posizioni sovrapposte.
func CheckOrder(current, wanted []int64) error {
	invalid := apperr.Conflict("collection_order_invalid",
		"le raccolte sono cambiate nel frattempo: ricarica e riprova")
	if len(current) != len(wanted) {
		return invalid
	}
	known := make(map[int64]bool, len(current))
	for _, id := range current {
		known[id] = true
	}
	for _, id := range wanted {
		if !known[id] {
			return invalid
		}
		delete(known, id)
	}
	return nil
}
