import { mapOk } from '@core/result/result';
import type { ChannelRepository } from '@domain/library/repositories/channel-repository';
import type { CollectionRepository } from '@domain/library/repositories/collection-repository';
import type { VideoRepository } from '@domain/library/repositories/video-repository';
import {
  toChannel,
  toCollection,
  toCollectionDetail,
  toReleaseOutcome,
  toTodayFeed,
  toVideo,
} from '../mappers/library-mappers';
import type { LibrarySource } from '../sources/library-source';

// Niente copia locale: ciò che Oggi mostra lo decide il server a un'ora precisa, e una
// lista di ieri riproposta come attuale è l'errore che questo prodotto esiste per evitare.
export function createRemoteChannelRepository(source: LibrarySource): ChannelRepository {
  return {
    async listChannels() {
      return mapOk(await source.fetchChannels(), (channels) => channels.map(toChannel));
    },
    async addChannel(reference) {
      return mapOk(await source.createChannel(reference), toChannel);
    },
    removeChannel: (channelId) => source.deleteChannel(channelId),
  };
}

export function createRemoteVideoRepository(source: LibrarySource): VideoRepository {
  return {
    async today() {
      return mapOk(await source.fetchToday(), toTodayFeed);
    },
    async getVideo(videoId) {
      return mapOk(await source.fetchVideo(videoId), toVideo);
    },
    async markWatched(videoId) {
      return mapOk(await source.markWatched(videoId), toVideo);
    },
    async releaseNow() {
      return mapOk(await source.releaseNow(), toReleaseOutcome);
    },
  };
}

export function createRemoteCollectionRepository(source: LibrarySource): CollectionRepository {
  return {
    async listCollections() {
      return mapOk(await source.fetchCollections(), (list) => list.map(toCollection));
    },
    async getCollection(collectionId) {
      return mapOk(await source.fetchCollection(collectionId), toCollectionDetail);
    },
    async createCollection(fields) {
      return mapOk(await source.createCollection(fields), toCollection);
    },
    async updateCollection(collectionId, fields) {
      return mapOk(await source.updateCollection(collectionId, fields), toCollection);
    },
    deleteCollection: (collectionId) => source.deleteCollection(collectionId),
    reorderCollections: (collectionIds) => source.reorderCollections(collectionIds),
    addVideo: (collectionId, videoId) => source.addToCollection(collectionId, videoId),
    removeVideo: (collectionId, videoId) => source.removeFromCollection(collectionId, videoId),
  };
}
