import { AppIcon } from '@shared/components/app-icon';
import styles from './feed-end.module.css';

// La fine del feed è un elemento disegnato, non uno spazio vuoto: è ciò che dice a chi
// guarda che può chiudere l'app.
export function FeedEnd({ title, text }: { title: string; text: string }) {
  return (
    <div className={styles.end}>
      <AppIcon name="done" className={styles.icon} />
      <div className={styles.text}>
        <p className={styles.title}>{title}</p>
        <p className={styles.detail}>{text}</p>
      </div>
    </div>
  );
}
