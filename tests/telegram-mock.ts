import http from 'node:http';

export type SentMessage = { chat_id: number; text: string; method: string; body: any };

export class TelegramMock {
  server: http.Server | null = null;
  url = '';
  sentMessages: SentMessage[] = [];
  webhookUrl: string | null = null;
  webhookSecret: string | null = null;
  botToken: string;
  botUsername: string;

  constructor(opts?: { token?: string; username?: string }) {
    this.botToken = opts?.token || 'test-bot-token';
    this.botUsername = opts?.username || 'testvillumbot';
  }

  clear() {
    this.sentMessages = [];
  }

  async start(): Promise<string> {
    return new Promise((resolve, reject) => {
      const srv = http.createServer((req, res) => {
        let body = '';
        req.on('data', (c) => (body += c));
        req.on('end', () => {
          const u = new URL(req.url || '/', 'http://localhost');
          const pathname = u.pathname;
          // Match /bot<token>/<method>
          const m = pathname.match(/^\/bot[^/]+\/(.+)$/);
          const method = m ? m[1] : pathname.replace(/^\//, '');
          let parsed: any = {};
          try { parsed = body ? JSON.parse(body) : {}; } catch {}
          const json = (obj: any) => {
            res.writeHead(200, { 'Content-Type': 'application/json' });
            res.end(JSON.stringify(obj));
          };
          if (method === 'getMe') {
            return json({ ok: true, result: { id: 123456, is_bot: true, first_name: 'TestBot', username: this.botUsername } });
          }
          if (method === 'sendMessage') {
            this.sentMessages.push({ chat_id: parsed.chat_id, text: parsed.text, method: 'sendMessage', body: parsed });
            return json({ ok: true, result: { message_id: this.sentMessages.length } });
          }
          if (method === 'sendDocument') {
            // multipart handled elsewhere; treat as ok
            this.sentMessages.push({ chat_id: parsed.chat_id ?? 0, text: '[document]', method: 'sendDocument', body: parsed });
            return json({ ok: true, result: { message_id: this.sentMessages.length } });
          }
          if (method === 'getChat') {
            const cid = parsed.chat_id ?? 0;
            return json({ ok: true, result: { id: cid, type: cid < 0 ? 'supergroup' : 'private', title: `Chat ${cid}`, username: '' } });
          }
          if (method === 'getUpdates') {
            return json({ ok: true, result: [] });
          }
          if (method === 'setWebhook') {
            this.webhookUrl = parsed.url || null;
            this.webhookSecret = parsed.secret_token || null;
            return json({ ok: true, result: true, description: 'Webhook was set' });
          }
          if (method === 'deleteWebhook') {
            this.webhookUrl = null;
            this.webhookSecret = null;
            return json({ ok: true, result: true });
          }
          // Fallback: treat any multipart sendDocument as success too (raw body not JSON)
          if (req.headers['content-type']?.includes('multipart/form-data')) {
            this.sentMessages.push({ chat_id: 0, text: '[document-multipart]', method: 'sendDocument', body: {} });
            return json({ ok: true, result: { message_id: this.sentMessages.length } });
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
    await new Promise<void>((r) => this.server!.close(() => r()));
    this.server = null;
  }
}
