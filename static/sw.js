const CACHE_NAME = 'villum-v1';
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
          .filter((k) => k !== CACHE_NAME)
          .map((k) => caches.delete(k))
      );
    })
  );
  self.clients.claim();
});

self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);

  if (event.request.method !== 'GET') return;

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

  if (url.pathname.startsWith('/static/') || url.pathname === '/app' || url.pathname === '/') {
    event.respondWith(
      caches.match(event.request).then((cached) => {
        return cached || fetch(event.request).then((response) => {
          return caches.open(CACHE_NAME).then((cache) => {
            if (url.pathname.startsWith('/static/')) {
              cache.put(event.request, response.clone());
            }
            return response;
          });
        });
      })
    );
    return;
  }

  event.respondWith(fetch(event.request).catch(() => {
    return new Response('Offline', { status: 503 });
  }));
});
