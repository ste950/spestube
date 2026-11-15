L'infrastruttura che tutto il resto dà per scontata: un esito che viaggia nel tipo di ritorno,
e un vocabolario chiuso di modi di fallire. Il perché di questa forma — e le alternative
scartate — è in
[Result e fallimenti](../rules/domain_layer.rule.md#%EF%B8%8F-result-e-fallimenti).

```ts
// web/src/core/result/failure.ts

// Il vocabolario è condiviso da tutti i soggetti: un soggetto ne dichiara uno proprio solo
// se ha un modo di fallire che non esiste altrove.
export type Failure =
  | { readonly kind: 'network'; readonly status: number | null; readonly cause: unknown }
  | { readonly kind: 'notFound' }
  | { readonly kind: 'parse'; readonly at: string; readonly cause: unknown }
  | { readonly kind: 'storage'; readonly cause: unknown };

// `cause` è `unknown` e non `Error`: quello che una libreria lancia non è necessariamente un
// `Error`, e chi assorbe deve poterlo riportare senza perderlo.
export const networkFailure = (status: number | null, cause: unknown): Failure => ({
  kind: 'network',
  status,
  cause,
});

export const notFoundFailure = (): Failure => ({ kind: 'notFound' });

// `at` dice quale forma non corrispondeva, non il messaggio della libreria: serve a trovare
// il confine che ha ceduto, e quello di una libreria di parsing non lo dice.
export const parseFailure = (at: string, cause: unknown): Failure => ({
  kind: 'parse',
  at,
  cause,
});

export const storageFailure = (cause: unknown): Failure => ({ kind: 'storage', cause });
```

```ts
// web/src/core/result/result.ts

import type { Failure } from './failure';

export type Result<T> =
  { readonly ok: true; readonly value: T } | { readonly ok: false; readonly failure: Failure };

export const ok = <T>(value: T): Result<T> => ({ ok: true, value });

export const fail = <T = never>(failure: Failure): Result<T> => ({ ok: false, failure });

// Il ramo di fallimento restituisce `result` **così com'è**: non contiene `T`, quindi è già un
// `Result<U>` valido e non c'è niente da ricostruire. È il vantaggio concreto dell'unione
// rispetto a una classe con `fold`, dove ogni esito andrebbe istanziato di nuovo.
export function mapOk<T, U>(result: Result<T>, fn: (value: T) => U): Result<U> {
  return result.ok ? ok(fn(result.value)) : result;
}
```

```ts
// web/src/core/exhaustive.ts

// Chiamarla in un ramo di default costringe il compilatore a verificare che quel ramo sia
// irraggiungibile: quando l'insieme cresce, il codice smette di compilare qui invece di
// ricadere in silenzio su un comportamento di ripiego.
export function assertNever(value: never): never {
  throw new Error(`caso non gestito: ${JSON.stringify(value)}`);
}
```

Come si consuma un `Result` — il restringimento lo fa il controllo di flusso, senza imparare
nessuna funzione nuova:

```ts
const result = await orderRepository.listOrders();
if (!result.ok) {
  return { orders: [], error: describeFailure(result.failure) };
}

// Qui `result.value` esiste, e prima di questa riga per il compilatore non esisteva.
return { orders: result.value, error: null };
```

E come si traduce una `Failure` in qualcosa che la UI sappia mostrare. È l'ultimo punto in cui
quel vocabolario esiste: sopra di qui non si parla più di `kind`.

```ts
// web/src/features/orders/state/describe-failure.ts

import { assertNever } from '@core/exhaustive';
import type { Failure } from '@core/result/failure';
import type { TranslationKey } from '@shared/i18n/keys';

export function describeFailure(failure: Failure): TranslationKey {
  switch (failure.kind) {
    case 'network':
      return 'common_error_network';
    case 'notFound':
      return 'orders_error_not_found';
    case 'parse':
    case 'storage':
      // Un dato illeggibile e uno storage rotto sono lo stesso fatto per chi guarda: non si
      // può mostrare niente, e riprovare non cambia nulla.
      return 'common_error_unexpected';
    default:
      return assertNever(failure);
  }
}
```

Punti chiave:

-   Un fallimento **previsto** è un esito e sta nel tipo di ritorno; le eccezioni restano alle
    librerie e si fermano dentro le sorgenti
    ([Errori](../rules/data_layer.rule.md#errori)).
-   Il vocabolario delle `Failure` vive in `core/result/` e attraversa l'app intatto: il
    repository non lo traduce, lo store lo traduce una volta sola.
-   Il `switch` su `kind` chiuso da `assertNever` è ciò che rende il vocabolario estendibile
    senza paura: nascere un modo nuovo di fallire rompe la compilazione dove manca il caso.
-   Il tipo `Result` non ha metodi. `mapOk` è una funzione libera, e ogni altra utilità che
    servisse va aggiunta allo stesso modo: un metodo non restringerebbe il tipo di `this`, che
    è esattamente ciò che serve qui.
