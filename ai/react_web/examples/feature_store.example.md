Lo store è il contenitore di stato di una feature: è il ruolo che negli altri stack ha il Bloc.
Vive in `features/<f>/state/`, riceve le porte di dominio quando viene creato e non sa niente
dei componenti che lo leggono ([Stato della
feature](../rules/feature_overview.rule.md#-stato-della-feature)).

```ts
// web/src/features/orders/state/orders-store.ts

import { create, type StoreApi, type UseBoundStore } from 'zustand';
import type { Order } from '@domain/orders/entities/order';
import type { OrderRepository } from '@domain/orders/repositories/order-repository';
import { sortOrders, type OrderSortCriterion } from '@domain/orders/use-cases/sort-orders';
import type { TranslationKey } from '@shared/i18n/keys';
import { describeFailure } from './describe-failure';

export type OrdersState = {
  readonly orders: readonly Order[];
  readonly criterion: OrderSortCriterion;
  readonly isLoading: boolean;
  readonly errorKey: TranslationKey | null;
  readonly load: () => Promise<void>;
  readonly setCriterion: (criterion: OrderSortCriterion) => void;
};

export type OrdersStore = UseBoundStore<StoreApi<OrdersState>>;

export function createOrdersStore(repository: OrderRepository): OrdersStore {
  return create<OrdersState>((set, get) => ({
    orders: [],
    criterion: 'recent',
    isLoading: false,
    errorKey: null,

    load: async () => {
      set({ isLoading: true, errorKey: null });

      const result = await repository.listOrders({
        statuses: ['draft', 'submitted', 'shipped'],
      });

      if (!result.ok) {
        // Nello stato ci va il **risultato**, non il `Result`: qui il vocabolario delle
        // `Failure` finisce, e sopra si parla solo di cosa mostrare.
        set({ isLoading: false, errorKey: describeFailure(result.failure) });

        return;
      }

      set({
        isLoading: false,
        orders: sortOrders({ orders: result.value, criterion: get().criterion }),
      });
    },

    setCriterion: (criterion) => {
      // L'ordinamento è uno use case di dominio, non un `sort` scritto qui: lo store decide
      // *quando* riordinare, non *come*.
      set({ criterion, orders: sortOrders({ orders: get().orders, criterion }) });
    },
  }));
}
```

Il modulo che lega lo store alle sue dipendenze reali è separato, e in una feature è l'unico che
importa `@bootstrap/dependencies`:

```ts
// web/src/features/orders/state/use-orders-store.ts

import { orderRepository } from '@bootstrap/dependencies';
import { createOrdersStore } from './orders-store';

export const useOrdersStore = createOrdersStore(orderRepository);
```

Un test non monta niente: costruisce lo store con un repository finto, chiama un'azione e legge
lo stato.

```ts
const repository: OrderRepository = {
  listOrders: async () => ok([draftOrder, shippedOrder]),
  getOrder: async () => fail(notFoundFailure()),
  changeStatus: async () => fail(notFoundFailure()),
};

const store = createOrdersStore(repository);
await store.getState().load();

expect(store.getState().orders.map((order) => order.id)).toEqual(['recente', 'vecchio']);
```

Punti chiave:

-   **Uno store per feature**, non uno per pagina: due pagine della stessa feature che
    mostrano gli stessi ordini li leggono dallo stesso posto.
-   La **fabbrica** riceve il repository, il modulo di legame sceglie quale. È la separazione
    che permette al test di non toccare mai il composition root.
-   Lo stato è di presentazione: entità già ordinate, un flag di caricamento, una chiave di
    traduzione. Nessun `Result` e nessuna `Failure` lo attraversano.
-   **Niente logica di dominio qui.** Un ordinamento, un calcolo o una selezione sono use case;
    se lo store cresce, è quasi sempre perché ne contiene uno.
-   I componenti leggono con **selettori**
    ([Pagina](./feature_page.example.md)), non prendendo lo stato intero.
-   `state/` può contenere anche view model di presentazione — `describeFailure` è uno di
    quelli: traduce un fatto di dominio in qualcosa che la UI sappia mostrare.
