package services

import (
	"database/sql"
	"strings"
	"time"

	"warkop/internal/dto"
	"warkop/internal/types"
	"warkop/internal/utils"
)

type KasbonService struct {
	db *sql.DB
}

const sisaUtangSQL = `COALESCE((SELECT SUM(jumlah) FROM pembayaran WHERE pelanggan_id=%s AND metode='kasbon'),0)
	- COALESCE((SELECT SUM(jumlah) FROM kasbon_bayar WHERE pelanggan_id=%s),0)` // pelanggan sudah milik satu klien, jadi jumlahnya ikut terpisah

func NewKasbonService(db *sql.DB) *KasbonService {
	return &KasbonService{db: db}
}

func (s *KasbonService) List(kid int64) ([]dto.PelangganResponse, error) {
	rows, err := s.db.Query("SELECT p.id,p.nama,p.telp,"+strings.ReplaceAll(sisaUtangSQL, "%s", "p.id")+" FROM pelanggan p WHERE p.klien_id=? ORDER BY p.nama", kid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []dto.PelangganResponse{}
	for rows.Next() {
		var p dto.PelangganResponse
		if err := rows.Scan(&p.ID, &p.Nama, &p.Telp, &p.SisaUtang); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *KasbonService) Create(kid int64, req dto.PelangganRequest) error {
	_, err := s.db.Exec("INSERT INTO pelanggan (klien_id,nama,telp) VALUES (?,?,?)", kid, strings.TrimSpace(req.Nama), strings.TrimSpace(req.Telp))
	return err
}

func (s *KasbonService) Lunasi(kid, pelangganID int64, req dto.LunasiRequest) error {
	return utils.WithTx(s.db, func(tx *sql.Tx) error {
		shiftID, err := activeShiftID(tx, kid)
		if err != nil {
			return err
		}
		// kunci pelanggan agar dua pelunasan bersamaan tidak melebihi sisa utang
		var exists int
		if tx.QueryRow("SELECT 1 FROM pelanggan WHERE id=? AND klien_id=? FOR UPDATE", pelangganID, kid).Scan(&exists) != nil {
			return types.NotFound("Pelanggan")
		}
		var sisa int
		if err := tx.QueryRow("SELECT "+strings.ReplaceAll(sisaUtangSQL, "%s", "?"), pelangganID, pelangganID).Scan(&sisa); err != nil {
			return err
		}
		if req.Jumlah > sisa {
			return types.BadRequest("Jumlah melebihi sisa utang")
		}
		_, err = tx.Exec("INSERT INTO kasbon_bayar (klien_id,pelanggan_id,jumlah,metode,shift_id,at) VALUES (?,?,?,?,?,?)",
			kid, pelangganID, req.Jumlah, req.Metode, shiftID, time.Now())
		return err
	})
}
