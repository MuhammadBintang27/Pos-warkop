package services

import (
	"database/sql"

	"warkop/internal/dto"
	"warkop/internal/types"
)

type TenantService struct {
	db  *sql.DB
	hub *EventHub
}

func NewTenantService(db *sql.DB, hub *EventHub) *TenantService {
	return &TenantService{db: db, hub: hub}
}

func (s *TenantService) List(kid int64) ([]dto.TenantResponse, error) {
	rows, err := s.db.Query("SELECT id,nama,komisi_persen,buka,milik_sendiri,printer_id FROM tenant WHERE klien_id=? AND aktif=1 ORDER BY id", kid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []dto.TenantResponse{}
	for rows.Next() {
		var t dto.TenantResponse
		if err := rows.Scan(&t.ID, &t.Nama, &t.KomisiPersen, &t.Buka, &t.MilikSendiri, &t.PrinterID); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// komisiEfektif: tenant milik sendiri selalu 100% untuk pemilik tempat, jadi tidak ada yang disetor.
func komisiEfektif(req dto.TenantRequest) float64 {
	if req.MilikSendiri {
		return 100
	}
	return req.KomisiPersen
}

func (s *TenantService) cekPrinter(kid int64, id *int64) error {
	if id == nil {
		return nil
	}
	var ada int
	if s.db.QueryRow("SELECT 1 FROM printer WHERE id=? AND klien_id=? AND aktif=1", *id, kid).Scan(&ada) != nil {
		return types.BadRequest("Printer tidak ditemukan")
	}
	return nil
}

func (s *TenantService) Create(kid int64, req dto.TenantRequest) error {
	if err := s.cekPrinter(kid, req.PrinterID); err != nil {
		return err
	}
	if _, err := s.db.Exec("INSERT INTO tenant (klien_id,nama,komisi_persen,milik_sendiri,printer_id) VALUES (?,?,?,?,?)",
		kid, req.Nama, komisiEfektif(req), req.MilikSendiri, req.PrinterID); err != nil {
		return err
	}
	s.hub.Publish(kid, "menu")
	return nil
}

func (s *TenantService) Update(kid, id int64, req dto.TenantRequest) error {
	if err := s.cekPrinter(kid, req.PrinterID); err != nil {
		return err
	}
	if _, err := s.db.Exec("UPDATE tenant SET nama=?, komisi_persen=?, milik_sendiri=?, printer_id=? WHERE id=? AND klien_id=? AND aktif=1",
		req.Nama, komisiEfektif(req), req.MilikSendiri, req.PrinterID, id, kid); err != nil {
		return err
	}
	var ada int
	if s.db.QueryRow("SELECT 1 FROM tenant WHERE id=? AND klien_id=? AND aktif=1", id, kid).Scan(&ada) != nil {
		return types.NotFound("Tenant")
	}
	s.hub.Publish(kid, "menu")
	return nil
}

// SetBuka membuka/menutup tenant tanpa mengubah status tersedia tiap menunya.
func (s *TenantService) SetBuka(kid, id int64, buka bool) error {
	if _, err := s.db.Exec("UPDATE tenant SET buka=? WHERE id=? AND klien_id=? AND aktif=1", buka, id, kid); err != nil {
		return err
	}
	var ada int
	if s.db.QueryRow("SELECT 1 FROM tenant WHERE id=? AND klien_id=? AND aktif=1", id, kid).Scan(&ada) != nil {
		return types.NotFound("Tenant")
	}
	s.hub.Publish(kid, "menu")
	return nil
}

func (s *TenantService) Delete(kid, id int64) error {
	var aktif, menu int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM tenant WHERE klien_id=? AND aktif=1", kid).Scan(&aktif); err != nil {
		return err
	}
	if aktif <= 1 {
		return types.BadRequest("Minimal harus ada satu tenant")
	}
	if err := s.db.QueryRow("SELECT COUNT(*) FROM menu WHERE klien_id=? AND tenant_id=? AND aktif=1", kid, id).Scan(&menu); err != nil {
		return err
	}
	if menu > 0 {
		return types.BadRequest("Tenant masih punya menu aktif. Pindahkan atau hapus menunya dulu")
	}
	res, err := s.db.Exec("UPDATE tenant SET aktif=0 WHERE id=? AND klien_id=? AND aktif=1", id, kid)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return types.NotFound("Tenant")
	}
	s.hub.Publish(kid, "menu")
	return nil
}
