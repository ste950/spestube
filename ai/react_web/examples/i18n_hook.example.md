Le etichette tradotte sono l'unica cosa che in questo stack attraversa l'albero dei componenti
con un `Context`: il tema passa dalle [variabili CSS](./theme_tokens.example.md), le
dipendenze di dominio dal [composition root](./bootstrap_dependencies.example.md). Il formato
dei cataloghi e le regole sulle chiavi sono nella skill
[Gestione delle etichette localizzate](../skills/localization_management/localization_management.skill.md).

## La scelta della lingua

```ts
// web/src/shared/i18n/catalogs.ts — aggiunta, accanto al tipo `Locale`

const supported: readonly Locale[] = ['it', 'en'];

const isLocale = (value: string): value is Locale =>
  (supported as readonly string[]).includes(value);

// La prima lingua del browser che il sito conosce, confrontata sulla lingua e non sulla
// regione: chi chiede `en-GB` riceve `en`, e nessuno resta con la lingua di ripiego perché la
// sua variante regionale non ha un catalogo.
export function pickLocale(requested: readonly string[], fallback: Locale): Locale {
  return requested.map((tag) => tag.split('-')[0] ?? '').find(isLocale) ?? fallback;
}
```

## Il contesto e l'hook

```ts
// web/src/shared/i18n/use-translate.ts

import { createContext, useContext } from 'react';
import type { TranslationKey } from './keys';

export type Translate = (key: TranslationKey, params?: Record<string, string | number>) => string;

export const TranslateContext = createContext<Translate>(() => '');

export const useTranslate = (): Translate => useContext(TranslateContext);
```

## Il provider

```tsx
// web/src/shared/i18n/i18n-provider.tsx

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
  // usa per pronunciarle, e una pagina in inglese letta con la pronuncia italiana non si capisce.
  useEffect(() => {
    document.documentElement.lang = locale;
  }, [locale]);

  return <TranslateContext value={translate}>{children}</TranslateContext>;
}
```

Due file e non uno perché il ricaricamento a caldo di Vite funziona solo su moduli che esportano
**soltanto** componenti: un file con il provider e l'hook insieme ricaricherebbe la pagina intera
a ogni modifica, e il linter lo segnala
([Confini fra livelli](./architecture_test.example.md)). Il contesto sta con l'hook, che è un
modulo senza componenti; il provider sta da solo.

## Dove si monta

In `main.tsx`, sopra il router, e da nessun'altra parte — vedi
[Token di tema](./theme_tokens.example.md#dove-si-monta):

```tsx
<I18nProvider locale={pickLocale(navigator.languages, defaultLocale)}>
  <RouterProvider router={router} />
</I18nProvider>
```

## Come si usa in un componente

```tsx
const t = useTranslate();

return (
  <>
    <title>{t('orders_title')}</title>
    <h1>{t('orders_title')}</h1>
    <p>{t('orders_line_count', { count: order.lines.length })}</p>
  </>
);
```

Punti chiave:

-   `TranslationKey` è un tipo, non una stringa qualunque: una chiave inesistente o rimasta
    dopo una rinomina non compila.
-   `t` prende i parametri come oggetto: l'ordine dei segnaposto in una frase cambia da lingua a
    lingua, e una lista posizionale legherebbe la traduzione all'ordine dell'originale.
-   La lingua la sceglie `main.tsx`, una volta, dalle preferenze del browser; il componente non
    sa quale sia. Un selettore di lingua, se il progetto lo vuole, cambia la `locale` del
    provider e non tocca nessun componente.
-   `<html lang>` segue la lingua scelta: è un requisito di
    [Accessibilità](../../quality/rules/accessibility.rule.md), non un dettaglio.
-   Il `<title>` di una pagina è un'etichetta come le altre: React lo porta nell'intestazione
    del documento, e passa da `t` come tutto il testo visibile.
-   Il provider è composizione dell'app, non stato di una feature: le feature leggono con
    `useTranslate()` e non montano provider.
