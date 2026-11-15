import { act, render, screen } from '@testing-library/react';
import { createMemoryRouter } from 'react-router';
import { RouterProvider } from 'react-router/dom';
import { beforeEach, describe, expect, it } from 'vitest';
import { AppLayout } from '@app/layouts/app-layout';
import { I18nProvider } from '@shared/i18n/i18n-provider';
import { setNavBadge, useNavBadges } from '@shared/navigation/nav-badges';

function renderAt(path: string) {
  const router = createMemoryRouter(
    [{ element: <AppLayout />, children: [{ path: '*', element: <p>pagina</p> }] }],
    { initialEntries: [path] },
  );
  return render(
    <I18nProvider locale="it">
      <RouterProvider router={router} />
    </I18nProvider>,
  );
}

describe('la barra delle sezioni', () => {
  beforeEach(() => {
    useNavBadges.setState({ counts: {} });
  });

  it('ha tre sezioni e nessuna quarta', () => {
    renderAt('/');
    const links = screen.getAllByRole('link').filter((link) => link.closest('nav') !== null);
    expect(links.map((link) => link.textContent)).toEqual(['Oggi', 'Da vedere', 'Canali']);
  });

  it('senza video non mostra nessun numero', () => {
    renderAt('/');
    expect(screen.getByRole('link', { name: 'Oggi' })).toBeInTheDocument();
  });

  it('accanto a Oggi mostra quanti video ci sono, e lo dice anche a chi ascolta', () => {
    renderAt('/canali');
    act(() => setNavBadge('/', 2));

    const today = screen.getByRole('link', { name: 'Oggi, 2 nuovi' });
    expect(today).toHaveTextContent('2');
    // Il numero si aggiorna, e torna a sparire quando Oggi si svuota.
    act(() => setNavBadge('/', 1));
    expect(screen.getByRole('link', { name: 'Oggi, 1 nuovo' })).toBeInTheDocument();
    act(() => setNavBadge('/', 0));
    expect(screen.getByRole('link', { name: 'Oggi' })).toBeInTheDocument();
  });

  it('tiene accesa Oggi mentre si guarda un video', () => {
    renderAt('/guarda/abc');
    expect(screen.getByRole('link', { name: 'Oggi' })).toHaveAttribute('aria-current', 'page');
    expect(screen.getByRole('link', { name: 'Canali' })).not.toHaveAttribute('aria-current');
  });
});
