import http from 'node:http';

export type SentMessage = { chat_id: number; text: string; method: string; body: any };

/**
 * Fields of a Bot API request. go-telegram/bot always sends multipart/form-data,
 * but the JSON fallback is kept so older payloads keep working.
 */
async function parseRequestFields(req: http.IncomingMessage, body: Buffer): Promise<Record<string, string>> {
  const contentType = req.headers['content-type'] || '';
  const fields: Record<string, string> = {};
  if (contentType.includes('multipart/form-data')) {
    const request = new Request('http://localhost/', {
      method: 'POST',
      headers: { 'content-type': contentType },
      body,
    });
    const form = await request.formData();
    for (const [key, value] of form.entries()) {
      const str = typeof value === 'string' ? value : `[file:${value.name || 'upload'}]`;
      if (fields[key] !== undefined) {
        // Preserve repeated keys (e.g. multiple `media` parts in sendMediaGroup)
        // by accumulating as a JSON array string.
        const existing = fields[key];
        try {
          const arr = JSON.parse(existing);
          if (Array.isArray(arr)) {
            arr.push(str);
            fields[key] = JSON.stringify(arr);
            continue;
          }
        } catch {
          // existing is not a JSON array
        }
        fields[key] = JSON.stringify([existing, str]);
      } else {
        fields[key] = str;
      }
    }
    return fields;
  }
  if (body.length > 0) {
    try {
      const parsed = JSON.parse(body.toString('utf8'));
      for (const [key, value] of Object.entries(parsed || {})) {
        fields[key] = typeof value === 'string' ? value : JSON.stringify(value);
      }
    } catch {
      // Not JSON and not multipart: leave fields empty.
    }
  }
  return fields;
}

function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export class TelegramMock {
  server: http.Server | null = null;
  url = '';
  sentMessages: SentMessage[] = [];
  /** Commands registered through setMyCommands. */
  commands: { command: string; description: string }[] = [];
  webhookUrl: string | null = null;
  webhookSecret: string | null = null;
  botToken: string;
  botUsername: string;
  /** Longest delay applied to an empty getUpdates long poll. */
  pollDelayMs = 1250;
  inlineAnswers: { query_id: string; results: any[]; button: any | null }[] = [];

  /** Fixture bytes served at GET /file/bot<token>/<file_path>. Keyed by file_path. */
  fileFixtures: Map<string, Buffer> = new Map();
  /** Default bytes returned when no explicit fixture is set for a file_path. */
  defaultFileBytes: Buffer = Buffer.from('test-file-bytes');
  /** Mapping file_id -> file_path for getFile responses (configurable). */
  filePathById: Map<string, string> = new Map();
  /** Mapping file_id -> file_unique_id for getFile responses. */
  fileUniqueIdById: Map<string, string> = new Map();

  constructor(opts?: { token?: string; username?: string }) {
    this.botToken = opts?.token || 'test-bot-token';
    this.botUsername = opts?.username || 'testvillumbot';
  }

  clear() {
    this.sentMessages = [];
    this.inlineAnswers = [];
  }

  /** Set the bytes that GET /file/bot<token>/<path> will return. */
  setFileFixture(filePath: string, data: Buffer | string): void {
    this.fileFixtures.set(filePath, typeof data === 'string' ? Buffer.from(data) : data);
  }

  /** Alias for setFileFixture for convenience. */
  setFileBytes(filePath: string, data: Buffer | string): void {
    this.setFileFixture(filePath, data);
  }

  /** Configure the getFile response for a given file_id. */
  setFileMeta(fileId: string, filePath: string, fileUniqueId?: string): void {
    this.filePathById.set(fileId, filePath);
    if (fileUniqueId) this.fileUniqueIdById.set(fileId, fileUniqueId);
  }

  /** Override the default file bytes used for unregistered paths. */
  setDefaultFileBytes(data: Buffer | string): void {
    this.defaultFileBytes = typeof data === 'string' ? Buffer.from(data) : data;
  }

  async start(): Promise<string> {
    return new Promise((resolve, reject) => {
      const srv = http.createServer((req, res) => {
        const chunks: Buffer[] = [];
        req.on('data', (c: Buffer) => chunks.push(c));
        req.on('end', async () => {
          const u = new URL(req.url || '/', 'http://localhost');
          const pathname = u.pathname;

          // Serve GET /file/bot<token>/<file_path> with fixture bytes.
          if (req.method === 'GET' && pathname.startsWith('/file/bot')) {
            // pathname is /file/bot<token>/<file_path>
            const prefix = `/file/bot${this.botToken}/`;
            const altPrefix = `/file/bot`;
            let filePath = '';
            if (pathname.startsWith(prefix)) {
              filePath = pathname.slice(prefix.length);
            } else if (pathname.startsWith(altPrefix + '/')) {
              // Fallback: extract after /file/bot.../  (tolerate token mismatch)
              const after = pathname.slice('/file/'.length);
              const slashIdx = after.indexOf('/');
              if (slashIdx !== -1) filePath = after.slice(slashIdx + 1);
            }
            // Strip leading slash if any
            filePath = filePath.replace(/^\//, '');
            const bytes = this.fileFixtures.get(filePath) ?? this.defaultFileBytes;
            res.writeHead(200, { 'Content-Type': 'application/octet-stream', 'Content-Length': String(bytes.length) });
            res.end(bytes);
            return;
          }

          // Match /bot<token>/<method>
          const m = pathname.match(/^\/bot[^/]+\/(.+)$/);
          const method = m ? m[1] : pathname.replace(/^\//, '');
          const fields = await parseRequestFields(req, Buffer.concat(chunks));
          const num = (v: string | undefined) => Number(v ?? 0);
          const json = (obj: any) => {
            res.writeHead(200, { 'Content-Type': 'application/json' });
            res.end(JSON.stringify(obj));
          };

          if (method === 'getMe') {
            return json({ ok: true, result: { id: 123456, is_bot: true, first_name: 'TestBot', username: this.botUsername } });
          }
          if (method === 'sendMessage') {
            this.sentMessages.push({ chat_id: num(fields.chat_id), text: fields.text ?? '', method: 'sendMessage', body: fields });
            return json({ ok: true, result: { message_id: this.sentMessages.length } });
          }
          // Single-media senders: photo/audio/video/document
          if (method === 'sendPhoto' || method === 'sendAudio' || method === 'sendVideo' || method === 'sendDocument') {
            const caption = fields.caption ?? '';
            // Determine text preview: caption if present, else file placeholder
            const fileKey = method === 'sendPhoto' ? 'photo' : method === 'sendAudio' ? 'audio' : method === 'sendVideo' ? 'video' : 'document';
            const fileVal = fields[fileKey] ?? fields.document ?? fields.photo ?? fields.audio ?? fields.video ?? '';
            // text is caption when available, otherwise the file marker
            const text = caption || fileVal || `[${method}]`;
            this.sentMessages.push({ chat_id: num(fields.chat_id), text, method, body: fields });
            return json({ ok: true, result: { message_id: this.sentMessages.length } });
          }
          if (method === 'sendMediaGroup') {
            // Parse multipart bodies for sendMediaGroup in detail: capture
            // chat_id, number of media JSON parts, each part's type/attach ref/caption,
            // and the uploaded binary parts (file0…).
            // parseRequestFields already collapsed duplicate `media` keys; we re-parse
            // the original body to get accurate arrays.
            let mediaJsonParts: string[] = [];
            let uploadedFiles: string[] = [];
            const contentType = req.headers['content-type'] || '';
            if (contentType.includes('multipart/form-data')) {
              try {
                const rawBody = Buffer.concat(chunks);
                const request = new Request('http://localhost/', {
                  method: 'POST',
                  headers: { 'content-type': contentType },
                  body: rawBody,
                });
                const form = await request.formData();
                // form.getAll available in Node 18+
                const allMedia: string[] = [];
                // Use getAll if present, otherwise iterate
                if (typeof (form as any).getAll === 'function') {
                  const vals = (form as any).getAll('media') as unknown[];
                  for (const v of vals) if (typeof v === 'string') allMedia.push(v);
                } else {
                  for (const [k, v] of form.entries()) if (k === 'media' && typeof v === 'string') allMedia.push(v);
                }
                mediaJsonParts = allMedia;
                for (const [k, v] of form.entries()) {
                  if (k.startsWith('file')) {
                    const name = typeof v === 'string' ? v : (v as File).name || k;
                    uploadedFiles.push(typeof v === 'string' ? k : `${k}:${name}`);
                  }
                }
              } catch {
                // fall back to fields-based parsing
              }
            }
            // Fallback when not multipart or parsing failed: derive from fields
            if (mediaJsonParts.length === 0 && fields.media) {
              try {
                const parsed = JSON.parse(fields.media);
                if (Array.isArray(parsed)) mediaJsonParts = parsed.map((x: any) => JSON.stringify(x));
                else mediaJsonParts = [fields.media];
              } catch {
                mediaJsonParts = [fields.media];
              }
            }
            if (uploadedFiles.length === 0) {
              for (const k of Object.keys(fields)) if (k.startsWith('file')) uploadedFiles.push(k);
            }

            // Build a structured body so specs can assert counts/types/refs/captions/files.
            // Go's SendMediaGroup sends `media` as a single JSON array string with N objects; expand it.
            let expandedParts: string[] = [];
            for (const raw of mediaJsonParts) {
              try {
                const parsed = JSON.parse(raw);
                if (Array.isArray(parsed)) {
                  for (const el of parsed) expandedParts.push(JSON.stringify(el));
                  continue;
                }
              } catch {}
              expandedParts.push(raw);
            }
            // If we expanded, use expanded for entries/counts
            const effectiveParts = expandedParts.length > 0 ? expandedParts : mediaJsonParts;
            let mediaEntries: any[] = [];
            for (const raw of effectiveParts) {
              try {
                mediaEntries.push(JSON.parse(raw));
              } catch {
                mediaEntries.push({ _raw: raw });
              }
            }

            const chatId = num(fields.chat_id);
            // caption for the group is embedded per-item per Go code; expose first caption as text
            const firstCaption = mediaEntries[0]?.caption ?? '';
            const enrichedBody: Record<string, any> = {
              ...fields,
              _mediaJsonParts: effectiveParts,
              _mediaEntries: mediaEntries,
              _uploadedFiles: uploadedFiles,
              _mediaCount: effectiveParts.length,
              _fileCount: uploadedFiles.length,
            };

            this.sentMessages.push({
              chat_id: chatId,
              text: firstCaption || `[mediaGroup:${effectiveParts.length}]`,
              method: 'sendMediaGroup',
              body: enrichedBody,
            });
            // Telegram returns an array of messages for media groups
            const result = effectiveParts.map((_, i) => ({ message_id: this.sentMessages.length + i }));
            return json({ ok: true, result });
          }
          if (method === 'getFile') {
            const fileId = fields.file_id || (fields as any).fileId || '';
            let filePath = this.filePathById.get(fileId) || '';
            if (!filePath) {
              // Default: files/<fileId>/file.bin — matches GetAPIBase + /file/bot<token>/<file_path>
              filePath = `files/${fileId || 'unknown'}/file.bin`;
            }
            const fileUniqueId = this.fileUniqueIdById.get(fileId) || `unique_${fileId || 'unknown'}`;
            // Ensure a fixture exists for this path so the subsequent GET succeeds
            if (!this.fileFixtures.has(filePath)) {
              // Keep default bytes available at this path
              this.fileFixtures.set(filePath, this.defaultFileBytes);
            }
            return json({ ok: true, result: { file_id: fileId, file_unique_id: fileUniqueId, file_path: filePath } });
          }
          if (method === 'getChat') {
            const cid = num(fields.chat_id);
            return json({ ok: true, result: { id: cid, type: cid < 0 ? 'supergroup' : 'private', title: `Chat ${cid}`, username: '' } });
          }
          if (method === 'getUpdates') {
            // Honor the long-poll timeout so the bot does not busy-loop, but
            // keep it short so test shutdown stays fast.
            const timeoutSec = num(fields.timeout);
            const wait = Math.min(Math.max(timeoutSec * 1000, 1000), this.pollDelayMs);
            await delay(wait);
            return json({ ok: true, result: [] });
          }
          if (method === 'setMyCommands') {
            try {
              this.commands = JSON.parse(fields.commands || '[]');
            } catch {
              this.commands = [];
            }
            return json({ ok: true, result: true });
          }
          if (method === 'answerCallbackQuery') {
            return json({ ok: true, result: true });
          }
          if (method === 'answerInlineQuery') {
            const qid = fields.inline_query_id || '';
            let results: any[] = [];
            try {
              results = JSON.parse(fields.results || '[]');
              if (!Array.isArray(results)) results = [];
            } catch {
              results = [];
            }
            let button: any | null = null;
            if (fields.button) {
              try {
                button = JSON.parse(fields.button);
              } catch {
                button = null;
              }
            }
            this.inlineAnswers.push({ query_id: qid, results, button });
            return json({ ok: true, result: true });
          }
          if (method === 'setWebhook') {
            this.webhookUrl = fields.url || null;
            this.webhookSecret = fields.secret_token || null;
            return json({ ok: true, result: true, description: 'Webhook was set' });
          }
          if (method === 'deleteWebhook') {
            this.webhookUrl = null;
            this.webhookSecret = null;
            return json({ ok: true, result: true });
          }
          return json({ ok: true, result: {} });
        });
      });
      srv.listen(0, '127.0.0.1', () => {
        const addr = srv.address() as any;
        this.url = `http://127.0.0.1:${addr.port}`;
        this.server = srv;
        resolve(this.url);
      });
      srv.on('error', reject);
    });
  }

  async stop(): Promise<void> {
    if (!this.server) return;
    const srv = this.server;
    srv.closeAllConnections();
    await new Promise<void>((r) => srv.close(() => r()));
    this.server = null;
  }
}
