import { describe, expect, it } from 'vitest';
import { networkFailure, rejectedFailure } from '@core/result/failure';
import { fail, ok } from '@core/result/result';
import type { Collection } from '@domain/library/entities/collection';
import type { CollectionRepository } from '@domain/library/repositories/collection-repository';
import { createCollectionsStore } from '@features/collections/state/collections-store';

const collection = (id: number, name: string): Collection => ({
  id,
  name,
  description: '',
  videoCount: 0,
  coverVideoId: null,
});

const repository = (overrides: Partial<CollectionRepository> = {}): CollectionRepository => ({
  listCollections: async () =>
    ok([collection(1, 'Orto'), collection(2, 'Cucina'), collection(3, 'Go')]),
  getCollection: async () => fail(networkFailure(null, null)),
  createCollection: async (fields) => ok(collection(9, fields.name)),
  updateCollection: async () => fail(networkFailure(null, null)),
  deleteCollection: async () => ok(undefined),
  reorderCollections: async () => ok(undefined),
  addVideo: async () => ok(undefined),
  removeVideo: async () => ok(undefined),
  ...overrides,
});

const names = (store: ReturnType<typeof createCollectionsStore>) =>
  store.getState().collections.map((c) => c.name);

describe('lo store delle raccolte', () => {
  it('carica le raccolte nell’ordine del server', async () => {
    const store = createCollectionsStore(repository());
    await store.getState().load();
    expect(names(store)).toEqual(['Orto', 'Cucina', 'Go']);
    expect(store.getState().hasLoaded).toBe(true);
  });

  it('spostare una raccolta la sposta subito e manda l’ordine completo', async () => {
    const sent: (readonly number[])[] = [];
    const store = createCollectionsStore(
      repository({
        reorderCollections: async (ids) => {
          sent.push(ids);
          return ok(undefined);
        },
      }),
    );
    await store.getState().load();
    await store.getState().move(3, 'up');

    expect(names(store)).toEqual(['Orto', 'Go', 'Cucina']);
    expect(sent).toEqual([[1, 3, 2]]);
  });

  it('agli estremi non parte nessuna richiesta', async () => {
    let calls = 0;
    const store = createCollectionsStore(
      repository({
        reorderCollections: async () => {
          calls += 1;
          return ok(undefined);
        },
      }),
    );
    await store.getState().load();
    await store.getState().move(1, 'up');
    await store.getState().move(3, 'down');
    expect(calls).toBe(0);
  });

  it('un ordine rifiutato torna com’era e dice perché', async () => {
    const store = createCollectionsStore(
      repository({
        reorderCollections: async () =>
          fail(rejectedFailure('collection_order_invalid', 'le raccolte sono cambiate')),
      }),
    );
    await store.getState().load();
    await store.getState().move(2, 'down');

    expect(names(store)).toEqual(['Orto', 'Cucina', 'Go']);
    expect(store.getState().error).toEqual({ text: 'le raccolte sono cambiate' });
  });

  it('una raccolta nuova va in fondo e chiude il form', async () => {
    const store = createCollectionsStore(repository());
    await store.getState().load();
    store.getState().openCreate();
    await store.getState().create({ name: 'Viaggi', description: '' });

    expect(names(store)).toEqual(['Orto', 'Cucina', 'Go', 'Viaggi']);
    expect(store.getState().isCreating).toBe(false);
  });

  it('senza nome non parte nessuna richiesta e il form resta aperto', async () => {
    let calls = 0;
    const store = createCollectionsStore(
      repository({
        createCollection: async () => {
          calls += 1;
          return ok(collection(9, 'x'));
        },
      }),
    );
    store.getState().openCreate();
    await store.getState().create({ name: '   ', description: 'qualcosa' });

    expect(calls).toBe(0);
    expect(store.getState().formError).toEqual({ key: 'collections_form_name_required' });
    expect(store.getState().isCreating).toBe(true);
  });

  it('un nome già usato resta nel form con le parole del server', async () => {
    const store = createCollectionsStore(
      repository({
        createCollection: async () =>
          fail(rejectedFailure('collection_exists', 'hai già una raccolta con questo nome')),
      }),
    );
    store.getState().openCreate();
    await store.getState().create({ name: 'Orto', description: '' });

    expect(store.getState().formError).toEqual({ text: 'hai già una raccolta con questo nome' });
    expect(store.getState().isCreating).toBe(true);
    expect(store.getState().isSaving).toBe(false);
  });
});
