import { rp, tgl } from '@/utils/format'

const CSS = `@page{size:58mm auto;margin:0}
body{font:12px/1.4 monospace;width:48mm;margin:0 auto;padding:4mm 0;color:#000}
h1{font-size:15px;margin:0;text-align:center}.c{text-align:center}.b{font-weight:bold}
.r{display:flex;justify-content:space-between;gap:6px}hr{border:0;border-top:1px dashed #000;margin:5px 0}
.big{font-size:20px;font-weight:bold;text-align:center}.sub{padding-left:10px;font-size:11px}
.judul{font-size:14px;font-weight:bold;text-align:center;border:2px solid #000;padding:2px;margin-bottom:4px}
.page{page-break-after:always}.page:last-child{page-break-after:auto}`

const esc = (s) => String(s ?? '').replace(/[&<>"']/g, (c) => `&#${c.charCodeAt(0)};`)

function printIframe(html) {
  const frame = document.createElement('iframe')
  frame.style.cssText = 'position:fixed;right:0;bottom:0;width:0;height:0;border:0'
  document.body.appendChild(frame)
  const doc = frame.contentWindow.document
  doc.open()
  doc.write(html)
  doc.close()
  frame.contentWindow.onafterprint = () => frame.remove()
  setTimeout(() => frame.contentWindow.print(), 100)
}

function printHtml(body) {
  printIframe(`<!doctype html><html><head><meta charset="utf-8"><style>${CSS}</style></head><body>${body}</body></html>`)
}

// Baris yang tercetak sama (nama, varian, harga) digabung di struk; catatan tidak dicetak di struk jadi tidak membedakan.
// Di sistem tetap terpisah per pengiriman agar dapur menerima tiket tambahan.
export function gabungItem(items) {
  const peta = new Map()
  for (const i of items) {
    const kunci = [i.nama, i.harga, i.opsi].join('|')
    const ada = peta.get(kunci)
    if (ada) ada.qty += i.qty
    else peta.set(kunci, { ...i })
  }
  return [...peta.values()]
}

// gabungTerkirim: tampilan keranjang. Baris yang sudah dikirim dan sama persis jadi satu (rows = baris aslinya);
// baris belum dikirim tetap sendiri supaya jelas mana yang masih harus dikirim.
export function gabungTerkirim(items) {
  const out = []
  const peta = new Map()
  for (const i of items) {
    if (!i.dikirim) {
      out.push(i)
      continue
    }
    const kunci = [i.menuId, i.harga, i.opsi, i.catatan].join('|')
    const ada = peta.get(kunci)
    if (ada) {
      ada.qty += i.qty
      ada.rows.push(i)
    } else {
      const g = { ...i, rows: [i] }
      peta.set(kunci, g)
      out.push(g)
    }
  }
  return out
}

function perTenant(items) {
  const grup = new Map()
  for (const i of items) {
    if (!grup.has(i.tenantId)) grup.set(i.tenantId, { id: i.tenantId, nama: i.tenantNama, items: [] })
    grup.get(i.tenantId).items.push(i)
  }
  return [...grup.values()]
}

const barisItem = (i) => `<div class="b">${i.qty}x ${esc(i.nama)}</div>` +
  (i.opsi ? `<div class="sub">${esc(i.opsi)}</div>` : '') +
  (i.catatan ? `<div class="sub b">* ${esc(i.catatan)}</div>` : '')

export function printStruk(p, usaha) {
  const baris = (i) => `<div class="r"><span>${i.qty}x ${esc(i.nama)}</span><span>${rp(i.harga * i.qty)}</span></div>` +
    (i.opsi ? `<div class="sub">${esc(i.opsi)}</div>` : '')
  const items = gabungItem(p.items).map(baris).join('')
  const bayar = p.pembayaran
    .map((b) => `<div class="r"><span>${esc(b.metode)}${b.pelanggan ? ' ' + esc(b.pelanggan) : ''}</span><span>${rp(b.diterima)}</span></div>` +
      (b.kembalian ? `<div class="r"><span>Kembali</span><span>${rp(b.kembalian)}</span></div>` : ''))
    .join('')
  printHtml(`<h1>${esc(usaha.namaUsaha)}</h1>
    <div class="c">${esc(usaha.alamat)}</div><hr>
    <div class="r"><span>${esc(p.mejaNama || 'Bawa pulang')}</span><span>#${p.id}</span></div>
    ${p.atasNama ? `<div>${esc(p.atasNama)}</div>` : ''}
    <div>${tgl(p.lunasAt || p.dibuatAt)}</div>${p.kasir ? `<div>Kasir: ${esc(p.kasir)}</div>` : ''}<hr>
    ${items}<hr>
    <div class="r"><span>Subtotal</span><span>${rp(p.subtotal)}</span></div>
    ${p.diskon ? `<div class="r"><span>Diskon</span><span>-${rp(p.diskon)}</span></div>` : ''}
    <div class="r b"><span>TOTAL</span><span>${rp(p.total)}</span></div><hr>
    ${bayar}<hr><div class="c">${esc(usaha.footerStruk)}</div>`)
}

export function printClosing(b, detail, usaha, tertunda) {
  const menu = (detail?.menu ?? [])
    .filter((m) => m.qty)
    .map((m) => `<div class="r"><span>${m.qty}x ${esc(m.nama)}</span><span>${rp(m.nilai)}</span></div>`)
    .join('')
  const jejak = (label, oleh, waktu) => (waktu ? `<div>${label}: ${esc(oleh)}</div><div class="sub">${tgl(waktu)}</div>` : '')
  printHtml(`<h1>${esc(usaha.namaUsaha)}</h1>
    <div class="judul">${b.milikSendiri ? 'REKAP CLOSING HARIAN' : 'BILL CLOSING TENANT'}</div>
    <div class="b">${esc(b.nama)}</div>
    <div>Tanggal: ${b.tanggal.split('-').reverse().join('/')}</div>
    <div>Jumlah bill: ${b.transaksi}</div><hr>
    ${menu || '<div class="c">-</div>'}<hr>
    <div class="r"><span>Penjualan</span><span>${rp(b.kotor)}</span></div>
    <div class="r"><span>Diskon</span><span>-${rp(b.diskon)}</span></div>
    ${b.milikSendiri
      ? `<div class="r b"><span>PENDAPATAN PEMILIK</span><span>${rp(b.komisi)}</span></div>`
      : `<div class="r"><span>Komisi ${b.komisiPersen}%</span><span>-${rp(b.komisi)}</span></div>
    <div class="r b"><span>SETOR TENANT</span><span>${rp(b.setor)}</span></div>`}
    ${detail?.talangan?.length ? `<div class="sub">Termasuk ${detail.talangan.length} bill belum dibayar pelanggan (${rp(detail.talangan.reduce((s, t) => s + t.kotor, 0))}), ditalangi pemilik</div>` : ''}
    ${detail?.batalQty ? `<div class="sub">Dibatalkan: ${detail.batalQty} porsi (${rp(detail.batalNilai)})</div>` : ''}
    ${tertunda ? `<div class="sub">Belum dibayar pelanggan: ${tertunda.bill} bill (${rp(tertunda.nilai)}), belum termasuk di atas</div>` : ''}<hr>
    ${jejak('Closing', b.dibuatOleh, b.dibuatAt)}
    ${jejak('Disetor', b.disetorOleh, b.disetorAt)}
    ${b.milikSendiri ? '' : `<hr><div class="r"><span>Tenant</span><span>Pemilik</span></div>
    <br><br><br>
    <div class="r"><span>(_________)</span><span>(_________)</span></div>`}`)
}

// Rekap setoran satu periode (fallback browser untuk cetakRekapTenant).
export function printRekapTenant(p, usaha) {
  const tglPendek = (s) => s.slice(8) + '/' + s.slice(5, 7)
  const jumlah = (k) => p.hari.reduce((s, h) => s + h[k], 0)
  const belum = p.hari.filter((h) => h.status === 'belum' || h.status === 'berubah').length
  const baris = p.hari
    .map((h) => `<div class="r"><span>${tglPendek(h.tanggal)} ${h.transaksi} bill${h.status === 'belum' || h.status === 'berubah' ? '*' : ''}</span><span>${rp(h.kotor)}</span></div>`)
    .join('')
  const tanggal = (s) => s.split('-').reverse().join('/')
  printHtml(`<h1>${esc(usaha.namaUsaha)}</h1>
    <div class="judul">${p.milikSendiri ? 'REKAP PENDAPATAN' : 'REKAP SETORAN TENANT'}</div>
    <div class="c b">${esc(p.tenant)}</div>
    <div class="c">${tanggal(p.dari)} - ${tanggal(p.sampai)}</div><hr>
    ${baris}<hr>
    <div class="r"><span>Jumlah bill</span><span>${jumlah('transaksi')}</span></div>
    <div class="r"><span>Penjualan</span><span>${rp(jumlah('kotor'))}</span></div>
    <div class="r"><span>Diskon</span><span>-${rp(jumlah('diskon'))}</span></div>
    ${p.milikSendiri
      ? `<div class="r b"><span>PENDAPATAN</span><span>${rp(jumlah('komisi'))}</span></div>`
      : `<div class="r"><span>Komisi ${p.komisiPersen}%</span><span>-${rp(jumlah('komisi'))}</span></div>
    <div class="r b"><span>SETOR TENANT</span><span>${rp(jumlah('setor'))}</span></div>`}
    ${belum ? `<div class="sub">* ${belum} hari belum di-closing atau berubah setelah closing; angkanya masih bisa berubah</div>` : ''}<hr>
    <div class="sub">Dicetak ${tgl(new Date())}</div>
    ${p.milikSendiri ? '' : `<div class="r"><span>Tenant</span><span>Pemilik</span></div>
    <br><br><br>
    <div class="r"><span>(_________)</span><span>(_________)</span></div>`}`)
}

// Laporan tutup shift (fallback browser untuk cetakShift).
export function printShift(d, usaha) {
  const baris = (k, v, cls = '') => `<div class="r ${cls}"><span>${k}</span><span>${v}</span></div>`
  printHtml(`<h1>${esc(usaha.namaUsaha)}</h1>
    <div class="judul">LAPORAN SHIFT #${d.id}</div>
    <div>Buka : ${tgl(d.bukaAt)} ${esc(d.kasir)}</div>
    <div>Tutup: ${d.tutupAt ? `${tgl(d.tutupAt)} ${esc(d.tutupOleh)}` : 'belum ditutup'}</div><hr>
    ${baris('Bill lunas', d.bill)}
    ${d.billBatal ? baris('Bill batal', d.billBatal) : ''}
    ${d.perMetode.map((m) => baris('&nbsp;&nbsp;' + esc(m.metode), rp(m.total))).join('')}
    ${baris('Penjualan', rp(d.penjualan), 'b')}<hr>
    ${baris('Kas awal', rp(d.kasAwal))}
    ${baris('+ Penjualan tunai', rp(d.penjualanTunai))}
    ${d.kasbonTunai ? baris('+ Cicilan kasbon tunai', rp(d.kasbonTunai)) : ''}
    ${baris('- Pengeluaran', rp(d.keluar))}
    ${d.pengeluaran.map((e) => baris('&nbsp;&nbsp;' + esc(e.keterangan), rp(e.jumlah))).join('')}<hr>
    ${d.kasSistem != null ? baris('Kas sistem', rp(d.kasSistem), 'b') : ''}
    ${d.kasHitung != null ? baris('Kas dihitung', rp(d.kasHitung)) : ''}
    ${d.selisih != null ? baris('SELISIH', rp(d.selisih), 'b') : ''}<hr>
    <div class="sub">Dicetak ${tgl(new Date())}</div>`)
}

// Tiket tenant yang sudah diantrekan ke printernya dicetak oleh stasiun cetak; perangkat ini hanya mencetak sisanya.
export const sisaLokal = (t) => ({ ...t, items: t.items.filter((i) => !t.antrean?.[i.tenantId]) })

// Satu halaman (satu tiket) per tenant, dalam satu pekerjaan cetak.
export function printDapur(t) {
  const halaman = perTenant(t.items).map((g) => `<div class="page">
    <div class="judul">${t.tambahan?.[g.id] ? 'PESANAN TAMBAHAN' : 'PESANAN BARU'}</div>
    <div class="big">${esc(t.mejaNama || 'BAWA PULANG')}</div>
    ${t.atasNama ? `<div class="c">${esc(t.atasNama)}</div>` : ''}
    <div class="c b">${esc(g.nama)}</div>
    <div class="c">${tgl(new Date())}</div>
    <div class="c">Dikirim oleh: ${esc(t.oleh)}</div><hr>${g.items.map(barisItem).join('')}</div>`)
  printHtml(halaman.join(''))
}

export function printBatalDapur(t) {
  const halaman = perTenant(t.items).map((g) => `<div class="page">
    <div class="big">*** BATAL ***</div>
    <div class="big">${esc(t.mejaNama || 'BAWA PULANG')}</div>
    ${t.atasNama ? `<div class="c">${esc(t.atasNama)}</div>` : ''}
    <div class="c b">${esc(g.nama)}</div>
    <div class="c">${tgl(new Date())}</div>
    <div class="c">Dibatalkan oleh: ${esc(t.oleh)}</div><hr>
    ${g.items.map((i) => `<div class="b">- ${i.qty}x ${esc(i.nama)}</div>` + (i.opsi ? `<div class="sub">${esc(i.opsi)}</div>` : '')).join('')}<hr>
    <div class="c b">JANGAN DIBUAT / HENTIKAN</div></div>`)
  printHtml(halaman.join(''))
}
