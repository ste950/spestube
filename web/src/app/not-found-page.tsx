import { Link } from 'react-router';
import { AppPageHeader } from '@shared/components/app-page-header';
import { useTranslate } from '@shared/i18n/use-translate';
import styles from './status-page.module.css';

export function NotFoundPage() {
  const t = useTranslate();
  return (
    <section className={styles.page}>
      <title>{`${t('common_not_found_title')} · ${t('common_app_name')}`}</title>
      <AppPageHeader title={t('common_not_found_title')} subtitle={t('common_not_found_text')} />
      <Link className={styles.link} to="/">
        {t('common_back_to_today')}
      </Link>
    </section>
  );
}
