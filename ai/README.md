# Regole per gli agenti AI

Questa cartella contiene le regole architetturali, le skill e gli esempi di codice che il
progetto segue. È una **copia** tenuta dentro il repo: SpesTube non dipende da nessun altro
repository, né per compilare né per essere capito da un agente.

## Cosa contiene

-   [`go/`](./go/CLAUDE.md): il backend Go a binario unico (`backend/`).
-   [`react_web/`](./react_web/CLAUDE.md): il frontend React con Vite (`web/`).
-   [`quality/`](./quality/CLAUDE.md): test di architettura e criteri di accessibilità.

Il `CLAUDE.md` alla radice importa i tre indici e dichiara ciò che è specifico di SpesTube,
comprese le deroghe (Docker al posto di systemd, frontend servito dal binario Go).

## Origine

Copiato dal repository `ai_framework`, commit `ea386d2`, il 15 novembre 2026. Rispetto
all'originale:

-   sono stati importati solo gli stack usati qui; Flutter, Python, React Native, l'autopilot
    e le cartelle `ci/`, `github/` e `project_docs/` non ci sono;
-   i link verso le cartelle non importate sono diventati testo semplice;
-   le sezioni "Come agganciarlo a un progetto" descrivono l'aggancio interno a questo repo;
-   i template `project_claude_md.example.md` sono stati tolti.

## Aggiornare le regole

Le regole si modificano direttamente qui, come qualsiasi altro file del repo. Per riportare
una novità dal framework d'origine si copia a mano il file interessato e si ricontrollano i
link: non c'è sincronizzazione automatica, ed è voluto.
