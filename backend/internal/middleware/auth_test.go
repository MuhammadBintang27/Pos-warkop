package middleware

import "testing"

func TestBolehTenant(t *testing.T) {
	for path, want := range map[string]bool{
		"/api/tenant-saya/pesanan":        true,
		"/api/tenant-saya/item/5/siap":    true,
		"/api/tenant-saya/laporan/harian": true,
		"/api/events":                     true,
		"/api/pesanan":                    false,
		"/api/pesanan/1/kirim":            false,
		"/api/riwayat":                    false,
		"/api/tenant":                     false,
		"/api/tenant-sayaX":               false,
		"/api/laporan/tenant/harian":      false,
		"/api/pesanan/events":             false,
	} {
		if got := bolehTenant(path); got != want {
			t.Errorf("bolehTenant(%q) = %v, want %v", path, got, want)
		}
	}
}
