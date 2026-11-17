import { networkFailure, notFoundFailure, parseFailure, rejectedFailure } from '../result/failure';
import { fail, ok, type Result } from '../result/result';

// La funzione di rete si riceve e non si nomina: questo modulo sa come si parla con
// un'API JSON, ma a parlare con l'esterno restano le sorgenti, che gli passano `fetch`.
export type Fetch = (input: string, init?: RequestInit) => Promise<Response>;

export type JsonApi = {
  // Il corpo, se c'è, viaggia come JSON. La risposta torna come `unknown`: darle una forma
  // è mestiere di chi chiama, con un controllo e non con una `as`.
  request(method: string, path: string, body?: unknown): Promise<Result<unknown>>;
};

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null;

// Legge `{ "error": { "code", "message" } }`, la forma di ogni errore dell'API.
function readError(body: unknown): { code: string; message: string } | null {
  if (!isRecord(body) || !isRecord(body.error)) return null;
  const { code, message } = body.error;
  return typeof code === 'string' && typeof message === 'string' ? { code, message } : null;
}

export function createJsonApi(send: Fetch, baseUrl: string): JsonApi {
  return {
    async request(method, path, body) {
      try {
        const response = await send(`${baseUrl}${path}`, {
          method,
          headers: {
            Accept: 'application/json',
            ...(body === undefined ? {} : { 'Content-Type': 'application/json' }),
          },
          ...(body === undefined ? {} : { body: JSON.stringify(body) }),
        });
        if (response.status === 204) return ok(null);

        // Un proxy davanti all'app può rispondere con una pagina HTML: leggere il corpo
        // come testo e poi interpretarlo tiene distinti «non è JSON» e «la rete è caduta».
        const text = await response.text();
        let parsed: unknown;
        try {
          parsed = JSON.parse(text);
        } catch (error) {
          return response.ok
            ? fail(parseFailure(`${method} ${path}`, error))
            : fail(networkFailure(response.status, error));
        }
        if (response.ok) return ok(parsed);

        const error = readError(parsed);
        if (error === null) return fail(networkFailure(response.status, parsed));
        // `not_found` è il percorso che l'API non conosce. Ogni altro rifiuto, anche un 404
        // o un 502 per YouTube che non risponde, arriva con parole che servono a chi guarda.
        if (response.status === 404 && error.code === 'not_found') return fail(notFoundFailure());
        return fail(rejectedFailure(error.code, error.message));
      } catch (error) {
        // `send` rifiuta solo quando la richiesta non è partita o non è tornata.
        return fail(networkFailure(null, error));
      }
    },
  };
}
