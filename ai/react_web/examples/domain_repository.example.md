La porta del dominio: dichiara **cosa serve al dominio**, nella lingua del dominio, e non dice
niente di come quei dati arrivino. Dietro può esserci HTTP, un file locale o un oggetto finto
in un test ([Entità, porte e use case](../rules/domain_layer.rule.md#-entità-porte-e-use-case)).

```ts
// web/src/domain/orders/repositories/order-repository.ts

import type { Result } from '@core/result/result';
import type { Order, OrderStatus } from '../entities/order';

export interface OrderRepository {
  listOrders(filter: OrderFilter): Promise<Result<readonly Order[]>>;

  getOrder(orderId: string): Promise<Result<Order>>;

  /** L'ordine aggiornato, non `void`: chi ha cambiato lo stato deve poterlo mostrare. */
  changeStatus(orderId: string, status: OrderStatus): Promise<Result<Order>>;
}

export type OrderFilter = {
  readonly statuses: readonly OrderStatus[];
};
```

Punti chiave:

-   Ogni metodo restituisce `Promise<Result<...>>`: il fallimento previsto è nella firma, e chi
    chiama non può ignorarlo senza che il compilatore glielo faccia notare.
-   I tipi che attraversano la porta sono **entità di dominio**, mai modelli di `data/` né
    forme pensate per la UI. Il filtro è un tipo di dominio anche lui: nessuna traccia di query
    string, paginazione della fonte o header.
-   L'`interface` è **strutturale**: l'implementazione di `data/` non dichiara di soddisfarla,
    lo verifica il composition root nel punto in cui annota il valore esportato (vedi
    [Composition root](./bootstrap_dependencies.example.md)).
-   La porta non sa che esistono delle sorgenti: il contratto della fonte vive in `data/`, ed è
    irraggiungibile da qui ([Perché il contratto sta in
    `data/`](../rules/data_layer.rule.md#perché-il-contratto-sta-in-data-e-non-in-domain)).
-   Un soggetto ha almeno una porta propria. Se un gruppo di entità non ne ha una, quasi sempre
    appartiene a un soggetto che esiste già.
