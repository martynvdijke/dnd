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

// Font Awesome icon per section, rendered before the tab label.
export const sectionIcons: Record<string, string> = {
  stats: 'fa-chart-simple',
  combat: 'fa-swords',
  spells: 'fa-wand-magic-sparkles',
  inventory: 'fa-bag-shopping',
  resources: 'fa-gauge-high',
  features: 'fa-star',
  feats: 'fa-medal',
  companions: 'fa-paw',
  crafting: 'fa-hammer',
  journal: 'fa-book',
  notes: 'fa-note-sticky',
  details: 'fa-circle-info',
  dice: 'fa-dice',
  party: 'fa-flag',
};

export function getSections(): string[] {
  return sections;
}
