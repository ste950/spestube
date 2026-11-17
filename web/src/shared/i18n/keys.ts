import { it } from './it';

// Il tipo nasce dal catalogo base, non da una lista scritta a mano.
export type TranslationKey = keyof typeof it;
