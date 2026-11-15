// Il sito non ha variabili `VITE_*`: lo serve lo stesso binario che espone l'API, quindi non
// c'è niente che cambi fra un'installazione e l'altra. Questo resta l'unico modulo che legge
// `import.meta.env`, per ciò che Vite scrive da sé.

// L'API vive sotto lo stesso indirizzo del sito, e in sviluppo ci arriva il proxy di Vite.
export const apiBaseUrl = '/api';

export const appVersion = __APP_VERSION__;

// Non è una variabile `VITE_*`: la scrive Vite dall'opzione `base`.
export const basePath = import.meta.env.BASE_URL;
