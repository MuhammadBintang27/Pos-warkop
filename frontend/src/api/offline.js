// Mode offline HP kasir (aplikasi Android): saat jaringan gagal, permintaan kasir dijawab "server mini" di HP ini
// dengan bentuk jawaban yang sama seperti server, dicatat ke antrean, lalu dikirim ulang berurutan saat online.
// Tiket dapur/batal dan struk dicetak HP langsung ke printer (utils/escpos.js). Server tetap penentu akhir:
// tiap aksi membawa X-Idem supaya tidak diproses dua kali, dan X-Offline supaya server tidak mencetak ulang tiketnya.
// Yang didukung sengaja terbatas (lihat RUTE); gabung, pisah, diskon, dan non-tunai tetap butuh internet.
import { reactive } from 'vue'
import { load, remove, save } from '@/utils/storage'
import { bisaCetakLangsung, cetakLangsung, escposStruk, escposTiket, logoRaster } from '@/utils/escpos'
import { gabungItem } from '@/utils/print'

export const offline = reactive({
  antrean: load('offline:antrean', []),
  galat: '', // aksi terdepan ditolak server; antrean berhenti sampai kasir mencoba lagi atau melewatinya
  sinkron: false,
  gagalCetak: [], // tiket yang tidak sampai ke printer: { label, printer, data, galat }
})

let data = load('offline:data', {}) // jawaban GET terakhir per URL, ditambah perubahan lokal
let peta = load('offline:peta', {}) // id sementara (negatif) -> id asli dari server
const simpanData = () => save('offline:data', data)
const simpanAntrean = () => {
  save('offline:antrean', offline.antrean)
  save('offline:peta', peta)
}

const user = () => load('user', null)
export const aktif = () => bisaCetakLangsung() && ['kasir', 'owner'].includes(user()?.role)

// data milik akun sebelumnya tidak boleh terbawa; antrean tidak dihapus (logout ditolak selama masih ada antrean)
export function lupakanData() {
  data = {}
  remove('offline:data')
}

const MSG_BATAL_ITEM = 'Item yang sudah dikirim ke dapur harus dibatalkan lewat pembatalan item (dengan alasan)'

function gagal(message, status = 400) {
  throw Object.assign(new Error(message), { isAxiosError: true, response: { status, data: { message } } })
}

let urut = 0
const sementara = () => -(Date.now() * 100 + (urut++ % 100))

const kunciPesanan = (id) => `/pesanan/${id}`
const denah = () => data['/meja'] ?? gagal('Denah meja belum tersimpan di HP ini, sambungkan internet sekali')

function ambil(id) {
  return data[kunciPesanan(id)] ?? gagal('Bill ini belum tersimpan di HP ini, sambungkan internet sekali', 404)
}
function bukaOpen(id) {
  const p = ambil(id)
  if (p.status !== 'open') gagal(`Pesanan sudah ${p.status}`)
  return p
}
function cariItem(p, itemId) {
  return p.items.find((i) => i.id === itemId) ?? gagal('Item tidak ditemukan', 404)
}

// simpan pesanan dan perbarui ringkasannya di denah, seperti yang akan dihitung server
function simpanPesanan(p) {
  p.subtotal = p.items.reduce((s, i) => s + i.harga * i.qty, 0)
  p.diskon = Math.min(p.diskon, p.subtotal)
  p.total = p.subtotal - p.diskon
  data[kunciPesanan(p.id)] = p
  const d = data['/meja']
  if (d) {
    for (const m of d.meja) m.pesanan = m.pesanan.filter((x) => x.id !== p.id)
    d.takeaway = d.takeaway.filter((x) => x.id !== p.id)
    if (p.status === 'open' && p.items.length) {
      const r = { id: p.id, atasNama: p.atasNama, total: p.total }
      const meja = p.mejaId && d.meja.find((m) => m.id === p.mejaId)
      if (meja) meja.pesanan.push(r)
      else if (!p.mejaId) d.takeaway.push(r)
    }
  }
  simpanData()
  return p
}

function cariMenu(id) {
  for (const k of data['/menu']?.kategori ?? []) {
    const m = k.menu.find((x) => x.id === id)
    if (m) return m
  }
  gagal('Menu tidak ditemukan', 404)
}

function ubahStok(menuId, delta) {
  const m = cariMenu(menuId)
  if (!m.trackStok || !delta) return
  if (m.stok + delta < 0) gagal(`Stok ${m.nama} tidak cukup (sisa ${m.stok})`)
  m.stok += delta
}

// salinan resolveOpsi/hargaAkhir di backend: opsi grup ganti-harga menggantikan harga dasar, opsi lain ditambahkan
function resolveOpsi(menu, ids = []) {
  const grup = menu.varian.map((g) => ({ ...g, dipilih: 0 }))
  let dasar = menu.harga
  let tambah = 0
  const nama = []
  const sudah = new Set()
  for (const oid of ids) {
    const g = grup.find((x) => x.opsi.some((o) => o.id === oid))
    if (!g || sudah.has(oid)) gagal('Pilihan varian tidak valid')
    sudah.add(oid)
    if (++g.dipilih > 1 && !g.multi) gagal(`Pilih satu saja untuk ${g.nama}`)
    const o = g.opsi.find((x) => x.id === oid)
    nama.push(o.nama)
    if (g.gantiHarga) dasar = o.tambahan
    else tambah += o.tambahan
  }
  for (const g of grup) if (g.wajib && !g.dipilih) gagal(`Pilih ${g.nama} dulu`)
  return { opsi: nama.join(', '), harga: dasar + tambah }
}

const sama = (i, k) => !i.dikirim && i.menuId === k.menuId && i.harga === k.harga && i.opsi === k.opsi && i.catatan === k.catatan
const kunciItem = (i) => ({ menuId: i.menuId, harga: i.harga, opsi: i.opsi, catatan: i.catatan })

// salinan sisipItem: digabung ke baris belum-dikirim yang sama persis, atau baris baru
function sisip(p, it) {
  const ada = p.items.find((i) => sama(i, it))
  if (ada) {
    if (ada.qty + it.qty > 99) gagal('Maksimal 99 porsi per baris pesanan')
    ada.qty += it.qty
    return ada
  }
  const baru = { ...it, id: sementara(), dikirim: false, diprosesAt: null, siapAt: null }
  p.items.push(baru)
  return baru
}

const printerById = (id) => (data['/printer'] ?? []).find((x) => x.id === id)

function kePrinter(printer, bytes, label) {
  try {
    cetakLangsung(printer, bytes)
  } catch (e) {
    offline.gagalCetak.push({ label, printer, data: bytes, galat: e.message })
  }
}

// tiket per tenant ke printer tenantnya; tenant tanpa printer tidak masuk antrean (klien memakai dialog cetak bila diatur)
function cetakTiket(tiket, batal) {
  tiket.antrean = {}
  const per = new Map()
  for (const i of tiket.items) {
    if (!per.has(i.tenantId)) per.set(i.tenantId, { nama: i.tenantNama, items: [] })
    per.get(i.tenantId).items.push({ qty: i.qty, nama: i.nama, opsi: i.opsi, catatan: i.catatan ?? '' })
  }
  for (const [tid, g] of per) {
    const tenant = (data['/tenant'] ?? []).find((t) => t.id === tid)
    const printer = tenant?.printerId && printerById(tenant.printerId)
    if (!printer) continue
    const judul = batal ? 'BATAL' : tiket.tambahan?.[tid] ? 'PESANAN TAMBAHAN' : 'PESANAN BARU'
    const p = { judul, mejaNama: tiket.mejaNama, atasNama: tiket.atasNama, oleh: tiket.oleh, tenant: g.nama, waktu: new Date(), items: g.items }
    kePrinter(printer, escposTiket(p, printer.lebarMm), `${judul} ${tiket.mejaNama || 'Bawa pulang'} (${g.nama})`)
    tiket.antrean[tid] = true
  }
  return tiket
}

function catatBatal(p, i, qty, alasan, sudahDibuat) {
  p.pembatalan.push({
    nama: i.nama, opsi: i.opsi, qty, harga: i.harga, alasan, sudahDikirim: i.dikirim,
    sudahDibuat, oleh: user()?.nama ?? '', at: new Date().toISOString(),
  })
}

const ambilData = (url) => () => (url in data ? data[url] : gagal('Data ini belum tersimpan di HP ini, sambungkan internet sekali'))

// [metode, pola URL, penangan, catat ke antrean?]; penangan menerima (angka dari URL..., body, entri antrean)
const RUTE = [
  ['get', /^\/menu$/, ambilData('/menu')],
  ['get', /^\/tenant$/, ambilData('/tenant')],
  ['get', /^\/meja$/, ambilData('/meja')],
  ['get', /^\/shift\/aktif$/, ambilData('/shift/aktif')],
  ['get', /^\/pengaturan$/, ambilData('/pengaturan')],
  ['get', /^\/printer$/, ambilData('/printer')],
  ['get', /^\/pesanan\/(-?\d+)$/, (id) => ambil(id)],
  ['get', /^\/cetak\/status$/, () => ({ antri: 0, macet: 0, stasiunAktif: true, daftar: [] })],

  ['post', /^\/pesanan$/, (b, e) => {
    const d = denah()
    const meja = b.mejaId && (d.meja.find((m) => m.id === b.mejaId) ?? gagal('Meja tidak ditemukan', 404))
    if (meja?.pesanan.length) return ambil(meja.pesanan[0].id) // seperti server: meja terisi memakai bill yang ada
    const p = {
      id: sementara(), mejaId: b.mejaId ?? null, mejaNama: meja?.nama ?? '', atasNama: (b.atasNama ?? '').trim(),
      status: 'open', diskon: 0, subtotal: 0, total: 0, dibuatAt: new Date().toISOString(), lunasAt: null,
      items: [], pembayaran: [], pembatalan: [],
    }
    e.sementara = p.id
    return simpanPesanan(p)
  }, true],

  ['post', /^\/pesanan\/(-?\d+)\/item$/, (id, b, e) => {
    const p = bukaOpen(id)
    const menu = cariMenu(b.menuId)
    const tenant = (data['/tenant'] ?? []).find((t) => t.id === menu.tenantId)
    if (tenant && !tenant.buka) gagal(`${tenant.nama} sedang tutup`)
    if (!menu.tersedia) gagal(`${menu.nama} sedang habis`)
    const { opsi, harga } = resolveOpsi(menu, b.opsiIds)
    ubahStok(menu.id, -b.qty)
    const it = sisip(p, {
      menuId: menu.id, tenantId: menu.tenantId, tenantNama: tenant?.nama ?? '', nama: menu.nama,
      harga, qty: b.qty, opsi, catatan: (b.catatan ?? '').trim(),
    })
    if (it.id < 0) Object.assign(e, { item: it.id, cocok: kunciItem(it) })
    return simpanPesanan(p)
  }, true],

  ['put', /^\/pesanan\/(-?\d+)\/item\/(-?\d+)$/, (id, itemId, b, e) => {
    const p = bukaOpen(id)
    const it = cariItem(p, itemId)
    const catatan = (b.catatan ?? '').trim()
    if (it.dikirim) {
      if (b.qty < it.qty) gagal(MSG_BATAL_ITEM)
      if (catatan !== it.catatan) gagal('Catatan item yang sudah dikirim tidak bisa diubah. Pesan ulang dengan catatan baru')
      if (b.qty === it.qty) return p
      // tambahan atas item terkirim jadi baris baru belum-dikirim, supaya dapur menerima tiket tambahan
      ubahStok(it.menuId, it.qty - b.qty)
      const baru = sisip(p, { ...it, qty: b.qty - it.qty })
      if (baru.id < 0) Object.assign(e, { item: baru.id, cocok: kunciItem(baru) })
    } else {
      ubahStok(it.menuId, it.qty - b.qty)
      Object.assign(it, { qty: b.qty, catatan })
    }
    return simpanPesanan(p)
  }, true],

  ['delete', /^\/pesanan\/(-?\d+)\/item\/(-?\d+)$/, (id, itemId) => {
    const p = bukaOpen(id)
    const it = cariItem(p, itemId)
    if (it.dikirim) gagal(MSG_BATAL_ITEM)
    ubahStok(it.menuId, it.qty)
    p.items = p.items.filter((i) => i !== it)
    return simpanPesanan(p)
  }, true],

  ['post', /^\/pesanan\/(-?\d+)\/kirim$/, (id) => {
    const p = bukaOpen(id)
    const sudah = new Set(p.items.filter((i) => i.dikirim).map((i) => i.tenantId))
    const baru = p.items.filter((i) => !i.dikirim)
    const tiket = {
      mejaNama: p.mejaNama, atasNama: p.atasNama, oleh: user()?.nama ?? '',
      items: baru.map((i) => ({ ...i })), tambahan: Object.fromEntries(baru.map((i) => [i.tenantId, sudah.has(i.tenantId)])),
    }
    baru.forEach((i) => (i.dikirim = true))
    simpanPesanan(p)
    return cetakTiket(tiket, false)
  }, true],

  ['post', /^\/pesanan\/(-?\d+)\/pindah$/, (id, b) => {
    const p = bukaOpen(id)
    const meja = denah().meja.find((m) => m.id === b.mejaId) ?? gagal('Meja tujuan tidak ditemukan', 404)
    if (meja.pesanan.some((x) => x.id !== p.id)) gagal('Meja tujuan sudah terisi, gunakan gabung')
    Object.assign(p, { mejaId: meja.id, mejaNama: meja.nama })
    return simpanPesanan(p)
  }, true],

  ['post', /^\/pesanan\/(-?\d+)\/batal$/, (id, b) => {
    const p = bukaOpen(id)
    const terkirim = p.items.filter((i) => i.dikirim)
    let alasan = (b.alasan ?? '').trim()
    if (terkirim.length && !alasan) gagal('Alasan pembatalan wajib diisi karena sudah ada item yang dikirim ke dapur')
    alasan ||= 'Dibatalkan sebelum dikirim'
    for (const i of p.items) {
      const sudahDibuat = i.dikirim && !!b.sudahDibuat
      if (!sudahDibuat) ubahStok(i.menuId, i.qty)
      catatBatal(p, i, i.qty, alasan, sudahDibuat)
    }
    p.status = 'batal'
    simpanPesanan(p)
    const tiket = terkirim.length ? cetakTiket({ mejaNama: p.mejaNama, atasNama: p.atasNama, oleh: user()?.nama ?? '', items: terkirim }, true) : null
    return { pesanan: p, tiket }
  }, true],

  ['post', /^\/pesanan\/(-?\d+)\/item\/(-?\d+)\/batal$/, (id, itemId, b) => {
    const alasan = (b.alasan ?? '').trim()
    if (!alasan) gagal('Alasan pembatalan wajib diisi')
    const p = bukaOpen(id)
    const it = cariItem(p, itemId)
    if (b.qty > it.qty) gagal('Jumlah pembatalan melebihi jumlah pesanan')
    const sudahDibuat = it.dikirim && !!b.sudahDibuat
    if (!sudahDibuat) ubahStok(it.menuId, b.qty)
    if (b.qty === it.qty) p.items = p.items.filter((i) => i !== it)
    else it.qty -= b.qty
    catatBatal(p, it, b.qty, alasan, sudahDibuat)
    simpanPesanan(p)
    const tiket = it.dikirim
      ? cetakTiket({ mejaNama: p.mejaNama, atasNama: p.atasNama, oleh: user()?.nama ?? '',
          items: [{ nama: it.nama, opsi: it.opsi, qty: b.qty, tenantId: it.tenantId, tenantNama: it.tenantNama }] }, true)
      : null
    return { pesanan: p, tiket }
  }, true],

  ['post', /^\/pesanan\/(-?\d+)\/bayar$/, (id, b) => {
    if (b.pembayaran.some((x) => x.metode !== 'tunai')) gagal('Saat offline hanya bisa bayar tunai')
    if (!data['/shift/aktif']) gagal('Buka shift dulu sebelum menerima pembayaran')
    const p = bukaOpen(id)
    if (!p.subtotal) gagal('Pesanan masih kosong')
    const sum = b.pembayaran.reduce((s, x) => s + x.jumlah, 0)
    if (sum !== p.total) gagal(`Total pembayaran (${sum}) harus sama dengan tagihan (${p.total})`)
    p.pembayaran = b.pembayaran.map((x) => {
      if (x.diterima > 0 && x.diterima < x.jumlah) gagal('Uang diterima kurang dari jumlah tunai')
      const diterima = x.diterima > 0 ? x.diterima : x.jumlah
      return { metode: 'tunai', jumlah: x.jumlah, diterima, kembalian: diterima - x.jumlah }
    })
    Object.assign(p, { status: 'lunas', lunasAt: new Date().toISOString(), kasir: user()?.nama ?? '' })
    return simpanPesanan(p)
  }, true],

  // struk dicetak HP sendiri; tidak perlu dikirim ulang ke server
  ['post', /^\/pesanan\/(-?\d+)\/cetak-struk$/, async (id) => {
    const p = ambil(id)
    if (p.status !== 'lunas') gagal('Struk hanya untuk pesanan yang sudah lunas')
    const u = data['/pengaturan'] ?? {}
    const printer = printerById(u.printerKasirId)
    if (!printer) return { antre: false }
    const struk = {
      usaha: u.namaUsaha, alamat: u.alamat, footer: u.footerStruk, id: p.id > 0 ? p.id : 'offline',
      mejaNama: p.mejaNama, atasNama: p.atasNama, kasir: p.kasir, waktu: p.lunasAt ?? p.dibuatAt, items: gabungItem(p.items),
      subtotal: p.subtotal, diskon: p.diskon, total: p.total,
      bayar: p.pembayaran.map((x) => ({ metode: x.metode, pelanggan: x.pelanggan ?? '', diterima: x.diterima, kembalian: x.kembalian })),
    }
    kePrinter(printer, escposStruk(struk, printer.lebarMm, await logoRaster(u.logo, printer.lebarMm)), `Struk ${p.mejaNama || 'Bawa pulang'}`)
    return { antre: true }
  }],
]

function cocokRute(cfg) {
  const url = cfg.url.replace(/^\/api/, '')
  for (const [metode, pola, fn, catat] of RUTE) {
    const m = cfg.method === metode && !cfg.params && url.match(pola)
    if (m) return { url, fn, catat, angka: m.slice(1).map(Number) }
  }
  return null
}

// pesanan yang dikembalikan server (langsung atau di dalam BatalResponse)
const pesananDari = (d) => (d?.pesanan ?? d)?.items && (d.pesanan ?? d).id && (d.pesanan ?? d)

// jawaban online yang berhasil disimpan untuk dipakai saat offline
function ingat(r, cfg, d) {
  if (cfg.method === 'get' && !r.url.startsWith('/pesanan/') && r.url !== '/cetak/status') {
    data[r.url] = d
    if (r.url === '/meja') buangBillLama(d)
    simpanData()
    return
  }
  const p = pesananDari(d)
  if (p) simpanPesanan(structuredClone(p))
}

// bill yang sudah tidak terbuka di denah dibuang (yang baru lunas/batal disimpan sebentar untuk cetak ulang struk)
function buangBillLama(d) {
  const buka = new Set([...d.meja.flatMap((m) => m.pesanan), ...d.takeaway].map((x) => x.id))
  const batas = Date.now() - 6 * 3600e3
  for (const k of Object.keys(data)) {
    const p = k.startsWith('/pesanan/') && data[k]
    if (!p || buka.has(p.id) || (p.status !== 'open' && Date.parse(p.lunasAt ?? p.dibuatAt) > batas)) continue
    delete data[k]
  }
}

async function jalankanLokal(r, cfg) {
  const body = typeof cfg.data === 'string' ? JSON.parse(cfg.data || '{}') : (cfg.data ?? {})
  const entri = { kunci: cfg.headers['X-Idem'], method: cfg.method, url: r.url, body, pemilik: user()?.id }
  const hasil = await r.fn(...r.angka, body, entri)
  if (r.catat) {
    offline.antrean.push(entri)
    simpanAntrean()
  }
  return structuredClone(hasil) // jangan bagikan objek yang disimpan ke layar
}

// Adapter axios: rute kasir dicoba ke server dulu (batas 10 detik); jaringan gagal -> dijawab lokal.
// Selama antrean belum habis, semua rute kasir dijawab lokal supaya urutan aksinya tetap sama di server.
let http
export function pasangOffline(h, bawaan) {
  http = h
  http.defaults.adapter = async (cfg) => {
    const r = !cfg.langsung && aktif() && cocokRute(cfg)
    if (!r) return bawaan(cfg)
    if (r.catat) cfg.headers['X-Idem'] = crypto.randomUUID()
    if (!offline.antrean.length) {
      try {
        const res = await bawaan({ ...cfg, timeout: cfg.timeout || 10000 })
        ingat(r, cfg, res.data?.data)
        return res
      } catch (e) {
        if (e.response) throw e // server menjawab: itu keputusan server, bukan masalah jaringan
      }
    }
    const d = await jalankanLokal(r, cfg)
    setTimeout(sinkron, 0)
    return { data: { data: d }, status: 200, statusText: 'OK', headers: {}, config: cfg, request: null }
  }
  window.addEventListener('online', sinkron)
  setInterval(sinkron, 15000)
}

const gantiId = (url) => url.replace(/-\d+/g, (s) => peta[s] ?? s)
const labelAksi = (e) =>
  ({ post: 'Kirim', put: 'Ubah', delete: 'Hapus' })[e.method] + ' ' + e.url.replace(/^\/pesanan\/-?\d+/, 'bill').replace(/-?\d+/g, '')

export async function sinkron() {
  if (offline.sinkron || !offline.antrean.length || offline.galat || !http) return
  offline.sinkron = true
  try {
    while (offline.antrean.length) {
      const e = offline.antrean[0]
      if (e.pemilik !== user()?.id) {
        offline.galat = 'Antrean offline milik akun lain di HP ini. Masuk dengan akun yang membuatnya untuk mengirimnya'
        return
      }
      let res
      try {
        res = await http.request({
          method: e.method, url: gantiId(e.url), data: e.body, langsung: true, timeout: 15000,
          headers: { 'X-Idem': e.kunci, 'X-Offline': '1' },
        })
      } catch (err) {
        if (err.response) offline.galat = `${labelAksi(e)}: ${err.response.data?.message || 'ditolak server'}`
        return // tanpa jawaban server: masih offline, dicoba lagi nanti
      }
      const d = res.data.data
      if (e.sementara) peta[e.sementara] = d.id
      if (e.item) {
        const row = pesananDari(d)?.items.find((i) => sama(i, e.cocok))
        if (row) peta[e.item] = row.id
      }
      offline.antrean.shift()
      simpanAntrean()
    }
    const selesai = peta
    peta = {}
    for (const k of Object.keys(data)) if (k.startsWith('/pesanan/-')) delete data[k]
    simpanData()
    simpanAntrean()
    window.dispatchEvent(new CustomEvent('offline:tersinkron', { detail: selesai }))
  } finally {
    offline.sinkron = false
  }
}

export function cobaLagi() {
  offline.galat = ''
  sinkron()
}

// aksi terdepan dibuang tanpa dikirim; kasir sudah diberi tahu bahwa server tidak akan mencatatnya
export function lewati() {
  offline.antrean.shift()
  simpanAntrean()
  cobaLagi()
}

export function cetakUlang(i) {
  const [g] = offline.gagalCetak.splice(i, 1)
  kePrinter(g.printer, g.data, g.label)
}
