/**
 * Shared application state.
 *
 * Centralizes mutable state that was previously scattered across app.ts
 * as module-level `let` bindings. Extracted modules import from here
 * instead of importing from app.ts, avoiding circular dependencies and
 * enabling proper reassignment.
 */
export let currentUser: { id: number; username: string; role: string } | null = null;
export let currentChar: any = null;
export let currentCampaign: any = null;
export let currentTab = 'stats';
export let allLocations: any[] = [];
export let allNPCs: any[] = [];

/** Roll mode applied to sheet checks/saves/attacks when no explicit advantage is passed. */
export type RollMode = 'normal' | 'advantage' | 'disadvantage';
export let rollMode: RollMode = 'normal';

export function setRollMode(mode: RollMode): void { rollMode = mode; }

export function setCurrentUser(u: typeof currentUser) { currentUser = u; }
export function setCurrentChar(c: any) { currentChar = c; }
export function setCurrentCampaign(c: any) { currentCampaign = c; }
export function setCurrentTab(t: string) { currentTab = t; }
export function setAllLocations(l: any[]) { allLocations = l; }
export function setAllNPCs(n: any[]) { allNPCs = n; }
