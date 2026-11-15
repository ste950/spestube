import { lazy } from 'react';
import { useParams, type RouteObject } from 'react-router';

const CollectionsPage = lazy(() =>
  import('./pages/collections-page').then((module) => ({ default: module.CollectionsPage })),
);
const CollectionPage = lazy(() =>
  import('./pages/collection-page').then((module) => ({ default: module.CollectionPage })),
);

function CollectionRoute() {
  // L'id è un numero; un segmento che non lo è diventa NaN, che la pagina tratta come una
  // raccolta che non esiste.
  const { collectionId = '' } = useParams<'collectionId'>();
  const id = /^\d+$/.test(collectionId) ? Number(collectionId) : Number.NaN;
  return <CollectionPage key={collectionId} collectionId={id} />;
}

export const collectionsRoutes: RouteObject[] = [
  { path: 'raccolte', element: <CollectionsPage /> },
  { path: 'raccolte/:collectionId', element: <CollectionRoute /> },
];
