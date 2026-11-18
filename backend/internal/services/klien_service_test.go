package services

import (
	"testing"
	"time"
)

func TestStatusLangganan(t *testing.T) {
	now := time.Date(2026, 11, 5, 15, 0, 0, 0, time.Local)
	tgl := func(y int, m time.Month, d int) *time.Time {
		v := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
		return &v
	}
	tests := []struct {
		name   string
		sampai *time.Time
		want   string
	}{
		{"belum ada langganan", nil, langgananBelum},
		{"berakhir kemarin", tgl(2026, 11, 4), langgananBerakhir},
		{"berakhir hari ini masih berlaku", tgl(2026, 11, 5), langgananSegera},
		{"7 hari lagi", tgl(2026, 11, 12), langgananSegera},
		{"8 hari lagi", tgl(2026, 11, 13), langgananAktif},
		{"setahun lagi", tgl(2027, 11, 5), langgananAktif},
	}
	for _, tc := range tests {
		if got := statusLangganan(tc.sampai, now); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}
