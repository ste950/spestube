import { channelRepository, settingsRepository, themeRepository } from '@bootstrap/dependencies';
import { applyTheme } from '@shared/theme/apply-theme';
import { createSettingsStore } from './settings-store';

export const useSettingsStore = createSettingsStore(
  channelRepository,
  settingsRepository,
  themeRepository,
  applyTheme,
);
