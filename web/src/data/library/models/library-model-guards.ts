import type { ChannelModel, ReleaseModel, TodayModel, VideoModel } from './library-models';

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
  typeof value.queue_reason === 'string' &&
  (value.queue_expires_at === null || typeof value.queue_expires_at === 'string') &&
  typeof value.queue_days_left === 'number';

export const isVideoList = (value: unknown): value is { videos: VideoModel[] } =>
  isRecord(value) && Array.isArray(value.videos) && value.videos.every(isVideoModel);

export const isTodayModel = (value: unknown): value is TodayModel =>
  isVideoList(value) && typeof (value as Record<string, unknown>).next_release_at === 'string';

export const isReleaseModel = (value: unknown): value is ReleaseModel =>
  isRecord(value) && typeof value.released === 'number' && typeof value.discarded === 'number';
