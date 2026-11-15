import { describe, expect, it } from 'vitest';
import {
  toChannel,
  toCollection,
  toCollectionDetail,
  toTodayFeed,
  toVideo,
} from '@data/library/mappers/library-mappers';
import {
  isCollectionDetailModel,
  isCollectionModel,
  isTodayModel,
  isVideoModel,
} from '@data/library/models/library-model-guards';
import type { VideoModel } from '@data/library/models/library-models';

const model = (overrides: Partial<VideoModel> = {}): VideoModel => ({
  id: 'v1',
  title: 'Antenne LoRa',
  channel_title: 'Officina Radio',
  published_at: '2026-11-04T15:52:59Z',
  duration_seconds: null,
  status: 'released',
  collection_ids: [],
  ...overrides,
});

describe('toVideo', () => {
  it('porta le date come date e le raccolte così come sono', () => {
    const video = toVideo(model({ collection_ids: [3, 7] }));
    expect(video.publishedAt.toISOString()).toBe('2026-11-04T15:52:59.000Z');
    expect(video.collectionIds).toEqual([3, 7]);
  });

  it("uno stato che l'app non conosce diventa archiviato", () => {
    // Anche quello di prima delle raccolte: un server vecchio può mandarlo ancora.
    expect(toVideo(model({ status: 'queued' })).status).toBe('archived');
  });
});

describe('le raccolte', () => {
  const collection = {
    id: 4,
    name: 'Orto',
    description: '',
    video_count: 1,
    cover_video_id: 'v1',
  };

  it('una raccolta vuota non ha copertina', () => {
    expect(toCollection({ ...collection, video_count: 0, cover_video_id: null })).toEqual({
      id: 4,
      name: 'Orto',
      description: '',
      videoCount: 0,
      coverVideoId: null,
    });
  });

  it('il dettaglio porta i video con la data di ingresso', () => {
    const detail = { collection, videos: [{ ...model(), added_at: '2026-11-10T08:00:00Z' }] };
    expect(isCollectionDetailModel(detail)).toBe(true);
    const mapped = toCollectionDetail(detail);
    expect(mapped.collection.coverVideoId).toBe('v1');
    expect(mapped.videos[0]?.addedAt.toISOString()).toBe('2026-11-10T08:00:00.000Z');
  });

  it.each([
    ['un id testuale', { ...collection, id: '4' }],
    ['una copertina numerica', { ...collection, cover_video_id: 1 }],
    ['senza conteggio', { ...collection, video_count: undefined }],
  ])('rifiutano %s', (_name, value) => {
    expect(isCollectionModel(value)).toBe(false);
  });

  it('rifiutano un video della raccolta senza data di ingresso', () => {
    expect(isCollectionDetailModel({ collection, videos: [model()] })).toBe(false);
  });
});

describe('toChannel', () => {
  it("un handle vuoto è un'assenza, e un errore di download è un segnale", () => {
    const base = { id: 'UCx', title: 'Officina Radio' };
    expect(toChannel({ ...base, handle: '', last_fetch_error: '' })).toEqual({
      ...base,
      handle: null,
      fetchFailed: false,
    });
    expect(
      toChannel({ ...base, handle: '@officina', last_fetch_error: 'YouTube non risponde' }),
    ).toEqual({ ...base, handle: '@officina', fetchFailed: true });
  });
});

describe('i controlli di forma', () => {
  it('accettano ciò che il server manda davvero', () => {
    const today = { videos: [model()], next_release_at: '2026-11-16T06:00:00Z' };
    expect(isTodayModel(today)).toBe(true);
    expect(toTodayFeed(today).nextReleaseAt.toISOString()).toBe('2026-11-16T06:00:00.000Z');
  });

  it.each([
    ['un campo mancante', { ...model(), title: undefined }],
    ['un tipo sbagliato', { ...model(), duration_seconds: '18:04' }],
    ['raccolte che non sono numeri', { ...model(), collection_ids: ['3'] }],
    ['null', null],
    ['una lista', [model()]],
  ])('rifiutano %s', (_name, value) => {
    expect(isVideoModel(value)).toBe(false);
  });

  it('rifiutano un feed senza il prossimo rilascio', () => {
    expect(isTodayModel({ videos: [model()] })).toBe(false);
  });
});
