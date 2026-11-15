// Come si mostra l'app: chiara, scura, o come il sistema del dispositivo.
export type ThemePreference = 'system' | 'light' | 'dark';

export const THEME_PREFERENCES: readonly ThemePreference[] = ['system', 'light', 'dark'];

export const DEFAULT_THEME: ThemePreference = 'system';
