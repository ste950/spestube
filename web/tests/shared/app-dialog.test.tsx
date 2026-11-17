import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { AppDialog } from '@shared/components/app-dialog';

describe('la finestra modale', () => {
  it('si apre come modale, con il titolo come nome', () => {
    render(
      <AppDialog title="Nuova raccolta" subject="Antenne LoRa" onClose={() => {}}>
        <p>contenuto</p>
      </AppDialog>,
    );
    const dialog = screen.getByRole('dialog', { name: 'Nuova raccolta' });
    expect(dialog).toHaveAttribute('open');
    expect(screen.getByText('Antenne LoRa')).toBeInTheDocument();
    expect(screen.getByText('contenuto')).toBeInTheDocument();
  });

  it('Esc chiede di chiudere, tranne mentre salva', () => {
    let closed = 0;
    const onClose = () => {
      closed += 1;
    };
    const { rerender } = render(
      <AppDialog title="Finestra" onClose={onClose}>
        <p>x</p>
      </AppDialog>,
    );
    fireEvent(screen.getByRole('dialog'), new Event('cancel', { cancelable: true }));
    expect(closed).toBe(1);

    rerender(
      <AppDialog title="Finestra" busy onClose={onClose}>
        <p>x</p>
      </AppDialog>,
    );
    fireEvent(screen.getByRole('dialog'), new Event('cancel', { cancelable: true }));
    expect(closed).toBe(1);
  });
});
