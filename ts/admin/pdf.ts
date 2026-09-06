// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
let pdfViewerDoc: any = null;
let pdfViewerPage = 1;
let pdfViewerScale = 1.5;
let pdfViewerUrl = '';
let pdfViewerTitle = '';
let pdfViewerLoaded = false;
let pdfViewerLoading = false;
const pdfViewerQueue: Array<() => void> = [];
function pdfViewerLoadLib(callback: () => void) {
  if (pdfViewerLoaded) { callback(); return; }
  if (pdfViewerLoading) { pdfViewerQueue.push(callback); return; }
  pdfViewerLoading = true;
  const s = document.createElement('script');
  s.src = 'https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.min.js';
  s.onload = () => {
    (window as any).pdfjsLib.GlobalWorkerOptions.workerSrc = 'https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.worker.min.js';
    pdfViewerLoaded = true; pdfViewerLoading = false; callback();
    pdfViewerQueue.forEach(fn => fn()); pdfViewerQueue.length = 0;
  };
  s.onerror = () => {
    pdfViewerLoading = false;
    const errEl = document.getElementById('pdfViewerError');
    if (errEl) { errEl.textContent = 'Failed to load PDF viewer library. Check your internet connection.'; errEl.style.display = 'block'; }
    const loading = document.getElementById('pdfViewerLoading');
    if (loading) loading.style.display = 'none';
    pdfViewerQueue.length = 0;
  };
  document.head.appendChild(s);
}
function pdfViewerRenderPage(num: number) {
  const doc = pdfViewerDoc; if (!doc) return;
  const canvas = document.getElementById('pdfViewerCanvas') as HTMLCanvasElement; if (!canvas) return;
  const ctx = canvas.getContext('2d'); if (!ctx) return;
  doc.getPage(num).then((page: any) => {
    const viewport = page.getViewport({ scale: pdfViewerScale });
    canvas.width = viewport.width; canvas.height = viewport.height; canvas.style.display = 'block';
    const loading = document.getElementById('pdfViewerLoading'); if (loading) loading.style.display = 'none';
    const err = document.getElementById('pdfViewerError'); if (err) err.style.display = 'none';
    page.render({ canvasContext: ctx, viewport: viewport });
    const info = document.getElementById('pdfViewerPageInfo'); if (info) info.textContent = num + ' / ' + doc.numPages;
    const prev = document.getElementById('pdfViewerPrevBtn') as HTMLButtonElement; const next = document.getElementById('pdfViewerNextBtn') as HTMLButtonElement;
    if (prev) prev.disabled = num <= 1; if (next) next.disabled = num >= doc.numPages;
    const zoom = document.getElementById('pdfViewerZoomLevel'); if (zoom) zoom.textContent = Math.round(pdfViewerScale * 100) + '%';
  });
}
function pdfViewerShowError(msg: string) {
  const el = document.getElementById('pdfViewerError'); if (el) { el.textContent = msg; el.style.display = 'block'; }
  const loading = document.getElementById('pdfViewerLoading'); if (loading) loading.style.display = 'none';
}
function pdfViewerFilenameFromUrl(url: string): string {
  const parts = url.split('/'); const last = parts[parts.length - 1] || 'document.pdf'; return decodeURIComponent(last);
}
expose('openPdfViewer', function (url: string, title?: string) {
  pdfViewerUrl = url; pdfViewerTitle = title || pdfViewerFilenameFromUrl(url);
  const modalEl = document.getElementById('pdfViewerModal'); if (!modalEl) return;
  document.getElementById('pdfViewerTitle')!.textContent = pdfViewerTitle;
  const loading = document.getElementById('pdfViewerLoading');
  if (loading) { loading.style.display = 'block'; loading.innerHTML = '<div class="spinner-border text-light mb-2" role="status"></div><p class="mb-0">Loading PDF...</p>'; }
  const canvas = document.getElementById('pdfViewerCanvas') as HTMLCanvasElement; if (canvas) canvas.style.display = 'none';
  const err = document.getElementById('pdfViewerError'); if (err) err.style.display = 'none';
  const info = document.getElementById('pdfViewerPageInfo'); if (info) info.textContent = '- / -';
  const prev = document.getElementById('pdfViewerPrevBtn') as HTMLButtonElement; const next = document.getElementById('pdfViewerNextBtn') as HTMLButtonElement;
  if (prev) prev.disabled = true; if (next) next.disabled = true;
  const zoom = document.getElementById('pdfViewerZoomLevel'); if (zoom) zoom.textContent = '100%';
  const modal = (window as any).bootstrap.Modal.getOrCreateInstance(modalEl); modal.show();
  pdfViewerScale = 1.5; pdfViewerPage = 1;
  if (pdfViewerDoc) { pdfViewerDoc.destroy(); pdfViewerDoc = null; }
  pdfViewerLoadLib(() => {
    (window as any).pdfjsLib.getDocument(url).promise.then((doc: any) => { pdfViewerDoc = doc; pdfViewerRenderPage(1); }).catch((err: any) => { pdfViewerShowError('Failed to load PDF: ' + (err.message || 'Unknown error')); });
  });
});
expose('pdfViewerPrevPage', function () { if (!pdfViewerDoc || pdfViewerPage <= 1) return; pdfViewerPage--; pdfViewerRenderPage(pdfViewerPage); });
expose('pdfViewerNextPage', function () { if (!pdfViewerDoc || pdfViewerPage >= pdfViewerDoc.numPages) return; pdfViewerPage++; pdfViewerRenderPage(pdfViewerPage); });
expose('pdfViewerZoomIn', function () { pdfViewerScale = Math.min(pdfViewerScale * 1.25, 5); if (pdfViewerDoc) pdfViewerRenderPage(pdfViewerPage); });
expose('pdfViewerZoomOut', function () { pdfViewerScale = Math.max(pdfViewerScale / 1.25, 0.25); if (pdfViewerDoc) pdfViewerRenderPage(pdfViewerPage); });
expose('pdfViewerFitToWidth', function () {
  if (!pdfViewerDoc) return; const canvas = document.getElementById('pdfViewerCanvas') as HTMLCanvasElement; if (!canvas) return;
  const container = canvas.parentElement; if (!container) return; const cw = container.clientWidth - 40;
  pdfViewerDoc.getPage(pdfViewerPage).then((page: any) => { const ov = page.getViewport({ scale: 1 }); pdfViewerScale = cw / ov.width; pdfViewerRenderPage(pdfViewerPage); });
});
expose('pdfViewerDownload', function () {
  if (pdfViewerUrl) { const a = document.createElement('a'); a.href = pdfViewerUrl; a.download = pdfViewerTitle.replace(/[^a-zA-Z0-9._-]/g, '_') + '.pdf'; document.body.appendChild(a); a.click(); document.body.removeChild(a); }
});
document.addEventListener('hidden.bs.modal', function (e: Event) {
  const target = e.target as HTMLElement;
  if (target && target.id === 'pdfViewerModal') { if (pdfViewerDoc) { pdfViewerDoc.destroy(); pdfViewerDoc = null; } pdfViewerPage = 1; pdfViewerScale = 1.5; }
});
document.addEventListener('keydown', function (e: KeyboardEvent) {
  const modalEl = document.getElementById('pdfViewerModal');
  if (modalEl && modalEl.classList.contains('show')) {
    if (e.key === 'ArrowLeft') { (window as any).pdfViewerPrevPage(); e.preventDefault(); }
    else if (e.key === 'ArrowRight') { (window as any).pdfViewerNextPage(); e.preventDefault(); }
  }
});
