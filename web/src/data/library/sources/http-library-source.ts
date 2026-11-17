import { createJsonApi } from '@core/http/json-api';
import { parseFailure } from '@core/result/failure';
import { fail, mapOk, ok, type Result } from '@core/result/result';
import {
  isChannelList,
  isChannelModel,
  isCollectionDetailModel,
  isCollectionList,
  isCollectionModel,
  isReleaseModel,
  isTodayModel,
  isVideoModel,
} from '../models/library-model-guards';
import type { LibrarySource } from './library-source';

// Da `unknown` a un modello: il controllo di forma è l'unico modo di passare.
function shaped<T>(
  raw: Result<unknown>,
  guard: (value: unknown) => value is T,
  at: string,
): Result<T> {
  if (!raw.ok) return raw;
  return guard(raw.value) ? ok(raw.value) : fail(parseFailure(at, raw.value));
}

const nothing = (raw: Result<unknown>): Result<void> => mapOk(raw, () => undefined);

export function createHttpLibrarySource(baseUrl: string): LibrarySource {
  // `fetch` si nomina qui e in nessun altro livello: è la funzione che il client JSON
  // riceve, avvolta perché non perda il suo `this`.
  const api = createJsonApi((input, init) => fetch(input, init), baseUrl);
  const video = (id: string) => `/videos/${encodeURIComponent(id)}`;
  const collection = (id: number) => `/collections/${encodeURIComponent(String(id))}`;
  const item = (collectionId: number, videoId: string) =>
    `${collection(collectionId)}/videos/${encodeURIComponent(videoId)}`;

  return {
    async fetchChannels() {
      const list = shaped(await api.request('GET', '/channels'), isChannelList, 'ChannelModel[]');
      return mapOk(list, (body) => body.channels);
    },
    async createChannel(reference) {
      const raw = await api.request('POST', '/channels', { channel: reference });
      return shaped(raw, isChannelModel, 'ChannelModel');
    },
    async deleteChannel(channelId) {
      return nothing(await api.request('DELETE', `/channels/${encodeURIComponent(channelId)}`));
    },
    async fetchToday() {
      return shaped(await api.request('GET', '/feed/today'), isTodayModel, 'TodayModel');
    },
    async fetchVideo(videoId) {
      return shaped(await api.request('GET', video(videoId)), isVideoModel, 'VideoModel');
    },
    async markWatched(videoId) {
      const raw = await api.request('POST', `${video(videoId)}/watched`);
      return shaped(raw, isVideoModel, 'VideoModel');
    },
    async releaseNow() {
      return shaped(await api.request('POST', '/dev/release'), isReleaseModel, 'ReleaseModel');
    },
    async fetchCollections() {
      const raw = await api.request('GET', '/collections');
      return mapOk(shaped(raw, isCollectionList, 'CollectionModel[]'), (body) => body.collections);
    },
    async fetchCollection(collectionId) {
      const raw = await api.request('GET', collection(collectionId));
      return shaped(raw, isCollectionDetailModel, 'CollectionDetailModel');
    },
    async createCollection(body) {
      const raw = await api.request('POST', '/collections', body);
      return shaped(raw, isCollectionModel, 'CollectionModel');
    },
    async updateCollection(collectionId, body) {
      const raw = await api.request('PUT', collection(collectionId), body);
      return shaped(raw, isCollectionModel, 'CollectionModel');
    },
    async deleteCollection(collectionId) {
      return nothing(await api.request('DELETE', collection(collectionId)));
    },
    async reorderCollections(collectionIds) {
      return nothing(await api.request('PUT', '/collections/order', { ids: collectionIds }));
    },
    async addToCollection(collectionId, videoId) {
      return nothing(await api.request('PUT', item(collectionId, videoId)));
    },
    async removeFromCollection(collectionId, videoId) {
      return nothing(await api.request('DELETE', item(collectionId, videoId)));
    },
  };
}
