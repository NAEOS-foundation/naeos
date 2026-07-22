const CACHE = 'naeos-v3';
const STATIC_CACHE = 'naeos-static-v3';

self.addEventListener('install', function (event) {
  self.skipWaiting();
  event.waitUntil(
    caches.open(STATIC_CACHE).then(function (cache) {
      return cache.addAll([
        '/',
        '/favicon.svg',
        '/manifest.json',
        '/images/icon-192.svg',
        '/images/icon-512.svg'
      ]);
    })
  );
});

self.addEventListener('activate', function (event) {
  event.waitUntil(
    caches.keys().then(function (keys) {
      return Promise.all(
        keys.filter(function (k) {
          return k !== CACHE && k !== STATIC_CACHE;
        }).map(function (k) {
          return caches.delete(k);
        })
      );
    })
  );
  return self.clients.claim();
});

self.addEventListener('fetch', function (event) {
  if (event.request.method !== 'GET') return;
  if (event.request.url.indexOf('chrome-extension') !== -1) return;

  const url = new URL(event.request.url);

  // Static assets: cache-first
  if (url.pathname.startsWith('/assets/') || url.pathname.endsWith('.css') || url.pathname.endsWith('.js')) {
    event.respondWith(
      caches.match(event.request).then(function (cached) {
        if (cached) return cached;
        return fetch(event.request).then(function (response) {
          return caches.open(STATIC_CACHE).then(function (cache) {
            cache.put(event.request, response.clone());
            return response;
          });
        });
      })
    );
    return;
  }

  // HTML pages: network-first with cache fallback
  if (event.request.headers.get('accept') && event.request.headers.get('accept').indexOf('text/html') !== -1) {
    event.respondWith(
      fetch(event.request).then(function (response) {
        return caches.open(CACHE).then(function (cache) {
          cache.put(event.request, response.clone());
          return response;
        });
      }).catch(function () {
        return caches.match(event.request);
      })
    );
    return;
  }

  // Other requests: cache-first
  event.respondWith(
    caches.match(event.request).then(function (cached) {
      return cached || fetch(event.request).then(function (response) {
        return caches.open(CACHE).then(function (cache) {
          if (url.origin === self.location.origin) {
            cache.put(event.request, response.clone());
          }
          return response;
        });
      });
    })
  );
});
