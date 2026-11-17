import { create, type StoreApi, type UseBoundStore } from 'zustand';
import type { Collection } from '@domain/library/entities/collection';
import type {
  CollectionFields,
  CollectionRepository,
} from '@domain/library/repositories/collection-repository';
import { moveCollection } from '@domain/library/use-cases/move-collection';
import { describeFailure, type Message } from '@shared/i18n/message';

export type CollectionsState = {
  readonly collections: readonly Collection[];
  readonly isLoading: boolean;
  readonly hasLoaded: boolean;
  readonly error: Message | null;
  // In modalità «ordina» ogni riga mostra le frecce per spostarla.
  readonly isReordering: boolean;
  readonly isCreating: boolean;
  readonly isSaving: boolean;
  readonly formError: Message | null;
  readonly load: () => Promise<void>;
  readonly toggleReorder: () => void;
  readonly move: (collectionId: number, direction: 'up' | 'down') => Promise<void>;
  readonly openCreate: () => void;
  readonly closeCreate: () => void;
  readonly create: (fields: CollectionFields) => Promise<void>;
};

export type CollectionsStore = UseBoundStore<StoreApi<CollectionsState>>;

export function createCollectionsStore(repository: CollectionRepository): CollectionsStore {
  return create<CollectionsState>((set, get) => ({
    collections: [],
    isLoading: false,
    hasLoaded: false,
    error: null,
    isReordering: false,
    isCreating: false,
    isSaving: false,
    formError: null,

    load: async () => {
      set({ isLoading: true, error: null });
      const result = await repository.listCollections();
      if (!result.ok) {
        set({ isLoading: false, error: describeFailure(result.failure) });
        return;
      }
      set({ isLoading: false, hasLoaded: true, collections: result.value });
    },

    toggleReorder: () => set({ isReordering: !get().isReordering }),

    move: async (collectionId, direction) => {
      const before = get().collections;
      const order = moveCollection({ order: before.map((c) => c.id), id: collectionId, direction });
      if (order.every((id, i) => id === before[i]?.id)) return;
      // La lista si sposta subito, e torna com'era se il server rifiuta: aspettare la
      // risposta per vedere la riga muoversi farebbe sembrare il tocco andato a vuoto.
      const byId = new Map(before.map((c) => [c.id, c]));
      const after = order.flatMap((id) => byId.get(id) ?? []);
      set({ collections: after, error: null });
      const saved = await repository.reorderCollections(order);
      if (!saved.ok) set({ collections: before, error: describeFailure(saved.failure) });
    },

    openCreate: () => set({ isCreating: true, formError: null }),

    closeCreate: () => set({ isCreating: false, formError: null }),

    create: async (fields) => {
      if (fields.name.trim() === '') {
        set({ formError: { key: 'collections_form_name_required' } });
        return;
      }
      set({ isSaving: true, formError: null });
      const created = await repository.createCollection(fields);
      if (!created.ok) {
        set({ isSaving: false, formError: describeFailure(created.failure) });
        return;
      }
      set({
        isSaving: false,
        isCreating: false,
        collections: [...get().collections, created.value],
      });
    },
  }));
}
