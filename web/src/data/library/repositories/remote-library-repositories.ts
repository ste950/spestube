import { mapOk } from '@core/result/result';
import type { ChannelRepository } from '@domain/library/repositories/channel-repository';
import type { VideoRepository } from '@domain/library/repositories/video-repository';
import { toChannel, toReleaseOutcome, toTodayFeed, toVideo } from '../mappers/library-mappers';
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
    async queue() {
      return mapOk(await source.fetchQueue(), (videos) => videos.map(toVideo));
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
