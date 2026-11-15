import { darkPalette, lightPalette, metrics, type Palette } from './tokens';

const kebab = (name: string): string => name.replace(/[A-Z]/g, (c) => `-${c.toLowerCase()}`);

const declarations = (group: string, values: Readonly<Record<string, string>>): string[] =>
  Object.entries(values).map(([name, value]) => `--${kebab(group)}-${kebab(name)}: ${value};`);

const scheme = (name: 'light' | 'dark', palette: Palette): string =>
  [`color-scheme: ${name};`, ...declarations('color', palette)].join(' ');

// `color.textMuted` diventa `--color-text-muted`, `size.minTarget` diventa `--size-min-target`:
// il nome della variabile si deriva, non si sceglie.
//
// Tre casi: senza scelta decide il sistema (la media query), con `data-theme` decide chi
// usa l'app. La regola con l'attributo viene dopo la media query, quindi vince sempre.
export function themeStylesheet(): string {
  const measures = Object.entries(metrics).flatMap(([group, values]) =>
    declarations(group, values),
  );
  return [
    `:root { ${measures.join(' ')} ${scheme('light', lightPalette)} }`,
    `@media (prefers-color-scheme: dark) { :root:not([data-theme='light']) { ${scheme('dark', darkPalette)} } }`,
    `:root[data-theme='dark'] { ${scheme('dark', darkPalette)} }`,
  ].join('\n');
}
