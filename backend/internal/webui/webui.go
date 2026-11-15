// Package webui contiene il frontend compilato e lo serve: è ciò che rende il backend
// un binario unico. La cartella dist/ la riempie la build del frontend; nel repo c'è
// solo un segnaposto.
package webui

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
)

//go:embed all:dist
var embedded embed.FS

// Il manifest della PWA ha un'estensione che la libreria standard non conosce: senza
// questa riga uscirebbe come testo semplice, e non tutti i browser lo accettano così.
func init() {
	if err := mime.AddExtensionType(".webmanifest", "application/manifest+json"); err != nil {
		panic(err)
	}
}

// Handler serve i file del frontend incorporati nel binario.
func Handler() http.Handler {
	dist, err := fs.Sub(embedded, "dist")
	if err != nil {
		// Non può succedere: dist è il percorso dell'embed qui sopra.
		panic(err)
	}
	return NewHandler(dist)
}

// NewHandler serve una cartella di file statici costruiti da Vite, con il ripiego su
// index.html per i percorsi che esistono solo nel router del client.
func NewHandler(dist fs.FS) http.Handler {
	files := http.FileServerFS(dist)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "metodo non ammesso", http.StatusMethodNotAllowed)
			return
		}
		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if name == "" {
			name = "index.html"
		}

		if info, err := fs.Stat(dist, name); err == nil && !info.IsDir() {
			// I file sotto assets/ hanno l'hash nel nome: non cambiano mai, e il browser
			// può tenerli per sempre. Tutto il resto (index.html, il service worker, il
			// manifest) va ricontrollato a ogni visita, o un aggiornamento non arriva.
			if strings.HasPrefix(name, "assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				w.Header().Set("Cache-Control", "no-cache")
			}
			files.ServeHTTP(w, r)
			return
		}

		// Un file che manca resta un 404: rispondere con index.html a una richiesta di
		// script nasconderebbe l'errore dietro una pagina che non si carica.
		if path.Ext(name) != "" && name != "index.html" {
			http.NotFound(w, r)
			return
		}
		index, err := fs.ReadFile(dist, "index.html")
		if err != nil {
			http.Error(w, "frontend non compilato: esegui la build di web/ e copia web/dist in backend/internal/webui/dist",
				http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(index)
	})
}
