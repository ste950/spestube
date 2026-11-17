import { createJsonApi } from '@core/http/json-api';
import { parseFailure } from '@core/result/failure';
import { fail, ok } from '@core/result/result';
import { isInstallationModel } from '../models/settings-models';
import type { SettingsSource } from './settings-source';

export function createHttpSettingsSource(baseUrl: string): SettingsSource {
  const api = createJsonApi((input, init) => fetch(input, init), baseUrl);

  return {
    async fetchInstallation() {
      const raw = await api.request('GET', '/meta');
      if (!raw.ok) return raw;
      return isInstallationModel(raw.value)
        ? ok(raw.value)
        : fail(parseFailure('InstallationModel', raw.value));
    },
  };
}
