import type { Result } from '@core/result/result';
import type { Collection, CollectionDetail } from '../entities/collection';

export type CollectionFields = {
  readonly name: string;
  readonly description: string;
};

export interface CollectionRepository {
  // Nell'ordine scelto dall'utente.
  listCollections(): Promise<Result<readonly Collection[]>>;
  getCollection(collectionId: number): Promise<Result<CollectionDetail>>;
  createCollection(fields: CollectionFields): Promise<Result<Collection>>;
  updateCollection(collectionId: number, fields: CollectionFields): Promise<Result<Collection>>;
  deleteCollection(collectionId: number): Promise<Result<void>>;
  // L'ordine completo, dalla prima all'ultima.
  reorderCollections(collectionIds: readonly number[]): Promise<Result<void>>;
  // Aggiungere un video che c'è già, o togliere uno che non c'è, non è un errore.
  addVideo(collectionId: number, videoId: string): Promise<Result<void>>;
  removeVideo(collectionId: number, videoId: string): Promise<Result<void>>;
}
