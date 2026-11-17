import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { beforeEach, describe, expect, it } from 'vitest';
import { createLocalThemeRepository } from '@data/appearance/repositories/local-theme-repository';
import {
  createLocalThemeSource,
  THEME_STORAGE_KEY,
} from '@data/appearance/sources/local-theme-source';

describe('il tema ricordato sul dispositivo', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('senza una scelta è quello del sistema', async () => {
    const repository = createLocalThemeRepository(createLocalThemeSource());
    expect(await repository.getTheme()).toEqual({ ok: true, value: 'system' });
  });

  it('ricorda la scelta', async () => {
    const repository = createLocalThemeRepository(createLocalThemeSource());
    await repository.setTheme('dark');
    expect(localStorage.getItem(THEME_STORAGE_KEY)).toBe('dark');
    expect(await repository.getTheme()).toEqual({ ok: true, value: 'dark' });
  });

  it('un valore sconosciuto torna al sistema', async () => {
    localStorage.setItem(THEME_STORAGE_KEY, 'seppia');
    const repository = createLocalThemeRepository(createLocalThemeSource());
    expect(await repository.getTheme()).toEqual({ ok: true, value: 'system' });
  });

  it('lo script in index.html legge la stessa chiave', () => {
    // Lo script applica il tema prima che parta l'app: se le due chiavi divergono, la
    // scelta si salva e non si applica più all'apertura.
    const html = readFileSync(join(import.meta.dirname, '..', '..', 'index.html'), 'utf8');
    expect(html).toContain(`localStorage.getItem('${THEME_STORAGE_KEY}')`);
  });
});
