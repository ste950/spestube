import { create, type StoreApi, type UseBoundStore } from 'zustand';
import type { Channel } from '@domain/library/entities/channel';
import type { ChannelRepository } from '@domain/library/repositories/channel-repository';
import type { SettingsRepository } from '@domain/settings/repositories/settings-repository';
import { describeFailure, type Message } from '@shared/i18n/message';

export type ChannelsState = {
  readonly channels: readonly Channel[];
  // C'è la chiave API: il campo accetta anche un @handle.
  readonly handleLookup: boolean;
  readonly isLoading: boolean;
  readonly hasLoaded: boolean;
  readonly error: Message | null;
  readonly isAdding: boolean;
  readonly addError: Message | null;
  readonly notice: Message | null;
  readonly load: () => Promise<void>;
  // Risponde `true` se il canale è stato aggiunto: chi chiama sa se svuotare il campo.
  readonly add: (reference: string) => Promise<boolean>;
  readonly remove: (channelId: string) => Promise<void>;
};

export type ChannelsStore = UseBoundStore<StoreApi<ChannelsState>>;

const byTitle = (a: Channel, b: Channel): number =>
  a.title.localeCompare(b.title, 'it', { sensitivity: 'base' });

export function createChannelsStore(
  channels: ChannelRepository,
  settings: SettingsRepository,
): ChannelsStore {
  return create<ChannelsState>((set, get) => ({
    channels: [],
    handleLookup: false,
    isLoading: false,
    hasLoaded: false,
    error: null,
    isAdding: false,
    addError: null,
    notice: null,

    load: async () => {
      set({ isLoading: true, error: null });
      const [list, installation] = await Promise.all([
        channels.listChannels(),
        settings.getInstallation(),
      ]);
      if (installation.ok) set({ handleLookup: installation.value.handleLookup });
      if (!list.ok) {
        set({ isLoading: false, error: describeFailure(list.failure) });
        return;
      }
      set({ isLoading: false, hasLoaded: true, channels: list.value });
    },

    add: async (reference) => {
      set({ isAdding: true, addError: null, notice: null });
      const result = await channels.addChannel(reference);
      if (!result.ok) {
        set({ isAdding: false, addError: describeFailure(result.failure) });
        return false;
      }
      set({
        isAdding: false,
        channels: [...get().channels, result.value].sort(byTitle),
        notice: { key: 'channels_added', params: { title: result.value.title } },
      });
      return true;
    },

    remove: async (channelId) => {
      set({ error: null, notice: null });
      const result = await channels.removeChannel(channelId);
      if (!result.ok) {
        set({ error: describeFailure(result.failure) });
        return;
      }
      set({ channels: get().channels.filter((channel) => channel.id !== channelId) });
    },
  }));
}
