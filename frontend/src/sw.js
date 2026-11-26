/* global FILES, VERSI */
// Service worker: tampilan aplikasi disimpan di perangkat supaya tetap terbuka saat internet putus.
// FILES (hasil build ber-hash) dan VERSI disisipkan oleh plugin di vite.config.js; data /api tidak disentuh di sini.
const CACHE = 'psd-app-' + VERSI

self.addEventListener('install', (e) => {
  e.waitUntil(
    caches
      .open(CACHE)
      .then((c) => c.addAll(FILES))
      .then(() => self.skipWaiting()),
  )
})

self.addEventListener('activate', (e) => {
  e.waitUntil(
    caches
      .keys()
      .then((keys) => Promise.all(keys.filter((k) => k.startsWith('psd-app-') && k !== CACHE).map((k) => caches.delete(k))))
      .then(() => self.clients.claim()),
  )
})

const bukanHalaman = /^\/(api|uploads|dl)\//

self.addEventListener('fetch', (e) => {
  const req = e.request
  const url = new URL(req.url)
  if (req.method !== 'GET' || url.origin !== location.origin) return
  if (req.mode === 'navigate') {
    if (bukanHalaman.test(url.pathname)) return
    // semua halaman SPA = index.html yang tersimpan; versi baru terpakai sekali buka setelah sw.js berubah
    e.respondWith(caches.match('/').then((r) => r || fetch(req)))
  } else if (FILES.includes(url.pathname)) {
    e.respondWith(caches.match(url.pathname).then((r) => r || fetch(req)))
  }
})
