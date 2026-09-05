import { showView } from './navigation';
import { expose } from './lib/expose';

expose('showFactions', function (): void {
  showView('factions');
  const el = document.getElementById('factionsContent')!;
  el.setAttribute('hx-get', '/htmx/factions');
  el.setAttribute('hx-trigger', 'load');
  el.setAttribute('hx-swap', 'innerHTML');
  el.innerHTML = '<div class="ornament">✧ Loading factions... ✧</div>';
  window.htmx.process(el);
});
