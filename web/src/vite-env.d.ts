/// <reference types="vite/client" />

// Toglie a `import.meta.env` la firma che accetta qualsiasi nome: il sito non legge nessuna
// variabile `VITE_*`, e con questa riga scriverne una per sbaglio non compila.
interface ViteTypeOptions {
  strictImportMetaEnv: unknown;
}

declare const __APP_VERSION__: string;
