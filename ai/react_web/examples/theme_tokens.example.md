I token di tema: un oggetto TypeScript come fonte unica, trasformato all'avvio in variabili CSS
che i fogli dei componenti leggono. Il perché di questa forma — e delle classi di utilità
scartate — è in [Tema e token](../rules/ui_conventions.rule.md#-tema-e-token).

## I token

```ts
// web/src/shared/theme/tokens.ts

import breakpointValues from './breakpoints.json';

// Il tipo della tavolozza è uno solo: un colore che esiste in uno dei due schemi e non
// nell'altro non compila, ed è il motivo per cui i token sono un oggetto e non un elenco di
// variabili scritte a mano in un foglio.
export type Palette = {
  readonly background: string;
  readonly surface: string;
  readonly text: string;
  readonly textMuted: string;
  readonly accent: string;
  readonly onAccent: string;
  readonly danger: string;
  readonly focus: string;
};

export const lightPalette: Palette = {
  background: '#ffffff',
  surface: '#f4f4f5',
  text: '#18181b',
  textMuted: '#52525b',
  accent: '#1d4ed8',
  onAccent: '#ffffff',
  danger: '#b91c1c',
  focus: '#1d4ed8',
};

export const darkPalette: Palette = {
  background: '#09090b',
  surface: '#18181b',
  text: '#fafafa',
  textMuted: '#a1a1aa',
  accent: '#60a5fa',
  onAccent: '#09090b',
  danger: '#f87171',
  focus: '#93c5fd',
};

// Le misure non cambiano con lo schema di colori. In `rem` e non in pixel: seguono la
// dimensione del testo scelta da chi legge, e un ingrandimento del carattere non rompe il
// layout.
export const metrics = {
  space: { xs: '0.25rem', sm: '0.5rem', md: '1rem', lg: '1.5rem', xl: '2.5rem' },
  radius: { sm: '0.375rem', md: '0.75rem' },
  fontSize: { body: '1rem', small: '0.875rem', title: '1.5rem' },
  font: { body: 'system-ui, sans-serif' },
  // Bersaglio minimo di un elemento interattivo: 48 pixel alla dimensione di testo predefinita.
  size: { minTarget: '3rem', focusRing: '0.1875rem' },
} as const;

// Le larghezze a cui il layout cambia. Una variabile CSS dentro una media query non vale, quindi
// nei fogli si scrivono come valore: il file JSON lo leggono sia questo modulo sia stylelint,
// che nelle media query ammette solo questi valori.
export const breakpoints: Readonly<Record<'tablet' | 'desktop', string>> = breakpointValues;
```

```json
// web/src/shared/theme/breakpoints.json
{ "tablet": "40rem", "desktop": "64rem" }
```

## Dai token al foglio

```ts
// web/src/shared/theme/theme-stylesheet.ts

import { darkPalette, lightPalette, metrics } from './tokens';

const kebab = (name: string): string => name.replace(/[A-Z]/g, (c) => `-${c.toLowerCase()}`);

const declarations = (group: string, values: Readonly<Record<string, string>>): string[] =>
  Object.entries(values).map(([name, value]) => `--${kebab(group)}-${kebab(name)}: ${value};`);

// `color.textMuted` diventa `--color-text-muted`, `size.minTarget` diventa `--size-min-target`:
// il nome della variabile si deriva, non si sceglie, e non c'è un secondo elenco da tenere
// allineato.
export function themeStylesheet(): string {
  const measures = Object.entries(metrics).flatMap(([group, values]) =>
    declarations(group, values),
  );

  // Lo schema lo sceglie il browser: nessun hook che lo legga, nessun re-render quando cambia.
  // `color-scheme` porta nello schema giusto anche ciò che il sito non disegna — le barre di
  // scorrimento, i controlli nativi dei form.
  return [
    `:root { color-scheme: light dark; ${[...measures, ...declarations('color', lightPalette)].join(' ')} }`,
    `@media (prefers-color-scheme: dark) { :root { ${declarations('color', darkPalette).join(' ')} } }`,
  ].join('\n');
}
```

```css
/* web/src/shared/theme/global.css — l'unico foglio non modulare del sito */

*,
*::before,
*::after {
  box-sizing: border-box;
}

body {
  margin: 0;
  background: var(--color-background);
  color: var(--color-text);
  font-family: var(--font-body);
  font-size: var(--font-size-body);
  line-height: 1.5;
}

/* Con la preferenza attiva le transizioni non si accorciano: si annullano, una volta sola e per
   tutti i componenti, invece di chiedere a ognuno di ricordarsene. */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }
}
```

## Dove si monta

```tsx
// web/src/main.tsx

import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { RouterProvider } from 'react-router/dom';

import { router } from '@app/router';
import { defaultLocale } from '@bootstrap/config';
import { I18nProvider } from '@shared/i18n/i18n-provider';
import { pickLocale } from '@shared/i18n/catalogs';
import { themeStylesheet } from '@shared/theme/theme-stylesheet';
import '@shared/theme/global.css';

// Prima del primo render: un componente che parte senza variabili si disegna senza colori per
// un istante, e quell'istante si vede.
const theme = document.createElement('style');
theme.textContent = themeStylesheet();
document.head.append(theme);

const root = document.getElementById('root');
if (root === null) throw new Error('index.html non contiene #root');

createRoot(root).render(
  <StrictMode>
    <I18nProvider locale={pickLocale(navigator.languages, defaultLocale)}>
      <RouterProvider router={router} />
    </I18nProvider>
  </StrictMode>,
);
```

## Come si usano in un componente

```css
/* web/src/features/orders/components/order-row.module.css */

.order-row {
  display: flex;
  gap: var(--space-sm);
  padding: var(--space-md);
  border-radius: var(--radius-md);
  background: var(--color-surface);
}

.reference {
  color: var(--color-text-muted);
  font-size: var(--font-size-small);
}

@media (width >= 40rem) {
  .order-row {
    padding: var(--space-lg);
  }
}
```

```tsx
import styles from './order-row.module.css';

// …
<li className={styles.orderRow}>
  <span className={styles.reference}>{order.reference}</span>
</li>
```

Punti chiave:

-   **Un oggetto, due uscite.** `tokens.ts` lo leggono il foglio generato all'avvio e il test
    di contrasto; nessuno scrive a mano una variabile CSS che esiste già.
-   **I componenti non sanno quale schema sia attivo.** Leggono `var(--color-surface)`, e il
    valore lo decide la media query del foglio generato. Non esiste un `useAppTheme()`: sul
    web il tema non ha bisogno di un `Context`, e lo stato di presentazione condiviso si riduce
    alle traduzioni ([Etichette tradotte nei componenti](./i18n_hook.example.md)).
-   **I valori esadecimali esistono in un posto solo**, `tokens.ts`. Dentro un `.module.css`
    sono un errore che stylelint rifiuta
    ([Confini fra livelli](./architecture_test.example.md)).
-   Un valore che serve davvero a runtime — una larghezza che segue una misura — si scrive in
    linea partendo da una variabile: `` style={{ width: `calc(var(--space-md) * ${n})` }} ``.
-   I punti di rottura sono l'unico valore che si ripete in un foglio, perché una variabile
    dentro `@media` non vale: per questo sono due, stanno in un JSON che anche stylelint legge,
    e una media query con un'altra larghezza non passa
    ([Confini fra livelli](./architecture_test.example.md)).
