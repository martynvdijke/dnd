export type ViewState =
  | 'campaignPicker'
  | 'characterPicker'
  | 'characters'
  | 'sheet'
  | 'dice'
  | 'compendium'
  | 'party'
  | 'encounter'
  | 'timeline'
  | 'combatTracker'
  | 'wiki'
  | 'oneshot'
  | 'factions'
  | 'shops'
  | 'singleEncounter'
  | 'campaignOverview';

export type SessionModeState = 'normal' | 'session';

export interface CharacterSummary {
  id: number;
  name: string;
  race: string;
  class: string;
  level: number;
  hp_current: number;
  hp_max: number;
  ac: number;
  portrait_url?: string;
}

export interface FABAction {
  id: string;
  label: string;
  icon: string;
  onclick: string;
}

export interface BottomSheetConfig {
  id: string;
  title: string;
  content: string;
  onDismiss?: () => void;
}

export interface NavItem {
  id: string;
  label: string;
  icon: string;
  view: ViewState;
  requiresRole?: 'admin' | 'dm';
  onclick: string;
}

export interface SessionModeStateMachine {
  state: SessionModeState;
  autoActivateDisabled: boolean;
  toggle: () => void;
  activate: () => void;
  deactivate: () => void;
  isActive: () => boolean;
}
