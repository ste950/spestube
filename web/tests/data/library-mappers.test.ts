import { describe, expect, it } from 'vitest';
import { toChannel, toTodayFeed, toVideo } from '@data/library/mappers/library-mappers';
import { isTodayModel, isVideoModel } from '@data/library/models/library-model-guards';
import type { VideoModel } from '@data/library/models/library-models';

const model = (overrides: Partial<VideoModel> = {}): VideoModel => ({
  id: 'v1',
  title: 'Antenne LoRa',
  channel_title: 'Officina Radio',
  published_at: '2026-11-04T15:52:59Z',
  duration_seconds: null,
  status: 'released',
  queue_reason: '',
  queue_expires_at: null,
  queue_days_left: 0,
  ...overrides,
});

describe('toVideo', () => {
  it('un video in Oggi non ha una permanenza in coda', () => {
    const video = toVideo(model());
    expect(video.queue).toBeNull();
    expect(video.publishedAt.toISOString()).toBe('2026-11-04T15:52:59.000Z');
  });

  it('un video in coda porta motivo, scadenza e giorni rimasti', () => {
    const video = toVideo(
      model({
        status: 'queued',
        queue_reason: "per l'orto",
        queue_expires_at: '2026-11-29T09:00:00Z',
        queue_days_left: 11,
      }),
    );
    expect(video.queue).toEqual({
      reason: "per l'orto",
      expiresAt: new Date('2026-11-29T09:00:00Z'),
      daysLeft: 11,
    });
  });

  it("uno stato che l'app non conosce diventa archiviato", () => {
    expect(toVideo(model({ status: 'qualcosa di nuovo' })).status).toBe('archived');
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
    ['null', null],
    ['una lista', [model()]],
  ])('rifiutano %s', (_name, value) => {
    expect(isVideoModel(value)).toBe(false);
  });

  it('rifiutano un feed senza il prossimo rilascio', () => {
    expect(isTodayModel({ videos: [model()] })).toBe(false);
  });
});
