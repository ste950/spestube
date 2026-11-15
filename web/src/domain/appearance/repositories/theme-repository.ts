import type { Result } from '@core/result/result';
import type { ThemePreference } from '../entities/theme';

// Il tema è una scelta del dispositivo e non dell'installazione: il telefono di sera e il
// computer di giorno possono volere cose diverse.
export interface ThemeRepository {
  getTheme(): Promise<Result<ThemePreference>>;
  setTheme(theme: ThemePreference): Promise<Result<void>>;
}
