import { create, type StoreApi, type UseBoundStore } from 'zustand';
import type { Video } from '@domain/library/entities/video';
import type { VideoRepository } from '@domain/library/repositories/video-repository';
import type { SettingsRepository } from '@domain/settings/repositories/settings-repository';
import { describeFailure, type Message } from '@shared/i18n/message';

export type TodayState = {
  readonly videos: readonly Video[];
  readonly nextReleaseAt: Date | null;
  readonly devTools: boolean;
  readonly isLoading: boolean;
  readonly error: Message | null;
  readonly isReleasing: boolean;
  readonly notice: Message | null;
  readonly load: () => Promise<void>;
  readonly releaseNow: () => Promise<void>;
};

export type TodayStore = UseBoundStore<StoreApi<TodayState>>;

// `announce` riceve il numero dei video in Oggi ogni volta che cambia: è il modo in cui la
// feature lo fa sapere a chi sta fuori (la barra delle sezioni) senza conoscerlo.
export function createTodayStore(
  videos: VideoRepository,
  settings: SettingsRepository,
  announce: (count: number) => void,
): TodayStore {
  return create<TodayState>((set, get) => ({
    videos: [],
    nextReleaseAt: null,
    devTools: false,
    isLoading: false,
    error: null,
    isReleasing: false,
    notice: null,

    load: async () => {
      set({ isLoading: true, error: null });
      const [today, installation] = await Promise.all([videos.today(), settings.getInstallation()]);
      // L'installazione è contorno: se non arriva gli strumenti di sviluppo restano
      // spenti, e il feed si mostra lo stesso.
      if (installation.ok) set({ devTools: installation.value.dev });
      if (!today.ok) {
        set({ isLoading: false, error: describeFailure(today.failure) });
        return;
      }
      set({
        isLoading: false,
        videos: today.value.videos,
        nextReleaseAt: today.value.nextReleaseAt,
      });
      announce(today.value.videos.length);
    },

    releaseNow: async () => {
      set({ isReleasing: true, notice: null, error: null });
      const outcome = await videos.releaseNow();
      if (!outcome.ok) {
        set({ isReleasing: false, error: describeFailure(outcome.failure) });
        return;
      }
      set({
        isReleasing: false,
        notice: {
          key: 'today_dev_released',
          params: { released: outcome.value.released, discarded: outcome.value.discarded },
        },
      });
      await get().load();
    },
  }));
}
