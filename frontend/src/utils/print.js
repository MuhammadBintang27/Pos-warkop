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

// Baris yang sama (menu, varian, catatan, harga) digabung di struk. Di sistem tetap terpisah per pengiriman agar dapur menerima tiket tambahan.
export function gabungItem(items) {
  const peta = new Map()
  for (const i of items) {
    const kunci = [i.menuId, i.harga, i.opsi, i.catatan].join('|')
    const ada = peta.get(kunci)
    if (ada) ada.qty += i.qty
    else peta.set(kunci, { ...i })
  }
  return [...peta.values()]
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
