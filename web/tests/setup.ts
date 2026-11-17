// Aggiunge a `expect` le asserzioni sul DOM (`toBeInTheDocument`, `toHaveAccessibleName`).
import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';
import { afterEach } from 'vitest';

// Senza le funzioni globali di Vitest la libreria non smonta da sé ciò che un test ha reso:
// lo si fa qui, o il componente di un test resterebbe nella pagina del successivo.
afterEach(cleanup);

// jsdom non implementa l'apertura di un `<dialog>`: qui se ne mette il minimo che i test
// possono osservare, l'attributo `open`. Il focus intrappolato e il resto della pagina
// inerte li fa il browser vero, e si verificano lì.
if (typeof HTMLDialogElement.prototype.showModal !== 'function') {
  HTMLDialogElement.prototype.showModal = function showModal(this: HTMLDialogElement) {
    this.setAttribute('open', '');
  };
  HTMLDialogElement.prototype.close = function close(this: HTMLDialogElement) {
    this.removeAttribute('open');
  };
}
