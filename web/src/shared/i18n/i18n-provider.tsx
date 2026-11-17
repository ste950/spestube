import { useEffect, useMemo, type ReactNode } from 'react';
import { catalogs, type Locale } from './catalogs';
import { TranslateContext, type Translate } from './use-translate';

export function I18nProvider({ locale, children }: { locale: Locale; children: ReactNode }) {
  const translate = useMemo<Translate>(() => {
    const catalog = catalogs[locale];
    return (key, params) =>
      catalog[key].replace(/\{(\w+)\}/g, (_, name: string) => String(params?.[name] ?? ''));
  }, [locale]);

  // La lingua del documento segue quella delle etichette: è quella che il lettore di schermo
  // usa per pronunciarle.
  useEffect(() => {
    document.documentElement.lang = locale;
  }, [locale]);

  return <TranslateContext value={translate}>{children}</TranslateContext>;
}
