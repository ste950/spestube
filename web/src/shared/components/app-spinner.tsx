import styles from './app-spinner.module.css';

// Il testo è il contenuto, non un'etichetta nascosta: chi vede e chi ascolta ricevono la
// stessa informazione, e non c'è un'animazione da spegnere per chi la preferisce ferma.
export function AppSpinner({ label }: { label: string }) {
  return (
    <p className={styles.spinner} role="status">
      {label}…
    </p>
  );
}
