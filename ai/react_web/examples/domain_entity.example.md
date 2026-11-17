Un'entità è un **valore**: un `type` con campi `readonly`, la logica derivata accanto come
funzione pura, e gli insiemi chiusi come unioni di letterali. Il perché di ognuna delle tre
scelte è in [Entità](../rules/domain_layer.rule.md#-entità).

```ts
// web/src/domain/orders/entities/order.ts

import { assertNever } from '@core/exhaustive';

export type OrderStatus = 'draft' | 'submitted' | 'shipped' | 'cancelled';

export type OrderLine = {
  readonly sku: string;
  readonly quantity: number;
  /** In centesimi: i prezzi non si tengono in virgola mobile. */
  readonly unitPriceCents: number;
};

export type Order = {
  readonly id: string;
  readonly reference: string;
  readonly status: OrderStatus;
  readonly placedAt: Date;
  readonly lines: readonly OrderLine[];
};

export function orderTotalCents(order: Order): number {
  return order.lines.reduce((total, line) => total + line.quantity * line.unitPriceCents, 0);
}

export function isOrderEditable(order: Order): boolean {
  switch (order.status) {
    case 'draft':
    case 'submitted':
      return true;
    case 'shipped':
    case 'cancelled':
      return false;
    default:
      // Uno stato nuovo non è modificabile o immodificabile per difetto: è una decisione, e
      // deve essere presa qui invece di essere ereditata da un ramo di ripiego.
      return assertNever(order.status);
  }
}
```

Punti chiave:

-   Nessun `class`, nessun costruttore, nessun metodo: un mapper costruisce un'entità con un
    oggetto letterale, e non deve istanziare niente per ogni elemento che arriva da una fonte.
-   `orderTotalCents` è una funzione esportata **accanto** al tipo, non un metodo. Chi la usa
    la importa; chi non la usa non la porta con sé.
-   `OrderStatus` è un'unione di letterali e non un `enum`: esiste solo a compile time, si
    confronta con `===`, ed è già serializzabile — il valore che arriva dalla fonte è già del
    tipo giusto una volta verificato.
-   Il `default` che chiama `assertNever` è il punto in cui aggiungere uno stato rompe la
    compilazione. Senza, il ramo mancante passerebbe inosservato fino a produzione.
-   I campi sono `readonly`, e `lines` è un `readonly OrderLine[]`: un'entità che arriva a una
    pagina non può essere modificata da chi la mostra.
-   Niente commento che ripeta il nome di un campo; il commento su `unitPriceCents` c'è perché
    l'unità di misura la firma non la dice ([Commenti nel
    Codice](../rules/code_comments.rule.md)).
