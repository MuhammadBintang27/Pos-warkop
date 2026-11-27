package services

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"warkop/internal/dto"
)

// penulis menyusun perintah ESC/POS untuk dokumen dua kolom (struk, closing).
type penulis struct {
	b     bytes.Buffer
	lebar int
}

func barisPenulis(lebarMM int) *penulis {
	lebar := karakterPerBaris[lebarMM]
	if lebar == 0 {
		lebar = 32
	}
	w := &penulis{lebar: lebar}
	w.b.Write([]byte{esc, 0x40})
	return w
}

func (w *penulis) baris(t string) { w.b.WriteString(ascii(t)); w.b.WriteByte('\n') }
func (w *penulis) rata(n byte)    { w.b.Write([]byte{esc, 0x61, n}) }
func (w *penulis) ukuran(n byte)  { w.b.Write([]byte{gs, 0x21, n}) }
func (w *penulis) garis()         { w.baris(strings.Repeat("-", w.lebar)) }
func (w *penulis) tebal(aktif bool) {
	v := byte(0)
	if aktif {
		v = 1
	}
	w.b.Write([]byte{esc, 0x45, v})
}

func (w *penulis) tengah(t string, lebar int) {
	w.rata(1)
	for _, l := range bungkus(t, lebar, "") {
		w.baris(l)
	}
	w.rata(0)
}

// kolom: teks kiri (boleh berbaris-baris) dan nilai rata kanan pada baris terakhir.
func (w *penulis) kolom(kiri, kanan string) {
	kiri, kanan = ascii(kiri), ascii(kanan)
	ruang := w.lebar - len(kanan) - 1
	if ruang < 1 {
		w.baris(kiri)
		w.baris(strings.Repeat(" ", max(w.lebar-len(kanan), 0)) + kanan)
		return
	}
	bagian := bungkus(kiri, ruang, "")
	for i, l := range bagian {
		if i < len(bagian)-1 {
			w.baris(l)
			continue
		}
		w.baris(l + strings.Repeat(" ", w.lebar-len(l)-len(kanan)) + kanan)
	}
}

func (w *penulis) akhir() []byte {
	w.b.Write([]byte{esc, 0x64, 4})      // dorong kertas supaya teks terakhir lolos pisau
	w.b.Write([]byte{gs, 0x56, 0x42, 0}) // potong sebagian
	return w.b.Bytes()
}

// rupiah memformat 12500 menjadi "Rp 12.500".
func rupiah(n int) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.Itoa(n)
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	out := "Rp " + strings.Join(parts, ".")
	if neg {
		return "-" + out
	}
	return out
}

func kapital(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// logo: hasil logoRaster, nil bila klien tidak punya logo.
func escposStruk(p dto.StrukPayload, lebarMM int, logo []byte) []byte {
	w := barisPenulis(lebarMM)
	if logo != nil {
		w.rata(1)
		w.b.Write(logo)
		w.rata(0)
	}
	w.tebal(true)
	w.ukuran(0x11)
	w.tengah(p.Usaha, w.lebar/2)
	w.ukuran(0)
	w.tebal(false)
	if p.Alamat != "" {
		w.tengah(p.Alamat, w.lebar)
	}
	w.garis()
	meja := p.MejaNama
	if meja == "" {
		meja = "Bawa pulang"
	}
	w.kolom(meja, "#"+strconv.FormatInt(p.ID, 10))
	if p.AtasNama != "" {
		w.baris(p.AtasNama)
	}
	w.baris(p.Waktu.In(time.Local).Format("02/01/06 15.04"))
	if p.Kasir != "" {
		w.baris("Kasir: " + p.Kasir)
	}
	w.garis()
	for _, i := range p.Items {
		w.kolom(strconv.Itoa(i.Qty)+"x "+i.Nama, rupiah(i.Harga*i.Qty))
		if i.Opsi != "" {
			for _, l := range bungkus(i.Opsi, w.lebar, "  ") {
				w.baris(l)
			}
		}
	}
	w.garis()
	w.kolom("Subtotal", rupiah(p.Subtotal))
	if p.Diskon > 0 {
		w.kolom("Diskon", "-"+rupiah(p.Diskon))
	}
	w.tebal(true)
	w.kolom("TOTAL", rupiah(p.Total))
	w.tebal(false)
	w.garis()
	if len(p.Bayar) == 0 {
		w.tebal(true)
		w.tengah("BELUM DIBAYAR", w.lebar)
		w.tebal(false)
	}
	for _, b := range p.Bayar {
		nama := kapital(b.Metode)
		if b.Pelanggan != "" {
			nama += " " + b.Pelanggan
		}
		w.kolom(nama, rupiah(b.Diterima))
		if b.Kembalian > 0 {
			w.kolom("Kembali", rupiah(b.Kembalian))
		}
	}
	w.garis()
	if p.Footer != "" {
		w.tengah(p.Footer, w.lebar)
	}
	return w.akhir()
}

func tanggalIndonesia(iso string) string {
	if t, err := time.Parse("2006-01-02", iso); err == nil {
		return t.Format("02/01/2006")
	}
	return iso
}

func escposClosing(p dto.ClosingPayload, lebarMM int) []byte {
	w := barisPenulis(lebarMM)
	w.tebal(true)
	w.ukuran(0x11)
	w.tengah(p.Usaha, w.lebar/2)
	w.ukuran(0)
	judul := "BILL CLOSING TENANT"
	if p.MilikSendiri {
		judul = "REKAP CLOSING HARIAN"
	}
	w.tengah(judul, w.lebar)
	w.baris(p.Tenant)
	w.tebal(false)
	w.baris("Tanggal: " + tanggalIndonesia(p.Tanggal))
	w.baris("Jumlah bill: " + strconv.Itoa(p.Transaksi))
	w.garis()
	ada := false
	for _, m := range p.Menu {
		if m.Qty > 0 {
			ada = true
			w.kolom(strconv.Itoa(m.Qty)+"x "+m.Nama, rupiah(m.Nilai))
		}
	}
	if !ada {
		w.tengah("-", w.lebar)
	}
	w.garis()
	w.kolom("Penjualan", rupiah(p.Kotor))
	w.kolom("Diskon", "-"+rupiah(p.Diskon))
	if p.MilikSendiri {
		w.tebal(true)
		w.kolom("PENDAPATAN PEMILIK", rupiah(p.Komisi))
	} else {
		w.kolom(labelKomisi(p.KomisiPersen, p.KomisiPerMenu), "-"+rupiah(p.Komisi))
		w.tebal(true)
		w.kolom("SETOR TENANT", rupiah(p.Setor))
	}
	w.tebal(false)
	if p.BatalQty > 0 {
		w.baris("Dibatalkan: " + strconv.Itoa(p.BatalQty) + " porsi (" + rupiah(p.BatalNilai) + ")")
	}
	if p.TalanganBill > 0 {
		for _, l := range bungkus("Termasuk "+strconv.Itoa(p.TalanganBill)+" bill belum dibayar pelanggan ("+rupiah(p.TalanganNilai)+"), ditalangi pemilik", w.lebar, "") {
			w.baris(l)
		}
	}
	if p.TertundaBill > 0 {
		for _, l := range bungkus("Belum dibayar pelanggan: "+strconv.Itoa(p.TertundaBill)+" bill ("+rupiah(p.TertundaNilai)+"), belum termasuk di atas", w.lebar, "") {
			w.baris(l)
		}
	}
	w.garis()
	if p.DibuatAt != nil {
		w.baris("Closing: " + p.DibuatOleh)
		w.baris("  " + p.DibuatAt.In(time.Local).Format("02/01/06 15.04"))
	}
	if p.DisetorAt != nil {
		w.baris("Disetor: " + p.DisetorOleh)
		w.baris("  " + p.DisetorAt.In(time.Local).Format("02/01/06 15.04"))
	}
	if !p.MilikSendiri {
		w.garis()
		w.kolom("Tenant", "Pemilik")
		w.baris("")
		w.baris("")
		w.kolom("(_________)", "(_________)")
	}
	return w.akhir()
}

func labelKomisi(persen float64, perMenu bool) string {
	if perMenu {
		return "Komisi per menu"
	}
	return "Komisi " + strconv.FormatFloat(persen, 'f', -1, 64) + "%"
}

// escposRekapTenant: struk bukti setor satu periode, satu baris per hari plus total; ditandatangani tenant dan pemilik.
func escposRekapTenant(p dto.RekapTenantPayload, lebarMM int) []byte {
	w := barisPenulis(lebarMM)
	w.tebal(true)
	w.ukuran(0x11)
	w.tengah(p.Usaha, w.lebar/2)
	w.ukuran(0)
	judul := "REKAP SETORAN TENANT"
	if p.MilikSendiri {
		judul = "REKAP PENDAPATAN"
	}
	w.tengah(judul, w.lebar)
	w.tengah(p.Tenant, w.lebar)
	w.tebal(false)
	w.tengah(tanggalIndonesia(p.Dari)+" - "+tanggalIndonesia(p.Sampai), w.lebar)
	w.garis()
	var bill, kotor, diskon, komisi, setor, belum int
	for _, h := range p.Hari {
		bill, kotor, diskon, komisi, setor = bill+h.Transaksi, kotor+h.Kotor, diskon+h.Diskon, komisi+h.Komisi, setor+h.Setor
		tanda := ""
		if h.Status == "belum" || h.Status == "berubah" {
			tanda, belum = "*", belum+1
		}
		w.kolom(tanggalIndonesia(h.Tanggal)[:5]+" "+strconv.Itoa(h.Transaksi)+" bill"+tanda, rupiah(h.Kotor))
	}
	w.garis()
	w.kolom("Jumlah bill", strconv.Itoa(bill))
	w.kolom("Penjualan", rupiah(kotor))
	w.kolom("Diskon", "-"+rupiah(diskon))
	w.tebal(true)
	if p.MilikSendiri {
		w.kolom("PENDAPATAN", rupiah(komisi))
	} else {
		w.tebal(false)
		w.kolom(labelKomisi(p.KomisiPersen, p.KomisiPerMenu), "-"+rupiah(komisi))
		w.tebal(true)
		w.kolom("SETOR TENANT", rupiah(setor))
	}
	w.tebal(false)
	if belum > 0 {
		for _, l := range bungkus("* "+strconv.Itoa(belum)+" hari belum di-closing atau berubah setelah closing; angkanya masih bisa berubah", w.lebar, "") {
			w.baris(l)
		}
	}
	w.garis()
	w.baris("Dicetak " + time.Now().Format("02/01/06 15.04"))
	if !p.MilikSendiri {
		w.kolom("Tenant", "Pemilik")
		w.baris("")
		w.baris("")
		w.kolom("(_________)", "(_________)")
	}
	return w.akhir()
}

// escposShift: laporan tutup shift: penjualan per metode lalu perhitungan kas laci.
func escposShift(usaha string, d *dto.ShiftDetail, lebarMM int) []byte {
	w := barisPenulis(lebarMM)
	w.tebal(true)
	w.ukuran(0x11)
	w.tengah(usaha, w.lebar/2)
	w.ukuran(0)
	w.tengah("LAPORAN SHIFT #"+strconv.FormatInt(d.ID, 10), w.lebar)
	w.tebal(false)
	w.baris("Buka : " + d.BukaAt.In(time.Local).Format("02/01/06 15.04") + " " + d.Kasir)
	if d.TutupAt != nil {
		w.baris("Tutup: " + d.TutupAt.In(time.Local).Format("02/01/06 15.04") + " " + d.TutupOleh)
	} else {
		w.baris("Tutup: belum ditutup")
	}
	w.garis()
	w.kolom("Bill lunas", strconv.Itoa(d.Bill))
	if d.BillBatal > 0 {
		w.kolom("Bill batal", strconv.Itoa(d.BillBatal))
	}
	for _, m := range d.PerMetode {
		w.kolom("  "+m.Metode, rupiah(m.Total))
	}
	w.kolom("Penjualan", rupiah(d.Penjualan))
	if d.Diskon > 0 {
		w.kolom("(sudah dipotong diskon)", rupiah(d.Diskon))
	}
	w.garis()
	w.kolom("Kas awal", rupiah(d.KasAwal))
	w.kolom("+ Penjualan tunai", rupiah(d.PenjualanTunai))
	if d.KasbonTunai > 0 {
		w.kolom("+ Cicilan kasbon tunai", rupiah(d.KasbonTunai))
	}
	w.kolom("- Pengeluaran", rupiah(d.Keluar))
	for _, e := range d.Pengeluaran {
		w.kolom("  "+e.Keterangan, rupiah(e.Jumlah))
	}
	w.garis()
	if d.KasSistem != nil {
		w.tebal(true)
		w.kolom("Kas sistem", rupiah(*d.KasSistem))
		w.tebal(false)
	}
	if d.KasHitung != nil {
		w.kolom("Kas dihitung", rupiah(*d.KasHitung))
	}
	if d.Selisih != nil {
		w.tebal(true)
		w.kolom("SELISIH", rupiah(*d.Selisih))
		w.tebal(false)
	}
	w.garis()
	w.baris("Dicetak " + time.Now().Format("02/01/06 15.04"))
	return w.akhir()
}
