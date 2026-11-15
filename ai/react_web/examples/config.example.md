Il solo modulo del sito che legge `import.meta.env`. Converte le variabili della build in
costanti tipizzate, fallisce subito se una manca, e le passa al composition root. Le regole
sono in [Configurazione](../rules/configuration.rule.md).

```ts
// web/src/bootstrap/config.ts

import type { Locale } from '@shared/i18n/catalogs';

// Nessun segreto passa da qui: tutto quello che questo modulo legge finisce nel JavaScript che
// ogni visitatore scarica.

// I nomi letti da questo modulo, e nessun altro: sono gli stessi dichiarati in `vite-env.d.ts`.
type Variable = 'VITE_API_BASE_URL' | 'VITE_DEFAULT_LOCALE';

// Una variabile presente ma vuota (`VITE_API_BASE_URL=`) è un errore di chi scrive il file, non
// un valore: si tratta come assente.
function read(name: Variable): string | undefined {
  const value = import.meta.env[name];

  return value === '' ? undefined : value;
}

function required(name: Variable): string {
  const value = read(name);
  if (value === undefined) {
    // All'avvio e con il nome: una variabile dimenticata si trova leggendo l'errore, non
    // seguendo un `undefined` fino alla richiesta che è partita verso `undefined/orders`.
    throw new Error(`Configurazione mancante: ${name} (vedi web/.env.example)`);
  }

  return value;
}

function url(name: Variable): string {
  const value = required(name);
  if (!URL.canParse(value)) {
    throw new Error(`Configurazione non valida: ${name} non è un URL (${value})`);
  }

  // Senza la barra finale: le sorgenti compongono `${apiBaseUrl}/orders`, e un URL che la
  // conserva produce `//orders`, che alcuni server accettano e altri no.
  return value.replace(/\/+$/, '');
}

function locale(name: Variable, fallback: Locale): Locale {
  const value = read(name);
  if (value === undefined) return fallback;
  if (value === 'it' || value === 'en') return value;

  throw new Error(`Configurazione non valida: ${name} deve essere 'it' o 'en' (${value})`);
}

export const apiBaseUrl = url('VITE_API_BASE_URL');

export const defaultLocale: Locale = locale('VITE_DEFAULT_LOCALE', 'it');

export const appVersion = __APP_VERSION__;

// Non è una variabile `VITE_*`: la scrive Vite dall'opzione `base`. Passa da qui lo stesso,
// perché questo resta l'unico modulo che legge `import.meta.env`.
export const basePath = import.meta.env.BASE_URL;
```

Il file che elenca le variabili, committato accanto a `package.json`:

```bash
# web/.env.example
#
# Copiare in .env.local per lo sviluppo. Tutto quello che c'è qui finisce nel sito pubblicato
# ed è leggibile da chiunque lo apra: nessun segreto, mai.

# L'API da cui il sito legge i dati, senza barra finale.
VITE_API_BASE_URL=http://localhost:8000

# La lingua delle etichette quando il browser non ne chiede una che il sito conosce.
VITE_DEFAULT_LOCALE=it
```

Chi lo consuma è solo il composition root, che passa i valori alle sorgenti:

```ts
// web/src/bootstrap/dependencies.ts — un estratto

import { apiBaseUrl } from './config';

export const orderRepository: OrderRepository = createCachedOrderRepository(
  createHttpOrderSource(apiBaseUrl),
  createLocalOrderCache(),
);
```

Punti chiave:

-   **Un modulo legge l'ambiente, tutti gli altri ricevono.** Una sorgente che legge
    `import.meta.env` da sé è una sorgente che un test non può costruire con un altro URL. La
    `grep` che lo verifica è in [Verifica delle Modifiche](../rules/verification.rule.md).
-   **Il controllo avviene al caricamento del modulo**, cioè all'apertura del sito: una build
    pubblicata senza una variabile si rompe subito e con un messaggio, invece di funzionare a
    metà.
-   I tipi in uscita sono stretti — un URL già normalizzato, un `Locale` dell'insieme chiuso —
    e nessun altro file si chiede più se il valore c'è.
-   `.env.example` è la documentazione della configurazione: ogni variabile ha un valore
    d'esempio e una riga che dice cos'è. I `.env.local` con i valori di chi sviluppa restano
    fuori dal repo.
