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

// Data yang dibutuhkan kasir saat offline; laporan dll sengaja tidak, supaya angka lama tidak dikira terbaru.
// Cache ini dihapus saat login/logout (stores/auth.js) karena kuncinya hanya URL, bukan akun.
const API_CACHE = 'psd-api'
const dataKasir = /^\/api\/(menu|tenant|meja|shift\/aktif|pengaturan|pesanan\/\d+)$/

// online: selalu ambil dari server lalu simpan; hanya saat jaringan gagal salinan terakhir dipakai
async function dataTerakhir(req) {
  try {
    const res = await fetch(req)
    if (res.ok) {
      const salin = res.clone()
      caches.open(API_CACHE).then((c) => c.put(req.url, salin))
    }
    return res
  } catch (err) {
    const simpan = await caches.match(req.url, { cacheName: API_CACHE })
    if (simpan) return simpan
    throw err
  }
}

self.addEventListener('fetch', (e) => {
  const req = e.request
  const url = new URL(req.url)
  if (req.method !== 'GET' || url.origin !== location.origin) return
  if (dataKasir.test(url.pathname)) return e.respondWith(dataTerakhir(req))
  if (req.mode === 'navigate') {
    if (bukanHalaman.test(url.pathname)) return
    // semua halaman SPA = index.html yang tersimpan; versi baru terpakai sekali buka setelah sw.js berubah
    e.respondWith(caches.match('/').then((r) => r || fetch(req)))
  } else if (FILES.includes(url.pathname)) {
    e.respondWith(caches.match(url.pathname).then((r) => r || fetch(req)))
  }
})
