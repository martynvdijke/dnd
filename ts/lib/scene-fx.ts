/**
 * Scene special-effect overlays. Shared by one-shot scene playback, spell
 * casting, and the realtime websocket handler. Effect names match the WLED
 * preset names in handlers/wled.go.
 */
export const SCENE_FX = ['fire', 'smoke', 'lightning', 'rain', 'snow', 'darkness', 'sparkle'];

export function playSceneEffect(effect: string): void {
  const name = (effect || '').toLowerCase();
  if (!SCENE_FX.includes(name)) return;
  const el = document.createElement('div');
  el.className = `scene-fx scene-fx-${name}`;
  document.body.appendChild(el);
  window.setTimeout(() => el.remove(), 4500);
}
