L'implementazione della porta di [Interfaccia Repository](./domain_repository.example.md),
servita da due sorgenti dello stesso soggetto: la fonte remota e la copia locale. Le sorgenti
restituiscono già `Result`, quindi qui non si traduce niente — si **decide**
([Sorgenti](../rules/data_layer.rule.md#-sorgenti)).

I contratti delle sorgenti e il modello che attraversa il confine. Stanno in `data/`, e nessuno
fuori dal soggetto può raggiungerli:

```ts
// web/src/data/orders/models/order-model.ts

// La forma che la fonte restituisce davvero: `status` è una stringa qualunque e la data è il
// testo che è arrivato. Interpretarli è mestiere del mapper.
export type OrderModel = {
  readonly id: string;
  readonly reference: string;
  readonly status: string;
  readonly placed_at: string;
  readonly lines: readonly OrderLineModel[];
};

export type OrderLineModel = {
  readonly sku: string;
  readonly quantity: number;
  readonly unit_price_cents: number;
};
```

```ts
// web/src/data/orders/sources/order-source.ts

import type { Result } from '@core/result/result';
import type { OrderModel } from '../models/order-model';

export interface OrderSource {
  fetchOrders(statuses: readonly string[]): Promise<Result<readonly OrderModel[]>>;
  fetchOrder(orderId: string): Promise<Result<OrderModel>>;
  updateStatus(orderId: string, status: string): Promise<Result<OrderModel>>;
}
```

```ts
// web/src/data/orders/sources/order-cache.ts

import type { Result } from '@core/result/result';
import type { OrderModel } from '../models/order-model';

export interface OrderCache {
  read(): Promise<Result<readonly OrderModel[]>>;
  write(orders: readonly OrderModel[]): Promise<Result<void>>;
}
```

Il mapper **interpreta**: default, insiemi chiusi, regole di dominio. È l'unico punto in cui un
modello diventa un'entità, e ci passano tutte le fonti.

```ts
// web/src/data/orders/mappers/order-mapper.ts

import type { Order, OrderStatus } from '@domain/orders/entities/order';
import type { OrderModel } from '../models/order-model';

const KNOWN_STATUSES: readonly OrderStatus[] = ['draft', 'submitted', 'shipped', 'cancelled'];

export function toOrder(model: OrderModel): Order {
  return {
    id: model.id,
    reference: model.reference,
    status: toStatus(model.status),
    placedAt: new Date(model.placed_at),
    lines: model.lines.map((line) => ({
      sku: line.sku,
      quantity: line.quantity,
      unitPriceCents: line.unit_price_cents,
    })),
  };
}

// Uno stato che il dominio non conosce diventa una bozza e non fa cadere la lettura: il
// backend può introdurne di nuovi prima che l'app sia aggiornata, e un ordine in meno sarebbe
// peggio di un ordine mostrato come non ancora inviato.
function toStatus(raw: string): OrderStatus {
  return KNOWN_STATUSES.find((status) => status === raw) ?? 'draft';
}
```

Il repository:

```ts
// web/src/data/orders/repositories/cached-order-repository.ts

import { mapOk, ok, type Result } from '@core/result/result';
import type { Order, OrderStatus } from '@domain/orders/entities/order';
import type { OrderFilter, OrderRepository } from '@domain/orders/repositories/order-repository';
import { toOrder } from '../mappers/order-mapper';
import type { OrderModel } from '../models/order-model';
import type { OrderCache } from '../sources/order-cache';
import type { OrderSource } from '../sources/order-source';

// Il tipo di ritorno annotato è il punto in cui TypeScript verifica la conformità: la
// conformità strutturale non si dichiara, si fa verificare dove il valore prende un tipo.
export function createCachedOrderRepository(
  source: OrderSource,
  cache: OrderCache,
): OrderRepository {
  async function listOrders(filter: OrderFilter): Promise<Result<readonly Order[]>> {
    const remote = await source.fetchOrders(filter.statuses);
    if (remote.ok) {
      await writeQuietly(cache, remote.value);

      return ok(remote.value.map(toOrder));
    }

    // La fonte è caduta: l'ultima lettura riuscita vale più di una pagina vuota. Se non
    // c'è nemmeno quella, si riporta il fallimento della **fonte** e non quello della cache —
    // è l'unico dei due che dica qualcosa a chi legge.
    const cached = await cache.read();
    if (!cached.ok || cached.value.length === 0) return remote;

    return ok(cached.value.map(toOrder));
  }

  async function getOrder(orderId: string): Promise<Result<Order>> {
    // Nessun ripiego sulla copia locale: chi apre un ordine preciso vuole il suo stato di
    // adesso, e una versione vecchia spacciata per attuale farebbe più danno di un errore.
    return mapOk(await source.fetchOrder(orderId), toOrder);
  }

  async function changeStatus(orderId: string, status: OrderStatus): Promise<Result<Order>> {
    // La copia locale resta vecchia fino alla prossima lettura: riscriverla da qui vorrebbe
    // dire fondere una lista con un singolo elemento, e sbagliare quella fusione è più
    // probabile che mostrare una lista di un istante fa.
    return mapOk(await source.updateStatus(orderId, status), toOrder);
  }

  return { listOrders, getOrder, changeStatus };
}

// Una copia locale che non si scrive è un peccato, non un guasto: la lettura è riuscita, e chi
// ha chiesto gli ordini li avrà comunque. Il nome del metodo dice che scartare quel `Result` è
// una scelta.
async function writeQuietly(cache: OrderCache, orders: readonly OrderModel[]): Promise<void> {
  await cache.write(orders);
}
```

Punti chiave:

-   Una **factory**, non una `class`: quello che si costruisce è una chiusura sulle sue
    dipendenze, e non un oggetto con un ciclo di vita.
-   Il nome dice la **strategia di composizione** (`cached`), non il trasporto: quale sia il
    trasporto lo sanno le sorgenti, e sostituirlo non tocca questo file.
-   Non c'è nessun `try/catch` e nessuna eccezione da tradurre: il confine è nella firma delle
    sorgenti, che restituiscono `Result`
    ([Errori](../rules/data_layer.rule.md#errori)).
-   Un fallimento che non richiede decisioni non costa codice: `mapOk` porta l'entità e lascia
    passare la `Failure` così com'è.
-   Il mapping è una chiamata sola, al mapper: due fonti, una conversione. Una sorgente che
    restituisse entità toglierebbe al repository il suo mestiere.
-   Niente logica di business qui: ordinamenti, selezioni e punteggi sono
    [use case](./use_case.example.md).
-   Il test costruisce il repository con due oggetti letterali annotati `OrderSource` e
    `OrderCache`, e copre i casi che la composizione ha creato: fonte a posto con cache
    aggiornata, fonte caduta con cache piena, fonte caduta con cache vuota.
