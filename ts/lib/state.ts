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
export let currentTab = 'stats';
export let allLocations: any[] = [];
export let allNPCs: any[] = [];

export function setCurrentUser(u: typeof currentUser) { currentUser = u; }
export function setCurrentChar(c: any) { currentChar = c; }
export function setCurrentTab(t: string) { currentTab = t; }
export function setAllLocations(l: any[]) { allLocations = l; }
export function setAllNPCs(n: any[]) { allNPCs = n; }
