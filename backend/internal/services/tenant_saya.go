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
	out := &dto.PesananMasukResponse{Items: []dto.ItemMasuk{}, Batal: []dto.BatalMasuk{}}
	if s.db.QueryRow("SELECT nama FROM tenant WHERE id=? AND klien_id=?", tenantID, kid).Scan(&out.Tenant) != nil {
		return nil, types.NotFound("Tenant")
	}
	// item lama (sebelum kolom dikirim_at ada) memakai waktu bill dibuat
	rows, err := s.db.Query(`SELECT i.id,i.pesanan_id,COALESCE(m.nama,''),p.atas_nama,i.nama,i.qty,i.opsi,i.catatan,
			COALESCE(i.dikirim_at,p.dibuat_at),i.dikirim_oleh,i.diproses_at,i.siap_at,
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
		if err := rows.Scan(&it.ID, &it.PesananID, &it.MejaNama, &it.AtasNama, &it.Nama, &it.Qty, &it.Opsi, &it.Catatan, &it.DikirimAt, &it.Oleh, &it.DiprosesAt, &it.SiapAt, &it.Tambahan); err != nil {
			return nil, err
		}
		out.Items = append(out.Items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	rows, err = s.db.Query(`SELECT b.pesanan_id,COALESCE(m.nama,''),p.atas_nama,b.nama,b.opsi,b.qty,b.alasan,b.sudah_dibuat,u.nama,b.at
		FROM pembatalan b JOIN pesanan p ON p.id=b.pesanan_id LEFT JOIN meja m ON m.id=p.meja_id JOIN users u ON u.id=b.oleh
		WHERE b.klien_id=? AND b.tenant_id=? AND b.sudah_dikirim=1 AND b.at>=? AND b.at<? ORDER BY b.at DESC, b.id DESC`,
		kid, tenantID, dari, sampai)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var b dto.BatalMasuk
		if err := rows.Scan(&b.PesananID, &b.MejaNama, &b.AtasNama, &b.Nama, &b.Opsi, &b.Qty, &b.Alasan, &b.SudahDibuat, &b.Oleh, &b.At); err != nil {
			return nil, err
		}
		out.Batal = append(out.Batal, b)
	}
	return out, rows.Err()
}

// SetSiap hanya menyentuh item milik tenant itu sendiri. Selesai berarti sudah dibuat, jadi diproses_at ikut terisi.
func (s *TenantService) SetSiap(kid, tenantID, itemID int64, siap bool) error {
	at := waktuStatus(siap)
	return s.setStatus(kid, tenantID, itemID, "siap_at=?, diproses_at=COALESCE(diproses_at,?)", at, at)
}

// SetProses: tenant mulai membuat item; kasir melihatnya dan pembatalan menganggap makanannya sudah dibuat.
func (s *TenantService) SetProses(kid, tenantID, itemID int64, proses bool) error {
	return s.setStatus(kid, tenantID, itemID, "diproses_at=?", waktuStatus(proses))
}

func waktuStatus(on bool) *time.Time {
	if !on {
		return nil
	}
	now := time.Now()
	return &now
}

func (s *TenantService) setStatus(kid, tenantID, itemID int64, set string, nilai ...any) error {
	res, err := s.db.Exec("UPDATE pesanan_item SET "+set+" WHERE id=? AND klien_id=? AND tenant_id=? AND dikirim=1", append(nilai, itemID, kid, tenantID)...)
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
