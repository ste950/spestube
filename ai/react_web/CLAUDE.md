# Framework AI — Sito React con Vite e React Router

Questo file è il punto di ingresso per gli agenti AI: elenca le regole trasversali, tutte le
skill disponibili e gli esempi di codice condivisi, con i loro percorsi e una breve
descrizione. Nessun contenuto sostanziale va scritto qui — solo link.

Il framework descrive un sito a pagina singola, costruito in file statici, con la stessa
architettura degli altri stack — feature per la presentazione, `domain/` e `data/` condivisi
per soggetto — espressa negli idiomi di TypeScript e del browser: `Result` come unione
discriminata, un composition root al posto del service locator, uno store leggero per lo stato
di presentazione, le rotte dichiarate dalla feature e composte dall'app, lo stile in CSS Modules
che leggono variabili nate dai token. È indipendente da qualsiasi progetto: nessuna regola,
skill o esempio deve citare un'applicazione specifica.

## 🔌 Come è agganciato a questo repo

Questa cartella è una copia dello stack, tenuta dentro il repo: il `CLAUDE.md` alla radice la
importa con `@ai/react_web/CLAUDE.md`. Non esiste una cartella sorella da clonare. Origine e
modalità di aggiornamento sono in [`ai/README.md`](../README.md).

Le regole qui dentro assumono il layout `web/`: il sito vive in `web/` alla radice del repo,
con il codice in `web/src/` e i test in `web/tests/`. Tutto ciò che è specifico di un progetto
(versione di Node, alias di import, prefisso delle variabili, dove viene pubblicato, servizi
esterni) resta nel `CLAUDE.md` del progetto.

## 📖 Regole Trasversali

-   [Struttura del Progetto](./rules/project_structure.rule.md): cosa copre lo stack (e cosa no), organizzazione di `web/`, perché le rotte stanno nella feature, alias di import, tabella delle dipendenze fra livelli e corrispondenze con Flutter e React Native.
-   [Verifica delle Modifiche](./rules/verification.rule.md): i check dello stack e dove gira ciascuno — in CI tutti, in locale il solo formattatore; perché `tsc -b` e non `--noEmit`, le dipendenze, i confini fra layer e la verifica nel browser.
-   [Convenzioni UI](./rules/ui_conventions.rule.md): token come variabili CSS, CSS Modules contro classi di utilità, punti di rottura, dimensione di componenti e fogli, accessibilità a partire dall'HTML.
-   [Commenti nel Codice](./rules/code_comments.rule.md): si commenta il perché e non il cosa; i tipi non si riscrivono a parole; ogni silenziamento porta la sua ragione; nessun riferimento a task, issue o PR dentro il codice.
-   [Configurazione](./rules/configuration.rule.md): le variabili `VITE_*` finiscono nel pacchetto e sono pubbliche, un solo modulo legge `import.meta.env`, una build per ambiente.
-   [Distribuzione](./rules/deployment.rule.md): la cartella `dist/`, il fallback su `index.html`, le due politiche di cache, il percorso di base e le intestazioni di sicurezza.

### Architettura: Feature e Layer Condivisi

-   [Struttura di una Feature](./rules/feature_overview.rule.md): layout di una feature (routes/pages/state/components), le rotte che la feature dichiara e `app/` compone, niente `loader`, lo store leggero al posto del Bloc e le regole di indipendenza fra feature.
-   [domain/ – Domain Layer](./rules/domain_layer.rule.md): entità, porte e use case per soggetto; `Result` come unione discriminata e il vocabolario delle `Failure`.
-   [data/ – Data Layer](./rules/data_layer.rule.md): sorgenti con il loro contratto, repository, modelli e mapper; il composition root al posto del service locator.

## 🧠 Lista delle Skill

-   **`skills/feature_creation/feature_creation.skill.md`**: Skill per creare una feature (rotte, pagine, store, componenti) e farla comporre dal router dell'app, che della feature vede solo `routes.tsx`.
-   **`skills/repository_creation/repository_creation.skill.md`**: Skill per dare a un soggetto una nuova fonte di dati — porta, modello con il suo controllo di forma, sorgente, mapper, repository e cablaggio nel composition root.
-   **`skills/use_case_creation/use_case_creation.skill.md`**: Skill per creare use case di dominio, come funzioni pure o fabbriche che ricevono le porte.
-   **`skills/shared_component_creation/shared_component_creation.skill.md`**: Skill per creare componenti condivisi sull'elemento HTML giusto, con il loro foglio CSS Module, accessibili e agnostici al dominio.
-   **`skills/localization_management/localization_management.skill.md`**: Skill per la gestione delle etichette localizzate con cataloghi tipizzati e prefisso della feature.
    -   [Formato dei cataloghi](./skills/localization_management/examples/i18n_catalog.example.md): i cataloghi di `shared/i18n/`, con il tipo delle chiavi che nasce dalla lingua base.

## 📦 Esempi Condivisi

Snippet di codice concreti, riusabili da tutte le skill. Usano tutti lo stesso dominio
inventato — il soggetto `orders` — e si incastrano fra loro: l'entità attraversa la sorgente,
il repository, lo use case, lo store, la pagina e il test. Messi insieme in un progetto Vite
compilano, passano lint, stylelint e test, e si costruiscono.

### Infrastruttura

I pezzi che le regole danno per scontati: un progetto nuovo parte da qui.

-   [Vite e TypeScript](./examples/vite_and_tsconfig.example.md): tre progetti TypeScript e una radice vuota, alias dichiarati una volta, CSS Modules in `camelCase`, test nella configurazione di Vite, versione di Node e prettier.
-   [Result e fallimenti](./examples/result_pattern.example.md): `Result<T>` come unione discriminata, il vocabolario delle `Failure` su `kind` e il `switch` chiuso da `assertNever`.
-   [Contratto UseCase](./examples/use_case_contract.example.md): gli alias `UseCase<In, Out>` e `AsyncUseCase<In, Out>`, che si soddisfano annotando invece che implementando.
-   [Configurazione all'avvio](./examples/config.example.md): `bootstrap/config.ts`, l'unico modulo che legge `import.meta.env`, con il controllo all'avvio e il file `.env.example`.
-   [Composition root](./examples/bootstrap_dependencies.example.md): `bootstrap/dependencies.ts`, dove si costruiscono repository e use case e si espongono annotati con i contratti di `domain/`.
-   [Token di tema](./examples/theme_tokens.example.md): la tavolozza e le misure in TypeScript, il foglio di variabili CSS generato all'avvio, il foglio globale e `main.tsx`.
-   [Etichette tradotte nei componenti](./examples/i18n_hook.example.md): la scelta della lingua, l'unico `Context` di questo stack e la lingua del documento.

### Layer condivisi e feature

-   [Entità di Dominio](./examples/domain_entity.example.md): `type` con campi `readonly`, logica derivata come funzione pura e insieme chiuso di stati come unione di letterali.
-   [Interfaccia Repository](./examples/domain_repository.example.md): la porta di dominio, che restituisce `Promise<Result<T>>` e non sa da dove arrivino i dati.
-   [Use Case](./examples/use_case.example.md): la regola di business nelle sue due forme — funzione pura, e fabbrica `make*` che riceve le porte.
-   [Sorgenti](./examples/source_impl.example.md): la sorgente HTTP con `fetch` e la copia locale in `localStorage`, con il controllo di forma che ferma `any` al confine.
-   [Implementazione di Repository](./examples/repository_impl.example.md): la factory che compone sorgente remota e copia locale, decide quando ripiegare e mappa i modelli in entità.
-   [Store di una feature](./examples/feature_store.example.md): la fabbrica dello store che riceve le porte, con il `Result` che si ferma qui e non entra nello stato.
-   [Rotte di una feature](./examples/feature_routes.example.md): `routes.tsx` con le pagine pigre e i componenti di montaggio, e `src/app/router.tsx` che le compone sotto il layout.
-   [Pagina](./examples/feature_page.example.md): la pagina che riceve i parametri già convertiti, legge lo store con i selettori e ha il suo foglio accanto.
-   [Componente Shared](./examples/shared_component.example.md): componente `App*` costruito su un `<button>` vero, con il foglio che usa solo variabili del tema.
-   [Componente Accessibile](./examples/accessible_component.example.md): i criteri di [Accessibilità](../quality/rules/accessibility.rule.md) sul web — etichetta obbligatoria, bersaglio minimo, stato oltre il colore, salto della navigazione, focus dopo una navigazione, movimento ridotto, coppie di contrasto.
-   [Test di uno Use Case](./examples/use_case_test.example.md): caso nominale e casi limite, con la porta finta come oggetto letterale annotato.

### Verifica e CI

-   [Confini fra livelli](./examples/architecture_test.example.md): la tabella delle dipendenze come regola di ESLint, i fogli di stile sotto stylelint, e il test che impedisce di disarmarli in silenzio.
-   [Job di CI](./examples/ci_job.example.yml): la verifica su macchina pulita, con gli stessi comandi di [Verifica delle Modifiche](./rules/verification.rule.md). È un `.yml` e non un `.example.md` perché si copia dentro `.github/` così com'è.

### Distribuzione e setup di un progetto

-   [Server statico](./examples/static_hosting.example.md): la configurazione di nginx con fallback su `index.html`, cache e intestazioni di sicurezza, e l'immagine che la contiene.

## 🔗 Materiale trasversale

Vale per questo stack come per gli altri, e **non** va riscritto qui.

-   [`quality/`](../quality/CLAUDE.md): i test che difendono l'architettura e i criteri di accessibilità.
