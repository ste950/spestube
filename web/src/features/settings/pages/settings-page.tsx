import { useEffect } from 'react';
import { Link } from 'react-router';
import { AppIcon } from '@shared/components/app-icon';
import { AppMessage } from '@shared/components/app-message';
import { AppPageHeader } from '@shared/components/app-page-header';
import { renderMessage } from '@shared/i18n/message';
import { useTranslate } from '@shared/i18n/use-translate';
import { ThemePicker } from '../components/theme-picker';
import { useSettingsStore } from '../state/use-settings-store';
import styles from './settings-page.module.css';

const channelCountLabel = (count: number) => {
  if (count === 0) return 'settings_channels_count_zero';
  return count === 1 ? 'settings_channels_count_one' : 'settings_channels_count_many';
};

// La pagina indice delle impostazioni: gruppi di voci, ognuno con un titolo. Una voce nuova
// è una riga in più in un gruppo, o un gruppo nuovo.
export function SettingsPage() {
  const t = useTranslate();
  const channelCount = useSettingsStore((state) => state.channelCount);
  const version = useSettingsStore((state) => state.version);
  const theme = useSettingsStore((state) => state.theme);
  const error = useSettingsStore((state) => state.error);
  const load = useSettingsStore((state) => state.load);
  const setTheme = useSettingsStore((state) => state.setTheme);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <section className={styles.page}>
      <title>{`${t('settings_title')} · ${t('common_app_name')}`}</title>
      <AppPageHeader title={t('settings_title')} />
      {error !== null && <AppMessage tone="error" text={renderMessage(t, error)} />}

      <section className={styles.group} aria-labelledby="impostazioni-contenuti">
        <h2 id="impostazioni-contenuti" className={styles.groupTitle}>
          {t('settings_group_content')}
        </h2>
        <div className={styles.card}>
          <Link to="/impostazioni/canali" className={styles.row}>
            <span className={styles.rowLabel}>{t('settings_channels')}</span>
            {channelCount !== null && (
              <span className={styles.rowValue}>
                {t(channelCountLabel(channelCount), { count: channelCount })}
              </span>
            )}
            <AppIcon name="forward" className={styles.chevron} />
          </Link>
        </div>
      </section>

      <section className={styles.group} aria-labelledby="impostazioni-aspetto">
        <h2 id="impostazioni-aspetto" className={styles.groupTitle}>
          {t('settings_group_appearance')}
        </h2>
        <div className={styles.cardPadded}>
          <ThemePicker value={theme} onChange={(next) => void setTheme(next)} />
        </div>
      </section>

      {version !== null && <p className={styles.version}>{t('settings_version', { version })}</p>}
    </section>
  );
}
