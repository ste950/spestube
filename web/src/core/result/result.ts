import type { Failure } from './failure';

export type Result<T> =
  { readonly ok: true; readonly value: T } | { readonly ok: false; readonly failure: Failure };

export const ok = <T>(value: T): Result<T> => ({ ok: true, value });

export const fail = <T = never>(failure: Failure): Result<T> => ({ ok: false, failure });

// Il ramo di fallimento restituisce `result` così com'è: non contiene `T`, quindi è già un
// `Result<U>` valido e non c'è niente da ricostruire.
export function mapOk<T, U>(result: Result<T>, fn: (value: T) => U): Result<U> {
  return result.ok ? ok(fn(result.value)) : result;
}
