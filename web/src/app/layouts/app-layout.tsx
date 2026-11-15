import { Suspense } from 'react';
import { Link, Outlet, useLocation } from 'react-router';
import { AppIcon, type AppIconName } from '@shared/components/app-icon';
import { AppSkipLink } from '@shared/components/app-skip-link';
import { AppSpinner } from '@shared/components/app-spinner';
import { AppVisuallyHidden } from '@shared/components/app-visually-hidden';
import type { TranslationKey } from '@shared/i18n/keys';
import { useTranslate } from '@shared/i18n/use-translate';
import { useNavBadges } from '@shared/navigation/nav-badges';
import styles from './app-layout.module.css';

type Tab = { to: string; label: TranslationKey; icon: AppIconName; alsoUnder?: string };

// Tre sezioni e nessuna quarta: niente ricerca, niente consigliati, niente profilo.
const TABS: readonly Tab[] = [
  // Guardare un video è una tappa di Oggi: mentre il player è aperto la sezione resta accesa.
  { to: '/', label: 'common_nav_today', icon: 'today', alsoUnder: '/guarda' },
  { to: '/da-vedere', label: 'common_nav_queue', icon: 'queue' },
  { to: '/canali', label: 'common_nav_channels', icon: 'channels' },
];

export function AppLayout() {
  const t = useTranslate();
  const { pathname } = useLocation();
  const counts = useNavBadges((state) => state.counts);

  return (
    <div className={styles.layout}>
      <AppSkipLink targetId="contenuto" label={t('common_skip_to_content')} />
      <main id="contenuto" className={styles.main} tabIndex={-1}>
        {/* Una pagina pigra non ancora scaricata sospende il render: il segnaposto sta dentro
            il layout, così la barra delle sezioni resta al suo posto. */}
        <Suspense fallback={<AppSpinner label={t('common_loading')} />}>
          <Outlet />
        </Suspense>
      </main>
      <nav className={styles.tabs} aria-label={t('common_nav_label')}>
        {TABS.map((tab) => {
          // La sezione attiva si calcola qui e non con `NavLink`: sotto /guarda deve restare
          // accesa Oggi, e `NavLink` segna come corrente solo il proprio percorso.
          const active =
            pathname === tab.to ||
            (tab.alsoUnder !== undefined && pathname.startsWith(tab.alsoUnder));
          const count = counts[tab.to] ?? 0;
          return (
            <Link
              key={tab.to}
              to={tab.to}
              className={active ? styles.tabActive : styles.tab}
              {...(active ? { 'aria-current': 'page' as const } : {})}
            >
              <span className={styles.tabGlyph}>
                <AppIcon name={tab.icon} className={styles.tabIcon} />
                {count > 0 && (
                  <span className={styles.badge} aria-hidden="true">
                    {count}
                  </span>
                )}
              </span>
              <span>
                {t(tab.label)}
                {/* Il numero sull'icona è solo da vedere: chi ascolta lo riceve qui, con la
                    parola che gli dà senso («Oggi, 2 nuovi»). */}
                {count > 0 && (
                  <AppVisuallyHidden>
                    {`, ${count} ${t(count === 1 ? 'common_nav_badge_one' : 'common_nav_badge_many')}`}
                  </AppVisuallyHidden>
                )}
              </span>
            </Link>
          );
        })}
      </nav>
    </div>
  );
}
