// @ts-nocheck — split from monolith
export let csrfToken = '';
export let apiToken = '';
export let currentUser: any = null;
export function setCsrfToken(v: string) { csrfToken = v; }
export function setApiToken(v: string) { apiToken = v; }
export function setCurrentUser(v: any) { currentUser = v; }
export function getCsrfToken() { return csrfToken; }
export function getApiToken() { return apiToken; }
export function getCurrentUser() { return currentUser; }

export async function api(method: string, path: string, body?: any): Promise<any> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (csrfToken) headers['X-CSRF-Token'] = csrfToken;
  if (apiToken && method !== 'GET' && method !== 'HEAD' && method !== 'OPTIONS') {
    headers['Authorization'] = `Bearer ${apiToken}`;
  }
  const opts: RequestInit = { method, headers, credentials: 'include' };
  if (body !== undefined) opts.body = JSON.stringify(body);
  const res = await fetch(path, opts);
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || 'Request failed');
  }
  return res.json();
}

// shared mutable state for compendium browser
export let currentSchemaId: number = 0;
export let currentSchemaFields: any[] = [];
export let currentSchemaPage: number = 1;
export let currentSchemaQuery: string = '';
export let currentSchemaName: string = '';
export let selectedEntryIds: Set<number> = new Set();
export let entryModalSchemaId = 0;
export let entryModalSchemaFields: any[] = [];
export let entryModalEditId: number | null = null;
export function setCurrentSchemaId(v: number) { currentSchemaId = v; }
export function setCurrentSchemaFields(v: any[]) { currentSchemaFields = v; }
export function setCurrentSchemaPage(v: number) { currentSchemaPage = v; }
export function setCurrentSchemaQuery(v: string) { currentSchemaQuery = v; }
export function setCurrentSchemaName(v: string) { currentSchemaName = v; }
export function setEntryModalSchemaId(v: number) { entryModalSchemaId = v; }
export function setEntryModalSchemaFields(v: any[]) { entryModalSchemaFields = v; }
export function setEntryModalEditId(v: number | null) { entryModalEditId = v; }

// logs / import / campaign shared
export let logRefreshInterval: any = null;
export function setLogRefreshInterval(v: any) { logRefreshInterval = v; }
export let importJsonData: { records: any[], filename: string } | null = null;
export let importMapping: { jsonField: string, schemaField: string, schemaLabel: string, required: boolean, preview: string }[] = [];
export function setImportJsonData(v: any) { importJsonData = v; }
export function setImportMapping(v: any) { importMapping = v; }
export let schemaEditId: number | null = null;
export function setSchemaEditId(v: number | null) { schemaEditId = v; }
export let aiEndpointEditId: number | null = null;
export function setAiEndpointEditId(v: number | null) { aiEndpointEditId = v; }
export let campaignEventEditId: number | null = null;
export function setCampaignEventEditId(v: number | null) { campaignEventEditId = v; }
