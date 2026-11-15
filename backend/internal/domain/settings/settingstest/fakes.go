// Package settingstest contiene il fake dell'archivio delle impostazioni.
package settingstest

import (
	"context"
	"time"
)

type Repository struct {
	Values map[string]string
}

func NewRepository() *Repository {
	return &Repository{Values: map[string]string{}}
}

func (r *Repository) All(context.Context) (map[string]string, error) {
	out := make(map[string]string, len(r.Values))
	for k, v := range r.Values {
		out[k] = v
	}
	return out, nil
}

func (r *Repository) Put(_ context.Context, values map[string]string, _ time.Time) error {
	for k, v := range values {
		r.Values[k] = v
	}
	return nil
}
