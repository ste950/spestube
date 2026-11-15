// Aggiunge a `expect` le asserzioni sul DOM (`toBeInTheDocument`, `toHaveAccessibleName`).
import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';
import { afterEach } from 'vitest';

// Senza le funzioni globali di Vitest la libreria non smonta da sé ciò che un test ha reso:
// lo si fa qui, o il componente di un test resterebbe nella pagina del successivo.
afterEach(cleanup);
