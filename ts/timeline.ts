// @ts-nocheck — legacy monolith, being refactored
import { showView } from './navigation';

declare const htmx: any;

(window as any).showTimeline = function () {
  showView('timeline');
  const el = document.getElementById('timelineContent')!;
  el.setAttribute('hx-get', '/htmx/timeline');
  el.setAttribute('hx-trigger', 'load');
  el.setAttribute('hx-swap', 'innerHTML');
  el.innerHTML = '<div class="ornament">✧ Loading timeline... ✧</div>';
  htmx.process(el);
};
