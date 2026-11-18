package services

import "testing"

func TestHargaAkhir(t *testing.T) {
	tests := []struct {
		name    string
		dasar   int
		pilihan []pilihanOpsi
		want    int
	}{
		{"tanpa varian memakai harga dasar", 5000, nil, 5000},
		{"opsi tambahan menambah harga dasar", 5000, []pilihanOpsi{{1000, false}, {2000, false}}, 8000},
		{"grup penentu harga: teh panas", 5000, []pilihanOpsi{{5000, true}}, 5000},
		{"grup penentu harga: teh es", 5000, []pilihanOpsi{{10000, true}}, 10000},
		{"penentu harga digabung tambahan", 5000, []pilihanOpsi{{10000, true}, {3000, false}}, 13000},
		{"urutan pilihan tidak mempengaruhi hasil", 5000, []pilihanOpsi{{3000, false}, {10000, true}}, 13000},
	}
	for _, tc := range tests {
		if got := hargaAkhir(tc.dasar, tc.pilihan); got != tc.want {
			t.Errorf("%s: got %d, want %d", tc.name, got, tc.want)
		}
	}
}
