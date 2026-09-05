// @ts-ignore — node builtins for test-time filesystem walk
import { readFileSync, readdirSync } from 'node:fs';
// @ts-ignore — node builtins
import { join } from 'node:path';
import { describe, it, expect } from 'vitest';

const allowlist = new Set([
  'ts/app.ts',
  'ts/admin.ts',
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
    expect([...allowlist].sort()).toEqual(['ts/admin.ts', 'ts/app.ts']);
  });
});
