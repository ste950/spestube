import type { Installation } from '@domain/settings/entities/installation';
import type { InstallationModel } from '../models/settings-models';

export const toInstallation = (model: InstallationModel): Installation => ({
  version: model.version,
  dev: model.dev,
  handleLookup: model.handle_lookup,
});
