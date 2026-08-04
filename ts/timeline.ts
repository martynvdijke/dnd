// @ts-nocheck — legacy monolith, being refactored
import { showView } from './navigation';
import { expose } from './lib/expose';

declare const htmx: any;

expose('showTimeline', function () {
  showView('timeline');
  const el = document.getElementById('timelineContent')!;
  el.setAttribute('hx-get', '/htmx/timeline');
  el.setAttribute('hx-trigger', 'load');
  el.setAttribute('hx-swap', 'innerHTML');
  el.innerHTML = '<div class="ornament">✧ Loading timeline... ✧</div>';
  htmx.process(el);
});
