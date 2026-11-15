---
description: Skill per creare una nuova feature di un sito React — pagine, store, componenti e il suo routes.tsx — e farla comporre dal router dell'app.
goal: Creare la cartella di una feature con le sue rotte, le pagine, lo store e i componenti, raggiungibile da un URL senza che le pagine dipendano dal router e senza che app/ veda altro che routes.tsx.
---

# Skill: Creazione Feature

Questa skill descrive come creare una feature in `web/src/features/<nome_feature>/` e come
renderla raggiungibile.

## 📋 Obiettivo

Creare il presentation layer di una funzionalità — percorsi, pagine, stato, componenti — e
farlo comporre da `src/app/router.tsx`, che della feature vede solo l'elenco delle rotte.

## 🧭 Prima di tutto: chi possiede le rotte

In questo stack **la feature dichiara i propri percorsi**, come in Flutter e al contrario di
React Native: il router è codice, e la feature gli consegna un elenco. `app/` lo compone sotto
i layout e non conosce nient'altro della feature. Il perché, e le alternative scartate, sono in
[Le rotte stanno nella feature](../../rules/project_structure.rule.md#-le-rotte-stanno-nella-feature).

## 🛠️ Procedura

1.  **Verifica il contesto**
    -   Leggi [Struttura di una Feature](../../rules/feature_overview.rule.md) e
        [Convenzioni UI](../../rules/ui_conventions.rule.md).
    -   Guarda le feature già presenti e riusa naming e struttura.
    -   Accertati che il **soggetto** di dominio esista già: entità, porta e implementazione
        stanno in `domain/<soggetto>/` e `data/<soggetto>/`, non nella feature. Se manca, parti
        da [Creazione di Repository e Sorgenti](../repository_creation/repository_creation.skill.md).

2.  **Crea la cartella della feature**
    -   `web/src/features/<nome_feature>/`, in `kebab-case`.
    -   `pages/`, `state/` e `components/` nascono con il primo file che le riempie: non si
        creano vuote. `routes.tsx` c'è sempre — il test di architettura lo pretende.

3.  **Crea lo store**, se la feature ha stato
    -   `state/<nome_feature>-store.ts`: una fabbrica `create<Feature>Store(porte)` che riceve
        le porte di dominio come argomenti.
    -   `state/use-<nome_feature>-store.ts`: due righe che passano le porte vere prese da
        `@bootstrap/dependencies`. È l'**unico** file della feature che importa quel modulo.
    -   [Vedi Esempio](../../examples/feature_store.example.md)

4.  **Crea le pagine**
    -   `pages/<nome>-page.tsx`, con il suo `pages/<nome>-page.module.css` accanto.
    -   La pagina riceve i parametri già convertiti come proprietà e legge lo stato con i
        selettori dello store; non chiama `useParams`.
    -   Un `<title>` e un solo `<h1>` per pagina, etichette da `useTranslate()` — [Vedi
        Esempio](../../examples/i18n_hook.example.md).
    -   Stili dal foglio, solo con variabili del tema — [Vedi
        Esempio](../../examples/theme_tokens.example.md).
    -   Ogni stato ha la sua resa: caricamento, errore, vuoto, pieno.
    -   [Vedi Esempio](../../examples/feature_page.example.md)

5.  **Dichiara le rotte**
    -   `routes.tsx` esporta `<nomeFeature>Routes: RouteObject[]`, con percorsi **relativi**
        (`orders`, `orders/:orderId`).
    -   Ogni pagina si importa con `lazy`; dove ci sono parametri, un componente di montaggio
        di poche righe li legge, li converte e rende la pagina.
    -   Niente `loader` né `action`: i dati li carica lo store.
    -   [Vedi Esempio](../../examples/feature_routes.example.md)

6.  **Aggiungi le rotte al router dell'app**
    -   In `src/app/router.tsx`, una riga: `...<nomeFeature>Routes` fra i figli del layout
        giusto. È l'unica modifica fuori dalla cartella della feature.

7.  **Verifica** — vedi sotto.

## ⚠️ Vincoli

-   **Solo presentation**: entità, porte e use case vanno nei layer condivisi. Una feature che
    contiene logica di dominio è uno use case mancante
    ([domain/](../../rules/domain_layer.rule.md)).
-   **Indipendenza**: una feature non importa mai un'altra feature. Si naviga con un `<Link>` e
    un percorso, e un percorso non è un import. Se serve condividere dati la condivisione passa
    dal dominio; se serve condividere un componente, quello va in `@shared/*`.
-   **Mai `@data/*`**: le implementazioni arrivano dal composition root attraverso le porte di
    `domain/` ([Composition root](../../examples/bootstrap_dependencies.example.md)).
-   **La feature non importa `@app/*`**: non sa sotto quale layout è montata.
-   **Niente rete, niente storage, niente `import.meta.env`** dentro store o componenti:
    stanno nelle sorgenti di `data/<soggetto>/sources/` e in `bootstrap/config.ts`.

## ✅ Verifica

I check dello stack — formattazione, lint e confini fra livelli, fogli di stile, `tsc`, suite,
build — girano nella CI e non qui: vedi [Verifica delle Modifiche](../../rules/verification.rule.md).
In locale resta il solo `npx prettier --write .`.

Le `grep` qui sotto non sono check: leggono, e servono mentre si lavora. Dalla cartella
`web/`:

```bash
grep -rn "@features/" src/features/<nome_feature> | grep -v "@features/<nome_feature>"
grep -rn "@data/\|@app/" src/features/<nome_feature>
grep -rn "@bootstrap/" src/features/<nome_feature> | grep -v "state/use-<nome_feature>-store"
```

Le tre `grep` devono essere vuote. Infine apri la pagina nel browser, a larghezza di telefono e
di desktop, attraversala con la sola tastiera e ricaricala sul suo URL
([Verifica nel browser](../../rules/verification.rule.md#%EF%B8%8F-verifica-nel-browser)).
