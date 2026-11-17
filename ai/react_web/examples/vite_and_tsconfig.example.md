La configurazione che le regole danno per scontata: alias di import dichiarati **una volta
sola**, `strict` con le due opzioni che lo completano, CSS Modules con le classi in
`camelCase`, e i test nello stesso programma della build. Il perché degli alias è in
[Import assoluti con alias](../rules/project_structure.rule.md#-import-assoluti-con-alias).

## TypeScript: tre progetti e una radice vuota

```json
// web/tsconfig.json
{
  "files": [],
  "references": [
    { "path": "./tsconfig.app.json" },
    { "path": "./tsconfig.test.json" },
    { "path": "./tsconfig.node.json" }
  ]
}
```

```json
// web/tsconfig.app.json
{
  "compilerOptions": {
    "tsBuildInfoFile": "./node_modules/.tmp/tsconfig.app.tsbuildinfo",
    "target": "es2023",
    "lib": ["ES2023", "DOM", "DOM.Iterable"],
    "module": "esnext",
    "moduleResolution": "bundler",
    "moduleDetection": "force",
    "jsx": "react-jsx",
    "types": ["vite/client"],
    "noEmit": true,
    "verbatimModuleSyntax": true,
    "allowArbitraryExtensions": true,
    "resolveJsonModule": true,
    "skipLibCheck": true,

    "strict": true,
    "noUncheckedIndexedAccess": true,
    "exactOptionalPropertyTypes": true,
    "noFallthroughCasesInSwitch": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "erasableSyntaxOnly": true,

    "paths": {
      "@app/*": ["./src/app/*"],
      "@bootstrap/*": ["./src/bootstrap/*"],
      "@core/*": ["./src/core/*"],
      "@domain/*": ["./src/domain/*"],
      "@data/*": ["./src/data/*"],
      "@features/*": ["./src/features/*"],
      "@shared/*": ["./src/shared/*"]
    }
  },
  "include": ["src"]
}
```

```json
// web/tsconfig.test.json
{
  "extends": "./tsconfig.app.json",
  "compilerOptions": {
    "tsBuildInfoFile": "./node_modules/.tmp/tsconfig.test.tsbuildinfo",
    "types": ["vite/client", "node"]
  },
  "include": ["src", "tests"]
}
```

```json
// web/tsconfig.node.json
{
  "compilerOptions": {
    "tsBuildInfoFile": "./node_modules/.tmp/tsconfig.node.tsbuildinfo",
    "target": "es2023",
    "lib": ["ES2023"],
    "module": "nodenext",
    "types": ["node"],
    "noEmit": true,
    "verbatimModuleSyntax": true,
    "resolveJsonModule": true,
    "skipLibCheck": true,
    "strict": true,
    "erasableSyntaxOnly": true
  },
  "include": ["vite.config.ts"]
}
```

I test hanno un progetto loro perché leggono i file del progetto con `node:fs` e hanno bisogno
dei tipi di Node — che nel codice del sito non devono esistere: con `"node"` fra i `types` di
`tsconfig.app.json`, un `process.env` scritto in un componente compilerebbe e si romperebbe
nel browser. Il progetto dei test eredita tutto il resto, quindi un test non è mai controllato
con regole più lasche del codice che prova.

Per il resto è la forma del template `react-ts` di Vite, con quattro aggiunte: `strict` scritto anche dove
la versione di TypeScript lo dà già per default — chi legge non deve sapere quale default vale —,
`noUncheckedIndexedAccess`, `exactOptionalPropertyTypes` e gli alias.

`erasableSyntaxOnly` viene dal template e conviene tenerlo: vieta la sintassi che TypeScript
deve tradurre invece di togliere — `enum`, `namespace`, i parametri-proprietà dei costruttori.
Sono esattamente le forme che le [regole del dominio](../rules/domain_layer.rule.md#-entità)
escludono già, e con questa opzione l'esclusione la verifica `tsc`.

## Vite: alias letti da `tsconfig`, test nella stessa configurazione

```ts
// web/vite.config.ts

/// <reference types="vitest/config" />
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

import packageJson from './package.json' with { type: 'json' };

export default defineConfig({
  plugins: [react()],

  // Gli alias non si ripetono qui: Vite li legge dai `paths` di tsconfig.app.json, e un alias
  // nuovo esiste per `tsc`, per la build e per i test nello stesso momento.
  resolve: { tsconfigPaths: true },

  // Il percorso sotto cui il sito è servito. Alla radice del dominio resta '/'; il router lo
  // riceve da `import.meta.env.BASE_URL` e nessun link lo scrive a mano.
  base: '/',

  css: {
    // `.order-row` nel foglio, `styles.orderRow` nel componente: il CSS resta nella sua
    // convenzione, TypeScript nella sua, e la forma con il trattino non esiste dal lato TS.
    modules: { localsConvention: 'camelCaseOnly' },
  },

  // La versione entra nella build da `package.json`, che è l'unico posto in cui si scrive.
  define: { __APP_VERSION__: JSON.stringify(packageJson.version) },

  test: {
    environment: 'jsdom',
    include: ['tests/**/*.test.{ts,tsx}'],
    setupFiles: ['tests/setup.ts'],
  },
});
```

```ts
// web/src/vite-env.d.ts

/// <reference types="vite/client" />

// Toglie a `import.meta.env` la firma che accetta qualsiasi nome: senza questa riga
// `import.meta.env.VITE_API_BASE_URLL` compila e vale `undefined`.
interface ViteTypeOptions {
  strictImportMetaEnv: unknown;
}

// Ogni variabile che `bootstrap/config.ts` legge è dichiarata qui: un nome sbagliato è un
// errore di `tsc`, non un `undefined` che arriva a una `fetch`.
interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string | undefined;
  readonly VITE_DEFAULT_LOCALE: string | undefined;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

declare const __APP_VERSION__: string;
```

```ts
// web/tests/setup.ts

// Aggiunge a `expect` le asserzioni sul DOM (`toBeInTheDocument`, `toHaveAccessibleName`).
import '@testing-library/jest-dom/vitest';
```

## Node e prettier

```text
# web/.nvmrc
24
```

```json
// web/.prettierrc.json
{ "singleQuote": true, "printWidth": 100 }
```

La versione di Node sta in `.nvmrc` perché la leggono sia chi sviluppa sia la
[CI](./ci_job.example.yml): scriverla nel workflow vorrebbe dire tenerne due. Il minimo è
**Node 22**, che è quello che Vitest e jsdom richiedono; `24` è la versione con supporto lungo.

La configurazione di prettier è quella con cui sono scritti gli esempi di questo stack: uno
snippet copiato da qui passa `prettier --check` così com'è.

## Gli script

```json
// web/package.json — la sezione "scripts"
{
  "scripts": {
    "dev": "vite",
    "build": "tsc -b && vite build",
    "preview": "vite preview",
    "test": "vitest"
  }
}
```

Punti chiave:

-   **`tsc -b`, non `tsc --noEmit`.** La radice ha `"files": []`: senza `-b`, `tsc` controlla
    l'insieme vuoto ed esce verde. Con `-b` segue i riferimenti e controlla `src/`, `tests/` e
    la configurazione di Vite, ciascuno con le sue librerie — il DOM per il sito, il DOM e Node
    per i test, Node per la configurazione. Vedi [Verifica delle Modifiche](../rules/verification.rule.md).
-   **I tipi di variabili in `import.meta.env` sono `string | undefined`**, anche quando la
    variabile è obbligatoria: è vero, e lo trasforma in un tipo stretto
    [`bootstrap/config.ts`](./config.example.md), una volta sola.
-   **Gli alias stanno in `tsconfig.app.json`** e solo lì, e `resolve.tsconfigPaths` li porta
    a Vite. Un secondo elenco in `vite.config.ts` funziona finché qualcuno non ne aggiorna uno
    solo: allora `tsc` è verde e la build non trova il modulo.
-   **I test usano la configurazione di Vite**: stessi alias, stessa trasformazione del JSX,
    stessi CSS Modules. Un test che passa con una configurazione diversa da quella della build
    verifica un altro programma.
-   `noUncheckedIndexedAccess` rende `string | undefined` anche le classi di un CSS Module
    (`styles.orderRow`): `className` lo accetta, ed è onesto — Vite non sa quali classi
    esistano nel foglio.
-   La versione entra con `define` e si dichiara in `vite-env.d.ts`: è una costante sostituita
    nella build, non una lettura di `package.json` a runtime, che nel browser non esiste.
