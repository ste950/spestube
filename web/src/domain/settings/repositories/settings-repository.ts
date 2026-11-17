import type { Result } from '@core/result/result';
import type { Installation } from '../entities/installation';

export interface SettingsRepository {
  getInstallation(): Promise<Result<Installation>>;
}
