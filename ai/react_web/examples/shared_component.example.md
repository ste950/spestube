Un componente di `src/shared/components/`: agnostico al dominio, con lo stile nel suo foglio e
l'accessibilità già dentro. Riceve quello che deve mostrare come proprietà già pronte — se per
usarlo servisse importare un tipo di `@domain/*`, sarebbe un componente di feature
([Riusabilità e Shared](../rules/ui_conventions.rule.md#%EF%B8%8F-riusabilità-e-shared)).

```tsx
// web/src/shared/components/app-button.tsx

import styles from './app-button.module.css';

type AppButtonProps = {
  label: string;
  onClick: () => void;
  variant?: 'primary' | 'quiet';
  disabled?: boolean;
  /** `submit` solo dentro un form: il default del browser è `submit`, ed è quasi sempre sbagliato. */
  type?: 'button' | 'submit';
};

export function AppButton({
  label,
  onClick,
  variant = 'primary',
  disabled = false,
  type = 'button',
}: AppButtonProps) {
  return (
    // Un `<button>` vero: la tastiera, il focus, l'annuncio come bottone e lo stato
    // disabilitato li dà l'HTML, e nessuno va ricostruito a mano.
    <button
      type={type}
      className={variant === 'primary' ? styles.primary : styles.quiet}
      onClick={onClick}
      disabled={disabled}
    >
      {label}
    </button>
  );
}
```

```css
/* web/src/shared/components/app-button.module.css */

.base {
  min-width: var(--size-min-target);
  min-height: var(--size-min-target);
  padding: var(--space-sm) var(--space-md);
  border: 1px solid var(--color-accent);
  border-radius: var(--radius-md);
  font: inherit;
  cursor: pointer;
}

.primary {
  composes: base;
  background: var(--color-accent);
  color: var(--color-on-accent);
}

.quiet {
  composes: base;
  background: transparent;
  color: var(--color-accent);
}

/* Il focus da tastiera si vede sempre, con la coppia di token che il test di contrasto
   verifica. `:focus-visible` e non `:focus`: chi clicca non vede l'anello, chi usa la tastiera
   sì. */
.base:focus-visible {
  outline: var(--size-focus-ring) solid var(--color-focus);
  outline-offset: var(--size-focus-ring);
}

.base:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}
```

Punti chiave:

-   Il prefisso `App` distingue i componenti condivisi dagli elementi HTML e da quelli di una
    libreria, e rende evidente in una pagina cosa viene da `shared/`.
-   **L'elemento giusto prima di tutto.** Un `<button>` è focalizzabile, si attiva con `Invio`
    e `Spazio`, si annuncia come bottone e smette di rispondere quando è `disabled`. Un `<div>`
    con `onClick` non fa nessuna di queste cose, e ricostruirle a mano con `role`, `tabIndex`
    e un gestore di tasti è il modo più costoso di arrivare a un bottone peggiore.
-   **`type="button"` per default.** Dentro un `<form>` il default dell'HTML è `submit`, e un
    bottone che invia il form senza volerlo è un difetto che si scopre in produzione.
-   **Nessun valore scritto a mano nel foglio**: colori, raggi, spaziature, bersaglio minimo e
    anello di focus sono variabili del tema. Il `1px` del bordo e l'`opacity` non sono valori
    del design ma della geometria, e stylelint li lascia passare
    ([Confini fra livelli](./architecture_test.example.md)).
-   **Le varianti sono classi composte** (`composes: base`) e non concatenazioni di stringhe nel
    componente: il componente sceglie una classe, il foglio sa cosa contiene.
-   Nessuna etichetta dentro il componente: il testo **arriva già tradotto** da chi lo usa. Un
    componente di `shared/` non conosce le chiavi di traduzione di nessuna feature.
-   `variant` è un'unione di letterali, non un booleano `isPrimary`: aggiungere una terza
    variante non costringe a rileggere ogni chiamata.
-   Il componente non tiene stato di dominio e non naviga: riceve `onClick` e lo chiama. Un
    bottone che porta a un'altra pagina non è un bottone: è un `<Link>`.
