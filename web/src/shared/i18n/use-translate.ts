import { createContext, useContext } from 'react';
import type { TranslationKey } from './keys';

export type Translate = (key: TranslationKey, params?: Record<string, string | number>) => string;

export const TranslateContext = createContext<Translate>(() => '');

export const useTranslate = (): Translate => useContext(TranslateContext);
