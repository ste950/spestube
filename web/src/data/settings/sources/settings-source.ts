import type { Result } from '@core/result/result';
import type { InstallationModel } from '../models/settings-models';

export interface SettingsSource {
  fetchInstallation(): Promise<Result<InstallationModel>>;
}
