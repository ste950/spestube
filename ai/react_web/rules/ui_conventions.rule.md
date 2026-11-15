# Convenzioni UI

Queste regole si applicano ogni volta che si crea o si modifica un componente, una pagina o un
foglio di stile.

## 🎨 Tema e token

-   **Prima di costruire qualsiasi pagina o componente**, leggere il design di riferimento del
    progetto — al percorso dichiarato nel suo `CLAUDE.md`, per convenzione `docs/design/` — per
    colori, tipografia, raggi, spaziature e pattern visivi.
-   Colori, famiglie e dimensioni di testo, raggi e spaziature vivono nei **token** di
    `src/shared/theme/tokens.ts`, e da nessun'altra parte. Un componente li usa, non li
    sceglie.
-   I token sono un **oggetto TypeScript**, non un elenco di variabili scritte a mano in un
    foglio di stile: i due schemi (chiaro e scuro) hanno lo stesso tipo, quindi un token che
    esiste in uno solo dei due non compila, e le coppie di contrasto si verificano in un test
    che importa l'oggetto ([Componente Accessibile](../examples/accessible_component.example.md)).
-   All'avvio, prima del primo render, `main.tsx` trasforma l'oggetto in **variabili CSS**
    (`color.textMuted` → `--color-text-muted`) dentro un foglio con due blocchi, uno per
    schema, scelti dal browser con `prefers-color-scheme`. Il passaggio fra chiaro e scuro lo
    fa il browser: nessun hook, nessun `Context`, nessun re-render.
-   I fogli dei componenti usano **solo** `var(--…)`. Niente valori esadecimali, niente
    `rgb()`, niente lunghezze scritte a mano per spaziature, raggi e dimensioni di testo.

La forma completa è in [Token di tema](../examples/theme_tokens.example.md).

### CSS Modules, non classi di utilità

Ogni componente ha il suo foglio accanto, `<componente>.module.css`, e lo importa: le classi
sono locali al file per costruzione, i valori arrivano dalle variabili del tema. L'alternativa —
le classi di utilità in stile Tailwind — è scartata, e non perché sia una cattiva libreria. È la
stessa decisione dello stack React Native, e per lo stesso motivo.

Il vincolo di questo stack è che **i valori stiano nei token e non nei componenti**, e la
differenza fra i due approcci è se quel vincolo si può verificare. In un foglio di stile, un
valore è una dichiarazione: `color: #1a1a1a` dentro `order-row.module.css` è un letterale in
una posizione nota, che stylelint trova e rifiuta. Con le classi di utilità, lo stile è una
stringa nel JSX che nessuno strumento guarda, e il valore arbitrario è sempre a una parentesi
di distanza (`bg-[#1a1a1a]`): il confine fra "token" e "valore scritto a mano" smette di
essere osservabile, e la regola torna a essere un promemoria.

Scartato anche il CSS scritto in JavaScript che si calcola a runtime: costa a ogni render
quello che un foglio statico costa una volta, e perde ciò che il CSS fa già da sé — le media
query, `:focus-visible`, `prefers-reduced-motion` — che qui servono proprio
([Accessibilità](#-accessibilità)).

Regole del foglio:

-   Le classi si scrivono in `kebab-case` nel CSS e si leggono in `camelCase` dal componente
    (`.order-row` → `styles.orderRow`): la conversione la fa Vite, configurata una volta in
    [Vite e TypeScript](../examples/vite_and_tsconfig.example.md).
-   **Un solo foglio globale**, `src/shared/theme/global.css`: il reset, `box-sizing`, il
    carattere del `body`. Tutto il resto è un modulo. Una regola globale scritta in un modulo
    (`:global(...)`) è uno stile che esce dal suo componente, e non si scrive.
-   Gli stili in linea restano ammessi solo per valori **calcolati a runtime** (una larghezza
    che segue una misura, una posizione che segue il puntatore), e anche lì la base arriva da
    una variabile del tema.
-   I punti di rottura del layout sono **pochi e nominati**, in
    `src/shared/theme/breakpoints.json` accanto ai token: un sito che si legge a larghezza di
    telefono e di desktop ne chiede due o tre, non uno per componente. Una variabile CSS dentro
    `@media` non vale, quindi nelle media query si scrivono come valore, e stylelint ammette
    solo quelli del file.

Il nome di una classe sbagliato in un componente (`styles.ordeRow`) **non** fallisce in
compilazione: Vite dichiara le classi di un modulo come un dizionario di stringhe. Il sintomo
è un elemento senza stile, e si vede nella [verifica nel
browser](./verification.rule.md#%EF%B8%8F-verifica-nel-browser), non in CI.

Il divieto dei valori scritti a mano non resta un promemoria: è una regola di stylelint su
tutti i `*.module.css` di `src/` — vedi [Confini fra livelli](../examples/architecture_test.example.md)
e [Test di Architettura](../../quality/rules/architecture_tests.rule.md).

## 📏 Dimensione dei Componenti

-   Ogni file di componente o pagina (`.tsx`) deve restare entro **100 righe di codice**, e
    così il suo foglio di stile (`.module.css`).
-   Il limite conta il **codice**, non i commenti, che restano incoraggiati dove spiegano una
    decisione — vedi [Commenti nel Codice](./code_comments.rule.md). Sfoltire i commenti non è
    il modo di rientrare nel limite: se il file è lungo, è il codice a dover essere diviso.
-   Il foglio è un file separato, ma non è un posto dove spostare complessità: se è lui a
    superare il limite, il componente sta disegnando troppe cose, e si estrae un
    sotto-componente con il suo foglio.
-   La misura, da `web/`:

    ```bash
    grep -vE '^\s*(//|/\*|\*|$)' <file> | wc -l
    ```

    E per trovare tutti quelli fuori scala:

    ```bash
    for f in $(find src/features src/shared/components src/app \( -name "*.tsx" -o -name "*.module.css" \)); do
      n=$(grep -vE '^\s*(//|/\*|\*|$)' "$f" | wc -l); [ "$n" -gt 100 ] && echo "$n $f";
    done
    ```

-   Lo scope è quello che la regola copre: **componenti, pagine e i loro fogli**. I file di
    `state/` non sono componenti e non rientrano nel limite — se uno store cresce troppo, il
    segnale è che contiene logica di dominio da estrarre in uno use case.
-   Un `routes.tsx` sta su un ordine di grandezza diverso: dichiara percorsi e monta pagine,
    quindi poche righe per rotta. Un componente di montaggio che si avvicina alla decina di
    righe sta facendo il lavoro di una pagina.

## ♿ Accessibilità

-   I criteri — contrasto, bersagli, etichette, focus, movimento — sono trasversali e stanno
    in [Accessibilità](../../quality/rules/accessibility.rule.md): valgono qui come in
    qualsiasi altro stack che disegni interfacce. L'API con cui il web li esprime è in
    [Componente Accessibile](../examples/accessible_component.example.md).
-   Sul web l'accessibilità **parte dall'HTML**: un `<button>` è già un bottone per la
    tastiera e per lo screen reader, un `<div onClick>` non lo è per nessuno dei due. Si usa
    l'elemento che ha il significato giusto, e gli attributi `aria-*` restano per quello che
    l'HTML non sa dire (lo stato di un pannello espandibile, un'etichetta di sola icona).
-   **Tutto si usa con la tastiera**, nell'ordine in cui si legge, e il focus è sempre
    visibile: lo stile di `:focus-visible` sta nei componenti di `src/shared/` e usa la coppia
    di token del focus. Togliere l'indicatore perché «non è nel design» è un difetto.
-   **Ogni pagina ha il suo titolo** (`<title>`, che React porta nell'intestazione del
    documento) e un solo `<h1>`: è l'unico modo in cui una navigazione senza ricaricamento si
    fa annunciare.
-   La lingua del documento (`<html lang>`) segue quella delle etichette mostrate.
-   Nascono nei token e nei componenti di `src/shared/`, dove si applicano una volta sola: il
    rapporto di contrasto si verifica sulle coppie di token, non su una pagina, e il bersaglio
    minimo appartiene al componente, non a chi lo usa.
-   Uno stato che si comunica con il solo colore non si comunica: va accompagnato da un testo
    o da un'icona con la sua etichetta.

## ♻️ Riusabilità e Shared

-   I componenti agnostici al dominio e riutilizzabili vanno in `src/shared/components/`.
-   Prima di crearne uno nuovo, valutare se ne esiste già uno in `shared/` che copre il caso.
-   Un componente di `shared/` non conosce nessuna entità: riceve quello che deve mostrare
    come proprietà già pronte. Se per usarlo serve importare un tipo di `@domain/*`, allora è
    un componente di feature.
