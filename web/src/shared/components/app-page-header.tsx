import styles from './app-page-header.module.css';

type AppPageHeaderProps = {
  title: string;
  subtitle?: string;
};

// L'intestazione di una schermata: un solo `<h1>`, che è ciò che il lettore di schermo
// annuncia all'arrivo sulla pagina.
export function AppPageHeader({ title, subtitle }: AppPageHeaderProps) {
  return (
    <header className={styles.header}>
      <h1 className={styles.title}>{title}</h1>
      {subtitle !== undefined && <p className={styles.subtitle}>{subtitle}</p>}
    </header>
  );
}
