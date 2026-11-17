import { mapOk } from '@core/result/result';
import {
  DEFAULT_THEME,
  THEME_PREFERENCES,
  type ThemePreference,
} from '@domain/appearance/entities/theme';
import type { ThemeRepository } from '@domain/appearance/repositories/theme-repository';
import type { ThemeSource } from '../sources/local-theme-source';

// Un valore sconosciuto, o nessun valore, è il default: un tema illeggibile non deve
// impedire all'app di partire.
const toTheme = (raw: string | null): ThemePreference =>
  THEME_PREFERENCES.find((theme) => theme === raw) ?? DEFAULT_THEME;

export function createLocalThemeRepository(source: ThemeSource): ThemeRepository {
  return {
    async getTheme() {
      return mapOk(await source.read(), toTheme);
    },
    setTheme: (theme) => source.write(theme),
  };
}
