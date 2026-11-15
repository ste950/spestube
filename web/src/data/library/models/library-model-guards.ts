import type {
  ChannelModel,
  CollectionDetailModel,
  CollectionModel,
  CollectionVideoModel,
  ReleaseModel,
  TodayModel,
  VideoModel,
} from './library-models';

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null;

// Codice e non una `as`: una `as` è un'affermazione che nessuno verifica, e farebbe arrivare
// `undefined` dentro un'entità lontano dal confine che avrebbe dovuto fermarlo.
export const isChannelModel = (value: unknown): value is ChannelModel =>
  isRecord(value) &&
  typeof value.id === 'string' &&
  typeof value.title === 'string' &&
  typeof value.handle === 'string' &&
  typeof value.last_fetch_error === 'string';

export const isChannelList = (value: unknown): value is { channels: ChannelModel[] } =>
  isRecord(value) && Array.isArray(value.channels) && value.channels.every(isChannelModel);

export const isVideoModel = (value: unknown): value is VideoModel =>
  isRecord(value) &&
  typeof value.id === 'string' &&
  typeof value.title === 'string' &&
  typeof value.channel_title === 'string' &&
  typeof value.published_at === 'string' &&
  (value.duration_seconds === null || typeof value.duration_seconds === 'number') &&
  typeof value.status === 'string' &&
  Array.isArray(value.collection_ids) &&
  value.collection_ids.every((id) => typeof id === 'number');

export const isVideoList = (value: unknown): value is { videos: VideoModel[] } =>
  isRecord(value) && Array.isArray(value.videos) && value.videos.every(isVideoModel);

export const isTodayModel = (value: unknown): value is TodayModel =>
  isVideoList(value) && typeof (value as Record<string, unknown>).next_release_at === 'string';

export const isReleaseModel = (value: unknown): value is ReleaseModel =>
  isRecord(value) && typeof value.released === 'number' && typeof value.discarded === 'number';

export const isCollectionModel = (value: unknown): value is CollectionModel =>
  isRecord(value) &&
  typeof value.id === 'number' &&
  typeof value.name === 'string' &&
  typeof value.description === 'string' &&
  typeof value.video_count === 'number' &&
  (value.cover_video_id === null || typeof value.cover_video_id === 'string');

export const isCollectionList = (value: unknown): value is { collections: CollectionModel[] } =>
  isRecord(value) && Array.isArray(value.collections) && value.collections.every(isCollectionModel);

const isCollectionVideoModel = (value: unknown): value is CollectionVideoModel =>
  isVideoModel(value) && typeof (value as Record<string, unknown>).added_at === 'string';

export const isCollectionDetailModel = (value: unknown): value is CollectionDetailModel =>
  isRecord(value) &&
  isCollectionModel(value.collection) &&
  Array.isArray(value.videos) &&
  value.videos.every(isCollectionVideoModel);
