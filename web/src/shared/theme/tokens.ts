import breakpointValues from './breakpoints.json';

// I colori del design di SpesTube (docs/design). Lo schema è uno solo, chiaro: è quello
// disegnato, e un tema scuro si aggiunge qui quando esiste un disegno da cui prenderlo.
export const palette = {
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
  track: '#E6E9EF',
  trackFill: '#8FA3C9',
  warning: '#B8642C',
  warningText: '#9A4C1A',
  warningTint: '#F1E4DA',
  warningBorder: '#E3C9B6',
  danger: '#A3271B',
  dangerTint: '#FBEAE7',
} as const;

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
