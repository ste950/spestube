# Design di riferimento

Le schermate di SpesTube come HTML statico, apribili nel browser senza server. Sono il
riferimento visivo: l'implementazione vive in `web/` e prende colori e misure dai token di
`web/src/shared/theme/`.

| File | Schermata |
| --- | --- |
| [`progetto.html`](./progetto.html) | Presentazione del progetto: problema, principi, perimetro |
| [`oggi.html`](./oggi.html) | Oggi: i video rilasciati e la fine del feed |
| [`guarda.html`](./guarda.html) | Guarda: player, motivo, note, schermata di stop |
| [`da-vedere.html`](./da-vedere.html) | Da vedere: coda con motivo e giorni rimanenti |
| [`canali.html`](./canali.html) | Canali e filtri |

## Token

| Ruolo | Valore |
| --- | --- |
| Principale | `#2557C7` (scuro `#1A3F94`, tinta `#E4ECFB`) |
| Sfondo | `#F3F4F7` |
| Superficie | `#FFFFFF` |
| Bordi | `#DCE0E8` (campi `#C5CBD6`, divisori `#ECEEF2`) |
| Testo | `#15181F` (secondario `#5A6170`, attenuato `#4A5160`) |
| Avviso scadenza | `#B8642C` (testo `#9A4C1A`) |
| Font | Inter per il testo, Space Grotesk per i titoli |
| Bersaglio tattile | almeno 44px |

## Differenze volute fra design e app

-   **Intestazione di Oggi**: non c'è. Data, titolo e "N video nuovi dai tuoi canali"
    occupavano spazio per dire ciò che la barra dice già: il numero dei video sta accanto
    alla voce Oggi.
-   **Miniature**: sempre a colori, come le pubblica YouTube. Il selettore "Nascoste / In
    B/N / Normali" del design non c'è, e i riquadri grigi delle schermate sono solo segnaposto.
-   **Filtro intelligente** (LLM): previsto per una versione successiva, non c'è.
-   **"Nascondi visualizzazioni e like"**: non è un'opzione, i contatori non esistono proprio.
-   **"Una volta a settimana"**: l'arrivo dei video è solo giornaliero, all'orario scelto.
-   **Etichetta di argomento** in Guarda: le raccolte per argomento arrivano dopo.
