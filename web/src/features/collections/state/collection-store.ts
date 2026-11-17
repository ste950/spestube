import { create, type StoreApi, type UseBoundStore } from 'zustand';
import type { CollectionDetail } from '@domain/library/entities/collection';
import type {
  CollectionFields,
  CollectionRepository,
} from '@domain/library/repositories/collection-repository';
import { describeFailure, type Message } from '@shared/i18n/message';

export type CollectionState = {
  readonly requestedId: number | null;
  readonly detail: CollectionDetail | null;
  readonly isLoading: boolean;
  readonly error: Message | null;
  readonly isEditing: boolean;
  readonly isSaving: boolean;
  readonly formError: Message | null;
  readonly load: (collectionId: number) => Promise<void>;
  readonly openEdit: () => void;
  readonly closeEdit: () => void;
  readonly update: (fields: CollectionFields) => Promise<void>;
  // Risponde `true` se la raccolta non c'è più: chi chiama sa che può andarsene.
  readonly remove: () => Promise<boolean>;
  readonly removeVideo: (videoId: string) => Promise<void>;
};

export type CollectionStore = UseBoundStore<StoreApi<CollectionState>>;

export function createCollectionStore(repository: CollectionRepository): CollectionStore {
  return create<CollectionState>((set, get) => ({
    requestedId: null,
    detail: null,
    isLoading: false,
    error: null,
    isEditing: false,
    isSaving: false,
    formError: null,

    load: async (collectionId) => {
      set({
        requestedId: collectionId,
        detail: null,
        isLoading: true,
        error: null,
        isEditing: false,
      });
      if (Number.isNaN(collectionId)) {
        set({ isLoading: false, error: { key: 'collection_not_found' } });
        return;
      }
      const result = await repository.getCollection(collectionId);
      // Una risposta lenta per una raccolta da cui si è già usciti non sovrascrive quella
      // aperta adesso.
      if (get().requestedId !== collectionId) return;
      if (!result.ok) {
        const notFound =
          result.failure.kind === 'rejected' && result.failure.code === 'collection_not_found';
        set({
          isLoading: false,
          error: notFound ? { key: 'collection_not_found' } : describeFailure(result.failure),
        });
        return;
      }
      set({ isLoading: false, detail: result.value });
    },

    openEdit: () => set({ isEditing: true, formError: null }),

    closeEdit: () => set({ isEditing: false, formError: null }),

    update: async (fields) => {
      const detail = get().detail;
      if (detail === null) return;
      if (fields.name.trim() === '') {
        set({ formError: { key: 'collections_form_name_required' } });
        return;
      }
      set({ isSaving: true, formError: null });
      const saved = await repository.updateCollection(detail.collection.id, fields);
      if (!saved.ok) {
        set({ isSaving: false, formError: describeFailure(saved.failure) });
        return;
      }
      set({ isSaving: false, isEditing: false, detail: { ...detail, collection: saved.value } });
    },

    remove: async () => {
      const detail = get().detail;
      if (detail === null) return false;
      const deleted = await repository.deleteCollection(detail.collection.id);
      if (!deleted.ok) {
        set({ error: describeFailure(deleted.failure) });
        return false;
      }
      return true;
    },

    removeVideo: async (videoId) => {
      const before = get().detail;
      if (before === null) return;
      // Il video sparisce subito dalla lista, e ci torna se il server rifiuta.
      const videos = before.videos.filter((video) => video.id !== videoId);
      const cover = videos[0]?.id ?? null;
      set({
        error: null,
        detail: {
          collection: { ...before.collection, videoCount: videos.length, coverVideoId: cover },
          videos,
        },
      });
      const removed = await repository.removeVideo(before.collection.id, videoId);
      if (!removed.ok) set({ detail: before, error: describeFailure(removed.failure) });
    },
  }));
}
