const CACHE_NAME = 'villum-v2';
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
