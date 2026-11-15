import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { describe, expect, it } from 'vitest';
import type { Video } from '@domain/library/entities/video';
import { VideoRow } from '@features/today/components/video-row';

const video: Video = {
  id: 'a0KLaUFyDMU',
  title: 'Cosa cambia tra farina 00, 0, 1, 2, integrale?',
  channelTitle: 'Geopop',
  publishedAt: new Date('2026-11-10T12:00:00Z'),
  durationSeconds: null,
  status: 'released',
  queue: null,
};

const renderRow = (overrides: Partial<Video> = {}) =>
  render(
    <MemoryRouter>
      <ul>
        <VideoRow video={{ ...video, ...overrides }} href="/guarda/a0KLaUFyDMU" />
      </ul>
    </MemoryRouter>,
  );

describe('la riga di un video', () => {
  it('mostra sempre la miniatura, così come la pubblica YouTube', () => {
    const { container } = renderRow();

    const image = container.querySelector('img');
    expect(image).not.toBeNull();
    expect(image?.getAttribute('src')).toBe('https://i.ytimg.com/vi/a0KLaUFyDMU/mqdefault.jpg');
    // Nessun filtro che la spenga o la porti in bianco e nero, né scritto in linea né
    // portato da una seconda classe.
    expect(image?.style.filter).toBe('');
    expect(image?.classList).toHaveLength(1);
  });

  it('è un collegamento al video, con il titolo come nome', () => {
    renderRow();
    const link = screen.getByRole('link', { name: /Cosa cambia tra farina/ });
    expect(link).toHaveAttribute('href', '/guarda/a0KLaUFyDMU');
  });

  it('mostra la durata solo quando è nota', () => {
    const unknown = renderRow();
    expect(screen.queryByText('18:04')).toBeNull();
    unknown.unmount();

    renderRow({ durationSeconds: 18 * 60 + 4 });
    expect(screen.getByText('18:04')).toBeInTheDocument();
  });
});
