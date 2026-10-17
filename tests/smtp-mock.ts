import net from 'node:net';

export type SmtpMessage = { from: string; to: string[]; data: string; raw: string };

export class SmtpMock {
  port: number = 0;
  messages: SmtpMessage[] = [];
  private server: net.Server | null = null;

  constructor() {}

  start(): Promise<number> {
    return new Promise((resolve, reject) => {
      const srv = net.createServer((socket) => {
        socket.setEncoding('utf8');
        socket.write('220 127.0.0.1 ESMTP villum-test\r\n');

        let buffer = '';
        let from = '';
        let to: string[] = [];
        let dataMode = false;
        let dataLines: string[] = [];
        let authStep: 'none' | 'plain-await' | 'login-user' | 'login-pass' = 'none';

        function extractAddress(s: string): string {
          // Try <addr> first
          const m = s.match(/<([^>]+)>/);
          if (m) return m[1].trim();
          // fallback: after colon
          const idx = s.indexOf(':');
          if (idx !== -1) return s.slice(idx + 1).trim().replace(/^<|>$/g, '').trim();
          return s.trim();
        }

        function handleLine(rawLine: string) {
          const line = rawLine; // already stripped \r
          const upper = line.toUpperCase();

          // Auth continuation steps take precedence
          if (authStep === 'plain-await') {
            authStep = 'none';
            socket.write('235 2.7.0 Authentication successful\r\n');
            return;
          }
          if (authStep === 'login-user') {
            authStep = 'login-pass';
            socket.write('334 UGFzc3dvcmQ6\r\n');
            return;
          }
          if (authStep === 'login-pass') {
            authStep = 'none';
            socket.write('235 2.7.0 Authentication successful\r\n');
            return;
          }

          if (dataMode) {
            if (line === '.') {
              dataMode = false;
              const raw = dataLines.join('\r\n');
              const data = raw; // full content including headers
              // push message
              const msg: SmtpMessage = { from, to: [...to], data, raw };
              // store
              (srv as any).__smtpMockInstance.messages.push(msg);
              dataLines = [];
              socket.write('250 OK queued\r\n');
            } else {
              // Handle dot-stuffing: client escapes leading dot as ..
              let l = line;
              if (l.startsWith('..')) l = l.slice(1);
              dataLines.push(l);
            }
            return;
          }

          if (upper.startsWith('EHLO') || upper.startsWith('HELO')) {
            socket.write('250-127.0.0.1\r\n250-AUTH PLAIN LOGIN\r\n250 OK\r\n');
          } else if (upper.startsWith('AUTH PLAIN')) {
            const parts = line.split(/\s+/);
            // AUTH PLAIN [b64]
            if (parts.length >= 3 && parts[2].length > 0) {
              socket.write('235 2.7.0 Authentication successful\r\n');
            } else {
              authStep = 'plain-await';
              socket.write('334 \r\n');
            }
          } else if (upper.startsWith('AUTH LOGIN')) {
            authStep = 'login-user';
            socket.write('334 VXNlcm5hbWU6\r\n');
          } else if (upper.startsWith('MAIL FROM:')) {
            from = extractAddress(line);
            socket.write('250 OK\r\n');
          } else if (upper.startsWith('RCPT TO:')) {
            const addr = extractAddress(line);
            to.push(addr);
            socket.write('250 OK\r\n');
          } else if (upper === 'DATA') {
            dataMode = true;
            dataLines = [];
            socket.write('354 End data with <CR><LF>.<CR><LF>\r\n');
          } else if (upper === 'RSET') {
            from = '';
            to = [];
            dataLines = [];
            dataMode = false;
            authStep = 'none';
            socket.write('250 OK\r\n');
          } else if (upper === 'NOOP') {
            socket.write('250 OK\r\n');
          } else if (upper === 'QUIT' || upper.startsWith('QUIT ')) {
            socket.write('221 Bye\r\n');
            socket.end();
          } else {
            socket.write('250 OK\r\n');
          }
        }

        socket.on('data', (chunk: string) => {
          buffer += chunk;
          // Process complete lines ending with \n
          let idx: number;
          while ((idx = buffer.indexOf('\n')) !== -1) {
            let rawLine = buffer.slice(0, idx);
            buffer = buffer.slice(idx + 1);
            if (rawLine.endsWith('\r')) rawLine = rawLine.slice(0, -1);
            handleLine(rawLine);
          }
        });

        socket.on('error', () => {});
      });

      // Expose instance for closure access
      (srv as any).__smtpMockInstance = this;

      srv.listen(0, '127.0.0.1', () => {
        const addr = srv.address() as net.AddressInfo;
        this.port = addr.port;
        this.server = srv;
        resolve(this.port);
      });
      srv.on('error', reject);
    });
  }

  stop(): Promise<void> {
    if (!this.server) return Promise.resolve();
    const srv = this.server;
    return new Promise<void>((res) => {
      // Close all open connections
      try { (srv as any).closeAllConnections?.(); } catch {}
      srv.close(() => res());
      this.server = null;
    });
  }

  clear(): void {
    this.messages = [];
  }
}
