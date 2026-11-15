import { AppIcon } from '@shared/components/app-icon';
import styles from './stop-screen.module.css';

// Prende il posto del player quando il video finisce: nessun video successivo, nessun conto
// alla rovescia. Il passo dopo lo decide chi guarda.
export function StopScreen({ title, text }: { title: string; text: string }) {
  return (
    <div className={styles.stop} role="status">
      <AppIcon name="stop" className={styles.icon} />
      <p className={styles.title}>{title}</p>
      <p className={styles.text}>{text}</p>
    </div>
  );
}
