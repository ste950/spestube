import { lazy } from 'react';
import type { RouteObject } from 'react-router';

const SettingsPage = lazy(() =>
  import('./pages/settings-page').then((module) => ({ default: module.SettingsPage })),
);
const ChannelsPage = lazy(() =>
  import('./pages/channels-page').then((module) => ({ default: module.ChannelsPage })),
);

// Le impostazioni sono una pagina indice con le sue sottopagine: oggi i canali, domani le
// altre voci, ognuna al suo indirizzo.
export const settingsRoutes: RouteObject[] = [
  { path: 'impostazioni', element: <SettingsPage /> },
  { path: 'impostazioni/canali', element: <ChannelsPage /> },
];
