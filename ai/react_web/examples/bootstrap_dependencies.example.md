Il composition root: l'unico modulo che costruisce le implementazioni di `data/` e le espone
**annotate con i contratti di `domain/`**. Prende il posto del service locator degli altri
stack, e il perché è in
[Iniezione delle dipendenze](../rules/data_layer.rule.md#-iniezione-delle-dipendenze).

```ts
// web/src/bootstrap/dependencies.ts

import { createHttpOrderSource } from '@data/orders/sources/http-order-source';
import { createLocalOrderCache } from '@data/orders/sources/local-order-cache';
import { createCachedOrderRepository } from '@data/orders/repositories/cached-order-repository';
import type { OrderRepository } from '@domain/orders/repositories/order-repository';
import { makeLoadEditableOrders } from '@domain/orders/use-cases/load-editable-orders';

import { apiBaseUrl } from './config';

// L'annotazione è tutto il lavoro di un contenitore: è qui che TypeScript verifica la
// conformità strutturale, ed è qui che il tipo esportato diventa il **contratto** invece
// dell'implementazione — chi consuma non può chiamare un metodo che la porta non dichiara.
export const orderRepository: OrderRepository = createCachedOrderRepository(
  createHttpOrderSource(apiBaseUrl),
  createLocalOrderCache(),
);

export const loadEditableOrders = makeLoadEditableOrders(orderRepository);
```

Chi lo consuma, e come. È l'unica riga di `bootstrap/` che una feature può toccare, e in una
feature si tocca **in un punto solo**: il modulo che lega lo store della feature alle sue
porte. Tutto il resto riceve.

```ts
// web/src/features/orders/state/use-orders-store.ts

import { orderRepository } from '@bootstrap/dependencies';
import { createOrdersStore } from './orders-store';

// La fabbrica dello store prende il repository come argomento ed è verificabile da sola; qui
// si dice solo **quale** repository, ed è l'unico file della feature che lo sa.
export const useOrdersStore = createOrdersStore(orderRepository);
```

In un test non si sostituisce niente a livello globale: si costruisce la stessa catena con gli
oggetti finti che servono, e questo modulo non entra mai in gioco.

```ts
const repository: OrderRepository = createCachedOrderRepository(fakeSource, fakeCache);
```

Punti chiave:

-   Una dipendenza mancante è un **errore di compilazione**, non una sorpresa all'avvio: non
    c'è nessuna risoluzione per tipo da ricordarsi di registrare.
-   Un modulo è già un singleton — viene valutato una volta sola — quindi `orderRepository` è
    l'unica istanza senza che nessuno debba dichiararlo.
-   **Niente di ciò che si testa importa questo file**: repository, use case e fabbriche di
    store ricevono le loro dipendenze come argomenti. L'unico consumatore dentro una feature è
    il modulo di due righe che lega lo store, e lì non c'è niente da provare.
-   Quello che varia fra ambienti (l'URL di base, le chiavi) sta in `bootstrap/config.ts` e
    arriva da qui: le sorgenti lo ricevono, non lo leggono.
-   Un `Context` di React non è il posto di queste dipendenze: obbligherebbe a rendere qualcosa
    per ottenere un repository. Resta il canale di ciò che è davvero di presentazione — vedi
    [Etichette tradotte nei componenti](./i18n_hook.example.md).
