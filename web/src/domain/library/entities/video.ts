// Gli stati in cui un video si può vedere. Quelli nuovi o scartati non lasciano mai il server.
export type VideoStatus = 'released' | 'queued' | 'watched' | 'archived';

// La permanenza in Da vedere: esiste solo finché il video è in coda.
export type QueueEntry = {
  readonly reason: string;
  readonly expiresAt: Date;
  readonly daysLeft: number;
};

// Niente visualizzazioni né like: non fanno parte di ciò che SpesTube sa di un video.
export type Video = {
  readonly id: string;
  readonly title: string;
  readonly channelTitle: string;
  readonly publishedAt: Date;
  // `null` se sconosciuta: senza chiave API il server non la conosce.
  readonly durationSeconds: number | null;
  readonly status: VideoStatus;
  readonly queue: QueueEntry | null;
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
export const thumbnailUrl = (video: Video): string =>
  `https://i.ytimg.com/vi/${encodeURIComponent(video.id)}/mqdefault.jpg`;

// Arrotondata al minuto, e mai zero: un video di venti secondi dura «1 min», non «0 min».
export function durationMinutes(video: Video): number | null {
  if (video.durationSeconds === null) return null;
  return Math.max(1, Math.round(video.durationSeconds / 60));
}

// Il tempo in coda già consumato, fra 0 e 1, per la barra che lo mostra.
export function queueElapsed(entry: QueueEntry, lifetimeDays: number): number {
  return Math.min(1, Math.max(0, 1 - entry.daysLeft / lifetimeDays));
}

// Quanto resta un video in Da vedere. Lo decide il server; qui serve solo a disegnare la barra.
export const QUEUE_LIFETIME_DAYS = 14;
