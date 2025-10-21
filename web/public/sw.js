const CACHE_NAME = 'easy-sync-v1';
const ASSETS = [
  '/',
  '/static/manifest.json'
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => cache.addAll(ASSETS))
  );
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) => Promise.all(
      keys.filter((k) => k !== CACHE_NAME).map((k) => caches.delete(k))
    ))
  );
});

self.addEventListener('fetch', (event) => {
  const { request } = event;
  // Network-first for API/WS; cache-first for static
  if (request.method === 'GET' && request.destination !== 'document') {
    event.respondWith(
      caches.match(request).then((cached) => {
        const fetchPromise = fetch(request).then((response) => {
          try {
            const copy = response.clone();
            caches.open(CACHE_NAME).then((cache) => cache.put(request, copy));
          } catch (e) {}
          return response;
        }).catch(() => cached || Promise.reject('offline'));
        return cached || fetchPromise;
      })
    );
  }
});