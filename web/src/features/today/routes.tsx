import { lazy } from 'react';
import type { RouteObject } from 'react-router';

const TodayPage = lazy(() =>
  import('./pages/today-page').then((module) => ({ default: module.TodayPage })),
);

// Oggi è la rotta indice: la feature dice «sono la pagina iniziale del mio layout», non
// sotto quale indirizzo vivrà.
export const todayRoutes: RouteObject[] = [{ index: true, element: <TodayPage /> }];
