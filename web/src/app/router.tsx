import { createBrowserRouter } from 'react-router';
import { basePath } from '@bootstrap/config';
import { channelsRoutes } from '@features/channels/routes';
import { queueRoutes } from '@features/queue/routes';
import { todayRoutes } from '@features/today/routes';
import { watchRoutes } from '@features/watch/routes';
import { AppLayout } from './layouts/app-layout';
import { NotFoundPage } from './not-found-page';
import { RouteErrorPage } from './route-error-page';

export const router = createBrowserRouter(
  [
    {
      element: <AppLayout />,
      // Un errore di render in una pagina si ferma qui, invece di lasciare una pagina bianca.
      errorElement: <RouteErrorPage />,
      children: [
        // Oggi è la pagina iniziale: l'app si apre su ciò che è arrivato, non su una vetrina.
        ...todayRoutes,
        ...watchRoutes,
        ...queueRoutes,
        ...channelsRoutes,
        { path: '*', element: <NotFoundPage /> },
      ],
    },
  ],
  { basename: basePath },
);
