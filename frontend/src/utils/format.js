export const rp = (n) => 'Rp' + Number(n || 0).toLocaleString('id-ID')

// Menu dengan grup penentu harga wajib: tampilkan harga terendah ("mulai") karena harga dasar tidak dipakai.
export function hargaTampil(m) {
  const g = m.varian.find((v) => v.gantiHarga && v.wajib)
  if (!g) return rp(m.harga)
  const harga = g.opsi.map((o) => o.tambahan)
  const min = Math.min(...harga)
  return (min === Math.max(...harga) ? '' : 'mulai ') + rp(min)
}

export const hari =(offset = 0) => new Date(new Date().setDate(new Date().getDate() + offset))

export const ymd = (d) => `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`

export const tgl = (d) => new Date(d).toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' })
