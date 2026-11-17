import type { UseCase } from '@core/use-cases/use-case';

export type MoveInput = {
  readonly order: readonly number[];
  readonly id: number;
  readonly direction: 'up' | 'down';
};

// Sposta una raccolta di un posto. Agli estremi non succede niente: la prima non sale e
// l'ultima non scende, e chi chiama lo capisce perché riceve lo stesso ordine.
export const moveCollection: UseCase<MoveInput, readonly number[]> = ({ order, id, direction }) => {
  const from = order.indexOf(id);
  const to = direction === 'up' ? from - 1 : from + 1;
  if (from === -1 || to < 0 || to >= order.length) return order;
  const next = [...order];
  next.splice(from, 1);
  next.splice(to, 0, id);
  return next;
};
