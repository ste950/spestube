// Le icone del design, disegnate a tratto. Sono decorative: il nome accessibile lo porta
// sempre l'elemento che le contiene, quindi qui sono nascoste ai lettori di schermo.
const PATHS = {
  today: ['M4 13h4l2 3h4l2-3h4', 'M4 13l2-8h12l2 8v6H4z'],
  queue: ['M7 4h10v16l-5-4-5 4z'],
  channels: ['M4 7h16', 'M4 12h16', 'M4 17h10'],
  back: ['M15 5l-7 7 7 7'],
  close: ['M6 6l12 12', 'M18 6L6 18'],
  done: ['M12 3a9 9 0 1 0 0 18a9 9 0 0 0 0-18z', 'M8 12.5l2.5 2.5L16 9.5'],
  stop: [
    'M7.5 6h9A1.5 1.5 0 0 1 18 7.5v9a1.5 1.5 0 0 1-1.5 1.5h-9A1.5 1.5 0 0 1 6 16.5v-9A1.5 1.5 0 0 1 7.5 6z',
  ],
} as const;

export type AppIconName = keyof typeof PATHS;

// `className` accetta `undefined` perché è ciò che un CSS Module restituisce per il tipo:
// Vite non sa quali classi esistano in un foglio.
type AppIconProps = { name: AppIconName; className?: string | undefined };

export function AppIcon({ name, className }: AppIconProps) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      width="1em"
      height="1em"
      fill="none"
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
