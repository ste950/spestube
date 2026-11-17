import { AppPageHeader } from '@shared/components/app-page-header';
import { useTranslate } from '@shared/i18n/use-translate';
import styles from './status-page.module.css';

// Qui arriva solo un errore di render, cioè un difetto dell'app. Gli errori previsti (la
// rete, un video che non c'è) sono stati di pagina e non passano da qui. Il collegamento è
// un `<a>` e non un `<Link>`: dopo un difetto conviene ricaricare l'app da capo.
export function RouteErrorPage() {
  const t = useTranslate();
  return (
    <section className={styles.page}>
      <title>{t('common_app_name')}</title>
      <AppPageHeader title={t('common_error_unexpected')} />
      <a className={styles.link} href="/">
        {t('common_back_to_today')}
      </a>
    </section>
  );
}
