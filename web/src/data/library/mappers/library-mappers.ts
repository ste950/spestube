import type { Channel } from '@domain/library/entities/channel';
import type { ReleaseOutcome, TodayFeed, Video, VideoStatus } from '@domain/library/entities/video';
import type { ChannelModel, ReleaseModel, TodayModel, VideoModel } from '../models/library-models';

export const toChannel = (model: ChannelModel): Channel => ({
  id: model.id,
  title: model.title,
  handle: model.handle === '' ? null : model.handle,
  fetchFailed: model.last_fetch_error !== '',
});

const KNOWN_STATUSES: readonly VideoStatus[] = ['released', 'queued', 'watched', 'archived'];

// Uno stato che l'app non conosce diventa «archiviato»: il video resta apribile, ma non
// compare né in Oggi né in Da vedere, dove un video di troppo sarebbe un errore visibile.
const toStatus = (raw: string): VideoStatus =>
  KNOWN_STATUSES.find((status) => status === raw) ?? 'archived';

export function toVideo(model: VideoModel): Video {
  const status = toStatus(model.status);
  return {
    id: model.id,
    title: model.title,
    channelTitle: model.channel_title,
    publishedAt: new Date(model.published_at),
    durationSeconds: model.duration_seconds,
    status,
    queue:
      status === 'queued' && model.queue_expires_at !== null
        ? {
            reason: model.queue_reason,
            expiresAt: new Date(model.queue_expires_at),
            daysLeft: model.queue_days_left,
          }
        : null,
  };
}

export const toTodayFeed = (model: TodayModel): TodayFeed => ({
  videos: model.videos.map(toVideo),
  nextReleaseAt: new Date(model.next_release_at),
});

export const toReleaseOutcome = (model: ReleaseModel): ReleaseOutcome => ({
  released: model.released,
  discarded: model.discarded,
});
