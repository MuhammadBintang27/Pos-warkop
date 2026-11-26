// Perintah printer thermal ESC/POS untuk dicetak HP kasir langsung ke printer (mode offline).
// Salinan escposTiket/escposStruk/logoRaster di backend (services/escpos*.go): ubah keduanya bersamaan.
const ESC = 0x1b
const GS = 0x1d
const KARAKTER = { 58: 32, 80: 48 }
const TITIK = { 58: 384, 80: 576 } // lebar area cetak 203 dpi

// printer hanya mengerti ASCII sederhana: aksen dibuang, karakter lain jadi "?"
export const ascii = (s) =>
  String(s ?? '')
    .normalize('NFD')
    .replace(/[̀-ͯ]/g, '')
    .replace(/[^\x20-\x7e]/gu, '?')

export function bungkus(teks, lebar, indent = '') {
  const baris = []
  let kini = indent
  for (const kata of ascii(teks).split(/\s+/).filter(Boolean)) {
    const ada = kini.trim() !== ''
    const calon = ada ? `${kini} ${kata}` : indent + kata
    if (calon.length > lebar && ada) {
      baris.push(kini)
      kini = indent + kata
    } else kini = calon
  }
  if (kini.trim()) baris.push(kini)
  return baris.length ? baris : [indent]
}

const dua = (n) => String(n).padStart(2, '0')
export const jamCetak = (d) => {
  const t = new Date(d)
  return `${dua(t.getDate())}/${dua(t.getMonth() + 1)}/${dua(t.getFullYear() % 100)} ${dua(t.getHours())}.${dua(t.getMinutes())}`
}

export const rupiah = (n) => (n < 0 ? '-' : '') + 'Rp ' + String(Math.abs(n)).replace(/\B(?=(\d{3})+(?!\d))/g, '.')

function penulis(lebarMM) {
  const lebar = KARAKTER[lebarMM] || 32
  const b = [ESC, 0x40]
  const w = {
    lebar,
    raw: (x) => {
      for (const v of x) b.push(v)
    },
    baris: (t) => {
      for (const c of ascii(t)) b.push(c.charCodeAt(0))
      b.push(10)
    },
    rata: (n) => b.push(ESC, 0x61, n), // 0 kiri, 1 tengah
    ukuran: (n) => b.push(GS, 0x21, n), // 0 normal, 0x11 dua kali
    tebal: (aktif) => b.push(ESC, 0x45, aktif ? 1 : 0),
    garis: () => w.baris('-'.repeat(lebar)),
    tengah: (t, l) => {
      w.rata(1)
      bungkus(t, l).forEach((x) => w.baris(x))
      w.rata(0)
    },
    // teks kiri boleh berbaris-baris, nilai rata kanan di baris terakhir
    kolom: (kiri, kanan) => {
      kiri = ascii(kiri)
      kanan = ascii(kanan)
      const ruang = lebar - kanan.length - 1
      if (ruang < 1) {
        w.baris(kiri)
        w.baris(' '.repeat(Math.max(lebar - kanan.length, 0)) + kanan)
        return
      }
      const bagian = bungkus(kiri, ruang)
      bagian.forEach((l, i) => w.baris(i < bagian.length - 1 ? l : l + ' '.repeat(lebar - l.length - kanan.length) + kanan))
    },
    akhir: () => {
      b.push(ESC, 0x64, 4) // dorong kertas supaya teks terakhir lolos pisau
      b.push(GS, 0x56, 0x42, 0) // potong sebagian
      return Uint8Array.from(b)
    },
  }
  return w
}

// p: { judul, mejaNama, atasNama, oleh, tenant, waktu, items: [{ qty, nama, opsi, catatan }] }
export function escposTiket(p, lebarMM) {
  const w = penulis(lebarMM)
  const batal = p.judul === 'BATAL'
  w.rata(1)
  w.tebal(true)
  w.ukuran(0x11)
  bungkus(p.judul, w.lebar / 2).forEach((l) => w.baris(l))
  bungkus(p.mejaNama || 'BAWA PULANG', w.lebar / 2).forEach((l) => w.baris(l))
  w.ukuran(0)
  if (p.atasNama) w.baris(p.atasNama)
  w.baris(p.tenant)
  w.tebal(false)
  w.baris(jamCetak(p.waktu))
  if (p.oleh) w.baris(`${batal ? 'Dibatalkan' : 'Dikirim'} oleh: ${p.oleh}`)
  w.rata(0)
  w.garis()
  for (const i of p.items) {
    w.tebal(true)
    bungkus(`${i.qty}x ${i.nama}`, w.lebar).forEach((l) => w.baris(l))
    w.tebal(false)
    if (i.opsi) bungkus(i.opsi, w.lebar, '  ').forEach((l) => w.baris(l))
    if (i.catatan) {
      w.tebal(true)
      bungkus(`* ${i.catatan}`, w.lebar, '  ').forEach((l) => w.baris(l))
      w.tebal(false)
    }
  }
  w.garis()
  if (batal) {
    w.rata(1)
    w.tebal(true)
    w.baris('JANGAN DIBUAT / HENTIKAN')
    w.tebal(false)
  }
  return w.akhir()
}

const kapital = (s) => (s ? s[0].toUpperCase() + s.slice(1) : s)

// p: { usaha, alamat, footer, id, mejaNama, atasNama, kasir, waktu, items: [{ qty, nama, opsi, harga }], subtotal, diskon, total,
//      bayar: [{ metode, pelanggan, diterima, kembalian }] }; logo: hasil logoRaster atau null
export function escposStruk(p, lebarMM, logo) {
  const w = penulis(lebarMM)
  if (logo) {
    w.rata(1)
    w.raw(logo)
    w.rata(0)
  }
  w.tebal(true)
  w.ukuran(0x11)
  w.tengah(p.usaha, w.lebar / 2)
  w.ukuran(0)
  w.tebal(false)
  if (p.alamat) w.tengah(p.alamat, w.lebar)
  w.garis()
  w.kolom(p.mejaNama || 'Bawa pulang', `#${p.id}`)
  if (p.atasNama) w.baris(p.atasNama)
  w.baris(jamCetak(p.waktu))
  if (p.kasir) w.baris(`Kasir: ${p.kasir}`)
  w.garis()
  for (const i of p.items) {
    w.kolom(`${i.qty}x ${i.nama}`, rupiah(i.harga * i.qty))
    if (i.opsi) bungkus(i.opsi, w.lebar, '  ').forEach((l) => w.baris(l))
  }
  w.garis()
  w.kolom('Subtotal', rupiah(p.subtotal))
  if (p.diskon > 0) w.kolom('Diskon', '-' + rupiah(p.diskon))
  w.tebal(true)
  w.kolom('TOTAL', rupiah(p.total))
  w.tebal(false)
  w.garis()
  for (const b of p.bayar) {
    w.kolom(kapital(b.metode) + (b.pelanggan ? ` ${b.pelanggan}` : ''), rupiah(b.diterima))
    if (b.kembalian > 0) w.kolom('Kembali', rupiah(b.kembalian))
  }
  w.garis()
  if (p.footer) w.tengah(p.footer, w.lebar)
  return w.akhir()
}

// Logo jadi gambar hitam putih paling lebar separuh kertas: ambang 50% pada kanal tergelap, transparan = kertas putih.
// Gambar yang gagal dibaca dilewati (null): struk tetap tercetak tanpa logo.
const logoCache = new Map()
export function logoRaster(url, lebarMM) {
  if (!url) return Promise.resolve(null)
  const kunci = `${lebarMM}|${url}`
  if (!logoCache.has(kunci)) logoCache.set(kunci, buatRaster(url, lebarMM).catch(() => null))
  return logoCache.get(kunci)
}

async function buatRaster(url, lebarMM) {
  const img = new Image()
  img.src = url
  await img.decode()
  const sw = img.naturalWidth
  const sh = img.naturalHeight
  if (!sw || !sh) return null
  const c = document.createElement('canvas')
  c.width = sw
  c.height = sh
  const ctx = c.getContext('2d')
  ctx.drawImage(img, 0, 0)
  const px = ctx.getImageData(0, 0, sw, sh).data
  const titik = TITIK[lebarMM] || 384
  const skala = Math.min(titik / 2 / sw, titik / 3 / sh, 1)
  const w = Math.max(Math.floor(sw * skala), 1)
  const h = Math.max(Math.floor(sh * skala), 1)
  const bw = (w + 7) >> 3
  const data = new Uint8Array(bw * h)
  for (let y = 0; y < h; y++) {
    const y0 = Math.floor((y * sh) / h)
    const y1 = Math.max(Math.floor(((y + 1) * sh) / h), y0 + 1)
    for (let x = 0; x < w; x++) {
      const x0 = Math.floor((x * sw) / w)
      const x1 = Math.max(Math.floor(((x + 1) * sw) / w), x0 + 1)
      let jumlah = 0
      let n = 0
      for (let sy = y0; sy < y1; sy++) {
        for (let sx = x0; sx < x1; sx++) {
          const i = (sy * sw + sx) * 4
          const a = px[i + 3]
          jumlah += (Math.min(px[i], px[i + 1], px[i + 2]) * a) / 255 + 255 - a
          n++
        }
      }
      if (jumlah / n < 128) data[y * bw + (x >> 3)] |= 0x80 >> (x & 7)
    }
  }
  return [GS, 0x76, 0x30, 0, bw & 0xff, bw >> 8, h & 0xff, h >> 8, ...data]
}

// Kirim langsung ke printer jaringan lewat aplikasi Android (port 9100). Melempar Error berisi alasan bila gagal.
export const bisaCetakLangsung = () => !!window.StasiunAndroid?.cetak

export function cetakLangsung(printer, bytes) {
  if (!bisaCetakLangsung()) throw new Error('Cetak langsung butuh aplikasi Android versi terbaru')
  let s = ''
  for (const v of bytes) s += String.fromCharCode(v)
  const galat = window.StasiunAndroid.cetak(printer.ip, printer.port || 9100, btoa(s))
  if (galat) throw new Error(`${printer.nama}: ${galat}`)
}
