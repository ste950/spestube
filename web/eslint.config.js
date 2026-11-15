import { readdirSync } from 'node:fs';

import js from '@eslint/js';
import reactHooks from 'eslint-plugin-react-hooks';
import { reactRefresh } from 'eslint-plugin-react-refresh';
import { defineConfig, globalIgnores } from 'eslint/config';
import globals from 'globals';
import tseslint from 'typescript-eslint';

import boundaries from './architecture/boundaries.json' with { type: 'json' };

const { layers, featuresRoot, everywhere, onlyInSources } = boundaries;

const sourcesOf = (dir) => [`${dir}/**/*.{ts,tsx}`];

// Tre gruppi con tre messaggi distinti invece di una lista sola: chi legge l'errore deve
// sapere quale regola ha attraversato, non solo che l'import è vietato.
// `inSources` dice se il blocco riguarda le sorgenti, a cui parlare con l'esterno è concesso.
// Un gruppo vuoto si scarta: finché il progetto non ha pacchetti in `onlyInSources`, ESLint
// rifiuterebbe una lista di divieti senza divieti.
const groupsFor = (layer, { inSources = false } = {}) =>
  [
    {
      group: layer.forbidden,
      message: `${layer.dir} non può dipendere da questo livello: vedi la tabella delle dipendenze.`,
    },
    {
      group: everywhere,
      message: 'Import relativo che esce dalla feature o dal soggetto: usa un alias assoluto.',
    },
    {
      group: inSources ? [] : onlyInSources,
      message: "Il codice che parla con l'esterno vive solo in src/data/<soggetto>/sources/.",
    },
  ].filter(({ group }) => group.length > 0);

const restrict = (patterns) => ({ 'no-restricted-imports': ['error', { patterns }] });

const OUTSIDE_SOURCES = 'Il browser e la rete si toccano solo in src/data/<soggetto>/sources/.';

export default defineConfig([
  globalIgnores(['dist', 'dev-dist']),

  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactRefresh.configs.vite(),
    ],
    languageOptions: { globals: globals.browser },
  },

  // Un blocco per livello. Gli insiemi di file sono disgiunti, e dove non lo sono — le
  // sorgenti dentro `data/`, una feature dentro `features/` — il blocco più interno **ripete**
  // i gruppi di quello esterno: `no-restricted-imports` non si somma fra configurazioni, e
  // l'ultima che vale per un file sostituisce le precedenti.
  ...Object.values(layers).map((layer) => ({
    files: sourcesOf(layer.dir),
    rules: restrict(groupsFor(layer)),
  })),

  // `fetch` e lo storage sono globali, non import: nessuna tabella di import li intercetta.
  // E `import.meta.env` non è né l'uno né l'altro: è sintassi, e si cerca sull'albero.
  {
    files: ['src/**/*.{ts,tsx}'],
    rules: {
      'no-restricted-globals': [
        'error',
        ...['fetch', 'XMLHttpRequest', 'localStorage', 'sessionStorage', 'indexedDB'].map(
          (name) => ({ name, message: OUTSIDE_SOURCES }),
        ),
      ],
      'no-restricted-properties': [
        'error',
        ...['fetch', 'localStorage', 'sessionStorage', 'indexedDB'].map((property) => ({
          object: 'window',
          property,
          message: OUTSIDE_SOURCES,
        })),
      ],
      'no-restricted-syntax': [
        'error',
        {
          selector: "MemberExpression[object.type='MetaProperty'][property.name='env']",
          message: "L'ambiente si legge solo in src/bootstrap/config.ts.",
        },
      ],
    },
  },

  {
    files: ['src/bootstrap/config.ts'],
    rules: { 'no-restricted-syntax': 'off' },
  },

  // Le sorgenti sono l'eccezione dichiarata: parlare con l'esterno è il loro mestiere, quindi
  // riprendono i divieti del loro livello senza `onlyInSources`.
  {
    files: [`${layers.data.dir}/*/sources/**/*.ts`],
    rules: {
      ...restrict(groupsFor(layers.data, { inSources: true })),
      'no-restricted-globals': 'off',
      'no-restricted-properties': 'off',
    },
  },

  // Una feature non importa un'altra feature. Il divieto è «tutte tranne la propria», quindi
  // va scritto una volta per feature: si genera leggendo la cartella, e nasce già valido
  // insieme alla feature successiva.
  ...readdirSync(featuresRoot, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map(({ name }) => ({
      files: [`${featuresRoot}/${name}/**/*.{ts,tsx}`],
      rules: restrict([
        ...groupsFor(layers.features),
        {
          group: ['@features/*', `!@features/${name}`],
          message: "Import di un'altra feature: ciò che due feature condividono passa dal dominio.",
        },
      ]),
    })),

  // `routes.tsx` esporta l'elenco delle rotte insieme ai componenti che le montano, e il
  // ricaricamento a caldo lo segnala. Qui il costo è solo un ricaricamento completo quando si
  // tocca un percorso, che è raro: si spegne su questi file, e solo su questi.
  {
    files: [`${featuresRoot}/*/routes.tsx`, 'src/app/router.tsx'],
    rules: { 'react-refresh/only-export-components': 'off' },
  },
]);
