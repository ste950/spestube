Il contratto base che ogni use case soddisfa. In questo stack uno use case è una **funzione**
(vedi [Use Case](../rules/domain_layer.rule.md#-use-case)), quindi il contratto è un alias di
tipo e non un'interfaccia da implementare.

```ts
// web/src/core/use-cases/use-case.ts

export type UseCase<In, Out> = (input: In) => Out;

// La variante asincrona esiste per gli use case che ricevono una porta: l'I/O è della porta,
// ma l'attesa attraversa lo use case e deve comparire nella sua firma.
export type AsyncUseCase<In, Out> = (input: In) => Promise<Out>;
```

Non c'è niente da implementare: il contratto si usa **annotando** il valore esportato, ed è lì
che TypeScript verifica la forma.

```ts
// web/src/domain/orders/use-cases/sort-orders.ts

import type { UseCase } from '@core/use-cases/use-case';
import type { Order } from '../entities/order';

export const sortOrders: UseCase<readonly Order[], readonly Order[]> = (orders) => [
  ...orders,
].sort(byPlacedAtDesc);
```

Punti chiave:

-   L'alias non aggiunge comportamento: serve a dare **un nome alla forma** e a far fallire la
    compilazione quando la firma di uno use case si allontana da quella degli altri.
-   Un use case con più ingressi prende **un oggetto**, non una lista di parametri: `In` resta
    un tipo solo, e aggiungere un ingresso non riordina le chiamate esistenti.
-   Uno use case che riceve delle porte non è questo tipo, ma la funzione che lo **restituisce**
    (`makeVerboSoggetto(porta): AsyncUseCase<In, Out>`) — vedi
    [Use Case](./use_case.example.md).
-   Niente classe con un solo metodo: in TypeScript una funzione è già un valore iniettabile e
    sostituibile, e la classe aggiungerebbe un `new` e nient'altro.
