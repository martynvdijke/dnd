import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';

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
