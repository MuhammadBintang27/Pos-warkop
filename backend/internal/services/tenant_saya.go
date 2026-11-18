package services

import (
	"time"

	"warkop/internal/dto"
	"warkop/internal/types"
)

// Antrean yang lebih lama dari ini dianggap terlupa dan tidak ditampilkan lagi.
const batasAntreanTenant = 24 * time.Hour

// PesananMasuk: antrean (belum selesai) lalu item yang selesai pada tanggal yang diminta (YYYY-MM-DD, kosong = hari ini).
// shortcut: semua item selesai satu hari dikirim sekaligus dan dipaginasi di layar; pindah ke paginasi server bila satu tenant ribuan item per hari.
func (s *TenantService) PesananMasuk(kid, tenantID int64, tanggal string) (*dto.PesananMasukResponse, error) {
	now := time.Now()
	dari, sampai, err := parseRange(tanggal, tanggal, now)
	if err != nil {
		return nil, err
	}
	out := &dto.PesananMasukResponse{Items: []dto.ItemMasuk{}}
	if s.db.QueryRow("SELECT nama FROM tenant WHERE id=? AND klien_id=?", tenantID, kid).Scan(&out.Tenant) != nil {
		return nil, types.NotFound("Tenant")
	}
	// item lama (sebelum kolom dikirim_at ada) memakai waktu bill dibuat
	rows, err := s.db.Query(`SELECT i.id,i.pesanan_id,COALESCE(m.nama,''),p.atas_nama,i.nama,i.qty,i.opsi,i.catatan,
			COALESCE(i.dikirim_at,p.dibuat_at),i.dikirim_oleh,i.siap_at,
			EXISTS(SELECT 1 FROM pesanan_item j WHERE j.pesanan_id=i.pesanan_id AND j.tenant_id=i.tenant_id AND j.dikirim=1 AND j.dikirim_at<i.dikirim_at)
		FROM pesanan_item i JOIN pesanan p ON p.id=i.pesanan_id LEFT JOIN meja m ON m.id=p.meja_id
		WHERE i.klien_id=? AND i.tenant_id=? AND i.dikirim=1 AND p.status<>'batal'
			AND ((i.siap_at IS NULL AND COALESCE(i.dikirim_at,p.dibuat_at)>=?) OR (i.siap_at>=? AND i.siap_at<?))
		ORDER BY i.siap_at IS NOT NULL, i.siap_at DESC, COALESCE(i.dikirim_at,p.dibuat_at), i.id`,
		kid, tenantID, now.Add(-batasAntreanTenant), dari, sampai)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var it dto.ItemMasuk
		if err := rows.Scan(&it.ID, &it.PesananID, &it.MejaNama, &it.AtasNama, &it.Nama, &it.Qty, &it.Opsi, &it.Catatan, &it.DikirimAt, &it.Oleh, &it.SiapAt, &it.Tambahan); err != nil {
			return nil, err
		}
		out.Items = append(out.Items, it)
	}
	return out, rows.Err()
}

// SetSiap hanya menyentuh item milik tenant itu sendiri.
func (s *TenantService) SetSiap(kid, tenantID, itemID int64, siap bool) error {
	var at *time.Time
	if siap {
		now := time.Now()
		at = &now
	}
	res, err := s.db.Exec("UPDATE pesanan_item SET siap_at=? WHERE id=? AND klien_id=? AND tenant_id=? AND dikirim=1", at, itemID, kid, tenantID)
	if err != nil {
		return err
	}
	// 0 baris juga terjadi bila nilainya tidak berubah, jadi pastikan itemnya memang tidak ada.
	if n, _ := res.RowsAffected(); n == 0 {
		var ok int
		if s.db.QueryRow("SELECT 1 FROM pesanan_item WHERE id=? AND klien_id=? AND tenant_id=? AND dikirim=1", itemID, kid, tenantID).Scan(&ok) != nil {
			return types.NotFound("Item")
		}
	}
	s.hub.Publish(kid, "dapur")
	return nil
}
