Le due sorgenti che il [repository](./repository_impl.example.md) compone: la fonte remota,
che parla HTTP, e la copia locale, che parla con lo storage del browser. Sono gli unici moduli
del sito in cui compaiono `fetch` e `localStorage`, e il punto in cui quello che arriva da
fuori — `any` per il compilatore — diventa un modello controllato
([Sorgenti](../rules/data_layer.rule.md#-sorgenti)).

## Il controllo di forma

```ts
// web/src/data/orders/models/order-model-guard.ts

import type { OrderLineModel, OrderModel } from './order-model';

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null;

const isLine = (value: unknown): value is OrderLineModel =>
  isRecord(value) &&
  typeof value.sku === 'string' &&
  typeof value.quantity === 'number' &&
  typeof value.unit_price_cents === 'number';

// Codice e non una `as`: una `as` è un'affermazione che nessuno verifica, e farebbe arrivare
// `undefined` dentro un'entità lontano dal confine che avrebbe dovuto fermarlo.
export const isOrderModel = (value: unknown): value is OrderModel =>
  isRecord(value) &&
  typeof value.id === 'string' &&
  typeof value.reference === 'string' &&
  typeof value.status === 'string' &&
  typeof value.placed_at === 'string' &&
  Array.isArray(value.lines) &&
  value.lines.every(isLine);
```

## La fonte remota

```ts
// web/src/data/orders/sources/http-order-source.ts

import { networkFailure, notFoundFailure, parseFailure } from '@core/result/failure';
import { fail, ok, type Result } from '@core/result/result';
import type { OrderModel } from '../models/order-model';
import { isOrderModel } from '../models/order-model-guard';
import type { OrderSource } from './order-source';

// La factory annota il tipo di ritorno: è lì che TypeScript verifica che l'oggetto soddisfi il
// contratto, e chi lo riceve conosce il contratto, non l'implementazione.
export function createHttpOrderSource(baseUrl: string): OrderSource {
  async function request(path: string, init?: RequestInit): Promise<Result<unknown>> {
    try {
      const response = await fetch(`${baseUrl}${path}`, {
        ...init,
        headers: { Accept: 'application/json', ...init?.headers },
      });
      if (response.status === 404) return fail(notFoundFailure());
      if (!response.ok) return fail(networkFailure(response.status, null));

      // `unknown` e non `any`: da qui in avanti il compilatore pretende un controllo prima di
      // ogni lettura.
      const body: unknown = await response.json();

      return ok(body);
    } catch (error) {
      // `fetch` rifiuta solo quando la richiesta non è partita o non è tornata — rete assente,
      // CORS, richiesta annullata — e `json()` quando il corpo non è JSON. Entrambi sono
      // esiti previsti, e si fermano qui.
      return fail(networkFailure(null, error));
    }
  }

  function one(raw: Result<unknown>, at: string): Result<OrderModel> {
    if (!raw.ok) return raw;

    return isOrderModel(raw.value) ? ok(raw.value) : fail(parseFailure(at, raw.value));
  }

  return {
    async fetchOrders(statuses) {
      const query = new URLSearchParams(statuses.map((status) => ['status', status]));
      const raw = await request(`/orders?${query}`);
      if (!raw.ok) return raw;

      return Array.isArray(raw.value) && raw.value.every(isOrderModel)
        ? ok(raw.value)
        : fail(parseFailure('OrderModel[]', raw.value));
    },

    async fetchOrder(orderId) {
      return one(await request(`/orders/${encodeURIComponent(orderId)}`), 'OrderModel');
    },

    async updateStatus(orderId, status) {
      const raw = await request(`/orders/${encodeURIComponent(orderId)}/status`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status }),
      });

      return one(raw, 'OrderModel');
    },
  };
}
```

## La copia locale

```ts
// web/src/data/orders/sources/local-order-cache.ts

import { parseFailure, storageFailure } from '@core/result/failure';
import { fail, ok, type Result } from '@core/result/result';
import type { OrderModel } from '../models/order-model';
import { isOrderModel } from '../models/order-model-guard';
import type { OrderCache } from './order-cache';

// Versionata: un sito aggiornato può trovare nel browser una copia scritta dalla versione
// precedente con un'altra forma, e cambiare la chiave è più semplice che migrarla.
const KEY = 'orders.cache.v1';

export function createLocalOrderCache(): OrderCache {
  return {
    async read() {
      try {
        const stored = localStorage.getItem(KEY);
        if (stored === null) return ok([]);

        const parsed: unknown = JSON.parse(stored);

        return Array.isArray(parsed) && parsed.every(isOrderModel)
          ? ok(parsed)
          : fail(parseFailure('OrderModel[] in cache', parsed));
      } catch (error) {
        // `localStorage` può lanciare anche solo a leggerlo: navigazione privata in alcuni
        // browser, storage disabilitato dall'utente, quota esaurita.
        return fail(storageFailure(error));
      }
    },

    async write(orders: readonly OrderModel[]): Promise<Result<void>> {
      try {
        localStorage.setItem(KEY, JSON.stringify(orders));

        return ok(undefined);
      } catch (error) {
        return fail(storageFailure(error));
      }
    },
  };
}
```

Punti chiave:

-   **Una sorgente per confine tecnico**: rete e storage sono due sorgenti anche se servono lo
    stesso dato. Quando usare l'una o l'altra lo decide il repository, non un `if` sepolto qui.
-   **Le sorgenti restituiscono modelli, mai entità**: il mapping è uno solo, nel repository,
    per tutte le fonti.
-   **Il `try/catch` vive qui e solo qui.** Quello che `fetch`, `json()` e `localStorage`
    lanciano diventa una `Failure`, e sopra questo file nessuno scrive più un `catch`.
-   **Una risposta che non è `ok` non è un'eccezione**: `fetch` non rifiuta per un 500, ed è
    la sorgente a trasformare lo stato HTTP in un esito.
-   **L'URL di base si riceve**, dal composition root che lo legge dalla
    [configurazione](./config.example.md): un test può costruire la sorgente verso un altro
    indirizzo senza toccare l'ambiente.
-   I segmenti di percorso che arrivano da fuori passano da `encodeURIComponent`, e i parametri
    di ricerca da `URLSearchParams`: un identificativo con una barra dentro non cambia
    l'indirizzo della richiesta.
