import { marked } from 'marked';
import DOMPurifyModule from 'dompurify';

let purifyInstance: any = null;
let hookInstalled = false;

const ALLOWED_TAGS = [
  'a', 'p', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
  'ul', 'ol', 'li', 'code', 'pre', 'blockquote',
  'table', 'thead', 'tbody', 'tr', 'th', 'td',
  'img', 'strong', 'em', 'b', 'i', 'hr', 'br', 'span', 'div',
];

const ALLOWED_ATTR = ['href', 'src', 'alt', 'title', 'colspan', 'rowspan'];

function getPurify(): any {
  if (purifyInstance) return purifyInstance;
  const m: any = DOMPurifyModule as any;
  if (typeof window !== 'undefined' && (window as any).document && typeof m === 'function') {
    try {
      const inst = m(window);
      if (inst && typeof inst.sanitize === 'function') {
        purifyInstance = inst;
        return purifyInstance;
      }
    } catch {}
  }
  if (m && typeof m.sanitize === 'function') {
    purifyInstance = m;
    return purifyInstance;
  }
  purifyInstance = m;
  return purifyInstance;
}

function ensureHook(): void {
  if (hookInstalled) return;
  hookInstalled = true;
  try {
    const purify = getPurify();
    purify.addHook('uponSanitizeAttribute', (_node: any, data: any) => {
      const name = (data.attrName || '').toLowerCase();
      const value = (data.attrValue || '').trim().toLowerCase();
      if (name.startsWith('on')) {
        data.keepAttr = false;
        return;
      }
      if (
        (name === 'href' || name === 'src' || name === 'xlink:href') &&
        (value.startsWith('javascript:') || value.startsWith('vbscript:') || value.startsWith('data:'))
      ) {
        data.keepAttr = false;
      }
    });
  } catch {}
}

function naiveSanitize(html: string): string {
  let out = html;
  // Remove forbidden tags and content
  out = out.replace(/<\s*(script|style|iframe|object|form|applet|embed|link|meta|base|noscript)[^>]*>[\s\S]*?<\s*\/\s*\1\s*>/gi, '');
  out = out.replace(/<\s*(script|style|iframe|object|form|applet|embed|link|meta|base|noscript)[^>]*\/?>/gi, '');
  // Remove on* attributes
  out = out.replace(/\s+on\w+\s*=\s*(?:"[^"]*"|'[^']*'|[^\s"'`>]+)/gi, '');
  // Remove javascript:/vbscript:/data: from href/src
  out = out.replace(/\s+(href|src|xlink:href)\s*=\s*("|\')\s*(javascript|vbscript|data):[^"']*("|\')/gi, '');
  out = out.replace(/\s+(href|src|xlink:href)\s*=\s*(javascript|vbscript|data):[^\s>]+/gi, '');
  // Remove style attributes
  out = out.replace(/\s+style\s*=\s*(?:"[^"]*"|'[^']*'|[^\s"'`>]+)/gi, '');
  return out;
}

function isPurifyWorking(): boolean {
  try {
    const p = getPurify();
    const probe = p.sanitize('<b>probe</b>', { ALLOWED_TAGS: ['b'], ALLOWED_ATTR: [] });
    return typeof probe === 'string' && probe.includes('<b>');
  } catch {
    return false;
  }
}

export function renderMarkdown(md: string): string {
  ensureHook();
  const raw = marked.parse(md) as string;
  if (!isPurifyWorking()) {
    return naiveSanitize(raw);
  }
  const purify = getPurify();
  return purify.sanitize(raw, {
    ALLOWED_TAGS,
    ALLOWED_ATTR,
    FORBID_TAGS: ['script', 'style', 'iframe', 'object', 'form', 'applet', 'embed', 'link', 'meta', 'base', 'noscript'],
    FORBID_ATTR: ['style'],
  }) as unknown as string;
}
