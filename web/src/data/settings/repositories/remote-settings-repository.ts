import { mapOk } from '@core/result/result';
import type { SettingsRepository } from '@domain/settings/repositories/settings-repository';
import { toInstallation } from '../mappers/settings-mappers';
import type { SettingsSource } from '../sources/settings-source';

export function createRemoteSettingsRepository(source: SettingsSource): SettingsRepository {
  return {
    async getInstallation() {
      return mapOk(await source.fetchInstallation(), toInstallation);
    },
  };
}
