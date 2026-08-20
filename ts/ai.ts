/**
 * AI Generation module — modal for generating text and images via AI endpoints.
 */
import { esc, toast, showModal, hideModal } from './lib/dom';
import { api } from './lib/api';
import { expose } from './lib/expose';

// ─── State ───

export let aiGenLastResult: string | null = null;
export let aiGenLastImageUrl: string | null = null;

// ─── Feature toggle ───
// Mirrors the backend site-wide AI setting. When false, all AI UI is
// suppressed (buttons hidden via the `ai-disabled` body class) and the
// generation modal refuses to open.
export let aiEnabled = true;

export function setAIEnabled(v: boolean): void {
  aiEnabled = v;
  document.body.classList.toggle('ai-disabled', !v);
}

const AI_DEFAULT_SYSTEM_PROMPT = 'You are a helpful assistant for a D&D website called villum. You help DMs create compelling narratives, NPCs, locations, items, and other TTRPG content. Be creative and concise.';

// ─── Modal ───

// Show the AI modal, queueing the show until any in-flight hide transition finishes
// (bootstrap silently swallows show() during a hide transition).
function aiShowModal(): void {
  const el = document.getElementById('aiGenModal')!;
  const bootstrap = (window as any).bootstrap;
  if (el.classList.contains('show')) return;
  const modal = bootstrap.Modal.getOrCreateInstance(el);
  if (modal._isTransitioning) {
    el.addEventListener('hidden.bs.modal', function onHidden() {
      el.removeEventListener('hidden.bs.modal', onHidden);
      bootstrap.Modal.getOrCreateInstance(el).show();
    });
    return;
  }
  modal.show();
}

export function openAIGenModal(mode: string, targetId: string, promptHint?: string, title?: string) {
  if (!aiEnabled) {
    toast('AI features are disabled', true);
    return;
  }
  aiGenLastResult = null;
  aiGenLastImageUrl = null;
  document.getElementById('aiGenResult')!.style.display = 'none';

  (document.getElementById('aiGenTargetId') as HTMLInputElement).value = targetId;
  (document.getElementById('aiGenMode') as HTMLInputElement).value = mode;
  (document.getElementById('aiGenModal') as any).querySelector('.modal-title')!.textContent = title || (mode === 'image' ? 'Generate Image with AI' : 'Generate Text with AI');

  const promptEl = document.getElementById('aiGenPrompt') as HTMLTextAreaElement;
  if (promptHint) promptEl.value = promptHint;

  // Pre-fill system prompt with default if empty
  const systemEl = document.getElementById('aiGenSystem') as HTMLTextAreaElement;
  if (!systemEl.value.trim()) {
    systemEl.value = AI_DEFAULT_SYSTEM_PROMPT;
  }

  // Show/hide fields based on mode
  document.getElementById('aiGenSystemField')!.style.display = mode === 'text' ? 'block' : 'none';
  document.getElementById('aiGenInsertBtn')!.style.display = mode === 'text' ? 'inline-block' : 'none';
  document.getElementById('aiGenReplaceBtn')!.style.display = mode === 'text' ? 'inline-block' : 'none';
  document.getElementById('aiGenSaveBtn')!.style.display = mode === 'image' ? 'inline-block' : 'none';

  aiShowModal();

  // Fetch endpoints for this mode
  fetchAIEndpoints(mode);
}

async function fetchAIEndpoints(type: string) {
  const select = document.getElementById('aiGenEndpoint') as HTMLSelectElement;
  try {
    const eps = await api('GET', '/api/ai/endpoints?type=' + encodeURIComponent(type));
    if (!eps || eps.length === 0) {
      select.innerHTML = '<option value="">No enabled endpoints available</option>';
      return;
    }
    select.innerHTML = eps.map((ep: any) =>
      `<option value="${ep.id}">${esc(ep.name)} (${ep.model})</option>`
    ).join('');
  } catch {
    select.innerHTML = '<option value="">Failed to load endpoints</option>';
  }
}

// ─── Inline AI generation triggers ───

export function initAIClickHandler() {
  document.addEventListener('click', function (e: MouseEvent) {
    if (!aiEnabled) return;
    const btn = (e.target as HTMLElement).closest('.ai-generate-btn') as HTMLElement;
    if (!btn) return;
    e.preventDefault();
    const mode = btn.getAttribute('data-ai-mode') || 'text';
    const targetId = btn.getAttribute('data-ai-target') || '';
    const hint = btn.getAttribute('data-ai-hint') || '';
    const title = btn.getAttribute('data-ai-title') || undefined;
    openAIGenModal(mode, targetId, hint, title);
  });
}

export async function generateWithAI() {
  const endpointId = parseInt((document.getElementById('aiGenEndpoint') as HTMLSelectElement).value);
  const prompt = (document.getElementById('aiGenPrompt') as HTMLTextAreaElement).value.trim();
  const mode = (document.getElementById('aiGenMode') as HTMLInputElement).value;

  if (!endpointId) { toast('Please select an AI endpoint', true); return; }
  if (!prompt) { toast('Please enter a prompt', true); return; }

  const btn = document.getElementById('aiGenGenerateBtn') as HTMLButtonElement;
  btn.disabled = true;
  btn.innerHTML = '<i class="fa-solid fa-spinner fa-spin me-1"></i>'
    + (mode === 'image' ? 'Generating image...' : 'Generating...');

  try {
    if (mode === 'text') {
      const system = (document.getElementById('aiGenSystem') as HTMLTextAreaElement).value.trim();
      const result: any = await api('POST', '/api/ai/generate/text', {
        endpoint_id: endpointId,
        prompt: prompt,
        system: system || undefined,
      });
      aiGenLastResult = result.text;
      document.getElementById('aiGenResultText')!.textContent = result.text;
      document.getElementById('aiGenResultText')!.style.display = 'block';
      document.getElementById('aiGenResultImage')!.style.display = 'none';
    } else {
      const result: any = await api('POST', '/api/ai/generate/image', {
        endpoint_id: endpointId,
        prompt: prompt,
      });
      const imageUrl = result.images?.[0] || result.image_url || '';
      aiGenLastImageUrl = imageUrl;
      const img = document.getElementById('aiGenResultImage') as HTMLImageElement;
      img.src = imageUrl;
      img.style.display = 'block';
      document.getElementById('aiGenResultText')!.style.display = 'none';
      if (!imageUrl) {
        toast('Image generated but no URL returned', true);
      }
    }
    document.getElementById('aiGenResult')!.style.display = 'block';
  } catch (e: any) {
    toast(e.message || 'Generation failed', true);
  } finally {
    btn.disabled = false;
    btn.innerHTML = '<i class="fa-solid fa-wand-magic-sparkles me-1"></i>'
      + (mode === 'image' ? 'Generate Image' : 'Generate');
  }
}

export function regenerateWithAI() {
  generateWithAI();
}

export function insertAIGenResult() {
  const targetId = (document.getElementById('aiGenTargetId') as HTMLInputElement).value;
  if (!targetId || !aiGenLastResult) return;
  const target = document.getElementById(targetId) as HTMLTextAreaElement | HTMLInputElement;
  if (target) {
    target.value += (target.value ? '\n' : '') + aiGenLastResult;
    const modal = (window as any).bootstrap.Modal.getInstance(document.getElementById('aiGenModal')!);
    if (modal) modal.hide();
    toast('Text inserted');
  }
}

export function replaceAIGenResult() {
  const targetId = (document.getElementById('aiGenTargetId') as HTMLInputElement).value;
  if (!targetId || !aiGenLastResult) return;
  const target = document.getElementById(targetId) as HTMLTextAreaElement | HTMLInputElement;
  if (target) {
    target.value = aiGenLastResult;
    const modal = (window as any).bootstrap.Modal.getInstance(document.getElementById('aiGenModal')!);
    if (modal) modal.hide();
    toast('Text replaced');
  }
}

export async function saveAIGenImage() {
  if (!aiGenLastImageUrl) { toast('No image to save — generate one first', true); return; }
  const targetId = (document.getElementById('aiGenTargetId') as HTMLInputElement).value;
  try {
    const resp: any = await api('POST', '/api/ai/save-image', { url: aiGenLastImageUrl });
    if (!resp || !resp.url) { toast('Save failed', true); return; }
    if (targetId) {
      const target = document.getElementById(targetId) as HTMLTextAreaElement | HTMLInputElement;
      if (target) target.value = resp.url;
    }
    const modal = (window as any).bootstrap.Modal.getInstance(document.getElementById('aiGenModal')!);
    if (modal) modal.hide();
    toast('Image saved to library');
  } catch (e: any) {
    toast(e.message || 'Save failed', true);
  }
}

// ─── Window registration (called from inline HTML onclick handlers) ───

expose('openAIGenModal', openAIGenModal);
expose('generateWithAI', generateWithAI);
expose('regenerateWithAI', regenerateWithAI);
expose('insertAIGenResult', insertAIGenResult);
expose('replaceAIGenResult', replaceAIGenResult);
expose('saveAIGenImage', saveAIGenImage);

// ─── Campaign Dashboard Helper ───

// (window as any).openCampaignDashboard already calls showCampaignDashboard
// which is still in app.ts — keep it there
