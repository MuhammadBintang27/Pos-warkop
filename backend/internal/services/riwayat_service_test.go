package services

import (
	"strings"
	"testing"
	"time"

	"warkop/internal/dto"
	"warkop/internal/types"
	"warkop/internal/utils"
)

func TestRiwayatFilterMenurutPeran(t *testing.T) {
	from, to := time.Now(), time.Now().Add(24*time.Hour)
	aktor := func(role types.Role) *utils.JWTPayload { return &utils.JWTPayload{UserID: 7, KlienID: 1, Role: role} }

	tests := []struct {
		name     string
		actor    *utils.JWTPayload
		q        dto.RiwayatQuery
		wantArgs int
		wantSQL  string
	}{
		{"owner melihat semua", aktor(types.RoleOwner), dto.RiwayatQuery{}, 3, ""},
		{"owner filter pembuat", aktor(types.RoleOwner), dto.RiwayatQuery{Oleh: 3}, 4, "p.dibuat_oleh=?"},
		{"kasir: dibuat atau dibayar dirinya", aktor(types.RoleKasir), dto.RiwayatQuery{}, 5, "b.diterima_oleh=?"},
		{"pelayan: hanya yang dibuatnya", aktor(types.RolePelayan), dto.RiwayatQuery{}, 4, "p.dibuat_oleh=?"},
		{"pelayan tidak bisa memilih pembuat lain", aktor(types.RolePelayan), dto.RiwayatQuery{Oleh: 3}, 4, "p.dibuat_oleh=?"},
		{"pencarian menambah 3 argumen", aktor(types.RoleOwner), dto.RiwayatQuery{Q: "budi"}, 6, "m.nama LIKE ?"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			where, args := riwayatFilter(tc.actor, tc.q, from, to)
			if len(args) != tc.wantArgs {
				t.Errorf("jumlah argumen = %d, want %d", len(args), tc.wantArgs)
			}
			if !strings.Contains(where, tc.wantSQL) {
				t.Errorf("SQL %q tidak memuat %q", where, tc.wantSQL)
			}
		})
	}

	where, args := riwayatFilter(aktor(types.RolePelayan), dto.RiwayatQuery{Oleh: 3}, from, to)
	if strings.Contains(where, "diterima_oleh") {
		t.Error("pelayan tidak boleh melihat pesanan yang hanya dibayar olehnya")
	}
	if args[len(args)-1] != int64(7) {
		t.Errorf("pelayan harus selalu dibatasi ke dirinya sendiri (id 7), dapat %v", args[len(args)-1])
	}
}
