package services

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"warkop/internal/dto"
)

func tiketUji() dto.TiketPayload {
	return dto.TiketPayload{
		Judul: "PESANAN BARU", MejaNama: "Meja 3", AtasNama: "Budi", Oleh: "Sari", Tenant: "Tenant A",
		Waktu: time.Date(2026, 11, 5, 10, 0, 0, 0, time.Local),
		Items: []dto.TiketItem{{Qty: 2, Nama: "Kopi Susu", Opsi: "Es", Catatan: "less sugar"}},
	}
}

func TestEscposDiawaliInisialisasiDanDiakhiriPotong(t *testing.T) {
	b := escposTiket(tiketUji(), 58)
	if !bytes.HasPrefix(b, []byte{0x1b, 0x40}) {
		t.Error("tiket harus diawali inisialisasi printer ESC @")
	}
	if !bytes.HasSuffix(b, []byte{0x1d, 0x56, 0x42, 0x00}) {
		t.Error("tiket harus diakhiri perintah potong")
	}
}

func TestEscposMemuatIsiTiket(t *testing.T) {
	s := string(escposTiket(tiketUji(), 58))
	for _, want := range []string{"PESANAN BARU", "Meja 3", "Budi", "Tenant A", "Dikirim oleh: Sari", "2x Kopi Susu", "Es", "* less sugar"} {
		if !strings.Contains(s, want) {
			t.Errorf("tidak memuat %q", want)
		}
	}
}

func TestEscposTiketBatal(t *testing.T) {
	p := tiketUji()
	p.Judul = "BATAL"
	s := string(escposTiket(p, 80))
	if !strings.Contains(s, "JANGAN DIBUAT / HENTIKAN") || !strings.Contains(s, "Dibatalkan oleh: Sari") {
		t.Error("tiket BATAL harus memuat peringatan dan nama pembatal")
	}
}

func TestAsciiMembuangAksenDanSimbol(t *testing.T) {
	if got := ascii("Café ☕ Ñandú"); got != "Cafe ? Nandu" {
		t.Errorf("got %q", got)
	}
}

func TestBungkusSesuaiLebarKertas(t *testing.T) {
	baris := bungkus("Nasi Goreng Spesial Seafood Pedas Level Lima Tanpa Sayur", 32, "")
	if len(baris) < 2 {
		t.Fatal("teks panjang harus dipecah")
	}
	for _, l := range baris {
		if len(l) > 32 {
			t.Errorf("baris melebihi lebar: %q", l)
		}
	}
}
