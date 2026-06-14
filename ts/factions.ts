// @ts-nocheck — legacy monolith, being refactored
import { showView } from './navigation';

declare const htmx: any;

(window as any).showFactions = function () {
  showView('factions');
  const el = document.getElementById('factionsContent')!;
  el.setAttribute('hx-get', '/htmx/factions');
  el.setAttribute('hx-trigger', 'load');
  el.setAttribute('hx-swap', 'innerHTML');
  el.innerHTML = '<div class="ornament">✧ Loading factions... ✧</div>';
  htmx.process(el);
};
