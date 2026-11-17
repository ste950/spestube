// Il vocabolario è condiviso da tutti i soggetti: un soggetto ne dichiara uno proprio solo
// se ha un modo di fallire che non esiste altrove.
export type Failure =
  | { readonly kind: 'network'; readonly status: number | null; readonly cause: unknown }
  | { readonly kind: 'notFound' }
  // Il server ha capito la richiesta e l'ha rifiutata, dicendo perché: `code` è il codice
  // stabile del contratto, `message` il testo già scritto per chi usa l'app.
  | { readonly kind: 'rejected'; readonly code: string; readonly message: string }
  | { readonly kind: 'parse'; readonly at: string; readonly cause: unknown }
  | { readonly kind: 'storage'; readonly cause: unknown };

// `cause` è `unknown` e non `Error`: quello che una libreria lancia non è necessariamente un
// `Error`, e chi assorbe deve poterlo riportare senza perderlo.
export const networkFailure = (status: number | null, cause: unknown): Failure => ({
  kind: 'network',
  status,
  cause,
});

export const notFoundFailure = (): Failure => ({ kind: 'notFound' });

export const rejectedFailure = (code: string, message: string): Failure => ({
  kind: 'rejected',
  code,
  message,
});

export const storageFailure = (cause: unknown): Failure => ({ kind: 'storage', cause });

// `at` dice quale forma non corrispondeva: serve a trovare il confine che ha ceduto.
export const parseFailure = (at: string, cause: unknown): Failure => ({
  kind: 'parse',
  at,
  cause,
});
