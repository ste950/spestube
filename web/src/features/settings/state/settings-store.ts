import { create, type StoreApi, type UseBoundStore } from 'zustand';
import { DEFAULT_THEME, type ThemePreference } from '@domain/appearance/entities/theme';
import type { ThemeRepository } from '@domain/appearance/repositories/theme-repository';
import type { ChannelRepository } from '@domain/library/repositories/channel-repository';
import type { SettingsRepository } from '@domain/settings/repositories/settings-repository';
import { describeFailure, type Message } from '@shared/i18n/message';

export type SettingsState = {
  readonly channelCount: number | null;
  readonly version: string | null;
  readonly theme: ThemePreference;
  readonly error: Message | null;
  readonly load: () => Promise<void>;
  readonly setTheme: (theme: ThemePreference) => Promise<void>;
};

export type SettingsStore = UseBoundStore<StoreApi<SettingsState>>;

// `apply` porta il tema sulla pagina: è presentazione, e la riceve chi lega lo store.
export function createSettingsStore(
  channels: ChannelRepository,
  installation: SettingsRepository,
  themes: ThemeRepository,
  apply: (theme: ThemePreference) => void,
): SettingsStore {
  return create<SettingsState>((set) => ({
    channelCount: null,
    version: null,
    theme: DEFAULT_THEME,
    error: null,

    load: async () => {
      const [list, info, theme] = await Promise.all([
        channels.listChannels(),
        installation.getInstallation(),
        themes.getTheme(),
      ]);
      // Ogni voce si mostra per conto suo: un conteggio che non arriva non nasconde il tema.
      set({
        channelCount: list.ok ? list.value.length : null,
        version: info.ok ? info.value.version : null,
        theme: theme.ok ? theme.value : DEFAULT_THEME,
        error: list.ok ? null : describeFailure(list.failure),
      });
    },

    setTheme: async (theme) => {
      // Il tema cambia subito anche se non si riesce a ricordarlo: chi l'ha scelto lo vede,
      // e alla prossima apertura tornerà quello di prima.
      apply(theme);
      set({ theme, error: null });
      const saved = await themes.setTheme(theme);
      if (!saved.ok) set({ error: describeFailure(saved.failure) });
    },
  }));
}
