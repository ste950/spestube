import type { Result } from '@core/result/result';
import type {
  ChannelModel,
  CollectionDetailModel,
  CollectionModel,
  ReleaseModel,
  TodayModel,
  VideoModel,
} from '../models/library-models';

export type CollectionBody = { readonly name: string; readonly description: string };

export interface LibrarySource {
  fetchChannels(): Promise<Result<readonly ChannelModel[]>>;
  createChannel(reference: string): Promise<Result<ChannelModel>>;
  deleteChannel(channelId: string): Promise<Result<void>>;
  fetchToday(): Promise<Result<TodayModel>>;
  fetchVideo(videoId: string): Promise<Result<VideoModel>>;
  markWatched(videoId: string): Promise<Result<VideoModel>>;
  releaseNow(): Promise<Result<ReleaseModel>>;
  fetchCollections(): Promise<Result<readonly CollectionModel[]>>;
  fetchCollection(collectionId: number): Promise<Result<CollectionDetailModel>>;
  createCollection(body: CollectionBody): Promise<Result<CollectionModel>>;
  updateCollection(collectionId: number, body: CollectionBody): Promise<Result<CollectionModel>>;
  deleteCollection(collectionId: number): Promise<Result<void>>;
  reorderCollections(collectionIds: readonly number[]): Promise<Result<void>>;
  addToCollection(collectionId: number, videoId: string): Promise<Result<void>>;
  removeFromCollection(collectionId: number, videoId: string): Promise<Result<void>>;
}
