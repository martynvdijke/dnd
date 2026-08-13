/**
 * FilePicker — reusable modal for selecting an uploaded image or uploading a new one.
 *
 * Usage:
 *   import { FilePicker } from './file-picker';
 *   const url = await FilePicker.pick();
 *   // => "/media/abc123.jpg"
 *
 * The modal shows existing uploads in a grid. Click one to select it.
 * Drag-and-drop an image onto the modal to upload it, then auto-select.
 */
import { esc } from './lib/dom';
import { getCsrfToken } from './lib/api';

export class FilePicker {
  private static modalEl: HTMLElement | null = null;
  private static resolve: ((url: string) => void) | null = null;
  private static reject: ((err: Error) => void) | null = null;

  static pick(): Promise<string> {
    return new Promise((resolve, reject) => {
      this.resolve = resolve;
      this.reject = reject;
      this.open();
    });
  }

  private static async open() {
    // Create modal if not exists
    if (!this.modalEl) {
      this.createModal();
    }

    // Show modal
    this.modalEl!.classList.remove('d-none');
    document.body.classList.add('modal-open');

    // Load uploads
    const backdrop = document.getElementById('fpBackdrop')!;
    backdrop.classList.remove('d-none');

    const grid = document.getElementById('fpGrid')!;
    grid.innerHTML = '<div class="text-muted p-4 text-center">Loading...</div>';

    try {
      const res = await fetch('/api/uploads', {
        credentials: 'include',
      });
      const uploads = await res.json();
      this.renderGrid(uploads);
    } catch (e: any) {
      grid.innerHTML = '<div class="text-danger p-4 text-center">Failed to load images</div>';
    }
  }

  private static createModal() {
    const el = document.createElement('div');
    el.id = 'filePickerModal';
    el.className = 'fp-modal';
    el.innerHTML = `
      <div id="fpBackdrop" class="fp-backdrop"></div>
      <div class="fp-dialog">
        <div class="fp-header">
          <h5 class="fp-title">Choose Image</h5>
          <button type="button" class="btn-close" id="fpCloseBtn" aria-label="Close"></button>
        </div>
        <div class="fp-body">
          <div class="fp-drop-zone" id="fpDropZone">
            <p class="fp-drop-text"><i class="fa-solid fa-cloud-arrow-up fa-2x mb-2"></i><br>Drop an image here to upload</p>
            <input type="file" id="fpFileInput" accept="image/*" class="d-none">
          </div>
          <div class="fp-grid" id="fpGrid">
            <div class="text-muted p-4 text-center">Loading...</div>
          </div>
        </div>
      </div>
    `;
    document.body.appendChild(el);
    this.modalEl = el;

    // Event listeners
    document.getElementById('fpCloseBtn')!.addEventListener('click', () => this.close());
    document.getElementById('fpBackdrop')!.addEventListener('click', () => this.close());

    // Drag-and-drop
    const dropZone = document.getElementById('fpDropZone')!;
    const fileInput = document.getElementById('fpFileInput') as HTMLInputElement;

    // Click on drop zone opens file picker
    dropZone.addEventListener('click', (e) => {
      if (e.target === dropZone || (e.target as HTMLElement).closest('.fp-drop-text')) {
        fileInput.click();
      }
    });

    fileInput.addEventListener('change', () => {
      if (fileInput.files && fileInput.files[0]) {
        this.uploadFile(fileInput.files[0]);
      }
    });

    dropZone.addEventListener('dragover', (e) => {
      e.preventDefault();
      dropZone.classList.add('fp-drop-zone-active');
    });

    dropZone.addEventListener('dragleave', () => {
      dropZone.classList.remove('fp-drop-zone-active');
    });

    dropZone.addEventListener('drop', (e) => {
      e.preventDefault();
      dropZone.classList.remove('fp-drop-zone-active');
      if (e.dataTransfer?.files && e.dataTransfer.files[0]) {
        this.uploadFile(e.dataTransfer.files[0]);
      }
    });

    // Close on escape
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && this.modalEl && !this.modalEl.classList.contains('d-none')) {
        this.close();
      }
    });
  }

  private static renderGrid(uploads: any[]) {
    const grid = document.getElementById('fpGrid')!;

    if (!uploads || uploads.length === 0) {
      grid.innerHTML = '<div class="text-muted p-4 text-center">No images yet. Drop one above!</div>';
      return;
    }

    grid.innerHTML = uploads.map((u) => {
      const thumbUrl = u.thumbnail_url || u.resized_url || u.url;
      return `
        <div class="fp-item" data-url="${esc(u.url)}">
          <img src="${esc(thumbUrl)}" alt="" loading="lazy">
          <div class="fp-check"><i class="fa-solid fa-check"></i></div>
          <button class="fp-share" data-id="${esc(String(u.id))}" data-url="${esc(u.url)}" title="Share file" onclick="event.stopPropagation();shareEntity('upload', ${Number(u.id)})"><i class="fa-solid fa-share-nodes"></i></button>
        </div>
      `;
    }).join('');

    // Click handler (delegation)
    grid.querySelectorAll('.fp-item').forEach((item) => {
      item.addEventListener('click', () => {
        const url = item.getAttribute('data-url') || '';
        this.select(url);
      });
    });
  }

  private static async uploadFile(file: File) {
    if (!file.type.startsWith('image/')) {
      return;
    }

    const dropZone = document.getElementById('fpDropZone')!;
    dropZone.classList.add('fp-uploading');

    try {
      const form = new FormData();
      form.append('image', file);
      const res = await fetch('/api/upload', {
        method: 'POST',
        headers: { 'X-CSRF-Token': getCsrfToken() },
        credentials: 'include',
        body: form,
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Upload failed');

      // Auto-select the uploaded image
      this.select(data.url);
    } catch (e: any) {
      dropZone.classList.remove('fp-uploading');
      // Re-fetch uploads to show current state
      const grid = document.getElementById('fpGrid')!;
      grid.innerHTML = '<div class="text-danger p-4 text-center">Upload failed. Try again.</div>';
      try {
        const res = await fetch('/api/uploads', { credentials: 'include' });
        const uploads = await res.json();
        this.renderGrid(uploads);
      } catch (_) { /* ignore */ }
    }
  }

  private static select(url: string) {
    this.close();
    if (this.resolve) {
      this.resolve(url);
    }
  }

  static close() {
    if (this.modalEl) {
      this.modalEl.classList.add('d-none');
    }
    document.body.classList.remove('modal-open');
    const backdrop = document.getElementById('fpBackdrop');
    if (backdrop) {
      backdrop.classList.add('d-none');
    }
    // Clean up file input
    const fileInput = document.getElementById('fpFileInput') as HTMLInputElement;
    if (fileInput) {
      fileInput.value = '';
    }
    // Reset drop zone
    const dropZone = document.getElementById('fpDropZone');
    if (dropZone) {
      dropZone.classList.remove('fp-uploading', 'fp-drop-zone-active');
    }
  }
}
