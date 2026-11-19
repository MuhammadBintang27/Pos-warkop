package services

import (
	"testing"
	"time"
)

func TestUmurKasbonFIFO(t *testing.T) {
	d := func(day int) time.Time { return time.Date(2026, 11, day, 10, 0, 0, 0, time.UTC) }
	entri := []kasbonEntri{{10000, d(1)}, {20000, d(5)}, {5000, d(9)}}

	if sisa, tertua := umurKasbon(entri, 0); sisa != 35000 || !tertua.Equal(d(1)) {
		t.Fatalf("belum bayar: sisa %d tertua %v", sisa, tertua)
	}
	// cicilan 15000 menutup kasbon tgl 1 dan 5000 dari tgl 5
	if sisa, tertua := umurKasbon(entri, 15000); sisa != 20000 || !tertua.Equal(d(5)) {
		t.Fatalf("cicil sebagian: sisa %d tertua %v", sisa, tertua)
	}
	if sisa, tertua := umurKasbon(entri, 35000); sisa != 0 || tertua != nil {
		t.Fatalf("lunas: sisa %d tertua %v", sisa, tertua)
	}
}
