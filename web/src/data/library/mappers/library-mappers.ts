import type { Channel } from '@domain/library/entities/channel';
import type {
  Collection,
  CollectionDetail,
  CollectionVideo,
} from '@domain/library/entities/collection';
import type { ReleaseOutcome, TodayFeed, Video, VideoStatus } from '@domain/library/entities/video';
import type {
  ChannelModel,
  CollectionDetailModel,
  CollectionModel,
  CollectionVideoModel,
  ReleaseModel,
  TodayModel,
  VideoModel,
} from '../models/library-models';

export const toChannel = (model: ChannelModel): Channel => ({
  id: model.id,
  title: model.title,
  handle: model.handle === '' ? null : model.handle,
  fetchFailed: model.last_fetch_error !== '',
});

const KNOWN_STATUSES: readonly VideoStatus[] = ['released', 'watched', 'archived'];

// Uno stato che l'app non conosce diventa «archiviato»: il video resta apribile, ma non
// compare in Oggi, dove un video di troppo sarebbe un errore visibile.
const toStatus = (raw: string): VideoStatus =>
  KNOWN_STATUSES.find((status) => status === raw) ?? 'archived';

export const toVideo = (model: VideoModel): Video => ({
  id: model.id,
  title: model.title,
  channelTitle: model.channel_title,
  publishedAt: new Date(model.published_at),
  durationSeconds: model.duration_seconds,
  status: toStatus(model.status),
  collectionIds: model.collection_ids,
});

export const toCollection = (model: CollectionModel): Collection => ({
  id: model.id,
  name: model.name,
  description: model.description,
  videoCount: model.video_count,
  coverVideoId: model.cover_video_id,
});

const toCollectionVideo = (model: CollectionVideoModel): CollectionVideo => ({
  ...toVideo(model),
  addedAt: new Date(model.added_at),
});

export const toCollectionDetail = (model: CollectionDetailModel): CollectionDetail => ({
  collection: toCollection(model.collection),
  videos: model.videos.map(toCollectionVideo),
});

export const toTodayFeed = (model: TodayModel): TodayFeed => ({
  videos: model.videos.map(toVideo),
  nextReleaseAt: new Date(model.next_release_at),
});

export const toReleaseOutcome = (model: ReleaseModel): ReleaseOutcome => ({
  released: model.released,
  discarded: model.discarded,
});
