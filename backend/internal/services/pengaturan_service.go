package services

import (
	"database/sql"
	"encoding/json"

	"warkop/internal/dto"
	"warkop/internal/types"
)

type PengaturanService struct {
	db  *sql.DB
	hub *EventHub
}

const kunciTampilan = "tampilan"

var defaultPengaturan = dto.PengaturanRequest{NamaUsaha: "Warkop", Theme: "aura", Primary: "amber", Mode: "light", DialogDapur: true}

func NewPengaturanService(db *sql.DB, hub *EventHub) *PengaturanService {
	return &PengaturanService{db: db, hub: hub}
}

func (s *PengaturanService) Get(kid int64) (*dto.PengaturanRequest, error) {
	var raw string
	out := defaultPengaturan
	err := s.db.QueryRow("SELECT nilai FROM pengaturan WHERE klien_id=? AND kunci=?", kid, kunciTampilan).Scan(&raw)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == nil {
		if err := json.Unmarshal([]byte(raw), &out); err != nil {
			return nil, err
		}
	}
	out.Kode = ""
	if kid > 0 && s.db.QueryRow("SELECT kode FROM klien WHERE id=?", kid).Scan(&out.Kode) != nil {
		out.Kode = ""
	}
	return &out, nil
}

func (s *PengaturanService) Update(kid int64, req dto.PengaturanRequest) error {
	if req.PrinterKasirID != nil {
		var ada int
		if s.db.QueryRow("SELECT 1 FROM printer WHERE id=? AND klien_id=? AND aktif=1", *req.PrinterKasirID, kid).Scan(&ada) != nil {
			return types.BadRequest("Printer kasir tidak ditemukan")
		}
	}
	req.Kode = ""
	raw, err := json.Marshal(req)
	if err != nil {
		return err
	}
	if _, err := s.db.Exec(`INSERT INTO pengaturan (klien_id,kunci,nilai) VALUES (?,?,?)
		ON DUPLICATE KEY UPDATE nilai=VALUES(nilai)`, kid, kunciTampilan, string(raw)); err != nil {
		return err
	}
	s.hub.Publish(kid, "pengaturan")
	return nil
}
