import { describe, expect, it } from 'vitest';
import { durationMinutes, isSaved, thumbnailUrl, type Video } from '@domain/library/entities/video';
import { moveCollection } from '@domain/library/use-cases/move-collection';

const video = (overrides: Partial<Video> = {}): Video => ({
  id: 'KfODG5REyss',
  title: 'Un video',
  channelTitle: 'Officina Radio',
  publishedAt: new Date('2026-11-04T15:52:59Z'),
  durationSeconds: null,
  status: 'released',
  collectionIds: [],
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
    expect(thumbnailUrl('KfODG5REyss')).toBe('https://i.ytimg.com/vi/KfODG5REyss/mqdefault.jpg');
  });
});

describe('isSaved', () => {
  it('un video è salvato se sta in almeno una raccolta', () => {
    expect(isSaved(video())).toBe(false);
    expect(isSaved(video({ collectionIds: [2] }))).toBe(true);
  });
});

describe('moveCollection', () => {
  const order = [1, 2, 3];

  it.each([
    ['su di un posto', 2, 'up', [2, 1, 3]],
    ['giù di un posto', 2, 'down', [1, 3, 2]],
    ['la prima non sale', 1, 'up', [1, 2, 3]],
    ["l'ultima non scende", 3, 'down', [1, 2, 3]],
    ['una sconosciuta non sposta niente', 9, 'up', [1, 2, 3]],
  ] as const)('%s', (_name, id, direction, expected) => {
    expect(moveCollection({ order, id, direction })).toEqual(expected);
  });

  it("non tocca l'ordine che riceve", () => {
    moveCollection({ order, id: 2, direction: 'up' });
    expect(order).toEqual([1, 2, 3]);
  });
});
