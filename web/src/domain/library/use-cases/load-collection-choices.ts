import { ok, type Result } from '@core/result/result';
import type { AsyncUseCase } from '@core/use-cases/use-case';
import type { Collection } from '../entities/collection';
import type { CollectionRepository } from '../repositories/collection-repository';
import type { VideoRepository } from '../repositories/video-repository';

export type CollectionChoices = {
  readonly collections: readonly Collection[];
  // Le raccolte che contengono già il video.
  readonly selected: readonly number[];
};

// Ciò che serve per chiedere «in quali raccolte salvi questo video?»: tutte le raccolte, e
// quelle in cui il video c'è già. Il video si rilegge, perché da un'altra schermata la sua
// appartenenza può essere cambiata.
export const makeLoadCollectionChoices =
  (
    collections: CollectionRepository,
    videos: VideoRepository,
  ): AsyncUseCase<string, Result<CollectionChoices>> =>
  async (videoId) => {
    const [list, video] = await Promise.all([
      collections.listCollections(),
      videos.getVideo(videoId),
    ]);
    if (!list.ok) return list;
    if (!video.ok) return video;
    return ok({ collections: list.value, selected: video.value.collectionIds });
  };
