// Il tema scelto si scrive come attributo della pagina, e il foglio del tema fa il resto.
// «Come il sistema» è l'assenza dell'attributo: decide la media query del browser.
export type AppliedTheme = 'system' | 'light' | 'dark';

export function applyTheme(theme: AppliedTheme): void {
  if (theme === 'system') {
    delete document.documentElement.dataset.theme;
  } else {
    document.documentElement.dataset.theme = theme;
  }
}
