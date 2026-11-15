import { lazy } from 'react';
import type { RouteObject } from 'react-router';

const ChannelsPage = lazy(() =>
  import('./pages/channels-page').then((module) => ({ default: module.ChannelsPage })),
);

export const channelsRoutes: RouteObject[] = [{ path: 'canali', element: <ChannelsPage /> }];
