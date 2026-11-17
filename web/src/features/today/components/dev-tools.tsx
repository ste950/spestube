import { AppButton } from '@shared/components/app-button';
import styles from './dev-tools.module.css';

type DevToolsProps = {
  title: string;
  text: string;
  actionLabel: string;
  busy: boolean;
  onRelease: () => void;
};

// Visibile solo su un'installazione di sviluppo. Un bottone che fa arrivare video a comando
// è il contrario del prodotto: per questo è dichiarato come strumento, e non somiglia al resto.
export function DevTools({ title, text, actionLabel, busy, onRelease }: DevToolsProps) {
  return (
    <aside className={styles.tools} aria-label={title}>
      <p className={styles.title}>{title}</p>
      <p className={styles.text}>{text}</p>
      <AppButton label={actionLabel} variant="quiet" onClick={onRelease} disabled={busy} />
    </aside>
  );
}
