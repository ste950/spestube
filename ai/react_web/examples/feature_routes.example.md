Le rotte di una feature e il router che le compone. La divisione — la feature possiede i propri
percorsi, `app/` li monta sotto i layout — è in
[Le rotte stanno nella feature](../rules/project_structure.rule.md#-le-rotte-stanno-nella-feature);
cosa può stare in `routes.tsx` è in
[La feature e le sue rotte](../rules/feature_overview.rule.md#-la-feature-e-le-sue-rotte).

## Le rotte della feature

```tsx
// web/src/features/orders/routes.tsx

import { lazy } from 'react';
import { useParams, type RouteObject } from 'react-router';

// Le pagine si caricano pigre: il codice della feature si scarica quando qualcuno ci arriva, e
// il confine della feature diventa il confine del pacchetto scaricato.
const OrdersPage = lazy(() =>
  import('./pages/orders-page').then((module) => ({ default: module.OrdersPage })),
);
const OrderDetailPage = lazy(() =>
  import('./pages/order-detail-page').then((module) => ({ default: module.OrderDetailPage })),
);

function OrderDetailRoute() {
  // I parametri di rotta sono stringhe, e la conversione avviene qui: la pagina riceve un
  // identificativo, e il dominio non sa che esiste un URL. Il segmento è obbligatorio nel
  // percorso, quindi c'è sempre; il tipo del router non lo sa.
  const { orderId = '' } = useParams<'orderId'>();

  return <OrderDetailPage orderId={orderId} />;
}

// Percorsi relativi: la feature dice `orders`, non sotto quale layout o prefisso vivrà.
export const ordersRoutes: RouteObject[] = [
  { path: 'orders', element: <OrdersPage /> },
  { path: 'orders/:orderId', element: <OrderDetailRoute /> },
];
```

## Il router dell'app

```tsx
// web/src/app/router.tsx

import { createBrowserRouter, Navigate } from 'react-router';
import { basePath } from '@bootstrap/config';
import { ordersRoutes } from '@features/orders/routes';
import { AppLayout } from './layouts/app-layout';
import { NotFoundPage } from './not-found-page';
import { RouteErrorPage } from './route-error-page';

export const router = createBrowserRouter(
  [
    {
      element: <AppLayout />,
      // Un errore di render in una pagina si ferma qui, dentro il layout, invece di lasciare
      // il visitatore davanti a una pagina bianca.
      errorElement: <RouteErrorPage />,
      children: [
        // La pagina iniziale è una rotta di feature come le altre: qui si dice solo quale.
        { index: true, element: <Navigate to="orders" replace /> },
        ...ordersRoutes,
        // L'ultimo: i percorsi che nessuna feature dichiara. È anche la pagina a cui arriva il
        // fallback del server per un URL sbagliato.
        { path: '*', element: <NotFoundPage /> },
      ],
    },
  ],
  { basename: basePath },
);
```

```tsx
// web/src/app/layouts/app-layout.tsx

import { Suspense } from 'react';
import { Outlet } from 'react-router';
import { AppSpinner } from '@shared/components/app-spinner';
import { AppSkipLink } from '@shared/components/app-skip-link';
import { useTranslate } from '@shared/i18n/use-translate';
import styles from './app-layout.module.css';

export function AppLayout() {
  const t = useTranslate();

  return (
    <div className={styles.layout}>
      <AppSkipLink targetId="contenuto" label={t('common_skip_to_content')} />
      <header className={styles.header}>{/* navigazione del sito */}</header>
      <main id="contenuto" className={styles.main}>
        {/* Una pagina pigra non ancora scaricata sospende il render: il segnaposto sta dentro
            il layout, così intestazione e navigazione restano al loro posto. */}
        <Suspense fallback={<AppSpinner label={t('common_loading')} />}>
          <Outlet />
        </Suspense>
      </main>
    </div>
  );
}
```

Punti chiave:

-   **`app/` importa `routes.tsx` e nient'altro della feature.** Né pagine, né store, né
    componenti: la tabella dei confini lo vieta ([Confini fra livelli](./architecture_test.example.md)).
-   **Il componente di montaggio è di poche righe**: legge e converte i parametri, rende la
    pagina. Niente repository, niente store, niente layout. Se cresce, sta facendo il lavoro
    di una pagina.
-   **Niente `loader` né `action`.** I dati li carica lo store della feature quando la pagina
    si monta; il router resta una tabella di indirizzi
    ([Niente `loader` né `action`](../rules/feature_overview.rule.md#niente-loader-né-action)).
-   **`basename` arriva da `bootstrap/config.ts`**, che lo legge dall'opzione `base` di Vite:
    un sito spostato sotto un percorso non riscrive nessun link, e `app/` non legge l'ambiente
    ([Distribuzione](../rules/deployment.rule.md#-il-percorso-di-base)).
-   **Un `errorElement` sul layout** cattura gli errori di render di tutte le pagine figlie.
    Gli errori previsti — la rete, una risorsa che non esiste — non arrivano fin qui: sono
    `Failure` dentro un `Result`, e la pagina li mostra come uno stato.
-   La pagina dei percorsi sconosciuti e quella d'errore vivono in `app/`: non appartengono a
    nessuna feature.
-   `routes.tsx` esporta un elenco e definisce componenti: il plugin di ESLint per il
    ricaricamento a caldo lo segnala, e la [configurazione del
    linter](./architecture_test.example.md) lo spegne su questi file con la sua ragione.
