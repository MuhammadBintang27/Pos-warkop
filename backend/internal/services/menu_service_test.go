package services

import (
	"testing"

	"warkop/internal/dto"
)

func TestCekGantiHarga(t *testing.T) {
	tests := []struct {
		name    string
		varian  []dto.GrupRequest
		wantErr bool
	}{
		{"tanpa varian", nil, false},
		{"satu grup penentu harga", []dto.GrupRequest{{GantiHarga: true}, {}}, false},
		{"dua grup penentu harga ditolak", []dto.GrupRequest{{GantiHarga: true}, {GantiHarga: true}}, true},
		{"penentu harga tidak boleh multi pilihan", []dto.GrupRequest{{GantiHarga: true, Multi: true}}, true},
		{"grup multi biasa boleh", []dto.GrupRequest{{Multi: true}, {Multi: true}}, false},
	}
	for _, tc := range tests {
		if err := cekGantiHarga(tc.varian); (err != nil) != tc.wantErr {
			t.Errorf("%s: err = %v, wantErr %v", tc.name, err, tc.wantErr)
		}
	}
}
