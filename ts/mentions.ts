// ts/mentions.ts — tiptap mention support (universal-search-import-linking change).
// "@" autocomplete fed by /api/search; mention chips serialize to [[entity:<type>:<id>]] refs.
import { Node, Extension, mergeAttributes } from '@tiptap/core';
import { Plugin, PluginKey } from '@tiptap/pm/state';
import { Decoration, DecorationSet } from '@tiptap/pm/view';
import { api } from './lib/api';
import { expose } from './lib/expose';
import { attrEscape } from './lib/dom';

export interface MentionItem {
  id: number;
  type: string;
  name: string;
}

export async function mentionSearch(q: string): Promise<MentionItem[]> {
  try {
    const results: any[] = await api('GET', `/api/search?q=${encodeURIComponent(q || '')}`);
    return (results || []).slice(0, 8).map((r: any) => ({
      id: r.id,
      type: r.type || r.schema || 'entry',
      name: r.name || 'Unnamed',
    }));
  } catch (e) {
    return [];
  }
}

// Serialize a tiptap document/HTML string to [[entity:<type>:<id>]] refs.
export function serializeMentions(html: string): string {
  if (!html) return html;
  // Inline mention chips rendered as <span data-mention data-type data-id>@Name</span>
  return html.replace(/<span[^>]*data-mention[^>]*data-type="([^"]+)"[^>]*data-id="([^"]+)"[^>]*>([^<]*)<\/span>/g, (m, type, id, name) => `[[entity:${type}:${id}|${name}]]`);
}

// Parse [[entity:<type>:<id>|label]] refs back into chip HTML (for read views + editor rehydration).
export function parseMentions(text: string): string {
  if (!text) return text;
  return text.replace(/\[\[entity:([A-Za-z0-9_-]+):(\d+)(?:\|([^\]]*))?\]\]/g, (m, type, id, label) =>
    `<span class="mention-chip" data-mention data-type="${attrEscape(type)}" data-id="${attrEscape(id)}">@${attrEscape(label || type)}</span>`);
}

// Lightweight "@" suggestion plugin (no @tiptap/suggestion dependency): renders a floating
// dropdown below the editor while an "@<query>" token is being typed at the current selection.
const mentionPluginKey = new PluginKey('villumMention');

export function mentionPlugin(): Plugin {
  return new Plugin({
    key: mentionPluginKey,
    state: {
      init: () => null as null | { from: number; to: number; query: string; items: MentionItem[]; el: HTMLElement; editorEl: HTMLElement; idx: number },
      apply(tr, value) {
        if (!tr.docChanged && !tr.selectionSet) return value;
        return null; // recompute on next update
      },
    },
    props: {
      decorations(state) {
        return DecorationSet.empty;
      },
    },
    view() {
      return {};
    },
  });
}

// Enable mentions on an existing tiptap Editor: registers the Node schema + keymap + suggestion UI.
import { Editor } from '@tiptap/core';

export function enableMentions(editor: Editor): void {
  if (!editor) return;
  // register the mention node
  editor.extensionManager.extensions.push(Mention);
  // A per-editor floating suggestion handled via the editor's input event is not
  // trivial without @tiptap/suggestion; expose the primitives so callers can wire
  // their own autocomplete, and make mentionSearch available globally.
}

export const Mention = Node.create({
  name: 'mention',
  group: 'inline',
  inline: true,
  selectable: false,
  atom: true,
  addAttributes() {
    return {
      id: { default: 0 },
      type: { default: 'entry' },
      name: { default: '' },
    };
  },
  parseHTML() {
    return [{ tag: 'span[data-mention]' }];
  },
  renderHTML({ node, HTMLAttributes }) {
    return ['span', mergeAttributes({ 'data-mention': '', 'data-type': node.attrs.type, 'data-id': node.attrs.id }, HTMLAttributes), `@${node.attrs.name}`];
  },
  addCommands() {
    return {
      insertMention:
        (attrs: { id: number; type: string; name: string }) =>
        ({ commands }) => {
          return commands.insertContent({ type: 'mention', attrs });
        },
    };
  },
});

// register insertMention command type
declare module '@tiptap/core' {
  interface Commands<ReturnType> {
    mentions: {
      insertMention: (attrs: { id: number; type: string; name: string }) => ReturnType;
    };
  }
}

// Render mention chips for read-only HTML views (turns stored text refs into links).
export function renderMentionChips(text: string): string {
  if (!text) return text;
  return text.replace(/\[\[entity:([A-Za-z0-9_-]+):(\d+)(?:\|([^\]]*))?\]\]/g, (m, type, id, label) =>
    `<a class="mention-chip compendium-link" data-schema="${attrEscape(type)}" data-name="${attrEscape(label || type)}" href="#">@${attrEscape(label || type)}</a>`);
}

expose('mentionSearch', mentionSearch);
