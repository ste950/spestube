import { createJsonApi } from '@core/http/json-api';
import { parseFailure } from '@core/result/failure';
import { fail, mapOk, ok, type Result } from '@core/result/result';
import {
  isChannelList,
  isChannelModel,
  isReleaseModel,
  isTodayModel,
  isVideoList,
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

export function createHttpLibrarySource(baseUrl: string): LibrarySource {
  // `fetch` si nomina qui e in nessun altro livello: è la funzione che il client JSON
  // riceve, avvolta perché non perda il suo `this`.
  const api = createJsonApi((input, init) => fetch(input, init), baseUrl);
  const video = (id: string) => `/videos/${encodeURIComponent(id)}`;

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
      const raw = await api.request('DELETE', `/channels/${encodeURIComponent(channelId)}`);
      return mapOk(raw, () => undefined);
    },
    async fetchToday() {
      return shaped(await api.request('GET', '/feed/today'), isTodayModel, 'TodayModel');
    },
    async fetchQueue() {
      const list = shaped(await api.request('GET', '/queue'), isVideoList, 'VideoModel[]');
      return mapOk(list, (body) => body.videos);
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
  };
}
