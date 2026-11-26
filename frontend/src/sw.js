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
      // termasuk psd-api dari versi lama: data offline kini disimpan api/offline.js
      .then((keys) => Promise.all(keys.filter((k) => k !== CACHE && k !== UPLOADS).map((k) => caches.delete(k))))
      .then(() => self.clients.claim()),
  )
})

const bukanHalaman = /^\/(api|uploads|dl)\//

// Data kasir untuk offline disimpan api/offline.js, bukan di sini. Unggahan (logo struk, foto menu) bernama unik
// per unggahan, jadi aman disimpan selamanya; logo tetap bisa dicetak HP saat offline.
const UPLOADS = 'psd-uploads'
async function unggahan(req) {
  const ada = await caches.match(req, { cacheName: UPLOADS })
  if (ada) return ada
  const res = await fetch(req)
  if (res.ok) {
    const salin = res.clone()
    caches.open(UPLOADS).then((c) => c.put(req, salin))
  }
  return res
}

self.addEventListener('fetch', (e) => {
  const req = e.request
  const url = new URL(req.url)
  if (req.method !== 'GET' || url.origin !== location.origin) return
  if (url.pathname.startsWith('/uploads/') && req.mode !== 'navigate') return e.respondWith(unggahan(req))
  if (req.mode === 'navigate') {
    if (bukanHalaman.test(url.pathname)) return
    // semua halaman SPA = index.html yang tersimpan; versi baru terpakai sekali buka setelah sw.js berubah
    e.respondWith(caches.match('/').then((r) => r || fetch(req)))
  } else if (FILES.includes(url.pathname)) {
    e.respondWith(caches.match(url.pathname).then((r) => r || fetch(req)))
  }
})
