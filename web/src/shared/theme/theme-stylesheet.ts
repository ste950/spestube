import { metrics, palette } from './tokens';

const kebab = (name: string): string => name.replace(/[A-Z]/g, (c) => `-${c.toLowerCase()}`);

const declarations = (group: string, values: Readonly<Record<string, string>>): string[] =>
  Object.entries(values).map(([name, value]) => `--${kebab(group)}-${kebab(name)}: ${value};`);

// `color.textMuted` diventa `--color-text-muted`, `size.minTarget` diventa `--size-min-target`:
// il nome della variabile si deriva, non si sceglie.
export function themeStylesheet(): string {
  const measures = Object.entries(metrics).flatMap(([group, values]) =>
    declarations(group, values),
  );
  return `:root { color-scheme: light; ${[...measures, ...declarations('color', palette)].join(' ')} }`;
}
