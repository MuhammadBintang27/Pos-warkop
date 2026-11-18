package services

import (
	"database/sql"
	"fmt"
	"time"

	"warkop/internal/config"
	"warkop/internal/dto"
	"warkop/internal/types"
	"warkop/internal/utils"
)

type ShiftService struct {
	db  *sql.DB
	hub *EventHub
}

func NewShiftService(db *sql.DB, hub *EventHub) *ShiftService {
	return &ShiftService{db: db, hub: hub}
}

// activeShiftID: satu laci uang = satu shift terbuka, dipakai bersama semua kasir.
func activeShiftID(q querier, kid int64) (int64, error) {
	var id int64
	err := q.QueryRow("SELECT id FROM shift WHERE klien_id=? AND tutup_at IS NULL ORDER BY id DESC LIMIT 1", kid).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, types.BadRequest("Buka shift dulu sebelum menerima pembayaran")
	}
	return id, err
}

func loadShift(q querier, kid, id int64) (*dto.ShiftResponse, error) {
	var sh dto.ShiftResponse
	err := q.QueryRow(`SELECT s.id,u.nama,s.buka_at,s.tutup_at,s.kas_awal,s.kas_sistem,s.kas_hitung,COALESCE(ut.nama,'')
		FROM shift s JOIN users u ON u.id=s.kasir_id LEFT JOIN users ut ON ut.id=s.tutup_oleh WHERE s.id=? AND s.klien_id=?`, id, kid).
		Scan(&sh.ID, &sh.Kasir, &sh.BukaAt, &sh.TutupAt, &sh.KasAwal, &sh.KasSistem, &sh.KasHitung, &sh.TutupOleh)
	if err == sql.ErrNoRows {
		return nil, types.NotFound("Shift")
	}
	if sh.KasSistem != nil && sh.KasHitung != nil {
		d := *sh.KasHitung - *sh.KasSistem
		sh.Selisih = &d
	}
	return &sh, err
}

// Aktif mengembalikan nil bila belum ada shift yang dibuka.
func (s *ShiftService) Aktif(kid int64) (*dto.ShiftResponse, error) {
	id, err := activeShiftID(s.db, kid)
	if err != nil {
		if _, ok := utils.AsHTTPError(err); ok {
			return nil, nil
		}
		return nil, err
	}
	return loadShift(s.db, kid, id)
}

func (s *ShiftService) Buka(kid, userID int64, req dto.BukaShiftRequest) (*dto.ShiftResponse, error) {
	var id int64
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		// kunci satu baris tetap: dua orang yang menekan "Buka shift" bersamaan akan antre, bukan membuat dua shift
		if _, err := tx.Exec("SELECT nilai FROM pengaturan WHERE klien_id=? AND kunci=? FOR UPDATE", kid, config.KunciShift); err != nil {
			return err
		}
		aktif, err := activeShiftID(tx, kid)
		if err == nil {
			var oleh string
			var buka time.Time
			if err := tx.QueryRow("SELECT u.nama,s.buka_at FROM shift s JOIN users u ON u.id=s.kasir_id WHERE s.id=?", aktif).Scan(&oleh, &buka); err != nil {
				return err
			}
			return types.BadRequest(fmt.Sprintf("Shift sudah dibuka oleh %s pukul %s. Pakai shift itu, atau tutup dulu sebelum membuka yang baru", oleh, buka.Format("15:04")))
		}
		if _, ok := utils.AsHTTPError(err); !ok {
			return err
		}
		res, err := tx.Exec("INSERT INTO shift (klien_id,kasir_id,buka_at,kas_awal) VALUES (?,?,?,?)", kid, userID, time.Now(), req.KasAwal)
		if err != nil {
			return err
		}
		id, err = res.LastInsertId()
		return err
	})
	if err != nil {
		return nil, err
	}
	s.hub.Publish(kid, "shift")
	return loadShift(s.db, kid, id)
}

// Tutup: kas sistem = kas awal + tunai masuk + pelunasan kasbon tunai - pengeluaran.
func (s *ShiftService) Tutup(kid, userID int64, req dto.TutupShiftRequest) (*dto.ShiftResponse, error) {
	var id int64
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		var err error
		if id, err = activeShiftID(tx, kid); err != nil {
			return err
		}
		_, err = tx.Exec(`UPDATE shift s SET tutup_at=?, tutup_oleh=?, kas_hitung=?, kas_sistem = s.kas_awal
			+ COALESCE((SELECT SUM(jumlah) FROM pembayaran WHERE shift_id=s.id AND metode='tunai'),0)
			+ COALESCE((SELECT SUM(jumlah) FROM kasbon_bayar WHERE shift_id=s.id AND metode='tunai'),0)
			- COALESCE((SELECT SUM(jumlah) FROM pengeluaran WHERE shift_id=s.id),0)
			WHERE s.id=? AND s.klien_id=?`, time.Now(), userID, req.KasHitung, id, kid)
		return err
	})
	if err != nil {
		return nil, err
	}
	s.hub.Publish(kid, "shift")
	return loadShift(s.db, kid, id)
}

func (s *ShiftService) CreatePengeluaran(kid, userID int64, req dto.PengeluaranRequest) error {
	kategori := req.Kategori
	if kategori == "" {
		kategori = "lainnya"
	}
	shiftID, err := activeShiftID(s.db, kid)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT INTO pengeluaran (klien_id,shift_id,keterangan,jumlah,kategori,dicatat_oleh,at) VALUES (?,?,?,?,?,?,?)",
		kid, shiftID, req.Keterangan, req.Jumlah, kategori, userID, time.Now())
	return err
}

func (s *ShiftService) ListPengeluaran(kid int64) ([]dto.PengeluaranResponse, error) {
	out := []dto.PengeluaranResponse{}
	shiftID, err := activeShiftID(s.db, kid)
	if err != nil {
		if _, ok := utils.AsHTTPError(err); ok {
			return out, nil
		}
		return nil, err
	}
	rows, err := s.db.Query(`SELECT p.id,p.keterangan,p.jumlah,p.kategori,p.at,COALESCE(u.nama,'') FROM pengeluaran p
		LEFT JOIN users u ON u.id=p.dicatat_oleh WHERE p.klien_id=? AND p.shift_id=? ORDER BY p.id DESC`, kid, shiftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p dto.PengeluaranResponse
		if err := rows.Scan(&p.ID, &p.Keterangan, &p.Jumlah, &p.Kategori, &p.At, &p.Oleh); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
