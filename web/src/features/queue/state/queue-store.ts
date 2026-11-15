import { create, type StoreApi, type UseBoundStore } from 'zustand';
import type { Video } from '@domain/library/entities/video';
import type { VideoRepository } from '@domain/library/repositories/video-repository';
import { describeFailure, type Message } from '@shared/i18n/message';

export type QueueState = {
  readonly videos: readonly Video[];
  readonly isLoading: boolean;
  readonly hasLoaded: boolean;
  readonly error: Message | null;
  readonly load: () => Promise<void>;
};

export type QueueStore = UseBoundStore<StoreApi<QueueState>>;

export function createQueueStore(videos: VideoRepository): QueueStore {
  return create<QueueState>((set) => ({
    videos: [],
    isLoading: false,
    hasLoaded: false,
    error: null,
    load: async () => {
      set({ isLoading: true, error: null });
      const result = await videos.queue();
      if (!result.ok) {
        set({ isLoading: false, error: describeFailure(result.failure) });
        return;
      }
      set({ isLoading: false, hasLoaded: true, videos: result.value });
    },
  }));
}
