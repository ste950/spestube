import { create, type StoreApi, type UseBoundStore } from 'zustand';
import type { Video } from '@domain/library/entities/video';
import type { VideoRepository } from '@domain/library/repositories/video-repository';
import { describeFailure, type Message } from '@shared/i18n/message';

export type WatchState = {
  readonly requestedId: string | null;
  readonly video: Video | null;
  readonly isLoading: boolean;
  readonly error: Message | null;
  // Il video è arrivato in fondo: al posto del player c'è la schermata di stop.
  readonly hasEnded: boolean;
  readonly isMarking: boolean;
  readonly load: (videoId: string) => Promise<void>;
  readonly end: () => void;
  readonly markWatched: () => Promise<void>;
};

export type WatchStore = UseBoundStore<StoreApi<WatchState>>;

export function createWatchStore(videos: VideoRepository): WatchStore {
  return create<WatchState>((set, get) => ({
    requestedId: null,
    video: null,
    isLoading: false,
    error: null,
    hasEnded: false,
    isMarking: false,

    load: async (videoId) => {
      set({ requestedId: videoId, video: null, isLoading: true, error: null, hasEnded: false });
      const result = await videos.getVideo(videoId);
      // Una risposta lenta per un video da cui si è già usciti non deve sovrascrivere
      // quello aperto adesso.
      if (get().requestedId !== videoId) return;
      if (!result.ok) {
        set({ isLoading: false, error: describeFailure(result.failure) });
        return;
      }
      set({ isLoading: false, video: result.value });
    },

    end: () => set({ hasEnded: true }),

    markWatched: async () => {
      const current = get().video;
      if (current === null) return;
      set({ isMarking: true, error: null });
      const result = await videos.markWatched(current.id);
      if (!result.ok) {
        set({ isMarking: false, error: describeFailure(result.failure) });
        return;
      }
      set({ isMarking: false, video: result.value });
    },
  }));
}
