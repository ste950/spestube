import { describe, expect, it } from 'vitest';
import { networkFailure, rejectedFailure } from '@core/result/failure';
import { fail, ok, type Result } from '@core/result/result';
import { createSaveSheetStore, type SaveSheetPorts } from '@shared/save-sheet/save-sheet-store';

const groups = [
  { id: 1, label: 'Orto', count: 2 },
  { id: 2, label: 'Cucina', count: 0 },
];

function sheet(overrides: Partial<SaveSheetPorts> = {}) {
  const changes: { itemId: string; selected: readonly number[] }[] = [];
  const store = createSaveSheetStore({
    load: async () => ok({ groups, selected: [1] }),
    setMember: async () => ok(undefined),
    create: async (label) => ok({ id: 7, label, count: 0 }),
    onChange: (itemId, selected) => changes.push({ itemId, selected }),
    ...overrides,
  });
  return { store, changes };
}

describe('il pannello «Salva in»', () => {
  it('si apre con le raccolte e quelle che contengono già il video', async () => {
    const { store } = sheet();
    await store.getState().open('v1', 'Antenne LoRa');

    expect(store.getState().itemId).toBe('v1');
    expect(store.getState().groups).toEqual(groups);
    expect(store.getState().selected).toEqual([1]);
  });

  it('una casella cambia subito, aggiorna il conteggio e lo dice a chi lo usa', async () => {
    const sent: [number, string, boolean][] = [];
    const { store, changes } = sheet({
      setMember: async (groupId, itemId, member) => {
        sent.push([groupId, itemId, member]);
        return ok(undefined);
      },
    });
    await store.getState().open('v1', 'Antenne LoRa');
    await store.getState().toggle(2);
    await store.getState().toggle(1);

    expect(sent).toEqual([
      [2, 'v1', true],
      [1, 'v1', false],
    ]);
    expect(store.getState().selected).toEqual([2]);
    expect(store.getState().groups.map((g) => g.count)).toEqual([1, 1]);
    expect(changes).toEqual([
      { itemId: 'v1', selected: [1, 2] },
      { itemId: 'v1', selected: [2] },
    ]);
  });

  it('una casella che il server rifiuta torna com’era', async () => {
    const { store, changes } = sheet({ setMember: async () => fail(networkFailure(null, null)) });
    await store.getState().open('v1', 'Antenne LoRa');
    await store.getState().toggle(2);

    expect(store.getState().selected).toEqual([1]);
    expect(store.getState().pending).toEqual([]);
    expect(store.getState().error).toEqual({ key: 'common_error_network' });
    expect(changes).toEqual([]);
  });

  it('mentre una casella aspetta il server, un secondo tocco non parte', async () => {
    let release: (value: Result<void>) => void = () => {};
    let calls = 0;
    const { store } = sheet({
      setMember: () => {
        calls += 1;
        return new Promise((resolve) => {
          release = resolve;
        });
      },
    });
    await store.getState().open('v1', 'Antenne LoRa');
    const first = store.getState().toggle(2);
    await store.getState().toggle(2);
    release(ok(undefined));
    await first;

    expect(calls).toBe(1);
    expect(store.getState().selected).toEqual([1, 2]);
  });

  it('una raccolta nuova compare in fondo, già spuntata', async () => {
    const { store, changes } = sheet();
    await store.getState().open('v1', 'Antenne LoRa');
    expect(await store.getState().create('Viaggi')).toBe(true);

    expect(store.getState().groups.map((g) => g.label)).toEqual(['Orto', 'Cucina', 'Viaggi']);
    expect(store.getState().selected).toEqual([1, 7]);
    expect(changes.at(-1)).toEqual({ itemId: 'v1', selected: [1, 7] });
  });

  it('un nome già usato non crea niente e dice perché', async () => {
    const { store } = sheet({
      create: async () => fail(rejectedFailure('collection_exists', 'nome già usato')),
    });
    await store.getState().open('v1', 'Antenne LoRa');
    expect(await store.getState().create('Orto')).toBe(false);

    expect(store.getState().groups).toEqual(groups);
    expect(store.getState().error).toEqual({ text: 'nome già usato' });
  });

  it('chiuso, non risponde più ai tocchi', async () => {
    let calls = 0;
    const { store } = sheet({
      setMember: async () => {
        calls += 1;
        return ok(undefined);
      },
    });
    await store.getState().open('v1', 'Antenne LoRa');
    store.getState().close();
    await store.getState().toggle(2);

    expect(store.getState().itemId).toBeNull();
    expect(calls).toBe(0);
  });
});
