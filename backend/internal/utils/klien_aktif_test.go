package utils

import (
	"testing"
	"time"
)

func TestTokenBerlaku(t *testing.T) {
	sejak := time.Date(2026, 11, 19, 10, 0, 0, 0, time.UTC)
	if !TokenBerlaku(99, sejak.Add(-time.Hour)) {
		t.Fatal("user tanpa pencabutan harus berlaku")
	}
	CabutToken(99, sejak)
	if TokenBerlaku(99, sejak.Add(-time.Second)) {
		t.Fatal("token sebelum pencabutan harus ditolak")
	}
	if !TokenBerlaku(99, sejak) {
		t.Fatal("login ulang di detik yang sama harus berlaku")
	}
}
