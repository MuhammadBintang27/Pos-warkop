package services

import (
	"testing"
	"time"
)

func TestHitungSetoran(t *testing.T) {
	tests := []struct {
		name                      string
		kotor, diskon, komisiMenu int
		persen                    float64
		perMenu                   bool
		komisi, setor             int
	}{
		{"komisi 10% setelah diskon", 100000, 10000, 0, 10, false, 9000, 81000},
		{"tanpa komisi", 50000, 0, 0, 0, false, 0, 50000},
		{"pembulatan setengah ke atas", 1005, 0, 0, 10, false, 101, 904},
		{"tanpa penjualan", 0, 0, 0, 15, false, 0, 0},
		{"persen mengabaikan tarif menu", 100000, 0, 30000, 10, false, 10000, 90000},
		{"per menu: diskon ditanggung tenant", 100000, 10000, 30000, 10, true, 30000, 60000},
		{"per menu: dibatasi penjualan setelah diskon", 20000, 15000, 8000, 0, true, 5000, 0},
		{"per menu: diskon penuh", 10000, 10000, 2000, 0, true, 0, 0},
	}
	for _, tc := range tests {
		komisi, setor := hitungSetoran(tc.kotor, tc.diskon, tc.komisiMenu, tc.persen, tc.perMenu)
		if komisi != tc.komisi || setor != tc.setor {
			t.Errorf("%s: komisi=%d setor=%d, want %d dan %d", tc.name, komisi, setor, tc.komisi, tc.setor)
		}
	}
}

func TestParseRange(t *testing.T) {
	wib := time.FixedZone("WIB", 7*3600)
	now := time.Date(2026, 11, 5, 15, 0, 0, 0, wib)

	from, to, err := parseRange("", "", now)
	if err != nil || from.Day() != 5 || to.Day() != 6 {
		t.Fatalf("default hari ini salah: %v %v %v", from, to, err)
	}
	from, to, err = parseRange("2026-10-31", "2026-11-02", now)
	if err != nil || !from.Equal(time.Date(2026, 10, 31, 0, 0, 0, 0, wib)) || !to.Equal(time.Date(2026, 11, 3, 0, 0, 0, 0, wib)) {
		t.Fatalf("rentang lintas bulan salah: %v %v %v", from, to, err)
	}
	for _, bad := range [][2]string{{"2026-11-05", "2026-11-04"}, {"abc", ""}, {"2024-01-01", "2026-11-05"}} {
		if _, _, err := parseRange(bad[0], bad[1], now); err == nil {
			t.Errorf("seharusnya error untuk %v", bad)
		}
	}
}
