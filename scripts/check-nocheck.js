import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';

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

function walk(dir) {
  const out = [];
  for (const e of readdirSync(dir, { withFileTypes: true })) {
    const p = join(dir, e.name);
    if (e.isDirectory()) out.push(...walk(p));
    else if (e.isFile() && p.endsWith('.ts')) out.push(p);
  }
  return out;
}

const files = walk('ts');
const offenders = [];
for (const f of files) {
  if (f.includes('nocheck-guard')) continue;
  const content = readFileSync(f, 'utf8');
  if (content.includes('@ts-' + 'nocheck')) {
    if (!allowlist.has(f)) offenders.push(f);
  }
}
if (offenders.length) {
  console.error('Disallowed // @ts-' + 'nocheck found in:');
  offenders.forEach((f) => console.error('  ' + f));
  console.error('\nAllowed: ' + [...allowlist].join(', '));
  process.exit(1);
} else {
  console.log('nocheck guard: ok (allowed files only)');
}
