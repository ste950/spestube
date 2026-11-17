Le regole di dipendenza fra livelli (vedi
[Struttura del Progetto](../rules/project_structure.rule.md)) si verificano con un `grep`, ma
un `grep` lo si esegue solo se ci si ricorda di farlo. Qui diventano **deterministiche**: un
import vietato fa fallire `npx eslint .`, un colore scritto a mano in un foglio fa fallire
`npx stylelint`. La forma generale è in
[Test di Architettura](../../quality/rules/architecture_tests.rule.md).

## 🔧 Perché i linter, e non solo un test

È la stessa scelta dello stack React Native, e per la stessa ragione: **ESLint legge già gli
import**, e `no-restricted-imports` accetta esattamente la forma in cui la tabella delle
dipendenze è scritta — una lista di prefissi vietati per cartella. Con gli
[alias assoluti](../rules/project_structure.rule.md#-import-assoluti-con-alias) obbligatori fra
livelli, il confronto è fra stringhe, senza risolvere percorsi e senza un plugin in più.

Il web aggiunge un secondo confine che React Native non ha: lo stile sta in **fogli CSS**, che
ESLint non legge. Lì il lavoro lo fa **stylelint**, e la regola è la stessa — un valore del
design scritto a mano dove dovrebbe esserci un token è un errore sulla riga in cui è scritto.

Quello che si guadagna rispetto a un test è **quando** il fallimento arriva: mentre si scrive,
nell'editor, sulla riga dell'import o della dichiarazione. Restano due condizioni perché sia
una garanzia e non una comodità — la configurazione è **committata nel progetto**, e i linter
**girano nella CI** (vedi [Job di CI](./ci_job.example.yml)) — e resta un buco, che è il motivo
per cui più sotto c'è comunque un test: **un linter che non analizza niente non dice niente.**
Una cartella rinominata fa smettere di valere i blocchi che la riguardano, in silenzio.

## 📋 La tabella, in un posto solo

```json
// web/architecture/boundaries.json
{
  "everywhere": ["../../**"],
  "onlyInSources": ["idb", "idb/**"],
  "featuresRoot": "src/features",
  "layers": {
    "app": {
      "dir": "src/app",
      "forbidden": [
        "@domain/**",
        "@data/**",
        "@core/**",
        "@bootstrap/**",
        "!@bootstrap/config",
        "@features/*/*",
        "!@features/*/routes",
        "@features/*/*/**"
      ]
    },
    "features": {
      "dir": "src/features",
      "forbidden": ["@data/**", "@app/**", "@bootstrap/**", "!@bootstrap/dependencies"]
    },
    "data": {
      "dir": "src/data",
      "forbidden": [
        "@features/**",
        "@shared/**",
        "@bootstrap/**",
        "@app/**",
        "react",
        "react-dom",
        "react-dom/**",
        "react-router",
        "react-router/**"
      ]
    },
    "domain": {
      "dir": "src/domain",
      "forbidden": [
        "@data/**",
        "@features/**",
        "@shared/**",
        "@bootstrap/**",
        "@app/**",
        "react",
        "react-dom",
        "react-dom/**",
        "react-router",
        "react-router/**",
        "zustand",
        "zustand/**"
      ]
    },
    "core": {
      "dir": "src/core",
      "forbidden": ["@domain/**", "@data/**", "@features/**", "@shared/**", "@app/**", "@bootstrap/**"]
    },
    "shared": {
      "dir": "src/shared",
      "forbidden": ["@domain/**", "@data/**", "@features/**", "@app/**", "@bootstrap/**"]
    }
  }
}
```

La sintassi dei gruppi è quella di `.gitignore`, ed è lei a rendere esprimibili le due
eccezioni della tabella. `"@bootstrap/**", "!@bootstrap/dependencies"` vieta tutto
`bootstrap/` tranne il modulo che le feature possono toccare. Per `app/` serve un passo in
più: `.gitignore` non riammette un file se è esclusa la cartella che lo contiene, quindi il
divieto si scrive **per livelli** — `@features/*/*` esclude i figli diretti di ogni feature,
`!@features/*/routes` riammette `routes`, `@features/*/*/**` esclude tutto quello che sta più
in basso. Il risultato è che `app/` vede di una feature solo il suo `routes.tsx`.

Tre voci non sono livelli e vanno lette a parte:

-   **`everywhere`** è l'unico divieto che vale ovunque: `../../**` è un import relativo che
    **esce** dalla feature o dal soggetto in cui si trova. I relativi restano comodi dentro
    l'unità (`../state/orders-store`), e sono vietati appena la attraversano.
-   **`onlyInSources`** sono i pacchetti che parlano con l'esterno: vietati a tutti i livelli e
    riammessi solo in `src/data/<soggetto>/sources/`. Un progetto aggiunge qui il proprio client
    HTTP e ogni SDK di terze parti che introduce.
-   **`src/bootstrap`** non compare, e l'assenza è voluta: è il composition root, l'unico modulo
    che può importare tutto.

## 🚧 La configurazione di ESLint

```js
// web/eslint.config.js

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
const groupsFor = (layer) => [
  {
    group: layer.forbidden,
    message: `${layer.dir} non può dipendere da questo livello: vedi la tabella delle dipendenze.`,
  },
  {
    group: everywhere,
    message: 'Import relativo che esce dalla feature o dal soggetto: usa un alias assoluto.',
  },
  {
    group: onlyInSources,
    message: "Il codice che parla con l'esterno vive solo in src/data/<soggetto>/sources/.",
  },
];

const restrict = (patterns) => ({ 'no-restricted-imports': ['error', { patterns }] });

const OUTSIDE_SOURCES = 'Il browser e la rete si toccano solo in src/data/<soggetto>/sources/.';

export default defineConfig([
  globalIgnores(['dist']),

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
      ...restrict(groupsFor(layers.data).slice(0, 2)),
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
```

## 🎨 La configurazione di stylelint

```js
// web/stylelint.config.js

import breakpoints from './src/shared/theme/breakpoints.json' with { type: 'json' };

const TOKEN = 'Valore scritto a mano: usa una variabile del tema (var(--…)).';
const UNITS = ['px', 'rem', 'em'];

/** @type {import('stylelint').Config} */
export default {
  extends: ['stylelint-config-standard'],
  rules: {
    // `composes` è la proprietà con cui un CSS Module compone le classi: stylelint non la
    // conosce, e senza questa riga la segnalerebbe come un refuso.
    'property-no-unknown': [true, { ignoreProperties: ['composes'] }],

    // Uno stile globale scritto dentro un modulo esce dal suo componente: l'unico foglio
    // globale è src/shared/theme/global.css.
    'selector-pseudo-class-disallowed-list': [
      ['global'],
      { message: 'Stile globale in un modulo.' },
    ],

    // Colori: tutti dai token, in nessuna delle loro forme.
    'color-no-hex': [true, { message: TOKEN }],
    'color-named': ['never', { message: TOKEN }],
    'function-disallowed-list': [
      ['rgb', 'rgba', 'hsl', 'hsla', 'hwb', 'lab', 'lch', 'oklab', 'oklch', 'color'],
      { message: TOKEN },
    ],

    // Spaziature, raggi e dimensioni del testo: nessuna unità scritta a mano. Il bordo da un
    // pixel e le proporzioni restano liberi, perché sono geometria e non design.
    'declaration-property-unit-disallowed-list': [
      { '/^(margin|padding|gap|row-gap|column-gap|inset|font-size|border-radius)/': UNITS },
      { message: TOKEN },
    ],

    // Le larghezze delle media query sono le sole ammesse, e arrivano dallo stesso file da cui
    // le legge tokens.ts: una variabile CSS dentro `@media` non vale, e un elenco ripetuto qui
    // divergerebbe al primo cambio.
    'media-feature-name-value-allowed-list': {
      width: Object.values(breakpoints),
    },
  },
};
```

## 🧪 Il test che difende i linter

I linter verificano i file che gli arrivano. Questo test verifica che gliene arrivino — ed è
l'unica parte della garanzia che un linter non può dare a sé stesso.

```ts
// web/tests/architecture/boundaries.test.ts

import { existsSync, readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

import boundaries from '../../architecture/boundaries.json';

// tests/architecture/ → tests/ → web/. Il percorso si deriva dalla posizione del file e non
// dalla cartella corrente: un percorso sbagliato non farebbe fallire questo test, lo farebbe
// passare su un insieme vuoto — cioè esattamente il difetto che è qui per intercettare.
const WEB = join(import.meta.dirname, '..', '..');

function filesUnder(dir: string, pattern: RegExp): string[] {
  if (!existsSync(join(WEB, dir))) return [];

  return readdirSync(join(WEB, dir), { withFileTypes: true }).flatMap((entry) => {
    if (entry.isDirectory()) return filesUnder(join(dir, entry.name), pattern);
    return pattern.test(entry.name) ? [join(dir, entry.name)] : [];
  });
}

describe('i confini che i linter difendono esistono ancora', () => {
  it.each(Object.entries(boundaries.layers))('%s contiene dei file', (_name, layer) => {
    // Una cartella rinominata o spostata non fa protestare ESLint: nessun file corrisponde più
    // al glob, e il blocco che la riguardava smette di valere senza dirlo a nessuno.
    expect(filesUnder(layer.dir, /\.tsx?$/)).not.toHaveLength(0);
  });

  it('esiste almeno una feature, con il suo routes.tsx', () => {
    // I blocchi per feature si generano leggendo questa cartella: se è vuota, l'indipendenza
    // fra feature è difesa da zero regole.
    const features = readdirSync(join(WEB, boundaries.featuresRoot), { withFileTypes: true })
      .filter((entry) => entry.isDirectory())
      .map((entry) => entry.name);

    expect(features).not.toHaveLength(0);
    expect(
      features.filter(
        (name) => !existsSync(join(WEB, boundaries.featuresRoot, name, 'routes.tsx')),
      ),
    ).toEqual([]);
  });

  it('stylelint ha dei fogli da leggere', () => {
    expect(filesUnder('src', /\.module\.css$/)).not.toHaveLength(0);
  });
});

const GUARDS = [
  'no-restricted-imports',
  'no-restricted-globals',
  'no-restricted-properties',
  'no-restricted-syntax',
];

it('nessun file spegne una regola di confine', () => {
  const sources = filesUnder('src', /\.tsx?$/);
  const sheets = filesUnder('src', /\.css$/);
  expect(sources).not.toHaveLength(0);

  // Un `eslint-disable` o uno `stylelint-disable` locale è il modo più rapido di attraversare
  // un confine lasciando la CI verde. Gli altri silenziamenti restano ammessi con la loro
  // ragione; questi no.
  const silenced = [
    ...sources.filter((file) =>
      GUARDS.some((guard) =>
        new RegExp(`eslint-disable[^\\n]*${guard}`).test(readFileSync(join(WEB, file), 'utf8')),
      ),
    ),
    ...sheets.filter((file) => readFileSync(join(WEB, file), 'utf8').includes('stylelint-disable')),
  ];

  expect(silenced).toEqual([]);
});
```

L'assenza di `expect` dentro i cicli è voluta: si raccolgono i file colpevoli e si confronta la
lista con quella vuota, così il fallimento **nomina i file** invece di fermarsi al primo.

## ✅ Dove gira

-   **`npm test`** esegue il test: sta in `web/tests/architecture/` insieme agli altri test di
    architettura, senza un comando dedicato.
-   **`npx eslint .`** e **`npx stylelint "src/**/*.css"`** eseguono i divieti.

Girano tutti nel [job di CI](./ci_job.example.yml) e non sulla macchina di chi scrive — vedi
Dove gira la verifica.

## 🚫 Cosa resta fuori

L'indipendenza fra **soggetti** di `domain/` non è né una regola del linter né un test. La
regola dice che un import fra soggetti diversi è da evitare e che, se serve davvero, il segnale
è che i due soggetti vanno fusi: è un giudizio, non un divieto, e imporlo produrrebbe un
`eslint-disable` il giorno in cui la fusione è la risposta giusta.

Resta fuori anche il nome di una classe sbagliato in un componente (`styles.ordeRow`): Vite
non dichiara quali classi esistano in un foglio, e nessuno dei due linter legge l'altro file.
Si vede nella [verifica nel browser](../rules/verification.rule.md#%EF%B8%8F-verifica-nel-browser).
