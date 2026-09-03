// Extracted from app.ts — Locations section (address-tech-debt-and-ux)
import { expose } from '../lib/expose';
import { currentChar, allLocations, setAllLocations } from '../lib/state';
import { esc, attrEscape, showModal, hideModal, toast } from '../lib/dom';
import { api } from '../lib/api';
import L from 'leaflet';

// ─── Locations ───

let locationMap: any = null;
let locationMarkers: any[] = [];
let pickMarker: any = null;
let editingLocId: number | null = null;

function initLocationMap() {
  const container = document.getElementById('locMapContainer');
  if (!container || locationMap) return;
  locationMap = L.map('locMapContainer', {
    center: [30, 0], zoom: 2,
    zoomControl: true, attributionControl: false,
  });
  L.tileLayer('https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png', {
    maxZoom: 19, subdomains: 'abcd',
  }).addTo(locationMap);
  setTimeout(() => locationMap.invalidateSize(), 200);
}

function clearLocationMarkers() {
  if (locationMarkers.length) {
    locationMarkers.forEach(m => locationMap.removeLayer(m));
    locationMarkers = [];
  }
}

async function renderLocations() {
  const sidebar = document.getElementById('locSidebar');
  if (!sidebar) return;
  initLocationMap();
  try {
    const links = await api('GET', `/api/characters/${currentChar.id}/locations`);
    if (!document.getElementById('locSidebar') || document.getElementById('locationList')) return;
    clearLocationMarkers();

    const linkedIds = new Set(links.map((l: any) => l.location_id));
    const withCoords = allLocations.filter((l: any) => l.latitude != null && l.longitude != null);
    const noCoords = allLocations.filter((l: any) => l.latitude == null || l.longitude == null);

    // Add markers for locations with coordinates
    withCoords.forEach((l: any) => {
      const isLinked = linkedIds.has(l.id);
      const linkInfo = links.find((x: any) => x.location_id === l.id);
      const color = isLinked ? '#8b0000' : '#b8963e';
      const marker = L.circleMarker([l.latitude, l.longitude], {
        radius: isLinked ? 10 : 7, fillColor: color, color: '#fff', weight: 2, opacity: 1, fillOpacity: 0.9,
      }).addTo(locationMap);
      marker.bindPopup(`
        <div style="font-family:'Playfair Display',serif;min-width:180px">
          <strong style="font-size:1.1rem">${esc(l.name)}</strong>
          <br><span style="color:#8b7355;font-style:italic">${esc(l.type)}</span>
          ${isLinked ? `<br><span style="color:#8b0000;font-weight:600">${esc(linkInfo.relationship)}</span>` : ''}
          ${l.description ? `<br><small>${esc(l.description).substring(0, 120)}</small>` : ''}
          <br><small style="color:#8b7355">${l.latitude.toFixed(4)}, ${l.longitude.toFixed(4)}</small>
        </div>`);
      marker.on('click', () => {
        sidebar.querySelectorAll('.loc-item').forEach(el => el.classList.remove('loc-active'));
        const item = document.getElementById('loc-sidebar-' + l.id);
        if (item) item.classList.add('loc-active');
      });
      locationMarkers.push(marker);
    });

    // Fit bounds if there are markers
    if (withCoords.length > 0) {
      const group = L.featureGroup(locationMarkers);
      locationMap.fitBounds(group.getBounds().pad(0.15));
    } else {
      locationMap.setView([30, 0], 2);
    }

    // Build sidebar
    let linkedHtml = links.length
      ? `<div class="list-group list-group-flush">${links.map((l: any) => {
          const loc = allLocations.find((x: any) => x.id === l.location_id);
          return `<div class="list-group-item loc-item" id="loc-sidebar-${l.location_id}" style="cursor:pointer;border-left:3px solid #8b0000"
            onclick="flyToLocation(${loc?.latitude ?? 'null'},${loc?.longitude ?? 'null'},'loc-sidebar-${l.location_id}')">
            <div class="fw-bold small">${esc(l.location_name)}</div>
            <div><span class="badge badge-gold" style="font-size:0.65rem">${esc(l.relationship)}</span>
              ${loc ? `<span class="text-muted" style="font-size:0.7rem">${esc(loc.type)}</span>` : ''}</div>
            ${l.notes ? `<div class="text-muted" style="font-size:0.7rem">${esc(l.notes)}</div>` : ''}
            <div class="mt-1"><button class="btn btn-sm btn-outline-danger py-0 px-1" style="font-size:0.65rem" onclick="event.stopPropagation();unlinkLocation(${l.id})"><i class="fa-solid fa-unlink"></i></button></div>
          </div>`;
        }).join('')}</div>`
      : '<div class="text-center text-muted py-4"><i class="fa-solid fa-map-pin fa-lg mb-2 d-block"></i><small>No linked locations</small></div>';

    let allHtml = noCoords.length > 0
      ? `<div class="list-group list-group-flush">${noCoords.map((l: any) =>
          `<div class="list-group-item loc-item" id="loc-${l.id}" style="cursor:pointer;opacity:0.7"
            onclick="showEditLocation(${l.id})">
            <div class="small">${esc(l.name)} <span class="text-muted" style="font-size:0.65rem">(${esc(l.type)})</span></div>
            ${l.description ? `<div class="text-muted" style="font-size:0.65rem">${esc(l.description).substring(0, 60)}</div>` : ''}
          </div>`).join('')}</div>`
      : '';

    sidebar.innerHTML = `
      <div class="p-2 border-bottom" style="background:var(--parchment-dark)">
        <small class="fw-bold text-muted">LINKED (${links.length})</small>
      </div>
      ${linkedHtml}
      <div class="p-2 border-bottom" style="background:var(--parchment-dark)">
        <small class="fw-bold text-muted">ALL LOCATIONS (${allLocations.length})</small>
      </div>
      ${allLocations.map((l: any) => {
        if (linkedIds.has(l.id)) return '';
        const hasCoord = l.latitude != null && l.longitude != null;
        return `<div class="list-group-item loc-item" id="loc-sidebar-${l.id}" style="cursor:pointer;border-left:3px solid ${hasCoord ? '#b8963e' : 'transparent'}"
          onclick="${hasCoord ? `flyToLocation(${l.latitude},${l.longitude},'loc-sidebar-${l.id}')` : `showEditLocation(${l.id})`}">
          <div class="small fw-bold">${esc(l.name)}</div>
          <div><span class="text-muted" style="font-size:0.65rem">${esc(l.type)}</span>
            ${hasCoord ? '<span class="text-muted" style="font-size:0.6rem"> · mapped</span>' : '<span class="text-muted" style="font-size:0.6rem"> · no coords</span>'}</div>
          ${l.description ? `<div class="text-muted" style="font-size:0.65rem">${esc(l.description).substring(0, 60)}</div>` : ''}
        </div>`;
      }).join('')}
      ${allLocations.length === 0 ? '<div class="text-center text-muted py-4"><i class="fa-solid fa-map fa-lg mb-2 d-block"></i><small>No locations yet</small></div>' : ''}`;
  } catch { sidebar.innerHTML = '<div class="text-center text-muted py-4">Could not load locations</div>'; }
}
expose('renderLocations', renderLocations);

function getLocSidebar(): HTMLElement { return document.getElementById('locSidebar')!; }

expose('flyToLocation', function (lat: number | null, lng: number | null, activeId: string) {
  if (lat != null && lng != null) {
    locationMap.setView([lat, lng], 8, { animate: true });
  }
  getLocSidebar().querySelectorAll('.loc-item').forEach(el => el.classList.remove('loc-active'));
  const item = document.getElementById(activeId);
  if (item) item.classList.add('loc-active');
});

// ─── Link / Unlink ───

expose('showLinkLocation', function () {
  showModal('Link Location', `
    <div class="mb-3"><label class="form-label">Search all locations</label>
      <input type="search" class="form-control" id="locSearchInput" placeholder="Search across all users..." autocomplete="off">
      <div id="locSearchResults" class="mt-1" style="max-height:30vh;overflow-y:auto"></div></div>
    <div class="mb-3"><label class="form-label">Location</label>
      <select class="form-select" id="linkLocId">${allLocations.map((l:any) => `<option value="${l.id}">${esc(l.name)} (${esc(l.type)})</option>`).join('')}</select></div>
    <div class="mb-3"><label class="form-label">Relationship</label>
      <select class="form-select" id="linkLocRel">
        <option value="current">Current Location</option><option value="hometown">Hometown</option><option value="visited">Visited</option>
        <option value="headquarters">Headquarters</option><option value="quest">Quest Location</option><option value="other">Other</option>
      </select></div>
    <div class="mb-3"><label class="form-label">Notes</label><textarea class="form-control" id="linkLocNotes" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveLinkLocation()"><i class="fa-solid fa-link me-1"></i>Link</button>
  `);
  let timer: ReturnType<typeof setTimeout> | null = null;
  const input = document.getElementById('locSearchInput') as HTMLInputElement;
  input.addEventListener('input', () => {
    if (timer) clearTimeout(timer);
    timer = setTimeout(async () => {
      const q = input.value.trim();
      const el = document.getElementById('locSearchResults')!;
      if (!q) { el.innerHTML = ''; return; }
      let results: any[] = [];
      try { results = await api('GET', `/api/locations/search?q=${encodeURIComponent(q)}`); } catch { results = []; }
      el.innerHTML = results.length ? results.map((l:any) => `
        <div class="cp-item d-flex justify-content-between align-items-center p-2 border-bottom">
          <div><span class="fw-bold">${esc(l.name)}</span>
            ${l.type ? `<span class="text-muted small ms-1">${esc(l.type)}</span>` : ''}
            ${l.description ? `<span class="text-muted small"> — ${esc(l.description).substring(0, 60)}</span>` : ''}</div>
          <button class="btn btn-sm btn-outline-primary" onclick="pickSearchedLocation(${l.id},'${attrEscape(l.name)}')">Use</button>
        </div>`).join('') : '<div class="text-muted small fst-italic p-2">No locations found.</div>';
    }, 250);
  });
});

expose('pickSearchedLocation', function (id: number, name: string) {
  const sel = document.getElementById('linkLocId') as HTMLSelectElement;
  if (![...sel.options].some((o) => +o.value === id)) {
    const opt = document.createElement('option');
    opt.value = String(id);
    opt.textContent = name + ' (searched)';
    sel.appendChild(opt);
  }
  sel.value = String(id);
  const input = document.getElementById('locSearchInput') as HTMLInputElement;
  if (input) input.value = name;
  toast(`Selected ${name}`);
});

expose('saveLinkLocation', async function () {
  await api('POST', `/api/characters/${currentChar.id}/locations`, {
    location_id: +(document.getElementById('linkLocId') as HTMLSelectElement).value,
    relationship: (document.getElementById('linkLocRel') as HTMLSelectElement).value,
    notes: (document.getElementById('linkLocNotes') as HTMLTextAreaElement).value,
  });
  hideModal();
  renderLocations();
  toast('Location linked');
});

expose('unlinkLocation', async function (id:number) {
  await api('DELETE', `/api/locations/link/${id}`);
  renderLocations();
});

// ─── Create / Edit ───

expose('showCreateLocation', function () {
  showModal('New Location', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="newLocName"></div>
    <div class="mb-3"><label class="form-label">Type</label>
      <select class="form-select" id="newLocType">
        <option value="region">Region</option><option value="city">City</option><option value="town">Town</option>
        <option value="dungeon">Dungeon</option><option value="tavern">Tavern</option><option value="temple">Temple</option>
        <option value="shop">Shop</option><option value="wilderness">Wilderness</option><option value="other">Other</option>
      </select></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="newLocDesc" rows="3"></textarea></div>
    <div class="row g-2 mb-3">
      <div class="col-6"><label class="form-label">Latitude</label><input class="form-control" id="newLocLat" type="number" step="any" placeholder="e.g. 51.5"></div>
      <div class="col-6"><label class="form-label">Longitude</label><input class="form-control" id="newLocLng" type="number" step="any" placeholder="e.g. -0.12"></div>
    </div>
    <button class="btn btn-outline-secondary btn-sm w-100 mb-2" onclick="pickFromMap('new')"><i class="fa-solid fa-crosshairs me-1"></i>Pick from Map</button>
    <button class="btn btn-primary w-100" onclick="saveNewLocation()"><i class="fa-solid fa-plus me-1"></i>Create</button>
  `);
});

expose('saveNewLocation', async function () {
  const lat = parseFloat((document.getElementById('newLocLat') as HTMLInputElement).value);
  const lng = parseFloat((document.getElementById('newLocLng') as HTMLInputElement).value);
  await api('POST', '/api/locations', {
    name: (document.getElementById('newLocName') as HTMLInputElement).value,
    type: (document.getElementById('newLocType') as HTMLSelectElement).value,
    description: (document.getElementById('newLocDesc') as HTMLTextAreaElement).value,
    latitude: isNaN(lat) ? null : lat,
    longitude: isNaN(lng) ? null : lng,
  });
  setAllLocations(await api('GET', '/api/locations'));
  await renderLocations();
  hideModal();
  toast('Location created');
});

expose('showEditLocation', async function (locId: number) {
  editingLocId = locId;
  const loc = allLocations.find((l: any) => l.id === locId);
  if (!loc) return;
  showModal('Edit Location', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="editLocName" value="${esc(loc.name)}"></div>
    <div class="mb-3"><label class="form-label">Type</label>
      <select class="form-select" id="editLocType">${['region','city','town','dungeon','tavern','temple','shop','wilderness','other'].map(t =>
        `<option value="${t}" ${t === loc.type ? 'selected' : ''}>${t.charAt(0).toUpperCase() + t.slice(1)}</option>`).join('')}</select></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="editLocDesc" rows="3">${esc(loc.description)}</textarea></div>
    <div class="row g-2 mb-3">
      <div class="col-6"><label class="form-label">Latitude</label><input class="form-control" id="editLocLat" type="number" step="any" value="${loc.latitude ?? ''}" placeholder="Optional"></div>
      <div class="col-6"><label class="form-label">Longitude</label><input class="form-control" id="editLocLng" type="number" step="any" value="${loc.longitude ?? ''}" placeholder="Optional"></div>
    </div>
    <button class="btn btn-outline-secondary btn-sm w-100 mb-2" onclick="pickFromMap('edit')"><i class="fa-solid fa-crosshairs me-1"></i>Pick from Map</button>
    <div class="d-flex gap-2">
      <button class="btn btn-primary flex-grow-1" onclick="saveEditLocation(${locId})"><i class="fa-solid fa-floppy-disk me-1"></i>Save</button>
      <button class="btn btn-outline-danger" onclick="deleteLocation(${locId})"><i class="fa-solid fa-trash"></i></button>
    </div>
  `);
});

expose('saveEditLocation', async function (locId: number) {
  const lat = parseFloat((document.getElementById('editLocLat') as HTMLInputElement).value);
  const lng = parseFloat((document.getElementById('editLocLng') as HTMLInputElement).value);
  await api('PUT', `/api/locations/${locId}`, {
    name: (document.getElementById('editLocName') as HTMLInputElement).value,
    type: (document.getElementById('editLocType') as HTMLSelectElement).value,
    description: (document.getElementById('editLocDesc') as HTMLTextAreaElement).value,
    latitude: isNaN(lat) ? null : lat,
    longitude: isNaN(lng) ? null : lng,
  });
  hideModal();
  setAllLocations(await api('GET', '/api/locations'));
  renderLocations();
  toast('Location updated');
});

expose('deleteLocation', async function (locId: number) {
  if (!confirm('Delete this location?')) return;
  await api('DELETE', `/api/locations/${locId}`);
  hideModal();
  setAllLocations(await api('GET', '/api/locations'));
  renderLocations();
  toast('Location deleted');
});

expose('pickFromMap', function (mode: string) {
  hideModal();
  toast('Click on the map to place a pin', false);
  if (pickMarker) locationMap.removeLayer(pickMarker);
  locationMap.once('click', function (e: any) {
    const lat = e.latlng.lat.toFixed(5);
    const lng = e.latlng.lng.toFixed(5);
    pickMarker = L.marker([lat, lng], {
      icon: L.divIcon({ className: '', html: '<i class="fa-solid fa-map-pin" style="color:#8b0000;font-size:2rem;text-shadow:0 1px 3px rgba(0,0,0,.5)"></i>', iconSize: [24, 24], iconAnchor: [12, 24] }),
    }).addTo(locationMap);
    if (mode === 'new') {
      (document.getElementById('newLocLat') as HTMLInputElement).value = lat;
      (document.getElementById('newLocLng') as HTMLInputElement).value = lng;
      (window as any).showCreateLocation();
    } else if (editingLocId) {
      (document.getElementById('editLocLat') as HTMLInputElement).value = lat;
      (document.getElementById('editLocLng') as HTMLInputElement).value = lng;
      (window as any).showEditLocation(editingLocId);
    }
  });
});
