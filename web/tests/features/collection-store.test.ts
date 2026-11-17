import { describe, expect, it } from 'vitest';
import { networkFailure, rejectedFailure } from '@core/result/failure';
import { fail, ok } from '@core/result/result';
import type { CollectionDetail, CollectionVideo } from '@domain/library/entities/collection';
import type { CollectionRepository } from '@domain/library/repositories/collection-repository';
import { createCollectionStore } from '@features/collections/state/collection-store';

const video = (id: string): CollectionVideo => ({
  id,
  title: `Video ${id}`,
  channelTitle: 'Officina Radio',
  publishedAt: new Date('2026-11-04T15:52:59Z'),
  durationSeconds: null,
  status: 'archived',
  collectionIds: [4],
  addedAt: new Date('2026-11-10T08:00:00Z'),
});

const detail: CollectionDetail = {
  collection: { id: 4, name: 'Orto', description: 'idee', videoCount: 2, coverVideoId: 'b' },
  videos: [video('b'), video('a')],
};

const repository = (overrides: Partial<CollectionRepository> = {}): CollectionRepository => ({
  listCollections: async () => ok([]),
  getCollection: async () => ok(detail),
  createCollection: async () => fail(networkFailure(null, null)),
  updateCollection: async (_id, fields) => ok({ ...detail.collection, ...fields }),
  deleteCollection: async () => ok(undefined),
  reorderCollections: async () => ok(undefined),
  addVideo: async () => ok(undefined),
  removeVideo: async () => ok(undefined),
  ...overrides,
});

describe('lo store di una raccolta', () => {
  it('carica la raccolta con i suoi video', async () => {
    const store = createCollectionStore(repository());
    await store.getState().load(4);
    expect(store.getState().detail).toEqual(detail);
  });

  it('una raccolta che non esiste ha il suo messaggio', async () => {
    const store = createCollectionStore(
      repository({
        getCollection: async () =>
          fail(rejectedFailure('collection_not_found', 'raccolta inesistente')),
      }),
    );
    await store.getState().load(4);
    expect(store.getState().error).toEqual({ key: 'collection_not_found' });

    // Un id che non è un numero non arriva nemmeno al server.
    await store.getState().load(Number.NaN);
    expect(store.getState().error).toEqual({ key: 'collection_not_found' });
  });

  it('togliere un video lo toglie subito e aggiorna conteggio e copertina', async () => {
    const removed: string[] = [];
    const store = createCollectionStore(
      repository({
        removeVideo: async (_id, videoId) => {
          removed.push(videoId);
          return ok(undefined);
        },
      }),
    );
    await store.getState().load(4);
    await store.getState().removeVideo('b');

    expect(removed).toEqual(['b']);
    expect(store.getState().detail?.videos.map((v) => v.id)).toEqual(['a']);
    expect(store.getState().detail?.collection).toMatchObject({ videoCount: 1, coverVideoId: 'a' });
  });

  it('un video che il server non toglie torna al suo posto', async () => {
    const store = createCollectionStore(
      repository({ removeVideo: async () => fail(networkFailure(null, null)) }),
    );
    await store.getState().load(4);
    await store.getState().removeVideo('b');

    expect(store.getState().detail).toEqual(detail);
    expect(store.getState().error).toEqual({ key: 'common_error_network' });
  });

  it('la modifica salva nome e descrizione e chiude il form', async () => {
    const store = createCollectionStore(repository());
    await store.getState().load(4);
    store.getState().openEdit();
    await store.getState().update({ name: 'Orto e giardino', description: '' });

    expect(store.getState().detail?.collection.name).toBe('Orto e giardino');
    expect(store.getState().detail?.videos).toHaveLength(2);
    expect(store.getState().isEditing).toBe(false);
  });

  it('eliminare risponde se la raccolta non c’è più', async () => {
    const ok_ = createCollectionStore(repository());
    await ok_.getState().load(4);
    expect(await ok_.getState().remove()).toBe(true);

    const refused = createCollectionStore(
      repository({ deleteCollection: async () => fail(networkFailure(null, null)) }),
    );
    await refused.getState().load(4);
    expect(await refused.getState().remove()).toBe(false);
    expect(refused.getState().error).toEqual({ key: 'common_error_network' });
  });
});
