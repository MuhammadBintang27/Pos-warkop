package services

import "testing"

func TestStatusPenutupan(t *testing.T) {
	tests := []struct {
		name                  string
		ada, disetor          bool
		snapKotor, snapDiskon int
		kotor, diskon         int
		want                  string
	}{
		{"belum di-closing", false, false, 0, 0, 100000, 0, statusBelum},
		{"sudah closing, angka sama", true, false, 100000, 5000, 100000, 5000, statusClosing},
		{"sudah disetor, angka sama", true, true, 100000, 5000, 100000, 5000, statusDisetor},
		{"ada transaksi baru setelah closing", true, false, 100000, 0, 120000, 0, statusBerubah},
		{"diskon berubah setelah closing", true, false, 100000, 0, 100000, 5000, statusBerubah},
		{"berubah setelah disetor tetap ditandai", true, true, 100000, 0, 120000, 0, statusBerubah},
	}
	for _, tc := range tests {
		got := statusPenutupan(tc.ada, tc.disetor, tc.snapKotor, tc.snapDiskon, tc.kotor, tc.diskon)
		if got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}
