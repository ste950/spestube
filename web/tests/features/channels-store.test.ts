import { describe, expect, it } from 'vitest';
import { rejectedFailure } from '@core/result/failure';
import { fail, ok } from '@core/result/result';
import type { Channel } from '@domain/library/entities/channel';
import type { ChannelRepository } from '@domain/library/repositories/channel-repository';
import type { SettingsRepository } from '@domain/settings/repositories/settings-repository';
import { createChannelsStore } from '@features/channels/state/channels-store';

const channel = (id: string, title: string): Channel => ({
  id,
  title,
  handle: null,
  fetchFailed: false,
});

const settings: SettingsRepository = {
  getInstallation: async () => ok({ version: 'test', dev: false, handleLookup: true }),
};

const repository = (overrides: Partial<ChannelRepository> = {}): ChannelRepository => ({
  listChannels: async () => ok([channel('UCs', 'Storia in Chiaro')]),
  addChannel: async () => ok(channel('UCo', 'officina Radio')),
  removeChannel: async () => ok(undefined),
  ...overrides,
});

describe('lo store dei canali', () => {
  it('carica i canali e sa se si può usare un handle', async () => {
    const store = createChannelsStore(repository(), settings);
    await store.getState().load();

    expect(store.getState().channels.map((c) => c.id)).toEqual(['UCs']);
    expect(store.getState().handleLookup).toBe(true);
    expect(store.getState().hasLoaded).toBe(true);
  });

  it('un canale aggiunto entra in ordine di nome, senza guardare le maiuscole', async () => {
    const store = createChannelsStore(repository(), settings);
    await store.getState().load();

    expect(await store.getState().add('UCo')).toBe(true);
    expect(store.getState().channels.map((c) => c.title)).toEqual([
      'officina Radio',
      'Storia in Chiaro',
    ]);
    expect(store.getState().notice).toEqual({
      key: 'channels_added',
      params: { title: 'officina Radio' },
    });
  });

  it("un'aggiunta rifiutata lascia la lista com'è e dice perché", async () => {
    const store = createChannelsStore(
      repository({
        addChannel: async () => fail(rejectedFailure('channel_exists', 'segui già questo canale')),
      }),
      settings,
    );
    await store.getState().load();

    expect(await store.getState().add('UCs')).toBe(false);
    expect(store.getState().channels).toHaveLength(1);
    expect(store.getState().addError).toEqual({ text: 'segui già questo canale' });
    expect(store.getState().isAdding).toBe(false);
  });

  it('un canale rimosso sparisce dalla lista solo se il server conferma', async () => {
    const refused = createChannelsStore(
      repository({ removeChannel: async () => fail(rejectedFailure('x', 'no')) }),
      settings,
    );
    await refused.getState().load();
    await refused.getState().remove('UCs');
    expect(refused.getState().channels).toHaveLength(1);
    expect(refused.getState().error).toEqual({ text: 'no' });

    const accepted = createChannelsStore(repository(), settings);
    await accepted.getState().load();
    await accepted.getState().remove('UCs');
    expect(accepted.getState().channels).toEqual([]);
  });
});
