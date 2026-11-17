import { create, type StoreApi, type UseBoundStore } from 'zustand';
import type { Result } from '@core/result/result';
import { describeFailure, type Message } from '../i18n/message';

// Un gruppo in cui un elemento può stare o non stare. Il pannello non sa che cosa sia: per
// le feature è una raccolta e l'elemento è un video.
export type SheetGroup = {
  readonly id: number;
  readonly label: string;
  readonly count: number;
};

// Ciò che il pannello chiede a chi lo usa. Le funzioni le scrive la feature con le sue
// porte di dominio: qui arrivano già pronte, e il pannello non sa da dove.
export type SaveSheetPorts = {
  readonly load: (
    itemId: string,
  ) => Promise<Result<{ groups: readonly SheetGroup[]; selected: readonly number[] }>>;
  readonly setMember: (groupId: number, itemId: string, member: boolean) => Promise<Result<void>>;
  readonly create: (label: string) => Promise<Result<SheetGroup>>;
  // Dopo ogni cambiamento riuscito, i gruppi che ora contengono l'elemento.
  readonly onChange: (itemId: string, selected: readonly number[]) => void;
};

export type SaveSheetState = {
  readonly itemId: string | null;
  readonly itemTitle: string;
  readonly groups: readonly SheetGroup[];
  readonly selected: readonly number[];
  readonly pending: readonly number[];
  readonly isLoading: boolean;
  readonly isCreating: boolean;
  readonly error: Message | null;
  readonly open: (itemId: string, itemTitle: string) => Promise<void>;
  readonly close: () => void;
  readonly toggle: (groupId: number) => Promise<void>;
  // Crea un gruppo nuovo e ci mette subito l'elemento. Risponde `true` se è andata.
  readonly create: (label: string) => Promise<boolean>;
};

export type SaveSheetStore = UseBoundStore<StoreApi<SaveSheetState>>;

const without = (list: readonly number[], id: number) => list.filter((x) => x !== id);

export function createSaveSheetStore(ports: SaveSheetPorts): SaveSheetStore {
  return create<SaveSheetState>((set, get) => ({
    itemId: null,
    itemTitle: '',
    groups: [],
    selected: [],
    pending: [],
    isLoading: false,
    isCreating: false,
    error: null,

    open: async (itemId, itemTitle) => {
      set({ itemId, itemTitle, groups: [], selected: [], isLoading: true, error: null });
      const result = await ports.load(itemId);
      if (get().itemId !== itemId) return;
      if (!result.ok) {
        set({ isLoading: false, error: describeFailure(result.failure) });
        return;
      }
      set({ isLoading: false, groups: result.value.groups, selected: result.value.selected });
    },

    close: () => set({ itemId: null, error: null, pending: [] }),

    toggle: async (groupId) => {
      const { itemId, selected, pending } = get();
      if (itemId === null || pending.includes(groupId)) return;
      const member = !selected.includes(groupId);
      // La casella cambia subito: aspettare il server per vedere la spunta farebbe sembrare
      // il tocco andato a vuoto. Se il server rifiuta, torna com'era.
      set({
        selected: member ? [...selected, groupId] : without(selected, groupId),
        pending: [...pending, groupId],
        error: null,
      });
      const result = await ports.setMember(groupId, itemId, member);
      const now = get();
      if (!result.ok) {
        set({
          selected: member ? without(now.selected, groupId) : [...now.selected, groupId],
          pending: without(now.pending, groupId),
          error: describeFailure(result.failure),
        });
        return;
      }
      set({
        pending: without(now.pending, groupId),
        groups: now.groups.map((g) =>
          g.id === groupId ? { ...g, count: g.count + (member ? 1 : -1) } : g,
        ),
      });
      ports.onChange(itemId, get().selected);
    },

    create: async (label) => {
      const itemId = get().itemId;
      if (itemId === null) return false;
      set({ isCreating: true, error: null });
      const created = await ports.create(label);
      if (!created.ok) {
        set({ isCreating: false, error: describeFailure(created.failure) });
        return false;
      }
      const group = created.value;
      set({ isCreating: false, groups: [...get().groups, group] });
      await get().toggle(group.id);
      return true;
    },
  }));
}
