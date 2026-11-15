import styles from './app-skip-link.module.css';

// Il primo elemento della pagina: chi naviga da tastiera salta la barra delle sezioni e
// arriva al contenuto. Resta fuori schermo finché non riceve il focus.
export function AppSkipLink({ targetId, label }: { targetId: string; label: string }) {
  return (
    <a className={styles.skipLink} href={`#${targetId}`}>
      {label}
    </a>
  );
}
