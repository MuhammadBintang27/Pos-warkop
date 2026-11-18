package services

import (
	"testing"

	"warkop/internal/dto"
)

func TestKomisiEfektif(t *testing.T) {
	if got := komisiEfektif(dto.TenantRequest{KomisiPersen: 10}); got != 10 {
		t.Errorf("tenant biasa harus memakai komisi yang diisi, got %v", got)
	}
	if got := komisiEfektif(dto.TenantRequest{KomisiPersen: 10, MilikSendiri: true}); got != 100 {
		t.Errorf("tenant milik sendiri harus 100%% walau komisi diisi lain, got %v", got)
	}
}
