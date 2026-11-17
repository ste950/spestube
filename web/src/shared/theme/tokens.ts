import breakpointValues from './breakpoints.json';

// Il tipo della tavolozza è uno solo: un colore che esiste in uno schema e non nell'altro
// non compila, ed è il motivo per cui i token sono un oggetto e non un elenco di variabili
// scritte a mano in un foglio.
export type Palette = {
  readonly background: string;
  readonly surface: string;
  readonly surfaceMuted: string;
  readonly border: string;
  readonly borderField: string;
  readonly borderDashed: string;
  readonly divider: string;
  readonly text: string;
  readonly textSoft: string;
  readonly textMuted: string;
  readonly accent: string;
  readonly accentDark: string;
  readonly accentTint: string;
  readonly onAccent: string;
  readonly focus: string;
  readonly player: string;
  readonly onPlayer: string;
  readonly onPlayerMuted: string;
  readonly playerAccent: string;
  readonly thumbnail: string;
  readonly badge: string;
  readonly scrim: string;
  readonly warningText: string;
  readonly warningTint: string;
  readonly warningBorder: string;
  readonly danger: string;
  readonly dangerTint: string;
};

// I colori del design di SpesTube (docs/design).
export const lightPalette: Palette = {
  background: '#F3F4F7',
  surface: '#FFFFFF',
  surfaceMuted: '#EEF0F4',
  border: '#DCE0E8',
  borderField: '#C5CBD6',
  borderDashed: '#B9C2D2',
  divider: '#ECEEF2',
  text: '#15181F',
  textSoft: '#4A5160',
  textMuted: '#5A6170',
  accent: '#2557C7',
  accentDark: '#1A3F94',
  accentTint: '#E4ECFB',
  onAccent: '#FFFFFF',
  focus: '#1A3F94',
  player: '#1D212B',
  onPlayer: '#FFFFFF',
  onPlayerMuted: '#C9CFDA',
  playerAccent: '#A9C4F5',
  thumbnail: '#C9CED8',
  badge: '#2B2F38',
  // Il velo dietro una finestra: il colore del testo, semitrasparente.
  scrim: '#15181F8C',
  warningText: '#9A4C1A',
  warningTint: '#F1E4DA',
  warningBorder: '#E3C9B6',
  danger: '#A3271B',
  dangerTint: '#FBEAE7',
};

// Lo schema scuro tiene la stessa voce del chiaro: superfici fredde con un filo di blu, e
// il blu dell'app schiarito, perché #2557C7 su un fondo scuro non si legge. Le coppie di
// contrasto le verifica un test.
export const darkPalette: Palette = {
  background: '#0F1218',
  surface: '#181C24',
  surfaceMuted: '#20252F',
  border: '#2A303C',
  borderField: '#3D4555',
  borderDashed: '#434C5D',
  divider: '#232833',
  text: '#E8EBF1',
  textSoft: '#B9C0CC',
  textMuted: '#A0A8B6',
  accent: '#8AADF4',
  accentDark: '#B9CDF8',
  accentTint: '#1D2A46',
  onAccent: '#0B1020',
  focus: '#B9CDF8',
  player: '#05070A',
  onPlayer: '#FFFFFF',
  onPlayerMuted: '#C9CFDA',
  playerAccent: '#A9C4F5',
  thumbnail: '#2A303C',
  badge: '#05070A',
  scrim: '#000000A6',
  warningText: '#F0B184',
  warningTint: '#33251B',
  warningBorder: '#5C4030',
  danger: '#F29387',
  dangerTint: '#3A1E1B',
};

// Le misure sono in `rem`: seguono la dimensione del testo scelta da chi legge.
export const metrics = {
  space: {
    xxs: '0.25rem',
    xs: '0.375rem',
    sm: '0.5rem',
    md: '0.75rem',
    lg: '1rem',
    xl: '1.125rem',
    xxl: '1.75rem',
  },
  radius: { xs: '0.25rem', sm: '0.5rem', md: '0.75rem', lg: '0.875rem', pill: '62.5rem' },
  fontSize: {
    badge: '0.6875rem',
    caption: '0.75rem',
    small: '0.8125rem',
    label: '0.875rem',
    body: '0.9375rem',
    // Il testo dentro un campo: sotto i 16 pixel Safari su iPhone ingrandisce la pagina
    // quando il campo prende il focus.
    field: '1rem',
    lead: '1.0625rem',
    heading: '1.375rem',
    display: '2.125rem',
  },
  font: {
    body: "'Inter Variable', Inter, system-ui, sans-serif",
    heading: "'Space Grotesk Variable', 'Space Grotesk', system-ui, sans-serif",
  },
  size: {
    // Bersaglio minimo di un elemento interattivo: 44 pixel alla dimensione di testo
    // predefinita, il minimo che il design si è dato.
    minTarget: '2.75rem',
    tabTarget: '3rem',
    tabMinWidth: '5.5rem',
    focusRing: '0.1875rem',
    icon: '1.375rem',
    thumbnailWidth: '7rem',
    // L'app è disegnata per il telefono: su uno schermo largo resta una colonna leggibile.
    pageMax: '32rem',
  },
} as const;

// Una variabile CSS dentro una media query non vale, quindi nei fogli queste larghezze si
// scrivono come valore: il file JSON lo leggono sia questo modulo sia stylelint.
export const breakpoints: Readonly<Record<'tablet' | 'desktop', string>> = breakpointValues;
