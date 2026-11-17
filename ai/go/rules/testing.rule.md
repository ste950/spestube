# Test

## 🧪 Tre livelli, una toolchain

Solo il package `testing` della libreria standard e `net/http/httptest`. Niente librerie di
asserzioni, niente generatori di mock: un confronto con `if got != want` e un `t.Errorf` che
dice entrambi i valori è tutto ciò che serve, e chiunque conosca Go lo legge.

| Livello | Dove | Cosa usa | Cosa dimostra |
|---|---|---|---|
| Dominio | `domain/<soggetto>/*_test.go` | fake dei repository | le regole di business, casi limite compresi |
| Data | `data/<soggetto>/*_test.go` | SQLite vero su file in `t.TempDir()` | le query, i vincoli, le migrazioni |
| API | `api/<risorsa>/*_test.go` | `httptest` + servizio vero su fake | il contratto HTTP: status, corpo, codici di errore |

Più il [test di architettura](../examples/architecture_test.example.md), che non testa un
package ma le dipendenze fra tutti.

## 📋 Table-driven

Quando lo stesso comportamento va provato su più ingressi, una tabella di casi con un nome
ciascuno e `t.Run(tc.name, ...)`. Il nome dice il caso, non il numero: quando fallisce,
`--- FAIL: TestMatch/regex_case_insensitive` è già metà della diagnosi. Vedi
[Test table-driven](../examples/table_test.example.md).

## 🎭 Fake, non mock

Un fake è un'implementazione vera e minima dell'interfaccia — una mappa in memoria al posto
della tabella. Si scrive una volta, in un package `<soggetto>test` dentro il soggetto
(`domain/<soggetto>/<soggetto>test/`), e si usa dai test del dominio e da quelli delle API.

Un mock che verifica «è stato chiamato `Save` una volta con questi argomenti» testa
l'implementazione invece del comportamento, e si rompe al primo refactoring che non cambia
niente di osservabile.

## ⏱️ Tempo e casualità

Il codice che dipende dall'ora riceve l'ora — un parametro `now time.Time`, o un campo
`now func() time.Time` nel servizio — invece di chiamare `time.Now()` dentro. Lo stesso per la
casualità. Un test che dipende dall'orologio vero fallisce a mezzanotte, o una volta su mille.

## ✅ Casi obbligatori

Per ogni regola di business: il caso nominale, **ogni** ramo di errore, i bordi (lista vuota,
zero, il valore al limite della soglia). Per ogni repository: lettura, scrittura, assenza,
violazione del vincolo di unicità. Per ogni endpoint: il caso buono e almeno un errore con il
suo `code`.

## 🚫 Divieti

-   Rete vera in un test: le sorgenti esterne si provano contro un `httptest.Server` che
    risponde con una fixture registrata.
-   `time.Sleep` per aspettare una goroutine: si sincronizza con un canale o un `WaitGroup`.
-   Test che dipendono dall'ordine di esecuzione o da file lasciati da un altro test.

## 🧭 Esempi

-   [Test table-driven](../examples/table_test.example.md)
-   [Test di un handler](../examples/handler_test.example.md)
-   [Test di architettura](../examples/architecture_test.example.md)
