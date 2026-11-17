import { existsSync, readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

import boundaries from '../../architecture/boundaries.json';

// tests/architecture/ → tests/ → web/. Il percorso si deriva dalla posizione del file e non
// dalla cartella corrente: un percorso sbagliato non farebbe fallire questo test, lo farebbe
// passare su un insieme vuoto — cioè esattamente il difetto che è qui per intercettare.
const WEB = join(import.meta.dirname, '..', '..');

function filesUnder(dir: string, pattern: RegExp): string[] {
  if (!existsSync(join(WEB, dir))) return [];

  return readdirSync(join(WEB, dir), { withFileTypes: true }).flatMap((entry) => {
    if (entry.isDirectory()) return filesUnder(join(dir, entry.name), pattern);
    return pattern.test(entry.name) ? [join(dir, entry.name)] : [];
  });
}

describe('i confini che i linter difendono esistono ancora', () => {
  it.each(Object.entries(boundaries.layers))('%s contiene dei file', (_name, layer) => {
    // Una cartella rinominata o spostata non fa protestare ESLint: nessun file corrisponde più
    // al glob, e il blocco che la riguardava smette di valere senza dirlo a nessuno.
    expect(filesUnder(layer.dir, /\.tsx?$/)).not.toHaveLength(0);
  });

  it('esiste almeno una feature, con il suo routes.tsx', () => {
    // I blocchi per feature si generano leggendo questa cartella: se è vuota, l'indipendenza
    // fra feature è difesa da zero regole.
    const features = readdirSync(join(WEB, boundaries.featuresRoot), { withFileTypes: true })
      .filter((entry) => entry.isDirectory())
      .map((entry) => entry.name);

    expect(features).not.toHaveLength(0);
    expect(
      features.filter(
        (name) => !existsSync(join(WEB, boundaries.featuresRoot, name, 'routes.tsx')),
      ),
    ).toEqual([]);
  });

  it('stylelint ha dei fogli da leggere', () => {
    expect(filesUnder('src', /\.module\.css$/)).not.toHaveLength(0);
  });
});

const GUARDS = [
  'no-restricted-imports',
  'no-restricted-globals',
  'no-restricted-properties',
  'no-restricted-syntax',
];

it('nessun file spegne una regola di confine', () => {
  const sources = filesUnder('src', /\.tsx?$/);
  const sheets = filesUnder('src', /\.css$/);
  expect(sources).not.toHaveLength(0);

  // Un `eslint-disable` o uno `stylelint-disable` locale è il modo più rapido di attraversare
  // un confine lasciando la CI verde. Gli altri silenziamenti restano ammessi con la loro
  // ragione; questi no.
  const silenced = [
    ...sources.filter((file) =>
      GUARDS.some((guard) =>
        new RegExp(`eslint-disable[^\\n]*${guard}`).test(readFileSync(join(WEB, file), 'utf8')),
      ),
    ),
    ...sheets.filter((file) => readFileSync(join(WEB, file), 'utf8').includes('stylelint-disable')),
  ];

  expect(silenced).toEqual([]);
});
