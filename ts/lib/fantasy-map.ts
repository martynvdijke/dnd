/**
 * Fantasy world-view helpers: locate the campaign's world map image and, when
 * there is none, generate a parchment basemap so the view never falls back to
 * an Earth tile layer.
 *
 * Coordinates stay the existing latitude/longitude degrees (the location editor
 * already speaks in degrees), so the world map spans the standard -180..180 /
 * -90..90 space. Upload a 2:1 image to avoid stretching.
 */

export const WORLD_MAP_NAME = 'World Map';

/** [[maxLat, minLng], [minLat, maxLng]] — Leaflet accepts either corner order. */
export const WORLD_BOUNDS: [[number, number], [number, number]] = [[90, -180], [-90, 180]];

export interface CampaignMapInfo {
  id: number;
  name: string;
  image_url?: string;
  width?: number;
  height?: number;
  grid_size?: number;
  grid_units?: string;
  is_active?: boolean;
  fog_of_war?: string;
}

/**
 * The campaign's world-map entry, if one has been created (image may be empty).
 * ponytail: identified by a fixed name inside campaign_maps; upgrade to a
 * dedicated kind/column if a battle-map UI ever surfaces the same table.
 */
export function findWorldMap(maps: CampaignMapInfo[] | null | undefined): CampaignMapInfo | null {
  if (!Array.isArray(maps)) return null;
  return maps.find((m) => m && m.name === WORLD_MAP_NAME) || null;
}

function drawCompass(ctx: CanvasRenderingContext2D, cx: number, cy: number, r: number): void {
  ctx.save();
  ctx.translate(cx, cy);
  ctx.strokeStyle = 'rgba(96,66,32,0.55)';
  ctx.fillStyle = 'rgba(96,66,32,0.55)';
  ctx.lineWidth = 1.5;
  ctx.beginPath();
  ctx.arc(0, 0, r, 0, Math.PI * 2);
  ctx.stroke();
  ctx.beginPath();
  ctx.arc(0, 0, r * 0.72, 0, Math.PI * 2);
  ctx.stroke();
  for (let i = 0; i < 4; i++) {
    const a = (i * Math.PI) / 2;
    const long = i % 2 === 0;
    const tip = long ? r : r * 0.62;
    ctx.beginPath();
    ctx.moveTo(Math.cos(a) * tip, Math.sin(a) * tip);
    ctx.lineTo(Math.cos(a + Math.PI / 4) * r * 0.34, Math.sin(a + Math.PI / 4) * r * 0.34);
    ctx.lineTo(Math.cos(a - Math.PI / 4) * r * 0.34, Math.sin(a - Math.PI / 4) * r * 0.34);
    ctx.closePath();
    ctx.fill();
  }
  ctx.font = `bold ${Math.round(r * 0.42)}px serif`;
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillText('N', 0, -r * 0.58);
  ctx.restore();
}

/**
 * Render a parchment basemap as a data URL. Returns null when a 2D canvas is
 * unavailable (e.g. non-browser environment), letting the caller fall back to
 * a plain parchment background colour.
 */
export function parchmentDataUrl(w = 1024, h = 512): string | null {
  if (typeof document === 'undefined') return null;
  const canvas = document.createElement('canvas');
  canvas.width = w;
  canvas.height = h;
  const ctx = canvas.getContext('2d');
  if (!ctx) return null;

  // Deterministic pseudo-random so the texture is stable between renders.
  let seed = 1337;
  const rnd = () => (seed = (seed * 1664525 + 1013904223) >>> 0) / 4294967296;

  ctx.fillStyle = '#e7d3a4';
  ctx.fillRect(0, 0, w, h);

  for (let i = 0; i < 700; i++) {
    const x = rnd() * w;
    const y = rnd() * h;
    const r = 6 + rnd() * 40;
    const g = ctx.createRadialGradient(x, y, 0, x, y, r);
    g.addColorStop(0, `rgba(150,118,72,${0.03 + rnd() * 0.06})`);
    g.addColorStop(1, 'rgba(150,118,72,0)');
    ctx.fillStyle = g;
    ctx.beginPath();
    ctx.arc(x, y, r, 0, Math.PI * 2);
    ctx.fill();
  }

  const vignette = ctx.createRadialGradient(w / 2, h / 2, Math.min(w, h) * 0.3, w / 2, h / 2, Math.max(w, h) * 0.62);
  vignette.addColorStop(0, 'rgba(120,88,44,0)');
  vignette.addColorStop(1, 'rgba(84,58,28,0.38)');
  ctx.fillStyle = vignette;
  ctx.fillRect(0, 0, w, h);

  ctx.strokeStyle = 'rgba(120,92,55,0.14)';
  ctx.lineWidth = 1;
  for (let x = 0; x <= w; x += w / 12) {
    ctx.beginPath();
    ctx.moveTo(x, 0);
    ctx.lineTo(x, h);
    ctx.stroke();
  }
  for (let y = 0; y <= h; y += h / 6) {
    ctx.beginPath();
    ctx.moveTo(0, y);
    ctx.lineTo(w, y);
    ctx.stroke();
  }

  ctx.strokeStyle = 'rgba(96,66,32,0.5)';
  ctx.lineWidth = 4;
  ctx.strokeRect(14, 14, w - 28, h - 28);
  ctx.lineWidth = 1;
  ctx.strokeRect(22, 22, w - 44, h - 44);

  drawCompass(ctx, w - 74, 74, 36);
  return canvas.toDataURL('image/png');
}
