import { describe, expect, it } from 'vitest';
import { darkPalette, lightPalette, type Palette } from '@shared/theme/tokens';

// Il rapporto di contrasto WCAG fra due colori #RRGGBB.
function luminance(hex: string): number {
  const channels = [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16) / 255);
  const [r = 0, g = 0, b = 0] = channels.map((c) =>
    c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4,
  );
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

function contrast(a: string, b: string): number {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return ((hi ?? 0) + 0.05) / ((lo ?? 0) + 0.05);
}

// Le coppie che l'app usa davvero per il testo: 4,5 per il testo normale, 3 per le icone
// e i bordi dei campi.
const TEXT_PAIRS: readonly [keyof Palette, keyof Palette][] = [
  ['text', 'background'],
  ['text', 'surface'],
  ['textSoft', 'surface'],
  ['textMuted', 'background'],
  ['textMuted', 'surface'],
  ['accent', 'background'],
  ['accent', 'surface'],
  ['onAccent', 'accent'],
  ['accentDark', 'accentTint'],
  ['danger', 'surface'],
  ['danger', 'background'],
  ['warningText', 'warningTint'],
  ['onPlayer', 'player'],
  ['onPlayerMuted', 'player'],
];

const UI_PAIRS: readonly [keyof Palette, keyof Palette][] = [
  ['focus', 'background'],
  ['focus', 'surface'],
];

describe.each([
  ['chiaro', lightPalette],
  ['scuro', darkPalette],
])('il tema %s', (_name, palette) => {
  it.each(TEXT_PAIRS)('%s su %s si legge (≥ 4,5)', (fg, bg) => {
    expect(contrast(palette[fg], palette[bg])).toBeGreaterThanOrEqual(4.5);
  });

  it.each(UI_PAIRS)('%s su %s si vede (≥ 3)', (fg, bg) => {
    expect(contrast(palette[fg], palette[bg])).toBeGreaterThanOrEqual(3);
  });
});
