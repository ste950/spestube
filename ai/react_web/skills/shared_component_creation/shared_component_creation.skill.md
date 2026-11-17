---
description: Skill per creare componenti React riutilizzabili in shared/ di un sito, con il loro foglio CSS Module, accessibili e agnostici al dominio.
goal: Guidare la creazione di componenti condivisi piccoli, costruiti sull'elemento HTML giusto, che prendono ogni valore visivo dalle variabili del tema e non conoscono nessuna entità né nessuna etichetta.
---

# Skill: Creazione Componenti Shared

Questa skill descrive come creare o modificare un componente in
`web/src/shared/components/`.

## 📋 Obiettivo

Creare un componente riutilizzabile da più feature, che riceve quello che deve mostrare come
proprietà già pronte e prende tutti i suoi valori visivi dalle variabili del tema.

## 🕐 Quando un componente è "shared"

Una sola domanda: **per usarlo serve importare un tipo di `@domain/*`?** Se sì, è un componente
di feature e va in `features/<f>/components/`. Se no, e più di una feature lo userebbe, va in
`shared/`. Un componente usato da una sola feature resta dov'è finché non arriva la seconda.

## 🛠️ Procedura

1.  **Verifica il contesto**
    -   Leggi [Convenzioni UI](../../rules/ui_conventions.rule.md) e i criteri di
        [Accessibilità](../../../quality/rules/accessibility.rule.md).
    -   Leggi il design di riferimento del progetto, al percorso dichiarato nel suo `CLAUDE.md`.
    -   **Guarda cosa c'è già** in `src/shared/components/`: prima di creare, valuta se un
        componente esistente copre il caso con una proprietà in più.

2.  **Scegli l'elemento HTML**
    -   Prima del codice, la domanda è quale elemento ha già il significato giusto: `<button>`
        per un'azione, `<a>` (cioè `<Link>`) per una navigazione, `<input>` con il suo
        `<label>` per un campo, `<dialog>` per una finestra modale.
    -   Un `<div>` con `onClick` non è mai la risposta: non si raggiunge con la tastiera e non
        si annuncia.

3.  **Crea il componente e il suo foglio**
    -   `web/src/shared/components/<nome>.tsx` e `<nome>.module.css`, in `kebab-case`; il
        componente ha il prefisso `App` (`AppButton`, `AppCard`, `AppTextField`).
    -   Una funzione, non una classe. Proprietà generiche — `label`, `value`, `onClick`,
        `variant`, `disabled` — e nessun nome che venga da una feature.
    -   Gli insiemi di varianti sono unioni di letterali, non booleani sommati; nel foglio
        diventano classi composte (`composes: base`).
    -   Entro **100 righe di codice** il componente, ed entro 100 il foglio. Se non ci stanno,
        si estrae un sotto-componente.
    -   [Vedi Esempio](../../examples/shared_component.example.md)

4.  **Prendi i valori dal tema**
    -   Colori, spaziature, raggi, dimensioni di testo, bersaglio minimo e anello di focus sono
        variabili CSS (`var(--color-accent)`) — [Vedi
        Esempio](../../examples/theme_tokens.example.md).
    -   **Nessun colore e nessuna spaziatura scritti a mano.** Se manca un token, si aggiunge a
        `tokens.ts`: non si scrive il valore nel foglio. Stylelint lo rifiuta comunque.
    -   Le media query usano solo i punti di rottura di `breakpoints.json`.

5.  **Cura l'accessibilità**
    -   Il focus da tastiera si vede: `:focus-visible` con la coppia di token del focus.
    -   Il bersaglio minimo appartiene al componente, non a chi lo usa.
    -   Un componente di sola icona ha un'etichetta **obbligatoria** e l'icona `aria-hidden`.
    -   Uno stato comunicato con il solo colore non è comunicato: va accompagnato da un testo o
        da un'icona.
    -   [Vedi Esempio](../../examples/accessible_component.example.md)

6.  **Rispetta le dipendenze**
    -   `shared/` può importare solo `@core/*` e React.
    -   Mai `@domain/*`, `@data/*`, `@features/*`, `@app/*`, `@bootstrap/*`.
    -   **Nessuna etichetta dentro il componente**: il testo arriva già tradotto da chi lo usa
        ([Gestione delle etichette
        localizzate](../localization_management/localization_management.skill.md)).

## ✅ Verifica

I check dello stack — formattazione, lint e confini fra livelli, fogli di stile, `tsc`, suite —
girano nella CI e non qui: vedi [Verifica delle Modifiche](../../rules/verification.rule.md).
In locale resta il solo `npx prettier --write .`.

Le `grep` qui sotto non sono check: leggono, e servono mentre si lavora. Dalla cartella
`web/`:

```bash
grep -rn "#[0-9a-fA-F]\{3,\}" src/shared/components
grep -rn "@domain/\|@data/\|@features/\|@app/\|@bootstrap/" src/shared
for f in $(find src/shared/components \( -name "*.tsx" -o -name "*.module.css" \)); do
  n=$(grep -vE '^\s*(//|/\*|\*|$)' "$f" | wc -l); [ "$n" -gt 100 ] && echo "$n $f";
done
```

Tutte e tre devono essere vuote. Se il componente ha comportamento non banale — uno stato
interno, una variante che cambia l'elemento — aggiungi un test con Testing Library, che
interroga per ruolo e nome accessibile come farebbe un lettore di schermo.
