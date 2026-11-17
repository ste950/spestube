import { describe, expect, it } from 'vitest';
import { notFoundFailure } from '@core/result/failure';
import { fail, ok, type Result } from '@core/result/result';
import type { Video } from '@domain/library/entities/video';
import type { VideoRepository } from '@domain/library/repositories/video-repository';
import { createWatchStore } from '@features/watch/state/watch-store';

const video: Video = {
  id: 'v1',
  title: 'Antenne LoRa',
  channelTitle: 'Officina Radio',
  publishedAt: new Date('2026-11-04T15:52:59Z'),
  durationSeconds: null,
  status: 'released',
  collectionIds: [],
};

const repository = (overrides: Partial<VideoRepository> = {}): VideoRepository => ({
  today: async () => ok({ videos: [], nextReleaseAt: new Date() }),
  getVideo: async () => ok(video),
  markWatched: async () => ok({ ...video, status: 'watched' }),
  releaseNow: async () => ok({ released: 0, discarded: 0 }),
  ...overrides,
});

describe('lo store di Guarda', () => {
  it('carica il video e riparte da capo a ogni video', async () => {
    const store = createWatchStore(repository());
    await store.getState().load('v1');
    store.getState().end();
    expect(store.getState().hasEnded).toBe(true);

    await store.getState().load('v1');
    // La schermata di stop non sopravvive al video successivo.
    expect(store.getState().hasEnded).toBe(false);
    expect(store.getState().video).toEqual(video);
  });

  it('un video che non esiste diventa un messaggio', async () => {
    const store = createWatchStore(repository({ getVideo: async () => fail(notFoundFailure()) }));
    await store.getState().load('manca');

    expect(store.getState().video).toBeNull();
    expect(store.getState().error).toEqual({ key: 'common_error_not_found' });
  });

  it('una risposta lenta non sovrascrive il video aperto dopo', async () => {
    let releaseSlow: (result: Result<Video>) => void = () => {};
    const slow = new Promise<Result<Video>>((resolve) => {
      releaseSlow = resolve;
    });
    const fast: Video = { ...video, id: 'v2', title: 'Il secondo' };
    const store = createWatchStore(
      repository({ getVideo: (id) => (id === 'v1' ? slow : Promise.resolve(ok(fast))) }),
    );

    const first = store.getState().load('v1');
    await store.getState().load('v2');
    releaseSlow(ok(video));
    await first;

    expect(store.getState().video).toEqual(fast);
  });

  it('segna il video come visto', async () => {
    const store = createWatchStore(repository());
    await store.getState().load('v1');
    await store.getState().markWatched();

    expect(store.getState().video?.status).toBe('watched');
    expect(store.getState().isMarking).toBe(false);
  });
});

describe('le raccolte del video aperto', () => {
  it('cambiano solo per il video aperto', async () => {
    const store = createWatchStore(repository());
    await store.getState().load('v1');

    store.getState().setCollections('altro', [9]);
    expect(store.getState().video?.collectionIds).toEqual([]);
    store.getState().setCollections('v1', [2, 9]);
    expect(store.getState().video?.collectionIds).toEqual([2, 9]);
  });
});
