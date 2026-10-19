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
  'downtime',
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
  downtime: 'fa-hourglass-half',
  journal: 'fa-book',
  notes: 'fa-note-sticky',
  details: 'fa-circle-info',
  dice: 'fa-dice',
  party: 'fa-flag',
};

export function getSections(): string[] {
  return sections;
}

// Tab category filters — declutter the sheet during play without removing tabs.
export const TAB_FILTERS: Array<{ id: string; label: string; sections: string[] }> = [
  { id: 'all', label: 'All', sections },
  { id: 'play', label: 'Play', sections: ['stats', 'combat', 'spells', 'inventory', 'resources', 'companions', 'dice'] },
  { id: 'build', label: 'Build', sections: ['stats', 'combat', 'spells', 'inventory', 'resources', 'features', 'feats', 'companions', 'crafting', 'details', 'dice'] },
  { id: 'world', label: 'World', sections: ['party', 'journal', 'downtime', 'notes', 'inventory', 'dice'] },
  { id: 'insights', label: 'Insights', sections: ['stats', 'journal', 'notes', 'dice'] },
];

export function sectionsForFilter(id: string): string[] {
  return TAB_FILTERS.find((f) => f.id === id)?.sections || sections;
}
