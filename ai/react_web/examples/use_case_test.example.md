Il test di [uno use case](./use_case.example.md): nessun render, nessun mock di moduli, nessuna
rete. Si chiama una funzione e si guarda cosa restituisce.

```ts
// web/tests/domain/orders/sort-orders.test.ts

import { describe, expect, it } from 'vitest';
import type { Order } from '@domain/orders/entities/order';
import { sortOrders } from '@domain/orders/use-cases/sort-orders';

function order(id: string, placedAt: string, unitPriceCents: number): Order {
  return {
    id,
    reference: id.toUpperCase(),
    status: 'submitted',
    placedAt: new Date(placedAt),
    lines: [{ sku: 'sku-1', quantity: 1, unitPriceCents }],
  };
}

const vecchio = order('vecchio', '2024-01-01T10:00:00Z', 5_000);
const recente = order('recente', '2024-03-01T10:00:00Z', 1_000);

describe('sortOrders', () => {
  it('con criterio "recent" mette per primo l’ordine più recente', () => {
    const sorted = sortOrders({ orders: [vecchio, recente], criterion: 'recent' });

    expect(sorted.map((o) => o.id)).toEqual(['recente', 'vecchio']);
  });

  it('con criterio "largest" mette per primo l’ordine di importo maggiore', () => {
    const sorted = sortOrders({ orders: [recente, vecchio], criterion: 'largest' });

    expect(sorted.map((o) => o.id)).toEqual(['vecchio', 'recente']);
  });

  it('restituisce una lista vuota quando non ci sono ordini', () => {
    expect(sortOrders({ orders: [], criterion: 'recent' })).toEqual([]);
  });

  it('non modifica la lista che riceve', () => {
    // Il caso limite esiste perché `sort` muta l’array: qui arriva da un’entità, e una
    // pagina che mostrasse la lista originale vedrebbe l’ordine cambiarle sotto.
    const originale = [vecchio, recente];

    sortOrders({ orders: originale, criterion: 'largest' });

    expect(originale.map((o) => o.id)).toEqual(['vecchio', 'recente']);
  });
});
```

Le entità di prova nascono dentro il file che le usa e si spostano in un
`order-fixtures.ts` accanto quando serve un secondo test: è quello che succede qui sotto.

Uno use case che riceve una porta si prova allo stesso modo: la porta è un **oggetto letterale
annotato con il contratto**, e il compilatore verifica che abbia tutti i metodi con le firme
giuste.

```ts
// web/tests/domain/orders/load-editable-orders.test.ts

import { expect, it } from 'vitest';
import { notFoundFailure } from '@core/result/failure';
import { fail, ok } from '@core/result/result';
import type { Order } from '@domain/orders/entities/order';
import type { OrderRepository } from '@domain/orders/repositories/order-repository';
import { makeLoadEditableOrders } from '@domain/orders/use-cases/load-editable-orders';
import { recente, vecchio } from './order-fixtures';

function repositoryReturning(orders: readonly Order[]): OrderRepository {
  return {
    listOrders: async () => ok(orders),
    getOrder: async () => fail(notFoundFailure()),
    changeStatus: async () => fail(notFoundFailure()),
  };
}

it('scarta gli ordini che il dominio non considera modificabili', async () => {
  const spedito: Order = { ...vecchio, id: 'spedito', status: 'shipped' };
  const loadEditableOrders = makeLoadEditableOrders(repositoryReturning([recente, spedito]));

  const result = await loadEditableOrders();

  expect(result.ok && result.value.map((o) => o.id)).toEqual(['recente']);
});

it('riporta il fallimento della porta senza tradurlo', async () => {
  const loadEditableOrders = makeLoadEditableOrders({
    listOrders: async () => fail(notFoundFailure()),
    getOrder: async () => fail(notFoundFailure()),
    changeStatus: async () => fail(notFoundFailure()),
  });

  const result = await loadEditableOrders();

  expect(result.ok).toBe(false);
});
```

Punti chiave:

-   Un test per comportamento, e il **nome è il commento**: se serve una riga per spiegare cosa
    sta provando, si sistema il nome ([Commenti nel
    Codice](../rules/code_comments.rule.md#-nei-test-il-nome-è-il-commento)).
-   I casi limite non sono facoltativi: lista vuota, valori a zero, e ogni caso nato da un
    difetto vero — quello merita il commento, perché il nome non può dire da dove viene.
-   La porta finta è un oggetto letterale **annotato**: quando il contratto cambia, è il test a
    smettere di compilare. Nessuna intercettazione di moduli, nessuna libreria di mock.
-   Il test non importa `@bootstrap/dependencies`: costruisce la catena che gli serve.
-   `describe`, `it` ed `expect` si importano da `vitest` invece di arrivare come globali: il
    file dice da solo con cosa gira, e `tsc` non ha bisogno di tipi globali in più.
-   Il file sta in `web/tests/domain/<soggetto>/` e rispecchia la posizione del sorgente.
