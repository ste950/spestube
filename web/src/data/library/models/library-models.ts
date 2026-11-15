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
  readonly queue_reason: string;
  readonly queue_expires_at: string | null;
  readonly queue_days_left: number;
};

export type TodayModel = {
  readonly videos: readonly VideoModel[];
  readonly next_release_at: string;
};

export type ReleaseModel = {
  readonly released: number;
  readonly discarded: number;
};
