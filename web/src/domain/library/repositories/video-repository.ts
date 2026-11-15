import type { Result } from '@core/result/result';
import type { ReleaseOutcome, TodayFeed, Video } from '../entities/video';

export interface VideoRepository {
  today(): Promise<Result<TodayFeed>>;
  queue(): Promise<Result<readonly Video[]>>;
  getVideo(videoId: string): Promise<Result<Video>>;
  markWatched(videoId: string): Promise<Result<Video>>;
  // Fa arrivare subito i video in attesa. Esiste solo su un'installazione di sviluppo.
  releaseNow(): Promise<Result<ReleaseOutcome>>;
}
