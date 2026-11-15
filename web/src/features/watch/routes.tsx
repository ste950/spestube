import { lazy } from 'react';
import { useParams, useSearchParams, type RouteObject } from 'react-router';

const WatchPage = lazy(() =>
  import('./pages/watch-page').then((module) => ({ default: module.WatchPage })),
);

function WatchRoute() {
  // Il segmento è obbligatorio nel percorso, quindi c'è sempre; il tipo del router non lo sa.
  const { videoId = '' } = useParams<'videoId'>();
  const [search] = useSearchParams();
  // `?raccolta=12` dice da quale raccolta si è arrivati; un valore che non è un id si ignora.
  const raw = search.get('raccolta') ?? '';
  const fromCollection = /^\d+$/.test(raw) ? Number(raw) : null;
  // La chiave fa ripartire la pagina da zero quando cambia il video: nessuno stato del
  // player di prima sopravvive nel successivo.
  return <WatchPage key={videoId} videoId={videoId} fromCollection={fromCollection} />;
}

export const watchRoutes: RouteObject[] = [{ path: 'guarda/:videoId', element: <WatchRoute /> }];
