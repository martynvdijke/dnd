/**
 * Campaign & character selection flow.
 *
 * After login the user picks a campaign first, then a character within it.
 * The selection is persisted in localStorage and cleared on logout. Picker
 * views render into #campaignPickerView and #characterPickerView; the main
 * app (characters list, sheet, party) becomes usable once a campaign is
 * selected, and a character is required to open the sheet.
 */

import { api } from './lib/api';
import { esc, toast } from './lib/dom';
import { expose } from './lib/expose';
import { setCurrentCampaign, setCurrentChar, currentCampaign, currentChar, currentUser } from './lib/state';
import { showView } from './navigation';

const LS_CAMPAIGN = 'villum_campaign';
const LS_CHARACTER = 'villum_character';

export interface CampaignSelection {
  id: number;
  name: string;
}

export function getStoredCampaign(): CampaignSelection | null {
  try {
    const raw = localStorage.getItem(LS_CAMPAIGN);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

export function getStoredCharacter(): { id: number; name: string } | null {
  try {
    const raw = localStorage.getItem(LS_CHARACTER);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

export function clearSelection(): void {
  localStorage.removeItem(LS_CAMPAIGN);
  localStorage.removeItem(LS_CHARACTER);
  setCurrentCampaign(null);
  setCurrentChar(null);
}

export async function loadCampaignPicker(): Promise<void> {
  const view = document.getElementById('campaignPickerView');
  if (!view) return;
  setCurrentCampaign(null);
  setCurrentChar(null);
  showView('campaignPicker');
  const container = document.getElementById('campaignPickerList')!;
  container.innerHTML = '<div class="text-center text-muted py-4"><i class="fa-solid fa-spinner fa-spin me-2"></i>Loading campaigns...</div>';
  try {
    const campaigns = await api('GET', '/api/campaigns/mine');
    if (!campaigns.length) {
      container.innerHTML = `
        <div class="empty-state text-center py-5">
          <i class="fa-solid fa-users fa-2x mb-3 d-block" aria-hidden="true"></i>
          <p class="text-muted mb-2">You are not part of any campaign yet.</p>
          <p class="text-muted small mb-3">Create a campaign or ask your Dungeon Master to add you.</p>
        </div>`;
      return;
    }
    container.innerHTML = campaigns.map((ca: any) => `
      <div class="col-md-6 col-lg-4">
        <div class="character-card campaign-picker-card" data-testid="campaign-picker-card" onclick="selectCampaign(${ca.id})">
          <div class="d-flex align-items-center gap-2 mb-1">
            <i class="fa-solid fa-flag" aria-hidden="true"></i>
            <div class="char-name">${esc(ca.name)}</div>
          </div>
          <div class="char-detail">
            ${ca.party_name ? `<i class="fa-solid fa-people-group me-1" aria-hidden="true"></i>${esc(ca.party_name)}` : ''}
            <span class="badge bg-secondary ms-1">${esc(ca.my_role || 'player')}</span>
          </div>
        </div>
      </div>
    `).join('');
  } catch (e: any) {
    container.innerHTML = '';
    toast(e.message, true);
  }
}

export async function selectCampaign(id: number): Promise<void> {
  try {
    const campaigns = await api('GET', '/api/campaigns/mine');
    const camp = campaigns.find((c: any) => c.id === id);
    if (!camp) {
      toast('Campaign not found', true);
      return;
    }
    localStorage.setItem(LS_CAMPAIGN, JSON.stringify({ id: camp.id, name: camp.name }));
    setCurrentCampaign(camp);
    setCurrentChar(null);
    await loadCharacterPicker(camp.id);
  } catch (e: any) {
    toast(e.message, true);
  }
}

export async function loadCharacterPicker(campaignId: number): Promise<void> {
  showView('characterPicker');
  const container = document.getElementById('characterPickerList')!;
  container.innerHTML = '<div class="text-center text-muted py-4"><i class="fa-solid fa-spinner fa-spin me-2"></i>Loading characters...</div>';
  try {
    const [chars, mine] = await Promise.all([
      api('GET', `/api/campaigns/${campaignId}/characters`),
      api('GET', '/api/characters').catch(() => [] as any[]),
    ]);
    // Characters that belong to the current user but are not assigned to a
    // campaign are surfaced under an "Unassigned" group so they are not lost.
    const assignedIds = new Set((chars as any[]).map((c: any) => c.id));
    const unassigned = (mine as any[]).filter((c: any) => !c.campaign_id && !assignedIds.has(c.id));
    const card = (c: any) => `
      <div class="col-md-6 col-lg-4">
        <div class="character-card" data-testid="character-picker-card" onclick="selectCharacter(${c.id})">
          <div class="d-flex align-items-center gap-2 mb-1">
            ${c.portrait_url ? `<img src="${esc(c.portrait_url)}" class="character-portrait" style="width:32px;height:32px;object-fit:cover;border-radius:50%" alt="">` : ''}
            <div class="char-name">${esc(c.name)}</div>
            ${c.owned
              ? '<span class="badge bg-success" title="You can edit this character"><i class="fa-solid fa-pen me-1" aria-hidden="true"></i>Owned</span>'
              : '<span class="badge bg-secondary" title="Read-only — shared character"><i class="fa-solid fa-eye me-1" aria-hidden="true"></i>Shared</span>'}
          </div>
          <div class="char-detail">
            ${c.race_color ? `<span class="badge" style="background:${c.race_color};color:#fff">${esc(c.race)}</span>` : esc(c.race)}
            ${esc(c.class)} · Level ${c.level}
          </div>
          <div class="char-hp mt-1">HP: ${c.hp_current}/${c.hp_max}</div>
        </div>
      </div>
    `;
    const campaignCards = (chars as any[]).map(card).join('');
    const unassignedCards = unassigned.map(card).join('');
    if (!campaignCards && !unassignedCards) {
      container.innerHTML = `
        <div class="empty-state text-center py-5">
          <i class="fa-solid fa-user fa-2x mb-3 d-block" aria-hidden="true"></i>
          <p class="text-muted mb-2">No characters in this campaign yet.</p>
        </div>`;
      return;
    }
    container.innerHTML = `
      ${campaignCards}
      ${unassignedCards ? `
        <div class="col-12">
          <hr class="my-4">
          <h5 class="text-muted"><i class="fa-solid fa-folder-open me-2" aria-hidden="true"></i>Unassigned</h5>
          <p class="text-muted small">Your characters not assigned to a campaign.</p>
        </div>
        ${unassignedCards}` : ''}
    `;
  } catch (e: any) {
    container.innerHTML = '';
    toast(e.message, true);
  }
}

export async function selectCharacter(id: number): Promise<void> {
  const campaignId = currentCampaign?.id ?? getStoredCampaign()?.id;
  if (!campaignId) {
    await loadCampaignPicker();
    return;
  }
  try {
    const [chars, mine] = await Promise.all([
      api('GET', `/api/campaigns/${campaignId}/characters`),
      api('GET', '/api/characters').catch(() => [] as any[]),
    ]);
    let ch = chars.find((c: any) => c.id === id) || mine.find((c: any) => c.id === id);
    if (!ch) {
      toast('Character not found', true);
      return;
    }
    localStorage.setItem(LS_CHARACTER, JSON.stringify({ id: ch.id, name: ch.name }));
    setCurrentChar(ch);
    (window as any).openChar?.(ch.id);
  } catch (e: any) {
    toast(e.message, true);
  }
}

/**
 * Validate the persisted selection against the server. Returns true when a
 * valid campaign is selected; renders the campaign picker otherwise.
 *
 * Admins bypass the picker entirely — they manage every campaign and
 * character, so the campaign-first gate only applies to regular users.
 */
export async function validateSelection(): Promise<boolean> {
  if (currentUser?.role === 'admin') return true;
  const stored = getStoredCampaign();
  if (!stored) {
    await loadCampaignPicker();
    return false;
  }
  try {
    const campaigns = await api('GET', '/api/campaigns/mine');
    const camp = campaigns.find((c: any) => c.id === stored.id);
    if (!camp) {
      // Stored campaign no longer accessible — reset and re-pick.
      clearSelection();
      await loadCampaignPicker();
      return false;
    }
    setCurrentCampaign(camp);
    const storedChar = getStoredCharacter();
    if (storedChar) {
      const chars = await api('GET', `/api/campaigns/${camp.id}/characters`);
      const ch = chars.find((c: any) => c.id === storedChar.id);
      if (ch) setCurrentChar(ch);
      else localStorage.removeItem(LS_CHARACTER);
    }
    return true;
  } catch (e: any) {
    toast(e.message, true);
    await loadCampaignPicker();
    return false;
  }
}

expose('loadCampaignPicker', loadCampaignPicker);
expose('loadCharacterPicker', loadCharacterPicker);
expose('selectCampaign', selectCampaign);
expose('selectCharacter', selectCharacter);
expose('validateSelection', validateSelection);
