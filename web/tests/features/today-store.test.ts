import { describe, expect, it } from 'vitest';
import { networkFailure, rejectedFailure } from '@core/result/failure';
import { fail, ok } from '@core/result/result';
import type { Video } from '@domain/library/entities/video';
import type { VideoRepository } from '@domain/library/repositories/video-repository';
import type { SettingsRepository } from '@domain/settings/repositories/settings-repository';
import { createTodayStore } from '@features/today/state/today-store';
import { clockDuration, nextReleaseLabel } from '@features/today/state/today-labels';

const video: Video = {
  id: 'v1',
  title: 'Antenne LoRa',
  channelTitle: 'Officina Radio',
  publishedAt: new Date('2026-11-04T15:52:59Z'),
  durationSeconds: null,
  status: 'released',
  collectionIds: [],
};
const nextReleaseAt = new Date('2026-11-16T06:00:00Z');

const videos = (overrides: Partial<VideoRepository> = {}): VideoRepository => ({
  today: async () => ok({ videos: [video], nextReleaseAt }),
  getVideo: async () => ok(video),
  markWatched: async () => ok(video),
  releaseNow: async () => ok({ released: 3, discarded: 1 }),
  ...overrides,
});

const settings = (overrides: Partial<SettingsRepository> = {}): SettingsRepository => ({
  getInstallation: async () => ok({ version: 'test', dev: true, handleLookup: false }),
  ...overrides,
});

describe('lo store di Oggi', () => {
  it('carica i video e il prossimo rilascio', async () => {
    const store = createTodayStore(videos(), settings(), () => {});
    await store.getState().load();

    const state = store.getState();
    expect(state.videos).toEqual([video]);
    expect(state.nextReleaseAt).toEqual(nextReleaseAt);
    expect(state.devTools).toBe(true);
    expect(state.isLoading).toBe(false);
    expect(state.error).toBeNull();
  });

  it("senza notizie dell'installazione gli strumenti di sviluppo restano spenti", async () => {
    const store = createTodayStore(
      videos(),
      settings({
        getInstallation: async () => fail(networkFailure(null, null)),
      }),
      () => {},
    );
    await store.getState().load();

    // Il contorno che non arriva non nasconde il feed.
    expect(store.getState().videos).toEqual([video]);
    expect(store.getState().devTools).toBe(false);
  });

  it('un feed che non arriva diventa un messaggio, non una lista vuota', async () => {
    const store = createTodayStore(
      videos({ today: async () => fail(networkFailure(null, null)) }),
      settings(),
      () => {},
    );
    await store.getState().load();

    expect(store.getState().error).toEqual({ key: 'common_error_network' });
    expect(store.getState().nextReleaseAt).toBeNull();
  });

  it('il rilascio forzato riporta il conteggio e ricarica il feed', async () => {
    let loads = 0;
    const store = createTodayStore(
      videos({
        today: async () => {
          loads += 1;
          return ok({ videos: [video], nextReleaseAt });
        },
      }),
      settings(),
      () => {},
    );
    await store.getState().releaseNow();

    expect(loads).toBe(1);
    expect(store.getState().notice).toEqual({
      key: 'today_dev_released',
      params: { released: 3, discarded: 1 },
    });
    expect(store.getState().isReleasing).toBe(false);
  });

  it('un rilascio rifiutato mostra le parole del server', async () => {
    const store = createTodayStore(
      videos({ releaseNow: async () => fail(rejectedFailure('x', 'non adesso')) }),
      settings(),
      () => {},
    );
    await store.getState().releaseNow();

    expect(store.getState().error).toEqual({ text: 'non adesso' });
    expect(store.getState().notice).toBeNull();
  });
});

describe('il numero dei video annunciato alla barra', () => {
  it('è quello dei video in Oggi, a ogni caricamento riuscito', async () => {
    const announced: number[] = [];
    let list: readonly Video[] = [video, { ...video, id: 'v2' }];
    const store = createTodayStore(
      videos({ today: async () => ok({ videos: list, nextReleaseAt }) }),
      settings(),
      (count) => announced.push(count),
    );

    await store.getState().load();
    list = [];
    await store.getState().load();

    // Anche lo zero si annuncia: è ciò che fa sparire il numero quando Oggi si svuota.
    expect(announced).toEqual([2, 0]);
  });

  it('resta quello di prima se il feed non arriva', async () => {
    const announced: number[] = [];
    const store = createTodayStore(
      videos({ today: async () => fail(networkFailure(null, null)) }),
      settings(),
      (count) => announced.push(count),
    );
    await store.getState().load();

    // Un errore di rete non dice che i video sono zero.
    expect(announced).toEqual([]);
  });
});

describe('le raccolte di un video in Oggi', () => {
  it('cambiano senza ricaricare, e il video resta in Oggi', async () => {
    const other: Video = { ...video, id: 'v2' };
    const store = createTodayStore(
      videos({ today: async () => ok({ videos: [video, other], nextReleaseAt }) }),
      settings(),
      () => {},
    );
    await store.getState().load();

    store.getState().setCollections('v1', [4]);
    expect(store.getState().videos.map((v) => v.collectionIds)).toEqual([[4], []]);
    store.getState().setCollections('v1', []);
    expect(store.getState().videos[0]?.collectionIds).toEqual([]);
    // Un video che non è in Oggi non cambia niente.
    store.getState().setCollections('altrove', [1]);
    expect(store.getState().videos).toHaveLength(2);
  });
});

describe('le etichette di Oggi', () => {
  it('dice se il prossimo arrivo è oggi o domani', () => {
    const now = new Date(2026, 10, 15, 6, 0);
    expect(nextReleaseLabel(new Date(2026, 10, 15, 7, 0), now).day).toBe('today_end_day_today');
    expect(nextReleaseLabel(new Date(2026, 10, 16, 7, 0), now).day).toBe('today_end_day_tomorrow');
    expect(nextReleaseLabel(new Date(2026, 10, 16, 7, 0), now).time).toBe('7:00');
  });

  it('scrive la durata come un orologio', () => {
    expect([45, 18 * 60 + 4, 3600 + 5 * 60 + 9].map(clockDuration)).toEqual([
      '0:45',
      '18:04',
      '1:05:09',
    ]);
  });
});
