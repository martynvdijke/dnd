// @ts-ignore — node builtins for test-time filesystem walk
import { readFileSync, readdirSync } from 'node:fs';
// @ts-ignore — node builtins
import { join } from 'node:path';
import { describe, it, expect } from 'vitest';

const allowlist = new Set([
  'ts/app.ts',
  'ts/party.ts',
  'ts/combat-tracker.ts',
  'ts/compendium.ts',
  'ts/encounter.ts',
  'ts/factions.ts',
  'ts/timeline.ts',
  'ts/characters/combat.ts',
  'ts/characters/inventory.ts',
  'ts/characters/resources.ts',
  'ts/characters/sheet.ts',
  'ts/characters/spells.ts',
  'ts/characters/stats.ts',
]);

function walk(dir: string): string[] {
  const out: string[] = [];
  for (const e of readdirSync(dir, { withFileTypes: true })) {
    const p = join(dir, e.name);
    if (e.isDirectory()) out.push(...walk(p));
    else if (e.isFile() && p.endsWith('.ts')) out.push(p);
  }
  return out;
}

describe('nocheck guard', () => {
  it('only allowlisted files contain @ts-nocheck', () => {
    const files = walk('ts');
    const offenders: string[] = [];
    const marker = '@ts-' + 'nocheck';
    for (const f of files) {
      if (f.includes('nocheck-guard')) continue;
      const content = readFileSync(f, 'utf8');
      if (content.includes(marker) && !allowlist.has(f)) {
        offenders.push(f);
      }
    }
    expect(offenders, 'Disallowed @ts-' + 'nocheck in: ' + offenders.join(', ')).toEqual([]);
  });

  it('allowlist matches current nocheck files', () => {
    const files = walk('ts').filter((f) => !f.includes('nocheck-guard'));
    const marker = '@ts-' + 'nocheck';
    const current = files.filter((f) => readFileSync(f, 'utf8').includes(marker)).sort();
    expect(current.sort()).toEqual([...allowlist].sort());
  });
});
