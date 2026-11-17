# Design di riferimento

Le schermate di SpesTube come HTML statico, apribili nel browser senza server.

**La fonte è l'artefatto "SpesTube" su claude.ai** (slug
`5fe24e6f-38c6-47e7-9439-c69f3c978d69`): il design si modifica lì, dove si vede e si commenta
sul canvas. Questa cartella ne è la copia dentro il repo, perché il repo resti leggibile
anche senza claude.ai. Si rigenera con il comando `/sync-design` di Claude Code, che scarica
le schermate dell'artefatto e le passa a [`sync_design.py`](./sync_design.py). Le pagine
`.html` qui dentro non si modificano a mano: la prossima sincronizzazione le riscriverebbe.

L'implementazione vive in `web/` e prende colori e misure dai token di
`web/src/shared/theme/tokens.ts`.

## Schermate

<!-- schermate:inizio -->

| File | Schermata |
| --- | --- |
| [`progetto.html`](./progetto.html) | Il progetto |
| [`oggi.html`](./oggi.html) | Oggi |
| [`guarda.html`](./guarda.html) | Guarda |
| [`salva.html`](./salva.html) | Salva in una raccolta |
| [`raccolte.html`](./raccolte.html) | Raccolte |
| [`raccolta.html`](./raccolta.html) | Raccolta |
| [`impostazioni.html`](./impostazioni.html) | Impostazioni |
| [`canali.html`](./canali.html) | Canali |
| [`oggi-scuro.html`](./oggi-scuro.html) | Oggi, tema scuro |

<!-- schermate:fine -->

## Token

| Ruolo | Chiaro | Scuro |
| --- | --- | --- |
| Principale | `#2557C7` (scuro `#1A3F94`, tinta `#E4ECFB`) | `#8AADF4` (chiaro `#B9CDF8`, tinta `#1D2A46`) |
| Sfondo | `#F3F4F7` | `#0F1218` |
| Superficie | `#FFFFFF` | `#181C24` |
| Bordi | `#DCE0E8` (campi `#C5CBD6`, divisori `#ECEEF2`) | `#2A303C` (campi `#3D4555`, divisori `#232833`) |
| Testo | `#15181F` (secondario `#5A6170`, attenuato `#4A5160`) | `#E8EBF1` (secondario `#A0A8B6`, attenuato `#B9C0CC`) |
| Pericolo | `#A3271B` | `#F29387` |
| Font | Inter per il testo, Space Grotesk per i titoli | uguale |
| Bersaglio tattile | almeno 44px | uguale |

Le coppie di colori che portano testo sono verificate da un test di contrasto
(`web/tests/shared/contrast.test.ts`), in entrambi i temi.

## Differenze volute fra design e app

-   **Miniature**: nelle schermate sono riquadri colorati; nell'app sono le miniature vere,
    sempre a colori, come le pubblica YouTube.
-   **Import da Google Takeout**: il collegamento c'è nella schermata Canali, ma nell'app non
    ancora.
-   **Filtro intelligente** (LLM): previsto per una versione successiva, non c'è.
