import type { Video } from './video';

// Una raccolta di video scelta dall'utente, come le playlist di YouTube: un nome, una
// descrizione facoltativa, un posto nell'ordine. Non scade.
export type Collection = {
  readonly id: number;
  readonly name: string;
  readonly description: string;
  readonly videoCount: number;
  // Il video aggiunto per ultimo, la cui miniatura fa da copertina. `null` se è vuota.
  readonly coverVideoId: string | null;
};

export type CollectionVideo = Video & {
  readonly addedAt: Date;
};

export type CollectionDetail = {
  readonly collection: Collection;
  readonly videos: readonly CollectionVideo[];
};

// Gli stessi limiti del server: superarli qui è un errore che si può dire prima di inviare.
export const COLLECTION_NAME_MAX = 80;
export const COLLECTION_DESCRIPTION_MAX = 500;
