import { lazy } from 'react';
import type { RouteObject } from 'react-router';

const QueuePage = lazy(() =>
  import('./pages/queue-page').then((module) => ({ default: module.QueuePage })),
);

export const queueRoutes: RouteObject[] = [{ path: 'da-vedere', element: <QueuePage /> }];
