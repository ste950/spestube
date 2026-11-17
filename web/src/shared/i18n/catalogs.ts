import { it } from './it';
import type { TranslationKey } from './keys';

// SpesTube parla solo italiano. La forma è quella di un sito con più lingue: aggiungerne
// una vuol dire un catalogo in più e una voce in questo tipo.
export type Locale = 'it';

export const catalogs: Record<Locale, Record<TranslationKey, string>> = { it };

export const defaultLocale: Locale = 'it';
