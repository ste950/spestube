import { describe, expect, it } from 'vitest';
import {
  durationMinutes,
  QUEUE_LIFETIME_DAYS,
  queueElapsed,
  thumbnailUrl,
  type Video,
} from '@domain/library/entities/video';

const video = (overrides: Partial<Video> = {}): Video => ({
  id: 'KfODG5REyss',
  title: 'Un video',
  channelTitle: 'Officina Radio',
  publishedAt: new Date('2026-11-04T15:52:59Z'),
  durationSeconds: null,
  status: 'released',
  queue: null,
  ...overrides,
});

describe('durationMinutes', () => {
  it.each([
    ['sconosciuta', null, null],
    ['venti secondi non sono zero minuti', 20, 1],
    ['arrotonda al minuto', 18 * 60 + 4, 18],
    ['arrotonda per eccesso dalla metà', 18 * 60 + 30, 19],
  ])('%s', (_name, seconds, expected) => {
    expect(durationMinutes(video({ durationSeconds: seconds }))).toBe(expected);
  });
});

describe('thumbnailUrl', () => {
  it("si deriva dall'id del video", () => {
    expect(thumbnailUrl(video())).toBe('https://i.ytimg.com/vi/KfODG5REyss/mqdefault.jpg');
  });
});

describe('queueElapsed', () => {
  const entry = (daysLeft: number) => ({ reason: 'serve', expiresAt: new Date(), daysLeft });

  it.each([
    ['appena entrato', 14, 0],
    ['a metà', 7, 0.5],
    ['scaduto', 0, 1],
    ['mai oltre il pieno', -3, 1],
    ['mai sotto zero', 20, 0],
  ])('%s', (_name, daysLeft, expected) => {
    expect(queueElapsed(entry(daysLeft), QUEUE_LIFETIME_DAYS)).toBeCloseTo(expected);
  });
});
