import { post } from '@/api/http'
import { printClosing, printRekapTenant, printStruk } from '@/utils/print'

// Bila printer kasir diatur di Pengaturan, dokumen dicetak server lewat stasiun cetak; bila tidak, lewat dialog cetak browser.
// Mengembalikan true bila masuk antrean printer.
export async function cetakStruk(pesanan, usaha) {
  if (usaha.printerKasirId) {
    const { antre } = await post(`/pesanan/${pesanan.id}/cetak-struk`)
    if (antre) return true
  }
  printStruk(pesanan, usaha)
  return false
}

export async function cetakClosing(b, detail, usaha, tertunda) {
  if (usaha.printerKasirId) {
    const { antre } = await post('/cetak/closing', {
      tenant: b.nama,
      tanggal: b.tanggal,
      milikSendiri: b.milikSendiri,
      transaksi: b.transaksi,
      kotor: b.kotor,
      diskon: b.diskon,
      komisiPersen: b.komisiPersen,
      komisi: b.komisi,
      setor: b.setor,
      batalQty: detail?.batalQty ?? 0,
      batalNilai: detail?.batalNilai ?? 0,
      talanganBill: detail?.talangan?.length ?? 0,
      talanganNilai: (detail?.talangan ?? []).reduce((s, t) => s + t.kotor, 0),
      tertundaBill: tertunda?.bill ?? 0,
      tertundaNilai: tertunda?.nilai ?? 0,
      dibuatOleh: b.dibuatOleh ?? '',
      dibuatAt: b.dibuatAt ?? null,
      disetorOleh: b.disetorOleh ?? '',
      disetorAt: b.disetorAt ?? null,
      menu: (detail?.menu ?? []).filter((m) => m.qty).map((m) => ({ qty: m.qty, nama: m.nama, nilai: m.nilai })),
    })
    if (antre) return true
  }
  printClosing(b, detail, usaha, tertunda)
  return false
}

export async function cetakRekapTenant(p, usaha) {
  if (usaha.printerKasirId) {
    const { antre } = await post('/cetak/rekap-tenant', p)
    if (antre) return true
  }
  printRekapTenant(p, usaha)
  return false
}
