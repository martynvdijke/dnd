let touchStartY = 0;
let touchCurrentY = 0;
let isPulling = false;

const PULL_THRESHOLD = 80;

export function initPullToRefresh(containerId: string, onRefresh: () => void): void {
  const container = document.getElementById(containerId);
  if (!container) return;

  let pullIndicator = document.createElement('div');
  pullIndicator.className = 'pull-to-refresh';
  pullIndicator.innerHTML = '<div class="spinner-border spinner-border-sm text-gold" role="status"><span class="visually-hidden">Loading...</span></div>';
  pullIndicator.style.display = 'none';
  container.prepend(pullIndicator);

  container.addEventListener('touchstart', (e) => {
    if (container.scrollTop <= 0) {
      touchStartY = e.touches[0].clientY;
      isPulling = true;
    }
  }, { passive: true });

  container.addEventListener('touchmove', (e) => {
    if (!isPulling) return;
    touchCurrentY = e.touches[0].clientY;
    const diff = touchCurrentY - touchStartY;
    if (diff > 0) {
      pullIndicator.style.display = 'flex';
      pullIndicator.style.opacity = String(Math.min(diff / PULL_THRESHOLD, 1));
    }
  }, { passive: true });

  container.addEventListener('touchend', () => {
    if (!isPulling) return;
    isPulling = false;
    const diff = touchCurrentY - touchStartY;
    if (diff > PULL_THRESHOLD) {
      pullIndicator.style.display = 'flex';
      pullIndicator.style.opacity = '1';
      onRefresh();
      setTimeout(() => {
        pullIndicator.style.display = 'none';
        pullIndicator.style.opacity = '0';
      }, 1000);
    } else {
      pullIndicator.style.display = 'none';
      pullIndicator.style.opacity = '0';
    }
    touchStartY = 0;
    touchCurrentY = 0;
  }, { passive: true });
}

export function initSwipeDismiss(containerId: string): void {
  const container = document.getElementById(containerId);
  if (!container) return;

  container.addEventListener('touchstart', (e) => {
    const target = e.target as HTMLElement;
    const toast = target.closest('.toast') as HTMLElement;
    if (!toast) return;
    touchStartY = e.touches[0].clientX;
    (toast as any)._swipeStart = touchStartY;
  }, { passive: true });

  container.addEventListener('touchmove', (e) => {
    const target = e.target as HTMLElement;
    const toast = target.closest('.toast') as HTMLElement;
    if (!toast || (toast as any)._swipeStart === undefined) return;
    const dx = e.touches[0].clientX - (toast as any)._swipeStart;
    if (dx > 50) {
      toast.style.transform = `translateX(${dx}px)`;
      toast.style.opacity = String(1 - dx / 300);
    }
  }, { passive: true });

  container.addEventListener('touchend', (e) => {
    const target = e.target as HTMLElement;
    const toast = target.closest('.toast') as HTMLElement;
    if (!toast || (toast as any)._swipeStart === undefined) return;
    const dx = e.changedTouches[0].clientX - (toast as any)._swipeStart;
    if (dx > 150) {
      toast.style.transition = 'all .3s';
      toast.style.transform = 'translateX(100%)';
      toast.style.opacity = '0';
      setTimeout(() => toast.remove(), 300);
    } else {
      toast.style.transition = 'all .3s';
      toast.style.transform = '';
      toast.style.opacity = '1';
      setTimeout(() => { toast.style.transition = ''; }, 300);
    }
    delete (toast as any)._swipeStart;
  }, { passive: true });
}

(window as any).initPullToRefresh = initPullToRefresh;
(window as any).initSwipeDismiss = initSwipeDismiss;
