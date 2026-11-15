import { storageFailure } from '@core/result/failure';
import { fail, ok, type Result } from '@core/result/result';

// La stessa chiave la legge lo script in testa a index.html, che applica il tema prima che
// parta l'app. Versionata: un formato nuovo cambia la chiave invece di migrare.
export const THEME_STORAGE_KEY = 'spestube.theme.v1';

export interface ThemeSource {
  read(): Promise<Result<string | null>>;
  write(value: string): Promise<Result<void>>;
}

export function createLocalThemeSource(): ThemeSource {
  return {
    async read() {
      try {
        return ok(localStorage.getItem(THEME_STORAGE_KEY));
      } catch (error) {
        // `localStorage` può lanciare anche solo a leggerlo: navigazione privata, storage
        // disabilitato.
        return fail(storageFailure(error));
      }
    },
    async write(value) {
      try {
        localStorage.setItem(THEME_STORAGE_KEY, value);
        return ok(undefined);
      } catch (error) {
        return fail(storageFailure(error));
      }
    },
  };
}
