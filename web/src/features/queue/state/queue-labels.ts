import type { TranslationKey } from '@shared/i18n/keys';

// La regola del plurale sta qui: «scade oggi», «scade domani», «ancora N giorni».
export function daysLeftLabel(daysLeft: number): TranslationKey {
  if (daysLeft <= 0) return 'queue_days_left_zero';
  return daysLeft === 1 ? 'queue_days_left_one' : 'queue_days_left_many';
}

// Dall'ultimo giorno la scadenza si mostra come avviso.
export const isExpiring = (daysLeft: number): boolean => daysLeft <= 1;
