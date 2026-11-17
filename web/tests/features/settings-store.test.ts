import { describe, expect, it } from 'vitest';
import { networkFailure, storageFailure } from '@core/result/failure';
import { fail, ok } from '@core/result/result';
import type { ThemePreference } from '@domain/appearance/entities/theme';
import type { ThemeRepository } from '@domain/appearance/repositories/theme-repository';
import type { ChannelRepository } from '@domain/library/repositories/channel-repository';
import type { SettingsRepository } from '@domain/settings/repositories/settings-repository';
import { createSettingsStore } from '@features/settings/state/settings-store';

const channels = (overrides: Partial<ChannelRepository> = {}): ChannelRepository => ({
  listChannels: async () =>
    ok([{ id: 'UCa', title: 'Officina Radio', handle: null, fetchFailed: false }]),
  addChannel: async () => fail(networkFailure(null, null)),
  removeChannel: async () => ok(undefined),
  ...overrides,
});

const installation: SettingsRepository = {
  getInstallation: async () => ok({ version: '0.1.0', dev: false, handleLookup: false }),
};

const themes = (overrides: Partial<ThemeRepository> = {}): ThemeRepository => ({
  getTheme: async () => ok('dark'),
  setTheme: async () => ok(undefined),
  ...overrides,
});

describe('lo store delle impostazioni', () => {
  it('carica il conteggio dei canali, la versione e il tema', async () => {
    const store = createSettingsStore(channels(), installation, themes(), () => {});
    await store.getState().load();

    expect(store.getState()).toMatchObject({ channelCount: 1, version: '0.1.0', theme: 'dark' });
  });

  it('un elenco di canali che non arriva non nasconde il resto', async () => {
    const store = createSettingsStore(
      channels({ listChannels: async () => fail(networkFailure(null, null)) }),
      installation,
      themes(),
      () => {},
    );
    await store.getState().load();

    expect(store.getState().channelCount).toBeNull();
    expect(store.getState().theme).toBe('dark');
    expect(store.getState().error).toEqual({ key: 'common_error_network' });
  });

  it('il tema scelto si applica subito e si ricorda', async () => {
    const applied: ThemePreference[] = [];
    const saved: ThemePreference[] = [];
    const store = createSettingsStore(
      channels(),
      installation,
      themes({
        setTheme: async (theme) => {
          saved.push(theme);
          return ok(undefined);
        },
      }),
      (theme) => applied.push(theme),
    );
    await store.getState().setTheme('light');

    expect(applied).toEqual(['light']);
    expect(saved).toEqual(['light']);
    expect(store.getState().theme).toBe('light');
  });

  it('se non si riesce a ricordarlo, il tema resta applicato e lo dice', async () => {
    const applied: ThemePreference[] = [];
    const store = createSettingsStore(
      channels(),
      installation,
      themes({ setTheme: async () => fail(storageFailure(new Error('quota'))) }),
      (theme) => applied.push(theme),
    );
    await store.getState().setTheme('dark');

    expect(applied).toEqual(['dark']);
    expect(store.getState().error).toEqual({ key: 'common_error_unexpected' });
  });
});
