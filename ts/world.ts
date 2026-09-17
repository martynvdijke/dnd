import { showView } from './navigation';
import { api } from './lib/api';
import { esc } from './lib/dom';
import { expose } from './lib/expose';
import { currentCampaign } from './lib/state';
import { FilePicker } from './file-picker';
import { WORLD_BOUNDS, WORLD_MAP_NAME, findWorldMap, parchmentDataUrl } from './lib/fantasy-map';
import L from 'leaflet';

let worldMap: any = null;
let worldOverlay: any = null;
let worldMarkers: any[] = [];

function getCampaignId(): number | null {
  return (currentCampaign as any)?.id ?? null;
}

function worldBounds(): any {
  return L.latLngBounds(WORLD_BOUNDS as any);
}

function initWorldMap(): void {
  const c = document.getElementById('worldMapContainer');
  if (!c) return;
  // worldContent is re-rendered on every visit, so drop the stale map/container.
  if (worldMap) { worldMap.remove(); worldMap = null; worldOverlay = null; }
  worldMap = L.map('worldMapContainer', {
    center: [0, 0], zoom: 0,
    crs: L.CRS.Simple,
    attributionControl: false,
    maxBounds: worldBounds().pad(0.25),
    minZoom: -2, maxZoom: 4, zoomSnap: 0.25,
  } as any);
  setTimeout(() => worldMap.invalidateSize(), 200);
}

/** Draw the DM's uploaded map, or a generated parchment when there is none. */
function setBasemap(imageUrl: string | null): void {
  if (!worldMap) return;
  if (worldOverlay) { worldMap.removeLayer(worldOverlay); worldOverlay = null; }
  const src = imageUrl || parchmentDataUrl();
  if (src) {
    worldOverlay = L.imageOverlay(src, WORLD_BOUNDS as any, { className: 'world-basemap' }).addTo(worldMap);
  }
  const c = document.getElementById('worldMapContainer');
  if (c) c.style.background = imageUrl ? '#1b1712' : '#e7d3a4';
}

function clearMarkers(): void {
  if (worldMap) worldMarkers.forEach((m: any) => worldMap.removeLayer(m));
  worldMarkers = [];
}

export async function showWorld(): Promise<void> {
  showView('world');
  const el = document.getElementById('worldContent')!;
  el.innerHTML = '<div class="ornament">Loading world...</div>';
  try {
    const cid = getCampaignId();
    const [locs, events, maps] = await Promise.all([
      api('GET', '/api/locations'),
      cid ? api('GET', `/api/timeline?campaign_id=${cid}`).catch(() => []) : Promise.resolve([]),
      cid ? api('GET', `/api/campaigns/${cid}/maps`).catch(() => []) : Promise.resolve([]),
    ]);
    const tl: any[] = Array.isArray(events) ? events : (events as any).events ?? [];
    const mapInfo = findWorldMap(maps as any[]);
    const mapImage = mapInfo?.image_url || null;
    // Group locations by type
    const byType: Record<string, any[]> = {};
    (locs as any[]).forEach((l: any) => { const t = l.type || 'other'; (byType[t] = byType[t] || []).push(l); });
    const withCoords = (locs as any[]).filter((l: any) => l.latitude != null && l.longitude != null);

    let html = `<div class="d-flex justify-content-between align-items-center mb-2">
      <small class="text-muted">${mapImage ? esc(mapInfo!.name) : 'Parchment world'}</small>
      <span>${mapImage ? `<button class="btn btn-sm btn-outline-secondary me-1" onclick="removeWorldMap()">Remove</button>` : ''}<button class="btn btn-sm btn-outline-primary" onclick="setWorldMap()"><i class="fa-solid fa-upload me-1"></i>${mapImage ? 'Change map' : 'Upload world map'}</button></span>
    </div>`;
    html += `<div id="worldMapContainer" style="height:380px;border-radius:8px;border:1px solid var(--border-light);margin-bottom:1rem"></div>`;
    // Directory grouped by type
    html += `<div class="row g-3">`;
    for (const [type, items] of Object.entries(byType)) {
      html += `<div class="col-md-4"><h6 class="text-capitalize"><i class="fa-solid fa-location-dot me-1"></i>${esc(type)} (${items.length})</h6><div class="list-group list-group-flush">`;
      for (const l of items) {
        const count = tl.filter((e: any) => e.linked_entity_type === 'location' && String(e.linked_entity_id) === String(l.id)).length;
        html += `<button class="list-group-item list-group-item-action py-2" onclick="showPlaceDetail(${l.id})"><span class="fw-bold small">${esc(l.name)}</span>${count ? ` <span class="badge bg-secondary ms-1">${count} events</span>` : ''}${l.description ? `<br><small class="text-muted">${esc(l.description).substring(0, 80)}</small>` : ''}</button>`;
      }
      html += `</div></div>`;
    }
    if ((locs as any[]).length === 0) html += `<div class="col-12"><p class="text-muted">No places yet.</p></div>`;
    html += `</div>`;
    html += `<div id="worldPlaceDetail" class="mt-3"></div>`;
    el.innerHTML = html;

    // expose data for detail view
    (window as any).__worldLocs = locs;
    (window as any).__worldEvents = tl;

    initWorldMap();
    setBasemap(mapImage);
    clearMarkers();
    withCoords.forEach((l: any) => {
      const m = L.circleMarker([l.latitude, l.longitude], { radius: 8, fillColor: '#8b0000', color: '#fff', weight: 2, fillOpacity: 0.9 }).addTo(worldMap);
      m.bindPopup(`<strong>${esc(l.name)}</strong><br><small>${esc(l.type)}</small>`);
      m.on('click', () => (window as any).showPlaceDetail(l.id));
      worldMarkers.push(m);
    });
    if (withCoords.length) {
      try { worldMap.fitBounds(L.featureGroup(worldMarkers).getBounds().pad(0.15), { maxZoom: 2 }); } catch {}
    } else {
      worldMap.fitBounds(worldBounds());
    }
  } catch (e: any) {
    el.innerHTML = `<div class="alert alert-danger">${esc(e.message)}</div>`;
  }
}

expose('showWorld', showWorld);

/** Pick an uploaded image and store it as this campaign's world map. */
expose('setWorldMap', async function (): Promise<void> {
  const cid = getCampaignId();
  if (!cid) return;
  let url: string;
  try { url = await FilePicker.pick(); } catch { return; }
  const maps = await api('GET', `/api/campaigns/${cid}/maps`).catch(() => []);
  const existing = findWorldMap(maps as any[]);
  if (existing) {
    await api('PUT', `/api/maps/${existing.id}`, { ...existing, image_url: url });
  } else {
    await api('POST', `/api/campaigns/${cid}/maps`, { name: WORLD_MAP_NAME, image_url: url });
  }
  await showWorld();
});

/** Drop the uploaded map and go back to the parchment basemap. */
expose('removeWorldMap', async function (): Promise<void> {
  const cid = getCampaignId();
  if (!cid) return;
  const maps = await api('GET', `/api/campaigns/${cid}/maps`).catch(() => []);
  const existing = findWorldMap(maps as any[]);
  if (!existing) return;
  await api('PUT', `/api/maps/${existing.id}`, { ...existing, image_url: '' });
  await showWorld();
});

expose('showPlaceDetail', async function (id: number): Promise<void> {
  const locs: any[] = (window as any).__worldLocs || [];
  const events: any[] = (window as any).__worldEvents || [];
  const loc = locs.find((l: any) => l.id === id);
  const detail = document.getElementById('worldPlaceDetail')!;
  if (!loc) { detail.innerHTML = '<p class="text-muted">Place not found.</p>'; return; }
  const linked = events.filter((e: any) => e.linked_entity_type === 'location' && String(e.linked_entity_id) === String(id));
  detail.innerHTML = `<div class="card"><div class="card-body"><h5>${esc(loc.name)} <small class="text-muted">${esc(loc.type)}</small></h5>${loc.description ? `<p class="small">${esc(loc.description)}</p>` : ''}${loc.latitude != null ? `<p class="small text-muted">${loc.latitude.toFixed(4)}, ${loc.longitude.toFixed(4)}</p>` : ''}<h6 class="mt-3">Timeline events (${linked.length})</h6>${linked.length ? linked.map((e: any) => `<div class="border-bottom py-1"><span class="fw-bold small">${esc(e.title)}</span> <small class="text-muted">${esc(e.event_date || '')} · ${esc(e.event_type || '')}</small>${e.description ? `<br><small>${esc(e.description)}</small>` : ''}</div>`).join('') : '<p class="small text-muted">No linked events.</p>'}<div id="placeRecapBacklinks" class="mt-3"><small class="text-muted">Loading session write-ups...</small></div></div></div>`;
  detail.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
  // Backlinks: recaps that link to this location
  try {
    const data: any = await api('GET', `/api/links/location/${id}`);
    const all: any[] = [...(data.outgoing || []), ...(data.backlinks || []), ...(data.links || [])];
    const recaps = all.filter((l: any) => l.source_type === 'recap');
    const container = document.getElementById('placeRecapBacklinks');
    if (!container) return;
    if (!recaps.length) { container.innerHTML = '<h6>Session write-ups (0)</h6><p class="small text-muted">No linked recaps.</p>'; return; }
    container.innerHTML = `<h6>Session write-ups (${recaps.length})</h6>` + recaps.map((l: any) => `<div class="border-bottom py-1"><a href="#" onclick="event.preventDefault();(window as any).renderRecaps ? (window as any).renderRecaps() : (window as any).showRecaps && (window as any).showRecaps()">${esc(l.source_title || `Recap #${l.source_id}`)}</a></div>`).join('');
  } catch {
    const c = document.getElementById('placeRecapBacklinks');
    if (c) c.innerHTML = '';
  }
});
