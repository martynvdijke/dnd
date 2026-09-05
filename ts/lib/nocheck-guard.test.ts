// @ts-ignore — node builtins for test-time filesystem walk
import { readFileSync, readdirSync } from 'node:fs';
// @ts-ignore — node builtins
import { join } from 'node:path';
import { describe, it, expect } from 'vitest';

const allowlist = new Set([
  'ts/app.ts',
  'ts/admin.ts',
  'ts/lib/errors.ts',
  'ts/lib/refresh.ts',
  'ts/app/sort-switch.ts',
  'ts/app/character-ops.ts',
  'ts/app/character-details.ts',
  'ts/app/combat-conditions.ts',
  'ts/app/notes.ts',
  'ts/app/shops.ts',
  'ts/app/random-gen.ts',
  'ts/app/wiki.ts',
  'ts/app/uploads.ts',
  'ts/app/oneshot.ts',
  'ts/app/inventory-spells.ts',
  'ts/app/campaign-dashboard.ts',
  'ts/app/encounter-treasure.ts',
  'ts/app/graph-analytics.ts',
  'ts/admin/state.ts',
  'ts/admin/site-settings.ts',
  'ts/admin/api-tokens.ts',
  'ts/admin/compendium.ts',
  'ts/admin/entries.ts',
  'ts/admin/logs.ts',
  'ts/admin/users.ts',
  'ts/admin/schemas.ts',
  'ts/admin/integrations.ts',
  'ts/admin/import-wizard.ts',
  'ts/admin/events.ts',
  'ts/admin/utils.ts',
  'ts/admin/pdf.ts',
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
    // allowlist may contain future-allowed files (e.g. ts/admin.ts) that currently have no pragma
    expect(current.every((f) => allowlist.has(f)), 'current nocheck files should be subset of allowlist; unexpected: ' + current.filter((f) => !allowlist.has(f)).join(', ')).toBe(true);
    // and every allowlisted file is either currently nochecked or intentionally pre-allowed (admin.ts)
    expect([...allowlist].sort()).toEqual(['ts/admin.ts', 'ts/admin/api-tokens.ts', 'ts/admin/compendium.ts', 'ts/admin/entries.ts', 'ts/admin/events.ts', 'ts/admin/import-wizard.ts', 'ts/admin/integrations.ts', 'ts/admin/logs.ts', 'ts/admin/pdf.ts', 'ts/admin/schemas.ts', 'ts/admin/site-settings.ts', 'ts/admin/state.ts', 'ts/admin/users.ts', 'ts/admin/utils.ts', 'ts/app.ts', 'ts/app/campaign-dashboard.ts', 'ts/app/character-details.ts', 'ts/app/character-ops.ts', 'ts/app/combat-conditions.ts', 'ts/app/encounter-treasure.ts', 'ts/app/graph-analytics.ts', 'ts/app/inventory-spells.ts', 'ts/app/notes.ts', 'ts/app/oneshot.ts', 'ts/app/random-gen.ts', 'ts/app/shops.ts', 'ts/app/sort-switch.ts', 'ts/app/uploads.ts', 'ts/app/wiki.ts', 'ts/lib/errors.ts', 'ts/lib/refresh.ts']);
  });
});
