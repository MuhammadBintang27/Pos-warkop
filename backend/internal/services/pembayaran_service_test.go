package services

import (
	"testing"

	"warkop/internal/dto"
	"warkop/internal/types"
)

func TestNormalizePembayaran(t *testing.T) {
	pelanggan := int64(1)
	tests := []struct {
		name      string
		total     int
		in        []dto.PembayaranInput
		wantErr   bool
		kembalian int
	}{
		{"tunai pas", 10000, []dto.PembayaranInput{{Metode: types.MetodeTunai, Jumlah: 10000}}, false, 0},
		{"tunai dengan kembalian", 10000, []dto.PembayaranInput{{Metode: types.MetodeTunai, Jumlah: 10000, Diterima: 20000}}, false, 10000},
		{"uang kurang", 10000, []dto.PembayaranInput{{Metode: types.MetodeTunai, Jumlah: 10000, Diterima: 5000}}, true, 0},
		{"split tunai + qris", 10000, []dto.PembayaranInput{
			{Metode: types.MetodeTunai, Jumlah: 4000, Diterima: 5000},
			{Metode: types.MetodeQris, Jumlah: 6000},
		}, false, 1000},
		{"total tidak sama", 10000, []dto.PembayaranInput{{Metode: types.MetodeQris, Jumlah: 9000}}, true, 0},
		{"kasbon tanpa pelanggan", 10000, []dto.PembayaranInput{{Metode: types.MetodeKasbon, Jumlah: 10000}}, true, 0},
		{"kasbon dengan pelanggan", 10000, []dto.PembayaranInput{{Metode: types.MetodeKasbon, Jumlah: 10000, PelangganID: &pelanggan}}, false, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rows, err := normalizePembayaran(tc.total, tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if err == nil && rows[0].Kembalian != tc.kembalian {
				t.Errorf("kembalian = %d, want %d", rows[0].Kembalian, tc.kembalian)
			}
		})
	}
}
