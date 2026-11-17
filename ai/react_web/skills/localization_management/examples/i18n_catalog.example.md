I cataloghi delle etichette, in `web/src/shared/i18n/`. Il catalogo della lingua base è la
**fonte di verità delle chiavi**: da lui nasce il tipo, e gli altri devono soddisfarlo.

```ts
// web/src/shared/i18n/it.ts

export const it = {
  common_retry: 'Riprova',
  common_loading: 'Caricamento in corso',
  common_skip_to_content: 'Vai al contenuto',
  common_error_network: 'Connessione non disponibile.',
  common_error_unexpected: 'Qualcosa è andato storto.',

  orders_title: 'Ordini',
  orders_empty: 'Non c’è ancora nessun ordine.',
  orders_error_not_found: 'Questo ordine non esiste più.',
  orders_total: 'Totale: {amount}',
  orders_line_count: '{count} articoli',
  orders_status_draft: 'Bozza',
  orders_status_submitted: 'Inviato',
  orders_status_shipped: 'Spedito',
  orders_status_cancelled: 'Annullato',
};
```

```ts
// web/src/shared/i18n/keys.ts

import { it } from './it';

// Il tipo nasce dal catalogo base, non da una lista scritta a mano: una chiave rinominata in un
// posto solo non esiste, perché il tipo la segue.
export type TranslationKey = keyof typeof it;
```

```ts
// web/src/shared/i18n/en.ts

import type { TranslationKey } from './keys';

// L'annotazione è il controllo: una chiave mancante o di troppo qui non compila. È il
// motivo per cui i cataloghi sono moduli TypeScript e non file JSON caricati a runtime.
export const en: Record<TranslationKey, string> = {
  common_retry: 'Retry',
  common_loading: 'Loading',
  common_skip_to_content: 'Skip to content',
  common_error_network: 'No connection available.',
  common_error_unexpected: 'Something went wrong.',

  orders_title: 'Orders',
  orders_empty: 'No orders yet.',
  orders_error_not_found: 'This order no longer exists.',
  orders_total: 'Total: {amount}',
  orders_line_count: '{count} items',
  orders_status_draft: 'Draft',
  orders_status_submitted: 'Submitted',
  orders_status_shipped: 'Shipped',
  orders_status_cancelled: 'Cancelled',
};
```

```ts
// web/src/shared/i18n/catalogs.ts

import { en } from './en';
import { it } from './it';
import type { TranslationKey } from './keys';

export type Locale = 'it' | 'en';

export const catalogs: Record<Locale, Record<TranslationKey, string>> = { it, en };
```

Punti chiave:

-   Le chiavi hanno il **prefisso della feature** (`orders_`) o `common_` quando non appartengono
    a nessuna: aprendo il catalogo si vede a colpo d'occhio a chi appartiene ogni riga, e una
    feature eliminata porta via le sue etichette.
-   I parametri sono **nominati** (`{amount}`), mai posizionali: l'ordine dei segnaposto cambia
    da lingua a lingua.
-   Una chiave con un plurale non si risolve con due chiavi e un `if` sparso nei componenti: si
    tiene la regola in un unico punto di presentazione e si passa il conteggio come parametro.
-   Nessun testo nel catalogo dipende dal dominio: qui ci sono etichette, non entità. Quale
    etichetta mostrare per un fallimento lo decide `state/`
    ([Result e fallimenti](../../../examples/result_pattern.example.md)).
-   Le lingue del progetto e quale sia la base le dichiara il `CLAUDE.md` del progetto: il
    framework impone la forma, non l'elenco.
