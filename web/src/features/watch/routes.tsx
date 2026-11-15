import { lazy } from 'react';
import { useParams, type RouteObject } from 'react-router';

const WatchPage = lazy(() =>
  import('./pages/watch-page').then((module) => ({ default: module.WatchPage })),
);

function WatchRoute() {
  // Il segmento è obbligatorio nel percorso, quindi c'è sempre; il tipo del router non lo sa.
  const { videoId = '' } = useParams<'videoId'>();
  // La chiave fa ripartire la pagina da zero quando cambia il video: nessuno stato del
  // player di prima sopravvive nel successivo.
  return <WatchPage key={videoId} videoId={videoId} />;
}

export const watchRoutes: RouteObject[] = [{ path: 'guarda/:videoId', element: <WatchRoute /> }];
