import { create } from 'zustand';

// I numeri mostrati accanto alle sezioni della barra, per percorso della sezione.
//
// È stato di presentazione condiviso, e sta qui perché ha due padroni che non si possono
// conoscere: lo scrive una feature, lo legge il layout dell'app. Non sa cosa conta: per lui
// un numero è un numero.
type NavBadgesState = {
  readonly counts: Readonly<Record<string, number>>;
  readonly setCount: (path: string, count: number) => void;
};

export const useNavBadges = create<NavBadgesState>((set) => ({
  counts: {},
  setCount: (path, count) => set((state) => ({ counts: { ...state.counts, [path]: count } })),
}));

export const setNavBadge = (path: string, count: number): void =>
  useNavBadges.getState().setCount(path, count);
