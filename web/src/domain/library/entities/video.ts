// Gli stati in cui un video si può vedere. Quelli nuovi o scartati non lasciano mai il server.
export type VideoStatus = 'released' | 'watched' | 'archived';

// Niente visualizzazioni né like: non fanno parte di ciò che SpesTube sa di un video.
export type Video = {
  readonly id: string;
  readonly title: string;
  readonly channelTitle: string;
  readonly publishedAt: Date;
  // `null` se sconosciuta: senza chiave API il server non la conosce.
  readonly durationSeconds: number | null;
  readonly status: VideoStatus;
  // Le raccolte che contengono il video. Vuota se non è salvato da nessuna parte.
  readonly collectionIds: readonly number[];
};

export type TodayFeed = {
  readonly videos: readonly Video[];
  readonly nextReleaseAt: Date;
};

export type ReleaseOutcome = {
  readonly released: number;
  readonly discarded: number;
};

// La miniatura a 16:9 che YouTube pubblica per ogni video. Si deriva dall'id: il server non
// la conserva e non fa da tramite.
export const thumbnailUrl = (videoId: string): string =>
  `https://i.ytimg.com/vi/${encodeURIComponent(videoId)}/mqdefault.jpg`;

// Arrotondata al minuto, e mai zero: un video di venti secondi dura «1 min», non «0 min».
export function durationMinutes(video: Video): number | null {
  if (video.durationSeconds === null) return null;
  return Math.max(1, Math.round(video.durationSeconds / 60));
}

export const isSaved = (video: Video): boolean => video.collectionIds.length > 0;
