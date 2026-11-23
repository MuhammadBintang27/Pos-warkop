package services

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"warkop/internal/dto"
)

func TestRupiah(t *testing.T) {
	tests := map[int]string{0: "Rp 0", 500: "Rp 500", 12500: "Rp 12.500", 1250000: "Rp 1.250.000", -5000: "-Rp 5.000"}
	for in, want := range tests {
		if got := rupiah(in); got != want {
			t.Errorf("rupiah(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestKolomRataKanan(t *testing.T) {
	w := barisPenulis(58)
	w.kolom("Kopi", "Rp 5.000")
	baris := strings.Split(strings.TrimRight(string(w.b.Bytes()[2:]), "\n"), "\n")
	if len(baris) != 1 || len(baris[0]) != 32 || !strings.HasPrefix(baris[0], "Kopi") || !strings.HasSuffix(baris[0], "Rp 5.000") {
		t.Errorf("kolom tidak rata kanan selebar kertas: %q", baris)
	}
}

func TestKolomNamaPanjangDibungkus(t *testing.T) {
	w := barisPenulis(58)
	w.kolom("Nasi Goreng Spesial Pedas Level Lima", "Rp 25.000")
	for _, l := range strings.Split(strings.TrimRight(string(w.b.Bytes()[2:]), "\n"), "\n") {
		if len(l) > 32 {
			t.Errorf("baris melebihi lebar kertas: %q", l)
		}
	}
}

func strukUji() dto.StrukPayload {
	return dto.StrukPayload{
		Judul: "STRUK", Usaha: "Warkop Maju", Alamat: "Jl. Mawar 1", Footer: "Terima kasih",
		ID: 12, MejaNama: "Meja 3", Kasir: "Sari", Waktu: time.Date(2026, 11, 5, 10, 0, 0, 0, time.Local),
		Items:    []dto.StrukItem{{Qty: 2, Nama: "Kopi Susu", Opsi: "Es", Harga: 12500}},
		Subtotal: 25000, Diskon: 5000, Total: 20000,
		Bayar: []dto.StrukBayar{{Metode: "tunai", Diterima: 50000, Kembalian: 30000}},
	}
}

func TestEscposStruk(t *testing.T) {
	b := escposStruk(strukUji(), 58)
	if !bytes.HasPrefix(b, []byte{0x1b, 0x40}) || !bytes.HasSuffix(b, []byte{0x1d, 0x56, 0x42, 0x00}) {
		t.Error("struk harus diawali inisialisasi dan diakhiri potong")
	}
	for _, want := range []string{"Warkop Maju", "#12", "2x Kopi Susu", "Rp 25.000", "-Rp 5.000", "TOTAL", "Rp 20.000", "Tunai", "Kembali", "Rp 30.000", "Terima kasih"} {
		if !bytes.Contains(b, []byte(want)) {
			t.Errorf("struk tidak memuat %q", want)
		}
	}
}

func TestEscposClosingTenantDanMilikSendiri(t *testing.T) {
	p := dto.ClosingPayload{
		Usaha: "Warkop Maju", Tenant: "Tenant A", Tanggal: "2026-11-05", Transaksi: 3, Kotor: 100000, Diskon: 5000,
		KomisiPersen: 10, Komisi: 9500, Setor: 85500, Menu: []dto.ClosingMenu{{Qty: 4, Nama: "Ayam", Nilai: 100000}},
	}
	b := escposClosing(p, 58)
	for _, want := range []string{"BILL CLOSING TENANT", "05/11/2026", "4x Ayam", "Komisi 10%", "SETOR TENANT", "Rp 85.500", "(_________)"} {
		if !bytes.Contains(b, []byte(want)) {
			t.Errorf("closing tenant tidak memuat %q", want)
		}
	}
	p.KomisiPerMenu = true
	if b = escposClosing(p, 58); !bytes.Contains(b, []byte("Komisi per menu")) {
		t.Error("closing tenant komisi per menu harus berlabel Komisi per menu")
	}
	p.MilikSendiri = true
	b = escposClosing(p, 58)
	if !bytes.Contains(b, []byte("PENDAPATAN PEMILIK")) || bytes.Contains(b, []byte("SETOR TENANT")) || bytes.Contains(b, []byte("(_________)")) {
		t.Error("closing milik sendiri tidak boleh punya setoran dan tanda tangan")
	}
}

func TestGabungStruk(t *testing.T) {
	items := []dto.ItemResponse{
		{MenuID: 1, Nama: "Kopi", Harga: 5000, Qty: 1, Opsi: "Es"},
		{MenuID: 1, Nama: "Kopi", Harga: 5000, Qty: 2, Opsi: "Es"},
		{MenuID: 1, Nama: "Kopi", Harga: 5000, Qty: 1, Opsi: "Panas"},
		{MenuID: 1, Nama: "Kopi", Harga: 5000, Qty: 4, Opsi: "Es", Catatan: "kurang gula"}, // catatan tidak tercetak
	}
	got := gabungStruk(items)
	if len(got) != 2 || got[0].Qty != 7 || got[1].Qty != 1 {
		t.Errorf("gabungStruk = %+v", got)
	}
}

func TestEscposRekapTenantMenjumlahkanHari(t *testing.T) {
	p := dto.RekapTenantPayload{
		Usaha: "Warkop Maju", Tenant: "Gorengan", Dari: "2026-11-12", Sampai: "2026-11-13", KomisiPersen: 20,
		Hari: []dto.RekapHari{
			{Tanggal: "2026-11-12", Transaksi: 3, Kotor: 50000, Diskon: 0, Komisi: 10000, Setor: 40000, Status: "closing"},
			{Tanggal: "2026-11-13", Transaksi: 2, Kotor: 30000, Diskon: 5000, Komisi: 5000, Setor: 20000, Status: "belum"},
		},
	}
	b := escposRekapTenant(p, 58)
	for _, want := range []string{"REKAP SETORAN TENANT", "12/11/2026 - 13/11/2026", "12/11 3 bill", "13/11 2 bill*", "Rp 80.000", "Rp 60.000", "1 hari belum", "(_________)"} {
		if !bytes.Contains(b, []byte(want)) {
			t.Errorf("rekap tidak memuat %q", want)
		}
	}
	p.MilikSendiri = true
	if b = escposRekapTenant(p, 58); bytes.Contains(b, []byte("SETOR TENANT")) || !bytes.Contains(b, []byte("REKAP PENDAPATAN")) {
		t.Error("rekap milik sendiri tidak boleh punya setoran")
	}
}

func TestEscposShiftMemuatPerhitunganKas(t *testing.T) {
	sistem, hitung, selisih := 330000, 325000, -5000
	tutup := time.Date(2026, 11, 18, 23, 0, 0, 0, time.Local)
	d := &dto.ShiftDetail{
		ShiftResponse: dto.ShiftResponse{ID: 7, Kasir: "Rina", TutupOleh: "Rina", BukaAt: tutup.Add(-16 * time.Hour), TutupAt: &tutup,
			KasAwal: 200000, KasSistem: &sistem, KasHitung: &hitung, Selisih: &selisih},
		Bill: 30, Penjualan: 400000, PenjualanTunai: 150000, Keluar: 20000,
		PerMetode:   []dto.MetodeTotal{{Metode: "tunai", Total: 150000}, {Metode: "qris", Total: 250000}},
		Pengeluaran: []dto.PengeluaranResponse{{Keterangan: "Galon", Jumlah: 20000}},
	}
	b := escposShift("Warkop Maju", d, 58)
	for _, want := range []string{"LAPORAN SHIFT #7", "18/11/26 23.00", "qris", "Kas awal", "Galon", "Rp 330.000", "SELISIH"} {
		if !bytes.Contains(b, []byte(want)) {
			t.Errorf("laporan shift tidak memuat %q", want)
		}
	}
}
