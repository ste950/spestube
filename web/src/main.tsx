import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { RouterProvider } from 'react-router/dom';
import { router } from '@app/router';
import { I18nProvider } from '@shared/i18n/i18n-provider';
import { defaultLocale } from '@shared/i18n/catalogs';
import { themeStylesheet } from '@shared/theme/theme-stylesheet';
// I font viaggiano con l'app: nessuna richiesta a un servizio esterno, e funzionano offline.
import '@fontsource-variable/inter';
import '@fontsource-variable/space-grotesk';
import '@shared/theme/global.css';

// Prima del primo render: un componente che parte senza variabili si disegna senza colori
// per un istante, e quell'istante si vede.
const theme = document.createElement('style');
theme.textContent = themeStylesheet();
document.head.append(theme);

const root = document.getElementById('root');
if (root === null) throw new Error('index.html non contiene #root');

createRoot(root).render(
  <StrictMode>
    <I18nProvider locale={defaultLocale}>
      <RouterProvider router={router} />
    </I18nProvider>
  </StrictMode>,
);
