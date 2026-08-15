// tabs.ts — shared character sheet section list (reorganize-sheet-tabs change)
// Single source of truth for sheet tabs. Campaign-scoped sections (locations,
// npcs, sessions, quests, graph, analytics) live in the party view instead.
export const sections = [
  'stats',
  'combat',
  'spells',
  'inventory',
  'resources',
  'features',
  'feats',
  'companions',
  'crafting',
  'journal',
  'notes',
  'details',
  'dice',
  'party',
];

export function getSections(): string[] {
  return sections;
}
