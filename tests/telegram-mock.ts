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
      fields[key] = typeof value === 'string' ? value : `[file:${value.name || 'upload'}]`;
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

  constructor(opts?: { token?: string; username?: string }) {
    this.botToken = opts?.token || 'test-bot-token';
    this.botUsername = opts?.username || 'testvillumbot';
  }

  clear() {
    this.sentMessages = [];
    this.inlineAnswers = [];
  }

  async start(): Promise<string> {
    return new Promise((resolve, reject) => {
      const srv = http.createServer((req, res) => {
        const chunks: Buffer[] = [];
        req.on('data', (c: Buffer) => chunks.push(c));
        req.on('end', async () => {
          const u = new URL(req.url || '/', 'http://localhost');
          const pathname = u.pathname;
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
          if (method === 'sendDocument') {
            this.sentMessages.push({ chat_id: num(fields.chat_id), text: '[document]', method: 'sendDocument', body: fields });
            return json({ ok: true, result: { message_id: this.sentMessages.length } });
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
