Una regola di business che, senza un posto suo, finirebbe dentro uno store o un componente.
Le due forme sono nella regola [Use Case](../rules/domain_layer.rule.md#-use-case): una
funzione pura quando non serve niente da fuori, una funzione che riceve le porte e restituisce
la funzione finale quando serve.

## Forma pura

```ts
// web/src/domain/orders/use-cases/sort-orders.ts

import type { UseCase } from '@core/use-cases/use-case';
import { assertNever } from '@core/exhaustive';
import { orderTotalCents, type Order } from '../entities/order';

export type OrderSortCriterion = 'recent' | 'largest';

export type SortOrdersInput = {
  readonly orders: readonly Order[];
  readonly criterion: OrderSortCriterion;
};

export const sortOrders: UseCase<SortOrdersInput, readonly Order[]> = ({ orders, criterion }) =>
  // Copia prima di ordinare: `sort` muta l'array che riceve, e questo arriva da un'entità.
  [...orders].sort(comparatorFor(criterion));

function comparatorFor(criterion: OrderSortCriterion): (a: Order, b: Order) => number {
  switch (criterion) {
    case 'recent':
      return (a, b) => b.placedAt.getTime() - a.placedAt.getTime();
    case 'largest':
      return (a, b) => orderTotalCents(b) - orderTotalCents(a);
    default:
      return assertNever(criterion);
  }
}
```

## Forma con porte

Quando lo use case ha bisogno di dati che stanno dietro una porta, la dipendenza è **esplicita
e passata da fuori**: una funzione `make*` la prende e restituisce lo use case vero. È quello
che permette a un test di passare un oggetto letterale al posto del repository.

```ts
// web/src/domain/orders/use-cases/load-editable-orders.ts

import { mapOk } from '@core/result/result';
import type { AsyncUseCase } from '@core/use-cases/use-case';
import type { Result } from '@core/result/result';
import { isOrderEditable, type Order } from '../entities/order';
import type { OrderRepository } from '../repositories/order-repository';

export function makeLoadEditableOrders(
  repository: OrderRepository,
): AsyncUseCase<void, Result<readonly Order[]>> {
  return async () => {
    const result = await repository.listOrders({ statuses: ['draft', 'submitted'] });

    // Il filtro c'è anche se il filtro di dominio è già passato alla porta: quali stati siano
    // modificabili lo decide `isOrderEditable`, e una fonte che ne restituisse uno in più non
    // deve poter far comparire un ordine modificabile per sbaglio.
    return mapOk(result, (orders) => orders.filter(isOrderEditable));
  };
}
```

Punti chiave:

-   Una responsabilità per use case: un caso d'uso, un output. `sortOrders` ordina e basta —
    non filtra, non impagina, non formatta.
-   Ritorna **entità di dominio**. Una percentuale per una barra di avanzamento, un'etichetta
    già tradotta o una data già formattata sono presentation e restano nella feature.
-   Nessuna `interface` per lo use case: sarebbe la copia dell'unica implementazione possibile.
    Se ti accorgi di volerla scrivere, quasi sempre quello che hai in mano è una porta.
-   La forma con porte non importa `@bootstrap/dependencies`: riceve, non si serve da sé. Chi
    la costruisce è il composition root, o il test.
-   File `<verbo>-<soggetto>.ts`, simbolo `<verbo><Soggetto>`; la variante con porte si chiama
    `make<Verbo><Soggetto>`.
-   Il test copre il caso nominale e i casi limite — [Test di uno use
    case](./use_case_test.example.md).
