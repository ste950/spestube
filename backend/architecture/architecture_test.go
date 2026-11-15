// Package architecture rende eseguibili le regole di dipendenza fra i livelli di
// internal/: un import vietato fa fallire go test come qualsiasi altro errore.
package architecture

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Gli import della libreria standard che fanno I/O: il dominio non li usa.
var ioPackages = []string{"database/sql", "net/http", "os", "os/exec", "net"}

type pkgImports struct {
	dir     string   // es. "internal/domain/catalog"
	imports []string // percorsi completi, come scritti nel sorgente
}

func TestLayerBoundaries(t *testing.T) {
	module := modulePath(t)
	pkgs := readImports(t)

	for _, p := range pkgs {
		layer, unit := layerOf(p.dir)
		for _, imp := range p.imports {
			if bad := violation(module, layer, unit, imp); bad != "" {
				t.Errorf("%s importa %s: %s", p.dir, imp, bad)
			}
		}
	}
}

// violation restituisce il motivo per cui l'import è vietato, o "" se è ammesso.
func violation(module, layer, unit, imp string) string {
	rel, internal := strings.CutPrefix(imp, module+"/internal/")
	if !internal {
		if layer == "domain" && !strings.HasSuffix(unit, "test") {
			for _, io := range ioPackages {
				if imp == io {
					return "il dominio non fa I/O"
				}
			}
			if strings.Contains(imp, ".") { // un modulo di terze parti: ha un dominio nel path
				return "il dominio non dipende da librerie esterne"
			}
		}
		return ""
	}
	target, targetUnit := layerOf("internal/" + rel)

	switch layer {
	case "core":
		if target != "core" {
			return "core/ non conosce i livelli applicativi"
		}
	case "domain":
		if target != "core" && target != "domain" {
			return "il dominio dipende solo da core/"
		}
		if target == "domain" && rootOf(targetUnit) != rootOf(unit) {
			return "due soggetti del dominio non si importano"
		}
	case "data":
		if target == "api" || target == "bootstrap" {
			return "data/ non conosce api/ né bootstrap/"
		}
		if target == "data" && rootOf(targetUnit) != rootOf(unit) && targetUnit != "migrations" {
			return "data/ importa solo il proprio soggetto"
		}
	case "api":
		if target == "data" || target == "bootstrap" {
			return "api/ parla solo con il dominio"
		}
		if target == "api" && unit != "" && targetUnit != unit {
			return "una risorsa non importa un'altra risorsa"
		}
	}
	return ""
}

// layerOf("internal/domain/catalog/catalogtest") = ("domain", "catalog/catalogtest").
func layerOf(dir string) (layer, unit string) {
	parts := strings.SplitN(strings.TrimPrefix(dir, "internal/"), "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}

func rootOf(unit string) string {
	root, _, _ := strings.Cut(unit, "/")
	return root
}

// readImports legge solo l'intestazione dei file (ImportsOnly): vede gli import
// ovunque siano scritti, e non scambia per un import una stringa in un commento.
func readImports(t *testing.T) []pkgImports {
	t.Helper()
	byDir := map[string]*pkgImports{}
	fset := token.NewFileSet()
	err := filepath.WalkDir("../internal", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		dir := filepath.ToSlash(strings.TrimPrefix(filepath.Dir(path), "../"))
		if byDir[dir] == nil {
			byDir[dir] = &pkgImports{dir: dir}
		}
		for _, s := range f.Imports {
			p, _ := strconv.Unquote(s.Path.Value)
			byDir[dir].imports = append(byDir[dir].imports, p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// Un test che non legge niente passa senza aver verificato nulla: un percorso
	// sbagliato deve fallire qui, non far diventare verde tutto.
	if len(byDir) == 0 {
		t.Fatal("nessun package trovato sotto ../internal")
	}
	out := make([]pkgImports, 0, len(byDir))
	for _, p := range byDir {
		out = append(out, *p)
	}
	return out
}

func modulePath(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("../go.mod")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(b), "\n") {
		if m, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.TrimSpace(m)
		}
	}
	t.Fatal("go.mod senza direttiva module")
	return ""
}
