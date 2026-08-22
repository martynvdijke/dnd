// {{VERSION}} is substituted at serve time by the /sw.js handler (see
// RegisterStaticRoutes) with the app version. Every release therefore
// produces different sw.js bytes, which is what makes browsers detect
// and install the updated service worker instead of serving the old
// cached bundle forever.
const CACHE_NAME = 'villum-v3-{{VERSION}}';
const CDN_CACHE = 'villum-cdn-v1';

// CDN origins to cache-first
const CDN_ORIGINS = [
  'cdn.jsdelivr.net',
  'cdnjs.cloudflare.com',
  'fonts.googleapis.com',
  'fonts.gstatic.com',
  'unpkg.com',
];

const APP_SHELL = [
  '/static/style.css',
  '/static/js/app.js',
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(APP_SHELL);
    })
  );
  self.skipWaiting();
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) => {
      return Promise.all(
        keys
          .filter((k) => k !== CACHE_NAME && k !== CDN_CACHE)
          .map((k) => caches.delete(k))
      );
    })
  );
  self.clients.claim();
});

self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);
  if (event.request.method !== 'GET') return;

  // CDN cache-first: try cache, fetch & update, fallback to stale
  if (CDN_ORIGINS.some(o => url.hostname === o)) {
    event.respondWith(
      caches.open(CDN_CACHE).then((cache) => {
        return cache.match(event.request).then((cached) => {
          const fetchPromise = fetch(event.request).then((response) => {
            if (response.ok) cache.put(event.request, response.clone());
            return response;
          }).catch(() => cached);
          // Return cached immediately if available, else wait for fetch
          return cached || fetchPromise;
        });
      })
    );
    return;
  }

  // Character detail: network-first with cache fallback
  if (url.pathname.startsWith('/api/characters/') && url.pathname.match(/^\/api\/characters\/\d+$/)) {
    event.respondWith(
      caches.open(CACHE_NAME).then((cache) => {
        return fetch(event.request)
          .then((response) => {
            cache.put(event.request, response.clone());
            return response;
          })
          .catch(() => {
            return caches.match(event.request).then((cached) => {
              return cached || new Response(
                JSON.stringify({ offline: true, message: 'Offline — cached data may be stale' }),
                { headers: { 'Content-Type': 'application/json' } }
              );
            });
          });
      })
    );
    return;
  }

  // Static assets + app shell: cache-first, cache-on-first-fetch, offline fallback
  if (url.pathname.startsWith('/static/') || url.pathname === '/app' || url.pathname === '/') {
    event.respondWith(
      caches.match(event.request).then((cached) => {
        if (cached) return cached;
        return fetch(event.request).then((response) => {
          if (response.ok && (url.pathname.startsWith('/static/') || url.pathname === '/' || url.pathname === '/app')) {
            return caches.open(CACHE_NAME).then((cache) => {
              cache.put(event.request, response.clone());
              return response;
            });
          }
          return response;
        }).catch(() => new Response('Offline', { status: 503 }));
      })
    );
    return;
  }

  // Everything else: network-only
  event.respondWith(fetch(event.request).catch(() => {
    return new Response('Offline', { status: 503 });
  }));
});

// ─── Web Push ───

// Payload JSON: {title, body, url, tag}. A malformed payload still shows a
// generic notification rather than dropping the message silently.
self.addEventListener('push', (event) => {
  let title = 'Villum';
  let body = 'You have a new notification.';
  let url = '/';
  let tag;
  try {
    const data = event.data ? event.data.json() : {};
    if (data.title) title = String(data.title);
    if (data.body) body = String(data.body);
    if (data.url) url = String(data.url);
    if (data.tag) tag = String(data.tag);
  } catch (err) {
    // Keep fallbacks; never throw out of a push handler.
  }
  event.waitUntil(
    self.registration.showNotification(title, {
      body,
      tag,
      icon: '/static/brand/logo-mark.svg',
      badge: '/static/brand/logo-mark.svg',
      data: { url },
    })
  );
});

// Click: focus an existing window at `url` if one exists, otherwise open it.
self.addEventListener('notificationclick', (event) => {
  event.notification.close();
  const target = (event.notification.data && event.notification.data.url) || '/';
  event.waitUntil(
    self.clients.matchAll({ type: 'window', includeUncontrolled: true }).then((clientList) => {
      for (const client of clientList) {
        if ('focus' in client) {
          client.focus();
          if ('navigate' in client && client.url !== new URL(target, self.location.origin).href) {
            try { client.navigate(target); } catch (err) { /* older browsers */ }
          }
          return;
        }
      }
      return self.clients.openWindow(target);
    })
  );
});
