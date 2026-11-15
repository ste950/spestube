I criteri stanno nella regola trasversale
[Accessibilità](../../quality/rules/accessibility.rule.md); qui c'è il modo in cui il web li
esprime. Tutto vive nei componenti, nei fogli e nei token di `src/shared/`, una volta sola: un
criterio applicato nel componente condiviso non deve essere ricordato da nessuna pagina.

La differenza da tenere a mente rispetto a React Native è opposta a quella che si aspetta chi
arriva da lì: **qui l'albero HTML c'è, e dice già quasi tutto.** Un `<button>` è un bottone,
un `<a href>` è un collegamento, un `<label>` dà il nome al suo campo. Gli attributi `aria-*`
servono per quello che l'HTML non sa dire, e un `role` messo su un `<div>` è quasi sempre il
segno che andava usato un altro elemento.

## 🏷️ Etichetta obbligatoria e bersaglio minimo

```tsx
// web/src/shared/components/app-icon-button.tsx

import type { ReactNode } from 'react';
import styles from './app-icon-button.module.css';

type AppIconButtonProps = {
  icon: ReactNode;
  /**
   * Cosa fa il bottone, non come è disegnato: «Elimina il documento», mai «cestino». Non è
   * opzionale di proposito: senza testo visibile è l'unica cosa che un lettore di schermo può
   * annunciare, e un parametro opzionale è un parametro che prima o poi non viene passato.
   * Arriva già tradotto: `shared/` non conosce le chiavi di nessuna feature.
   */
  label: string;
  onClick: () => void;
  disabled?: boolean;
};

export function AppIconButton({ icon, label, onClick, disabled = false }: AppIconButtonProps) {
  return (
    <button
      type="button"
      className={styles.target}
      aria-label={label}
      onClick={onClick}
      disabled={disabled}
    >
      {/* L'icona è disegno: il nome del bottone lo dà `aria-label`, e annunciarla due volte
          raddoppierebbe l'annuncio. */}
      <span aria-hidden="true">{icon}</span>
    </button>
  );
}
```

```css
/* web/src/shared/components/app-icon-button.module.css */

/* Il disegno dell'icona e il bersaglio sono due misure indipendenti: l'icona resta quella che
   il design chiede, il riquadro cliccabile non scende sotto il minimo. */
.target {
  display: inline-grid;
  place-items: center;
  min-width: var(--size-min-target);
  min-height: var(--size-min-target);
  padding: 0;
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: inherit;
  cursor: pointer;
}

.target:focus-visible {
  outline: var(--size-focus-ring) solid var(--color-focus);
  outline-offset: var(--size-focus-ring);
}
```

Gli **8 pixel fra due bersagli adiacenti** restano un fatto di layout: si ottengono con il
`gap` del contenitore (`var(--space-sm)`), non allargando i bottoni.

## 🚦 Stato che non si affida al solo colore

```tsx
// web/src/shared/components/app-status-badge.tsx

import type { ReactNode } from 'react';
import styles from './app-status-badge.module.css';

type AppStatusBadgeProps = {
  label: string;
  icon: ReactNode;
  tone: 'neutral' | 'positive' | 'danger';
};

export function AppStatusBadge({ label, icon, tone }: AppStatusBadgeProps) {
  return (
    <span className={styles[tone]}>
      <span aria-hidden="true">{icon}</span>
      {label}
    </span>
  );
}
```

Il colore è il terzo canale, non il primo: icona e testo dicono lo stato anche in scala di
grigi. Il tono è un **insieme chiuso** che il foglio traduce in variabili del tema, non un
colore passato dal chiamante: chi usa il componente dice *che cosa* sta mostrando, il foglio
decide *come*.

La grafica **decorativa** si toglie dall'albero di accessibilità con `aria-hidden="true"`, e
un'immagine decorativa con `alt=""` — l'attributo vuoto, non assente: senza `alt` il lettore
annuncia il nome del file.

## ⏭️ Saltare la navigazione

Chi naviga da tastiera attraversa intestazione e menu a ogni cambio di pagina, prima di
arrivare al contenuto. Il primo elemento focalizzabile del layout è un collegamento che li
salta, invisibile finché non riceve il focus:

```tsx
// web/src/shared/components/app-skip-link.tsx

import styles from './app-skip-link.module.css';

export function AppSkipLink({ targetId, label }: { targetId: string; label: string }) {
  return (
    <a className={styles.skipLink} href={`#${targetId}`}>
      {label}
    </a>
  );
}
```

```css
/* web/src/shared/components/app-skip-link.module.css */

.skip-link {
  position: absolute;
  inset-inline-start: var(--space-sm);
  inset-block-start: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  background: var(--color-accent);
  color: var(--color-on-accent);
  transform: translateY(-200%);
}

.skip-link:focus-visible {
  transform: none;
  outline: var(--size-focus-ring) solid var(--color-focus);
}
```

## 🧭 Struttura e focus

-   **Punti di riferimento**: `<header>`, `<nav>`, `<main>`, `<footer>` nel layout, una volta
    sola. Sono l'indice con cui un lettore di schermo salta da una zona all'altra.
-   **Un titolo per pagina**: `<title>` nel componente della pagina, che React porta
    nell'intestazione del documento, e un solo `<h1>`. Le intestazioni scendono senza saltare
    livelli.
-   **Il focus dopo una navigazione.** Il browser lo sposta da solo quando carica una pagina
    nuova; un router no, e senza intervento il focus resta sul collegamento appena premuto,
    in una pagina che non c'è più. Il layout lo porta sul contenuto quando cambia il percorso:

    ```tsx
    // web/src/app/layouts/app-layout.tsx — frammento

    const { pathname } = useLocation();
    const main = useRef<HTMLElement>(null);

    useEffect(() => {
      main.current?.focus();
    }, [pathname]);

    // …
    <main id="contenuto" ref={main} tabIndex={-1} className={styles.main}>
    ```

-   **I dialoghi sono `<dialog>`** aperti con `showModal()`: il browser porta il focus dentro,
    rende inerte il resto della pagina e chiude con `Esc`. Una finestra modale ricostruita con
    dei `<div>` deve rifare a mano tutte e tre le cose, e di solito ne dimentica una.
-   **L'ordine di tabulazione è l'ordine del sorgente.** `tabIndex` maggiore di zero non si usa:
    riordina la tastiera rispetto a quello che si vede, che è il difetto che la regola vieta.

## 🎞️ Movimento che si può spegnere

Il foglio globale annulla animazioni e transizioni quando il sistema chiede movimento ridotto
([Token di tema](./theme_tokens.example.md)). Un componente che anima qualcosa di ampio fa un
passo in più: l'animazione si **dichiara** solo per chi non ha chiesto il contrario.

```css
/* un frammento di modulo */

@media (prefers-reduced-motion: no-preference) {
  .panel {
    transition: transform 200ms ease-out;
  }
}
```

Un'animazione guidata da JavaScript legge la stessa preferenza con
`matchMedia('(prefers-reduced-motion: reduce)')`, e lo fa in un componente di `shared/`, non in
ogni pagina che anima.

## 🎨 Contrasto: le coppie stanno nel tema

Il tema non sa quale primo piano finirà su quale sfondo: l'elenco delle combinazioni ammesse è
parte del tema quanto i colori. Le coppie si dichiarano **per nome di token**, una volta sola,
e si verificano su entrambe le tavolozze:

```ts
// web/src/shared/theme/contrast.ts

import type { Palette } from './tokens';

export type ContrastPair = {
  readonly foreground: keyof Palette;
  readonly background: keyof Palette;
  /** 4.5 per il testo normale, 3 per il testo grande, la grafica significativa e il focus. */
  readonly minimumRatio: 4.5 | 3;
};

export const contrastPairs: readonly ContrastPair[] = [
  { foreground: 'text', background: 'background', minimumRatio: 4.5 },
  { foreground: 'textMuted', background: 'surface', minimumRatio: 4.5 },
  { foreground: 'onAccent', background: 'accent', minimumRatio: 4.5 },
  { foreground: 'danger', background: 'background', minimumRatio: 4.5 },
  { foreground: 'focus', background: 'background', minimumRatio: 3 },
];

const channel = (value: number): number =>
  value <= 0.03928 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;

// Niente accesso per indice: con `noUncheckedIndexedAccess` ogni elemento sarebbe
// `number | undefined`, e il controllo che ne uscirebbe non aggiungerebbe nulla.
const luminance = (hex: string): number => {
  const at = (offset: number) => channel(parseInt(hex.slice(offset, offset + 2), 16) / 255);

  return 0.2126 * at(1) + 0.7152 * at(3) + 0.0722 * at(5);
};

export function contrastRatio(foreground: string, background: string): number {
  const a = luminance(foreground);
  const b = luminance(background);

  return (Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05);
}
```

Dichiarare le coppie per nome e non per valore chiude il difetto più silenzioso di questo
controllo: una coppia copiata dal tema chiaro dentro quello scuro passa il test verificando due
volte i colori sbagliati. Qui la coppia è una sola, e i colori li prende da ciascuna tavolozza.

## ✅ Cosa si verifica da solo

I criteri con un numero diventano test e girano con `npm test` come gli altri, in
`web/tests/architecture/` insieme ai [confini fra livelli](./architecture_test.example.md).

```ts
// web/tests/architecture/contrast.test.ts

import { describe, expect, it } from 'vitest';
import { contrastPairs, contrastRatio } from '@shared/theme/contrast';
import { darkPalette, lightPalette } from '@shared/theme/tokens';

it('il tema dichiara le coppie da verificare', () => {
  // Una lista vuota passerebbe i test qui sotto senza aver letto un colore.
  expect(contrastPairs).not.toHaveLength(0);
});

describe.each([
  ['chiaro', lightPalette],
  ['scuro', darkPalette],
])('tema %s', (_name, palette) => {
  it.each(contrastPairs)('$foreground su $background raggiunge il minimo', (pair) => {
    expect(
      contrastRatio(palette[pair.foreground], palette[pair.background]),
    ).toBeGreaterThanOrEqual(pair.minimumRatio);
  });
});
```

```tsx
// web/tests/architecture/icon-button.test.tsx

import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { render, screen } from '@testing-library/react';
import { expect, it } from 'vitest';
import { AppIconButton } from '@shared/components/app-icon-button';

it('un bottone di sola icona si annuncia con la sua etichetta', () => {
  render(<AppIconButton icon="×" label="Chiudi il pannello" onClick={() => {}} />);

  // Il nome accessibile è quello che il DOM calcola davvero, non una proprietà dichiarata: se
  // l'icona finisse nel nome, il test lo vedrebbe.
  expect(screen.getByRole('button', { name: 'Chiudi il pannello' })).toBeInTheDocument();
});

it('il bersaglio del bottone di sola icona non scende sotto il minimo', () => {
  const sheet = readFileSync(
    join(import.meta.dirname, '../../src/shared/components/app-icon-button.module.css'),
    'utf8',
  );

  expect(sheet).toMatch(/min-width:\s*var\(--size-min-target\)/);
  expect(sheet).toMatch(/min-height:\s*var\(--size-min-target\)/);
});
```

Due precisazioni su cosa questi test dimostrano davvero:

-   il test del bersaglio verifica il **foglio dichiarato**, non il riquadro misurato: jsdom non
    esegue il layout, quindi un bottone che dichiara il minimo e poi viene compresso da un
    contenitore passa lo stesso. La misura vera esiste solo in un browser, ed è parte della
    [verifica nel browser](../rules/verification.rule.md#%EF%B8%8F-verifica-nel-browser);
-   l'**etichetta obbligatoria** è garantita due volte: `label` è una proprietà richiesta,
    quindi un bottone di sola icona senza etichetta **non compila**, e il test sul nome
    accessibile verifica che arrivi davvero a chi la legge.

Quello che resta è giudizio e resta in revisione: la qualità dell'etichetta, l'ordine di
lettura, il secondo canale oltre il colore. La regola trasversale dice quali sono e perché non
si forzano in un test.
