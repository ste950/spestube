// Le forme che l'API restituisce davvero: date come testo, stato come stringa qualunque.
// Interpretarle è mestiere dei mapper.
export type ChannelModel = {
  readonly id: string;
  readonly title: string;
  readonly handle: string;
  readonly last_fetch_error: string;
};

export type VideoModel = {
  readonly id: string;
  readonly title: string;
  readonly channel_title: string;
  readonly published_at: string;
  readonly duration_seconds: number | null;
  readonly status: string;
  readonly collection_ids: readonly number[];
};

export type CollectionModel = {
  readonly id: number;
  readonly name: string;
  readonly description: string;
  readonly video_count: number;
  readonly cover_video_id: string | null;
};

export type CollectionVideoModel = VideoModel & {
  readonly added_at: string;
};

export type CollectionDetailModel = {
  readonly collection: CollectionModel;
  readonly videos: readonly CollectionVideoModel[];
};

export type TodayModel = {
  readonly videos: readonly VideoModel[];
  readonly next_release_at: string;
};

export type ReleaseModel = {
  readonly released: number;
  readonly discarded: number;
};
