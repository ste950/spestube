import type { TranslationKey } from '@shared/i18n/keys';

// La regola del plurale sta in un punto solo, e i componenti passano il conteggio.
export function countLabel(count: number): TranslationKey {
  if (count === 0) return 'collections_count_zero';
  return count === 1 ? 'collections_count_one' : 'collections_count_many';
}
