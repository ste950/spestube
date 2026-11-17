---
description: Skill per isolare un servizio o un formato di terze parti dietro una sorgente di data/ — interfaccia, payload, adattatore HTTP e test contro una fixture registrata.
goal: Confinare rete, formati e librerie altrui in un package, con errori tradotti in apperr e test che non toccano la rete.
---
# Skill: Integrazione di una Sorgente Esterna

## Obiettivo

Che la rete lenta, il formato che cambia e la libreria di terzi tocchino **un** package, e che
il resto del codice veda solo payload e errori con un genere.

## Quando usarla

-   Il backend deve leggere da un sito, un'API, un formato (RSS, HTML, OPML) o scrivere verso
    un servizio esterno (notifiche, webhook).
-   Una libreria di terze parti sta per essere importata fuori da `data/`.

## Procedura

1.  **Verifica il contesto**
    -   Leggi [Struttura del progetto](../../rules/project_structure.rule.md): l'interfaccia della
        sorgente sta in `data/<soggetto>/sources/`, non nel dominio.

2.  **Interfaccia e payload**
    -   Un'interfaccia con i soli metodi usati, e un tipo payload con la forma del dato come
        arriva — non l'entità del dominio.

3.  **Adattatore**
    -   Client HTTP con timeout, `User-Agent` esplicito, limite sulla lettura del corpo,
        `NewRequestWithContext`.
    -   Errori di rete e status inattesi → `apperr.Unavailable`; input rifiutato → `Invalid`.
    -   Se serve una libreria (un parser), la si importa **solo** qui.
    -   [Vedi Esempio](../../examples/external_source.example.md)

4.  **Composizione**
    -   Il repository del soggetto riceve la sorgente nel costruttore e converte payload in
        entità; il composition root costruisce l'adattatore con i valori della `Config`.

5.  **Test**
    -   Fixture registrata dalla fonte vera in `testdata/` accanto al test (rileggila: niente
        token, niente dati personali), servita da un `httptest.Server`. Casi: risposta buona,
        status di errore, corpo malformato, timeout.

## Vincoli

-   Nessun import della libreria di terzi fuori da `sources/`.
-   Nessun segreto nel codice o nelle fixture: le credenziali arrivano dalla `Config`.
-   Nessuna rete vera nei test.

## Verifica

-   In locale `gofmt -w .`; `go mod tidy` se è entrata una dipendenza, da giustificare nella PR.
