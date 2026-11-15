import type { Result } from '@core/result/result';
import type { ChannelModel, ReleaseModel, TodayModel, VideoModel } from '../models/library-models';

export interface LibrarySource {
  fetchChannels(): Promise<Result<readonly ChannelModel[]>>;
  createChannel(reference: string): Promise<Result<ChannelModel>>;
  deleteChannel(channelId: string): Promise<Result<void>>;
  fetchToday(): Promise<Result<TodayModel>>;
  fetchQueue(): Promise<Result<readonly VideoModel[]>>;
  fetchVideo(videoId: string): Promise<Result<VideoModel>>;
  markWatched(videoId: string): Promise<Result<VideoModel>>;
  releaseNow(): Promise<Result<ReleaseModel>>;
}
