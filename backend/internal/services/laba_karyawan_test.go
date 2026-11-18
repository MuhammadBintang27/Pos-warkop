package services

import (
	"math"
	"testing"

	"warkop/internal/dto"
)

func TestRataBaru(t *testing.T) {
	tests := []struct {
		name          string
		stok          int
		rata          float64
		jumlah, total int
		want          float64
	}{
		{"stok kosong: harga baru langsung jadi rata-rata", 0, 0, 1000, 120000, 120},
		{"campur dengan stok lama", 1000, 100, 1000, 140000, 120},
		{"restock tanpa harga tidak mengubah rata-rata", 500, 100, 1000, 0, 100},
		{"stok minus tidak ikut menimbang", -200, 100, 1000, 50000, 50},
	}
	for _, tc := range tests {
		if got := rataBaru(tc.stok, tc.rata, tc.jumlah, tc.total); math.Abs(got-tc.want) > 0.0001 {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestHitungSlip(t *testing.T) {
	pokok, netto, err := hitungSlip("bulanan", 2000000, 0, 100000, 50000, 20000, 300000)
	if err != nil || pokok != 2000000 || netto != 1830000 {
		t.Errorf("bulanan: pokok=%d netto=%d err=%v", pokok, netto, err)
	}
	pokok, netto, err = hitungSlip("harian", 80000, 24, 0, 0, 0, 0)
	if err != nil || pokok != 1920000 || netto != 1920000 {
		t.Errorf("harian: pokok=%d netto=%d err=%v", pokok, netto, err)
	}
	if _, _, err = hitungSlip("harian", 80000, 0, 0, 0, 0, 0); err == nil {
		t.Error("karyawan harian tanpa hari kerja harus ditolak")
	}
	if _, _, err = hitungSlip("bulanan", 1000000, 0, 0, 0, 900000, 200000); err == nil {
		t.Error("netto negatif harus ditolak")
	}
}

func TestHitungLaba(t *testing.T) {
	r := &dto.LabaRugiResponse{PenjualanSendiri: 1000000, KomisiTenant: 100000, HPP: 400000, TotalPengeluaran: 150000, Gaji: 300000}
	hitungLaba(r)
	if r.Pendapatan != 1100000 || r.LabaKotor != 700000 || r.LabaBersih != 250000 {
		t.Errorf("laba salah: %+v", r)
	}
}
