export let csrfToken = '';
export let apiToken = '';
export let currentUser: any = null;
export function setCsrfToken(v: string) { csrfToken = v; }
export function setApiToken(v: string) { apiToken = v; }
export function setCurrentUser(v: any) { currentUser = v; }
export function getCsrfToken() { return csrfToken; }
export function getApiToken() { return apiToken; }
export function getCurrentUser() { return currentUser; }
export function clearApiToken() { apiToken = ''; }

type UnauthorizedHandler = () => Promise<void> | void;

let onUnauthorized: UnauthorizedHandler | null = null;

// setUnauthorizedHandler registers a callback that re-provisions the API token.
// The admin shell wires this to its bootstrap so a stale token self-heals on a
// 401 instead of wedging every mutation.
export function setUnauthorizedHandler(handler: UnauthorizedHandler | null): void {
  onUnauthorized = handler;
}

function isMutation(method: string): boolean {
  const m = method.toUpperCase();
  return m !== 'GET' && m !== 'HEAD' && m !== 'OPTIONS';
}

function sendRequest(method: string, path: string, body?: any): Promise<Response> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (csrfToken) headers['X-CSRF-Token'] = csrfToken;
  if (apiToken && isMutation(method)) {
    headers['Authorization'] = `Bearer ${apiToken}`;
  }
  const opts: RequestInit = { method, headers, credentials: 'include' };
  if (body !== undefined) opts.body = JSON.stringify(body);
  return fetch(path, opts);
}

export async function api(method: string, path: string, body?: any): Promise<any> {
  let res = await sendRequest(method, path, body);
  // GETs bypass the API-token check, so a stale token only shows up as a 401 on
  // a mutation. Re-provision once and retry rather than forcing a reload.
  if (res.status === 401 && isMutation(method) && apiToken && onUnauthorized) {
    try {
      await onUnauthorized();
    } catch {
      // Keep the original 401 if re-provisioning fails.
    }
    res = await sendRequest(method, path, body);
  }
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
