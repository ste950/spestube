import { describe, expect, it } from 'vitest';
import { createJsonApi, type Fetch } from '@core/http/json-api';

const respond = (status: number, body: string): Fetch => {
  return async () => new Response(status === 204 ? null : body, { status });
};

describe('createJsonApi', () => {
  it('restituisce il corpo di una risposta riuscita, senza dargli una forma', async () => {
    const api = createJsonApi(respond(200, '{"videos":[]}'), '/api');
    expect(await api.request('GET', '/feed/today')).toEqual({ ok: true, value: { videos: [] } });
  });

  it('manda il corpo come JSON, verso il percorso sotto la base', async () => {
    const calls: { url: string; init: RequestInit | undefined }[] = [];
    const send: Fetch = async (url, init) => {
      calls.push({ url, init });
      return new Response('{}', { status: 201 });
    };
    await createJsonApi(send, '/api').request('POST', '/channels', { channel: '@x' });

    expect(calls).toHaveLength(1);
    expect(calls[0]?.url).toBe('/api/channels');
    expect(calls[0]?.init?.method).toBe('POST');
    expect(calls[0]?.init?.body).toBe('{"channel":"@x"}');
  });

  it('una risposta senza corpo è un successo', async () => {
    const api = createJsonApi(respond(204, ''), '/api');
    expect(await api.request('DELETE', '/channels/UCx')).toEqual({ ok: true, value: null });
  });

  it("porta fuori il codice e le parole dell'errore del server", async () => {
    const body = '{"error":{"code":"channel_exists","message":"segui già questo canale"}}';
    const result = await createJsonApi(respond(409, body), '/api').request('POST', '/channels', {});
    expect(result).toEqual({
      ok: false,
      failure: { kind: 'rejected', code: 'channel_exists', message: 'segui già questo canale' },
    });
  });

  it('un 404 di una risorsa resta un rifiuto con le sue parole', async () => {
    const body = '{"error":{"code":"channel_not_found","message":"questo canale non esiste"}}';
    const result = await createJsonApi(respond(404, body), '/api').request('POST', '/channels', {});
    expect(result.ok === false && result.failure.kind).toBe('rejected');
  });

  it("un percorso che l'API non conosce è «non trovato»", async () => {
    const body = '{"error":{"code":"not_found","message":"risorsa inesistente"}}';
    const result = await createJsonApi(respond(404, body), '/api').request('GET', '/nope');
    expect(result).toEqual({ ok: false, failure: { kind: 'notFound' } });
  });

  it('una pagina HTML al posto del JSON è un guasto di rete, con il suo status', async () => {
    // È ciò che risponde il proxy davanti all'app quando la sessione è scaduta.
    const result = await createJsonApi(respond(502, '<html>Bad Gateway</html>'), '/api').request(
      'GET',
      '/feed/today',
    );
    expect(result.ok === false && result.failure).toMatchObject({ kind: 'network', status: 502 });
  });

  it('un 200 che non è JSON è un dato illeggibile', async () => {
    const result = await createJsonApi(respond(200, '<html>login</html>'), '/api').request(
      'GET',
      '/feed/today',
    );
    expect(result.ok === false && result.failure.kind).toBe('parse');
  });

  it('la rete che manca è un fallimento, non un errore lanciato', async () => {
    const send: Fetch = async () => {
      throw new TypeError('Failed to fetch');
    };
    const result = await createJsonApi(send, '/api').request('GET', '/feed/today');
    expect(result.ok === false && result.failure).toMatchObject({ kind: 'network', status: null });
  });
});
