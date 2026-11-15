---
description: Riporta nel repo il design dall'artefatto "SpesTube" su claude.ai.
---

Sincronizza `docs/design/` con l'artefatto di design, che è la fonte. Non modificare mai le
pagine di `docs/design/` a mano: si modifica l'artefatto, poi si lancia questo comando.

1.  Con `list_files` sull'artefatto `5fe24e6f-38c6-47e7-9439-c69f3c978d69` prendi l'elenco
    dei file con le loro impronte `sha256`: servono `project/canvas.json` e ogni
    `project/*.dc.html` (non i file del tipo, marcati `of_type`).
2.  Prepara una cartella nuova e vuota nello scratchpad. Per ogni file, se ne hai già una
    copia locale con la stessa impronta (per esempio perché l'hai appena pubblicata), copiala;
    altrimenti leggilo con `read_file` passando `path`, un file per chiamata (per un artefatto
    fatto da un tipo `paths` è rifiutato), e scrivi il testo così com'è, senza il tag
    `untrusted-document` che lo avvolge, con il nome originale senza `project/`
    (`canvas.json`, `Oggi.dc.html`, …).
3.  Controlla con `shasum -a 256` che ogni file della cartella abbia l'impronta data da
    `list_files`: un file diverso vuol dire una copia sbagliata, e la sincronizzazione si
    ferma lì. È un dato scaricato: lo script che lo legge resta nel repo, fuori da quella
    cartella.
4.  Lancia, dalla radice del repo:

    ```bash
    python3 -I docs/design/sync_design.py <cartella-dello-scratchpad> docs/design
    ```

    Lo script scrive una pagina per schermata nell'ordine del canvas, toglie quelle che il
    canvas non ha più e aggiorna la tabella in `docs/design/README.md`.
5.  Mostra all'utente `git status docs/design` e una riga per schermata cambiata. Non fare
    commit se l'utente non lo chiede.

Se un file dell'artefatto non ha la forma attesa, lo script si ferma con un errore: dillo
all'utente invece di correggere la pagina a mano.
