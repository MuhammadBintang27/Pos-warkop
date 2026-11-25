import { ref } from 'vue'
import { defineStore } from 'pinia'
import { del, get, post, put } from '@/api/http'

export const useKasirStore = defineStore('kasir', () => {
  const kategori = ref([])
  const tenant = ref([])
  const cetak = ref({ antri: 0, macet: 0, stasiunAktif: true, daftar: [] })
  const denah = ref({ meja: [], takeaway: [] })
  const pesanan = ref(null)
  const draft = ref(null)
  const shift = ref(null)

  const pid = () => pesanan.value.id
  const path = (suffix = '') => `/pesanan/${pid()}${suffix}`

  const loadMenu = async () => (kategori.value = (await get('/menu')).kategori)
  const loadCetak = async () => (cetak.value = await get('/cetak/status'))
  const loadTenant = async () => (tenant.value = await get('/tenant'))
  const loadDenah = async () => (denah.value = await get('/meja'))
  const loadShift = async () => (shift.value = (await get('/shift/aktif')) ?? null)

  async function refreshPesanan() {
    if (!pesanan.value) return
    try {
      const p = await get(path())
      pesanan.value = p.status === 'open' ? p : null
    } catch {
      pesanan.value = null
    }
  }

  async function bukaPesanan(id) {
    pesanan.value = await get(`/pesanan/${id}`)
    draft.value = null
  }

  // Meja kosong hanya dicatat di sisi klien; pesanannya dibuat server saat item pertama masuk.
  async function bukaMeja(meja) {
    if (meja.pesanan.length) return bukaPesanan(meja.pesanan[0].id)
    pesanan.value = null
    draft.value = { mejaId: meja.id, mejaNama: meja.nama, atasNama: '' }
  }

  const baru = async (atasNama) => {
    pesanan.value = null
    draft.value = { mejaId: null, mejaNama: '', atasNama }
  }

  const tutup = () => {
    pesanan.value = null
    draft.value = null
  }

  async function addItem(body) {
    if (pesanan.value) return (pesanan.value = await post(path('/item'), body))
    const d = draft.value
    pesanan.value = await post('/pesanan', { mejaId: d.mejaId, atasNama: d.atasNama })
    draft.value = null
    try {
      pesanan.value = await post(path('/item'), body)
    } catch (e) {
      if (!pesanan.value.items.length) {
        await post(path('/batal'), {}).catch(() => {})
        pesanan.value = null
        draft.value = d
      }
      throw e
    }
  }

  const updateItem = async (item, qty, catatan) => (pesanan.value = await put(path(`/item/${item.id}`), { qty, catatan }))
  const hapusItem = async (item) => (pesanan.value = await del(path(`/item/${item.id}`)))
  const setDiskon = async (diskon) => (pesanan.value = await put(path('/diskon'), { diskon }))
  const kirim = () => post(path('/kirim'))
  const pindah = async (mejaId) => (pesanan.value = await post(path('/pindah'), { mejaId }))
  const gabung = async (kePesananId) => (pesanan.value = await post(path('/gabung'), { kePesananId }))
  const pisah = async (items) => (pesanan.value = await post(path('/pisah'), { items }))

  // kedua pembatalan mengembalikan tiket BATAL untuk dapur bila ada item yang sudah terkirim
  async function batal(body) {
    const res = await post(path('/batal'), body)
    pesanan.value = null
    return res.tiket
  }

  // item bisa baris gabungan (item.rows): jumlahnya dibatalkan dari baris terbaru dulu, tiketnya disatukan
  async function batalItem(item, body) {
    let sisa = body.qty
    let tiket = null
    for (const row of [...(item.rows ?? [item])].reverse()) {
      if (sisa <= 0) break
      const qty = Math.min(sisa, row.qty)
      const res = await post(path(`/item/${row.id}/batal`), { ...body, qty })
      pesanan.value = res.pesanan
      sisa -= qty
      if (!res.tiket) continue
      if (!tiket) tiket = res.tiket
      else {
        tiket.items = [...(tiket.items ?? []), ...(res.tiket.items ?? [])]
        Object.assign(tiket.antrean ?? (tiket.antrean = {}), res.tiket.antrean)
      }
    }
    return tiket
  }

  async function bayar(pembayaran) {
    const lunas = await post(path('/bayar'), { pembayaran })
    pesanan.value = null
    return lunas
  }

  return {
    kategori, tenant, cetak, denah, pesanan, draft, shift,
    loadMenu, loadTenant, loadCetak, loadDenah, loadShift, refreshPesanan,
    bukaMeja, bukaPesanan, baru, tutup,
    addItem, updateItem, hapusItem, setDiskon, kirim, pindah, gabung, pisah, batal, batalItem, bayar,
  }
})
