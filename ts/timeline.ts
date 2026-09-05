import { showView } from './navigation';
import { expose } from './lib/expose';

expose('showTimeline', function (): void {
  showView('timeline');
  const el = document.getElementById('timelineContent')!;
  el.setAttribute('hx-get', '/htmx/timeline');
  el.setAttribute('hx-trigger', 'load');
  el.setAttribute('hx-swap', 'innerHTML');
  el.innerHTML = '<div class="ornament">✧ Loading timeline... ✧</div>';
  window.htmx.process(el);
});
