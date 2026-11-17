// Le icone del design, disegnate a tratto. Sono decorative: il nome accessibile lo porta
// sempre l'elemento che le contiene, quindi qui sono nascoste ai lettori di schermo.
const PATHS = {
  today: ['M4 13h4l2 3h4l2-3h4', 'M4 13l2-8h12l2 8v6H4z'],
  save: ['M7 4h10v16l-5-4-5 4z'],
  collections: ['M4 6h11', 'M4 11h11', 'M4 16h6', 'M15 14.5v5l4.5-2.5z'],
  settings: [
    'M4 7h9.5',
    'M18.5 7H20',
    'M16 4.5a2.5 2.5 0 1 1 0 5a2.5 2.5 0 1 1 0-5z',
    'M4 17h1.5',
    'M10.5 17H20',
    'M8 14.5a2.5 2.5 0 1 1 0 5a2.5 2.5 0 1 1 0-5z',
  ],
  back: ['M15 5l-7 7 7 7'],
  forward: ['M9 5l7 7-7 7'],
  up: ['M6 15l6-6 6 6'],
  down: ['M6 9l6 6 6-6'],
  plus: ['M12 5v14', 'M5 12h14'],
  close: ['M6 6l12 12', 'M18 6L6 18'],
  done: ['M12 3a9 9 0 1 0 0 18a9 9 0 0 0 0-18z', 'M8 12.5l2.5 2.5L16 9.5'],
  stop: [
    'M7.5 6h9A1.5 1.5 0 0 1 18 7.5v9a1.5 1.5 0 0 1-1.5 1.5h-9A1.5 1.5 0 0 1 6 16.5v-9A1.5 1.5 0 0 1 7.5 6z',
  ],
} as const;

export type AppIconName = keyof typeof PATHS;

// `className` accetta `undefined` perché è ciò che un CSS Module restituisce per il tipo:
// Vite non sa quali classi esistano in un foglio.
type AppIconProps = {
  name: AppIconName;
  className?: string | undefined;
  /** Riempita invece che a tratto: per lo stato «attivo» di un'icona, come un video salvato. */
  filled?: boolean;
};

export function AppIcon({ name, className, filled = false }: AppIconProps) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      width="1em"
      height="1em"
      fill={filled ? 'currentColor' : 'none'}
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      focusable="false"
    >
      {PATHS[name].map((d) => (
        <path key={d} d={d} />
      ))}
    </svg>
  );
}
