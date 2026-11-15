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
