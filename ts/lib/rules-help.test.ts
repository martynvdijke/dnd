import { describe, it, expect } from 'vitest';
import { getRulesHelp, rulesHelpButton, RULES } from './rules-help';

describe('rules-help', () => {
  it('returns known rule text', () => {
    expect(getRulesHelp('ac').title).toBe('Armor Class');
    expect(getRulesHelp('death-saves').title).toBe('Death Saves');
    expect(Object.keys(RULES).length).toBeGreaterThan(5);
  });

  it('falls back to a generic message for unknown keys', () => {
    const help = getRulesHelp('does-not-exist');
    expect(help.title).toBe('Rules Help');
    expect(help.body).toContain('No rules explanation');
  });

  it('renders a help button carrying the rule key', () => {
    const html = rulesHelpButton('exhaustion');
    expect(html).toContain('data-rule="exhaustion"');
    expect(html).toContain("showRulesHelp('exhaustion')");
  });
});
