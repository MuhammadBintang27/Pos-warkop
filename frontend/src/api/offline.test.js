import { beforeAll, describe, expect, it, vi } from 'vitest'
import axios from 'axios'

// lingkungan HP kasir: localStorage, akun kasir, dan aplikasi Android yang bisa cetak langsung
const simpan = new Map()
globalThis.localStorage = {
  getItem: (k) => simpan.get(k) ?? null,
  setItem: (k, v) => simpan.set(k, String(v)),
  removeItem: (k) => simpan.delete(k),
}
const cetak = vi.fn(() => '')
const peristiwa = []
globalThis.window = { StasiunAndroid: { cetak }, addEventListener() {}, dispatchEvent: (e) => peristiwa.push(e) }
globalThis.CustomEvent = class extends Event {
  constructor(type, o) {
    super(type)
    this.detail = o?.detail
  }
}
localStorage.setItem('user', JSON.stringify({ id: 9, nama: 'Kasir Satu', role: 'kasir' }))

const SERVER = {
  'get /menu': {
    kategori: [{ menu: [{ id: 1, nama: 'Kopi', harga: 5000, tenantId: 7, tersedia: true, trackStok: false, stok: 0,
      varian: [{ id: 1, nama: 'Suhu', wajib: true, multi: false, gantiHarga: false, opsi: [{ id: 11, nama: 'Es', tambahan: 1000 }] }] }] }],
  },
  'get /tenant': [{ id: 7, nama: 'Dapur', buka: true, printerId: 3 }],
  'get /printer': [{ id: 3, nama: 'Printer Dapur', ip: '10.0.0.9', port: 9100, lebarMm: 58 }],
  'get /meja?detail=1': {
    area: [],
    meja: [{ id: 5, nama: 'Meja 5', pesanan: [] }, { id: 6, nama: 'Meja 6', pesanan: [{ id: 55, atasNama: '', total: 5000 }] }],
    takeaway: [],
    detail: [{ id: 55, mejaId: 6, mejaNama: 'Meja 6', atasNama: '', status: 'open', diskon: 0, subtotal: 5000, total: 5000, items: [], pembayaran: [], pembatalan: [] }],
  },
  'get /shift/aktif': { id: 1 },
  'post /pesanan': { id: 100, items: [] },
  'post /pesanan/100/item': { id: 100, items: [{ id: 900, menuId: 1, harga: 6000, opsi: 'Es', catatan: '', dikirim: false }] },
  'put /pesanan/100/item/900': { id: 100, items: [{ id: 900, menuId: 1, harga: 6000, opsi: 'Es', catatan: '', dikirim: false, qty: 2 }] },
  'post /pesanan/100/kirim': { items: [] },
  'post /pesanan/100/bayar': { id: 100, status: 'lunas', items: [] },
}
let putus = false
const tolak = new Set() // aksi yang ditolak server dengan 400
const terkirim = []
const server = async (cfg) => {
  if (putus) throw Object.assign(new Error('Network Error'), { code: 'ERR_NETWORK' })
  if (tolak.has(`${cfg.method} ${cfg.url}`)) {
    throw Object.assign(new Error('Bad Request'), { response: { status: 400, data: JSON.stringify({ message: 'Meja tidak ditemukan' }), headers: {}, config: cfg } })
  }
  terkirim.push({ url: cfg.url, idem: cfg.headers['X-Idem'], offline: cfg.headers['X-Offline'] })
  // seperti adapter xhr sungguhan: isi masih teks JSON, axios yang mengurainya sesudah adapter
  return { data: JSON.stringify({ data: SERVER[`${cfg.method} ${cfg.url}`] }), status: 200, headers: {}, config: cfg }
}

let http, mod
beforeAll(async () => {
  mod = await import('./offline')
  http = axios.create({ baseURL: '/api' })
  mod.pasangOffline(http, server)
})
const data = (p) => p.then((r) => r.data.data)

describe('mode offline HP kasir', () => {
  it('menjawab lokal saat jaringan putus, mencetak tiket, lalu mengirim ulang berurutan dengan id asli', async () => {
    for (const u of ['/menu', '/tenant', '/printer', '/meja?detail=1', '/shift/aktif']) await http.get(u)

    putus = true
    // detail bill yang datang bersama denah bisa dibuka saat offline
    expect((await data(http.get('/pesanan/55'))).mejaNama).toBe('Meja 6')
    expect(await data(http.get('/meja'))).not.toHaveProperty('detail')
    const bill = await data(http.post('/pesanan', { mejaId: 5, atasNama: '' }))
    expect(bill.id).toBeLessThan(0)
    let p = await data(http.post(`/pesanan/${bill.id}/item`, { menuId: 1, qty: 1, opsiIds: [11], catatan: '' }))
    expect(p.items[0]).toMatchObject({ nama: 'Kopi', opsi: 'Es', harga: 6000, dikirim: false })
    p = await data(http.put(`/pesanan/${bill.id}/item/${p.items[0].id}`, { qty: 2, catatan: '' }))
    expect(p.total).toBe(12000)
    expect((await data(http.get('/meja'))).meja[0].pesanan).toEqual([{ id: bill.id, atasNama: '', total: 12000 }])

    const tiket = await data(http.post(`/pesanan/${bill.id}/kirim`))
    expect(tiket.antrean).toEqual({ 7: true })
    expect(cetak).toHaveBeenCalledWith('10.0.0.9', 9100, expect.any(String))
    expect(atob(cetak.mock.calls[0][2])).toContain('2x Kopi')

    await expect(http.post(`/pesanan/${bill.id}/bayar`, { pembayaran: [{ metode: 'qris', jumlah: 12000 }] })).rejects.toMatchObject({
      response: { data: { message: 'Saat offline hanya bisa bayar tunai' } },
    })
    const lunas = await data(http.post(`/pesanan/${bill.id}/bayar`, { pembayaran: [{ metode: 'tunai', jumlah: 12000, diterima: 20000 }] }))
    expect(lunas.pembayaran[0].kembalian).toBe(8000)
    expect((await data(http.get('/meja'))).meja[0].pesanan).toEqual([])
    expect(mod.offline.antrean).toHaveLength(5)
    const kunci = mod.offline.antrean.map((e) => e.kunci)

    putus = false
    terkirim.length = 0
    await vi.waitFor(async () => {
      await mod.sinkron()
      expect(mod.offline.antrean).toHaveLength(0)
    })
    expect(terkirim.map((t) => t.url)).toEqual([
      '/pesanan', '/pesanan/100/item', '/pesanan/100/item/900', '/pesanan/100/kirim', '/pesanan/100/bayar',
    ])
    expect(terkirim.every((t) => t.offline === '1')).toBe(true)
    expect(terkirim.map((t) => t.idem)).toEqual(kunci) // kunci sama dengan percobaan pertama: server tidak memproses dua kali
    expect(peristiwa.at(-1).detail[bill.id]).toBe(100)
  })

  it('selama antrean belum habis, aksi kasir tetap dijawab lokal supaya urutannya terjaga', async () => {
    putus = true
    const bill = await data(http.post('/pesanan', { mejaId: null, atasNama: 'Andi' }))
    putus = false
    terkirim.length = 0
    const p = await data(http.post(`/pesanan/${bill.id}/item`, { menuId: 1, qty: 1, opsiIds: [11], catatan: '' }))
    expect(p.id).toBe(bill.id)
    expect(terkirim.filter((t) => t.url.includes('/item'))).toHaveLength(0)
  })

  it('Lewati pada bill buatan offline ikut membuang aksi lanjutannya, karena tanpa bill itu semuanya pasti ditolak', async () => {
    await vi.waitFor(async () => {
      await mod.sinkron()
      expect(mod.offline.antrean).toHaveLength(0)
    })
    putus = true
    const bill = await data(http.post('/pesanan', { mejaId: null, atasNama: 'Budi' }))
    await http.post(`/pesanan/${bill.id}/item`, { menuId: 1, qty: 1, opsiIds: [11], catatan: '' })
    await http.post(`/pesanan/${bill.id}/kirim`)
    tolak.add('post /pesanan')
    putus = false
    await vi.waitFor(async () => {
      await mod.sinkron()
      expect(mod.offline.galat).toContain('Meja tidak ditemukan')
    })
    expect(mod.aksiTerkait()).toBe(3)
    mod.lewati()
    expect(mod.offline.antrean).toHaveLength(0)
    tolak.clear()
  })

  it('tiket yang gagal dicetak tersimpan permanen sampai dicetak ulang atau ditutup', async () => {
    putus = true
    cetak.mockReturnValueOnce('Kertas printer habis')
    const bill = await data(http.post('/pesanan', { mejaId: null, atasNama: 'Cici' }))
    await http.post(`/pesanan/${bill.id}/item`, { menuId: 1, qty: 1, opsiIds: [11], catatan: '' })
    await http.post(`/pesanan/${bill.id}/kirim`)
    expect(mod.offline.gagalCetak).toHaveLength(1)
    expect(JSON.parse(localStorage.getItem('offline:gagalCetak'))[0].galat).toContain('Kertas printer habis')
    mod.cetakUlang(0)
    expect(mod.offline.gagalCetak).toHaveLength(0)
    expect(JSON.parse(localStorage.getItem('offline:gagalCetak'))).toEqual([])
  })
})
