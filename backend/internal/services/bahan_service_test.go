package services

import "testing"

func TestStatusBahan(t *testing.T) {
	tests := []struct {
		name          string
		stok, minimum int
		want          string
	}{
		{"cukup", 500, 100, statusOK},
		{"tepat di batas minimum", 100, 100, statusTipis},
		{"di bawah minimum", 50, 100, statusTipis},
		{"tanpa minimum tidak menipis", 5, 0, statusOK},
		{"habis", 0, 100, statusHabis},
		{"habis tanpa minimum", 0, 0, statusHabis},
	}
	for _, tc := range tests {
		if got := statusBahan(tc.stok, tc.minimum); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}
