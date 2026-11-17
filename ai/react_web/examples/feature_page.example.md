Una pagina vive nella feature e non sa da quale URL è stata raggiunta: la monta il componente
di montaggio di [`routes.tsx`](./feature_routes.example.md), che le passa i parametri già
convertiti. Legge lo stato dallo [store della feature](./feature_store.example.md) con i
selettori, e prende lo stile dal suo foglio.

```tsx
// web/src/features/orders/pages/orders-page.tsx

import { useEffect } from 'react';
import { AppMessage } from '@shared/components/app-message';
import { useTranslate } from '@shared/i18n/use-translate';
import { OrderRow } from '../components/order-row';
import { useOrdersStore } from '../state/use-orders-store';
import styles from './orders-page.module.css';

export function OrdersPage() {
  const t = useTranslate();

  // Un selettore per valore: un componente che si sottoscrive allo stato intero si ridisegna
  // anche per i campi che non mostra.
  const orders = useOrdersStore((state) => state.orders);
  const isLoading = useOrdersStore((state) => state.isLoading);
  const errorKey = useOrdersStore((state) => state.errorKey);
  const load = useOrdersStore((state) => state.load);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <section className={styles.page} aria-busy={isLoading}>
      <title>{t('orders_title')}</title>
      <h1 className={styles.title}>{t('orders_title')}</h1>

      {errorKey !== null && <AppMessage tone="error" text={t(errorKey)} />}

      {errorKey === null && !isLoading && orders.length === 0 && (
        <AppMessage tone="neutral" text={t('orders_empty')} />
      )}

      <ul className={styles.list}>
        {orders.map((order) => (
          // Il collegamento al dettaglio sta nella riga, ed è un `<Link>`: un percorso non è
          // un import, e un link vero si apre in un'altra scheda.
          <OrderRow key={order.id} order={order} href={`/orders/${order.id}`} />
        ))}
      </ul>
    </section>
  );
}
```

```css
/* web/src/features/orders/pages/orders-page.module.css */

.page {
  display: grid;
  gap: var(--space-md);
  padding: var(--space-md);
}

.title {
  margin: 0;
  font-size: var(--font-size-title);
}

.list {
  display: grid;
  gap: var(--space-sm);
  margin: 0;
  padding: 0;
  list-style: none;
}

@media (width >= 64rem) {
  .page {
    padding: var(--space-xl);
  }
}
```

Il componente di riga, specifico della feature perché conosce l'entità:

```tsx
// web/src/features/orders/components/order-row.tsx

import { Link } from 'react-router';
import { orderTotalCents, type Order } from '@domain/orders/entities/order';
import { useTranslate } from '@shared/i18n/use-translate';
import { formatCents } from '../state/format-cents';
import styles from './order-row.module.css';

export function OrderRow({ order, href }: { order: Order; href: string }) {
  const t = useTranslate();

  return (
    <li className={styles.orderRow}>
      <Link to={href} className={styles.link}>
        {order.reference}
      </Link>
      <span className={styles.status}>{t(`orders_status_${order.status}`)}</span>
      <span className={styles.total}>{formatCents(orderTotalCents(order))}</span>
    </li>
  );
}
```

Punti chiave:

-   **La pagina non legge l'URL.** Niente `useParams` qui: i parametri arrivano dal componente
    di montaggio, già convertiti, come proprietà. La stessa pagina si può montare da un altro
    percorso senza cambiarla.
-   **Lo stato arriva da selettori** sullo store della feature; nessun `useState` che duplichi
    ciò che lo store già tiene. Il caricamento parte all'arrivo sulla pagina, dall'azione dello
    store, non da un `loader` del router.
-   **Ogni stato ha la sua resa**: caricamento (`aria-busy`), errore, lista vuota, lista. Uno
    stato dimenticato è una pagina bianca che nessuno sa spiegare.
-   **Titolo e intestazione**: `<title>` arriva nell'intestazione del documento, `<h1>` è uno
    solo. È ciò che fa annunciare il cambio di pagina a chi usa un lettore di schermo
    ([Convenzioni UI](../rules/ui_conventions.rule.md#-accessibilità)).
-   **Navigare è un `<Link>`**, non un `onClick`: è un `<a href>` vero, che si apre in un'altra
    scheda e si annuncia come collegamento.
-   **Lo stile sta nel foglio accanto** e usa solo variabili del tema; la media query usa uno
    dei [punti di rottura](./theme_tokens.example.md) dichiarati. Il foglio conta nelle 100
    righe come il componente.
-   **Ogni etichetta passa da `t(...)`**, anche quelle composte con un valore di dominio
    (`orders_status_${order.status}`): il tipo delle chiavi le verifica tutte, perché lo stato
    è un'unione chiusa.
