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

type Place = { pathname: string; search: string };

type Tab = {
  to: string;
  label: TranslationKey;
  icon: AppIconName;
  isActive: (place: Place) => boolean;
};

const under = (pathname: string, root: string) =>
  pathname === root || pathname.startsWith(`${root}/`);

// Un video aperto da una raccolta porta con sé `?raccolta=`: è così che Guarda sa dove
// tornare, e la barra quale sezione tenere accesa.
const fromCollection = (search: string) => new URLSearchParams(search).has('raccolta');

// Tre sezioni e nessuna quarta: niente ricerca, niente consigliati, niente profilo.
const TABS: readonly Tab[] = [
  // Guardare un video è una tappa della sezione da cui lo si è aperto.
  {
    to: '/',
    label: 'common_nav_today',
    icon: 'today',
    isActive: ({ pathname, search }) =>
      pathname === '/' || (under(pathname, '/guarda') && !fromCollection(search)),
  },
  {
    to: '/raccolte',
    label: 'common_nav_collections',
    icon: 'collections',
    isActive: ({ pathname, search }) =>
      under(pathname, '/raccolte') || (under(pathname, '/guarda') && fromCollection(search)),
  },
  {
    to: '/impostazioni',
    label: 'common_nav_settings',
    icon: 'settings',
    isActive: ({ pathname }) => under(pathname, '/impostazioni'),
  },
];

export function AppLayout() {
  const t = useTranslate();
  const { pathname, search } = useLocation();
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
          // La sezione attiva si calcola qui e non con `NavLink`: sotto /guarda dipende da
          // dove si viene, e `NavLink` segna come corrente solo il proprio percorso.
          const active = tab.isActive({ pathname, search });
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
