---
description: Skill per scrivere i test ai tre livelli — dominio con fake, data su SQLite vero, API con httptest — più il test di architettura, con la sola libreria standard.
goal: Coprire regole di business, query e contratto HTTP con test leggibili, deterministici e senza librerie di asserzioni o mock.
---
# Skill: Creazione dei Test

## Obiettivo

Test che dicono cosa si è rotto dal nome del caso, che non dipendono dall'ora o dalla rete, e
che sopravvivono a un refactoring che non cambia il comportamento.

## Quando usarla

-   Ogni task che aggiunge o cambia un comportamento.
-   Un bug: prima il test che lo riproduce, poi la correzione.

## Procedura

1.  **Verifica il contesto**
    -   Leggi [Test](../../rules/testing.rule.md) e scegli il livello: dominio, data o API.

2.  **Dominio**
    -   `<file>_test.go` accanto al servizio, package `<soggetto>_test`. Tabella di casi con
        nome, fake da `<soggetto>test`, orologio fisso.
    -   [Vedi Esempio](../../examples/table_test.example.md)

3.  **Data**
    -   Database vero in `t.TempDir()` con le migrazioni vere; casi obbligatori: lettura,
        scrittura, assenza, unicità.
    -   [Vedi Esempio](../../examples/table_test.example.md)

4.  **API**
    -   `httptest.Server` sulla risorsa vera; status, corpo e `error.code`.
    -   [Vedi Esempio](../../examples/handler_test.example.md)

5.  **Sorgenti esterne**
    -   `httptest.Server` con una fixture in `testdata/`.
    -   [Vedi Esempio](../../examples/external_source.example.md)

6.  **Architettura**
    -   Se nasce un livello o un soggetto nuovo con regole proprie, si estende il test di
        architettura invece di affidarsi alla memoria.
    -   [Vedi Esempio](../../examples/architecture_test.example.md)

## Vincoli

-   Solo `testing` e `httptest`: niente librerie di asserzioni o di mock.
-   `t.Fatalf` quando continuare non ha senso, `t.Errorf` per raccogliere più differenze; il
    messaggio dice ottenuto e atteso.
-   Niente `time.Sleep`, niente rete, niente ordine fra i test.

## Verifica

-   Comando di lavoro: `go test ./<package>/ -run <NomeTest> -v`.
-   La suite completa, con `-race`, gira in CI.
