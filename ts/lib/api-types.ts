export interface Character {
  id: number;
  name: string;
  race?: string;
  class?: string;
  level?: number;
  hp_current?: number;
  hp_max?: number;
  ac?: number;
  portrait_url?: string;
  can_edit?: boolean;
  [key: string]: unknown;
}

export interface Campaign {
  id: number;
  name: string;
  description?: string;
  [key: string]: unknown;
}

export interface CompendiumEntry {
  id: number;
  schema_id?: number;
  data: Record<string, unknown>;
  [key: string]: unknown;
}

export interface CompendiumSchema {
  id: number;
  type_name: string;
  display_name: string;
  entry_count: number;
  entries?: CompendiumEntry[];
  [key: string]: unknown;
}

export interface Paginated<T> {
  items: T[];
  total: number;
  page?: number;
  per_page?: number;
}

export interface ApiError {
  error: string;
}

export interface InventoryItem {
  id: number;
  name: string;
  category?: string;
  quantity?: number;
  weight?: number;
  equipped?: boolean;
  [key: string]: unknown;
}

export interface Spell {
  id: number;
  name: string;
  level?: number;
  school?: string;
  prepared?: boolean;
  [key: string]: unknown;
}

export interface CompendiumSearchResult {
  id: number;
  type: string;
  name: string;
  [key: string]: unknown;
}
