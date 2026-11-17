import { fireEvent, render, screen } from '@testing-library/react';
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
  collectionIds: [],
};

const renderRow = (overrides: Partial<Video> = {}, onSave: () => void = () => {}) =>
  render(
    <MemoryRouter>
      <ul>
        <VideoRow
          video={{ ...video, ...overrides }}
          href="/guarda/a0KLaUFyDMU"
          saveLabel="Salva in una raccolta: farina"
          onSave={onSave}
        />
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

  it('ha un bottone per salvarlo, fuori dal collegamento', () => {
    let asked = 0;
    renderRow({}, () => {
      asked += 1;
    });

    const button = screen.getByRole('button', { name: 'Salva in una raccolta: farina' });
    // Dentro il link non si raggiungerebbe da tastiera, e un tocco aprirebbe il video.
    expect(button.closest('a')).toBeNull();
    fireEvent.click(button);
    expect(asked).toBe(1);
  });

  it('il segnalibro si riempie quando il video è in una raccolta', () => {
    const empty = renderRow();
    expect(empty.container.querySelector('button svg')).toHaveAttribute('fill', 'none');
    empty.unmount();

    const saved = renderRow({ collectionIds: [3] });
    expect(saved.container.querySelector('button svg')).toHaveAttribute('fill', 'currentColor');
  });

  it('mostra la durata solo quando è nota', () => {
    const unknown = renderRow();
    expect(screen.queryByText('18:04')).toBeNull();
    unknown.unmount();

    renderRow({ durationSeconds: 18 * 60 + 4 });
    expect(screen.getByText('18:04')).toBeInTheDocument();
  });
});
