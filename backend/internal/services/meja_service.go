package services

import (
	"database/sql"
	"fmt"
	"strings"

	"warkop/internal/dto"
	"warkop/internal/types"
	"warkop/internal/utils"
)

type MejaService struct {
	db  *sql.DB
	hub *EventHub
}

func NewMejaService(db *sql.DB, hub *EventHub) *MejaService {
	return &MejaService{db: db, hub: hub}
}

func (s *MejaService) Denah(kid int64) (*dto.DenahResponse, error) {
	out := &dto.DenahResponse{Area: []dto.AreaResponse{}, Meja: []*dto.MejaResponse{}, Takeaway: []dto.PesananRingkas{}}
	byID := map[int64]*dto.MejaResponse{}

	rows, err := s.db.Query("SELECT id,nama,urutan FROM area WHERE klien_id=? AND aktif=1 ORDER BY urutan,id", kid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var a dto.AreaResponse
		if err := rows.Scan(&a.ID, &a.Nama, &a.Urutan); err != nil {
			rows.Close()
			return nil, err
		}
		out.Area = append(out.Area, a)
	}
	rows.Close()

	rows, err = s.db.Query("SELECT id,nama,area_id,pos_x,pos_y FROM meja WHERE klien_id=? AND aktif=1 ORDER BY id", kid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		m := &dto.MejaResponse{Pesanan: []dto.PesananRingkas{}}
		if err := rows.Scan(&m.ID, &m.Nama, &m.AreaID, &m.PosX, &m.PosY); err != nil {
			rows.Close()
			return nil, err
		}
		out.Meja, byID[m.ID] = append(out.Meja, m), m
	}
	rows.Close()

	rows, err = s.db.Query(`SELECT p.id, p.meja_id, p.atas_nama, COALESCE(SUM(i.harga*i.qty),0)-p.diskon
		FROM pesanan p LEFT JOIN pesanan_item i ON i.pesanan_id=p.id
		WHERE p.klien_id=? AND p.status='open' GROUP BY p.id HAVING COUNT(i.id) > 0 ORDER BY p.id`, kid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p dto.PesananRingkas
		var mejaID *int64
		if err := rows.Scan(&p.ID, &mejaID, &p.AtasNama, &p.Total); err != nil {
			return nil, err
		}
		if mejaID == nil {
			out.Takeaway = append(out.Takeaway, p)
		} else if m := byID[*mejaID]; m != nil {
			m.Pesanan = append(m.Pesanan, p)
		}
	}
	return out, rows.Err()
}

// cekArea memastikan area (bila diisi) milik klien yang sama dan masih aktif.
func cekArea(q querier, kid int64, areaID *int64) error {
	if areaID == nil {
		return nil
	}
	var ada int
	if q.QueryRow("SELECT 1 FROM area WHERE id=? AND klien_id=? AND aktif=1", *areaID, kid).Scan(&ada) != nil {
		return types.BadRequest("Area tidak ditemukan")
	}
	return nil
}

func (s *MejaService) Create(kid int64, req dto.MejaRequest) error {
	if err := cekArea(s.db, kid, req.AreaID); err != nil {
		return err
	}
	if _, err := s.db.Exec("INSERT INTO meja (klien_id,nama,area_id) VALUES (?,?,?)", kid, req.Nama, req.AreaID); err != nil {
		return err
	}
	s.hub.Publish(kid, "meja")
	return nil
}

func namaMeja(awalan string, mulai, jumlah int, ada map[string]bool) []string {
	var baru []string
	for n := mulai; n < mulai+jumlah; n++ {
		if nama := fmt.Sprintf("%s %d", awalan, n); !ada[nama] {
			baru = append(baru, nama)
		}
	}
	return baru
}

func (s *MejaService) Generate(kid int64, req dto.GenerateMejaRequest) (*dto.GenerateMejaResponse, error) {
	awalan := strings.TrimSpace(req.Awalan)
	if awalan == "" {
		return nil, types.BadRequest("Awalan nama meja wajib diisi")
	}
	var baru []string
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		if err := cekArea(tx, kid, req.AreaID); err != nil {
			return err
		}
		rows, err := tx.Query("SELECT nama FROM meja WHERE klien_id=? AND aktif=1", kid)
		if err != nil {
			return err
		}
		ada := map[string]bool{}
		for rows.Next() {
			var nama string
			if err := rows.Scan(&nama); err != nil {
				rows.Close()
				return err
			}
			ada[nama] = true
		}
		rows.Close()

		baru = namaMeja(awalan, req.Mulai, req.Jumlah, ada)
		for _, nama := range baru {
			if _, err := tx.Exec("INSERT INTO meja (klien_id,nama,area_id) VALUES (?,?,?)", kid, nama, req.AreaID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.hub.Publish(kid, "meja")
	return &dto.GenerateMejaResponse{Dibuat: len(baru), Dilewati: req.Jumlah - len(baru)}, nil
}

func (s *MejaService) SimpanPosisi(kid int64, req dto.PosisiRequest) error {
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		for _, m := range req.Meja {
			if _, err := tx.Exec("UPDATE meja SET pos_x=?, pos_y=? WHERE id=? AND klien_id=? AND aktif=1", m.PosX, m.PosY, m.ID, kid); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	s.hub.Publish(kid, "meja")
	return nil
}

func (s *MejaService) Update(kid, id int64, req dto.MejaRequest) error {
	if err := cekArea(s.db, kid, req.AreaID); err != nil {
		return err
	}
	// pindah area: posisi lama milik denah area lain, jadi dikosongkan supaya disusun ulang di denah barunya
	if _, err := s.db.Exec(`UPDATE meja SET nama=?,
		pos_x=IF(area_id <=> ?, pos_x, NULL), pos_y=IF(area_id <=> ?, pos_y, NULL), area_id=?
		WHERE id=? AND klien_id=?`, req.Nama, req.AreaID, req.AreaID, req.AreaID, id, kid); err != nil {
		return err
	}
	s.hub.Publish(kid, "meja")
	return nil
}

func (s *MejaService) Delete(kid, id int64) error {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM pesanan p WHERE p.klien_id=? AND p.meja_id=? AND p.status='open'
		AND EXISTS (SELECT 1 FROM pesanan_item WHERE pesanan_id=p.id)`, kid, id).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return types.BadRequest("Meja masih punya pesanan terbuka")
	}
	res, err := s.db.Exec("UPDATE meja SET aktif=0 WHERE id=? AND klien_id=? AND aktif=1", id, kid)
	if err != nil {
		return err
	}
	if c, _ := res.RowsAffected(); c == 0 {
		return types.NotFound("Meja")
	}
	s.hub.Publish(kid, "meja")
	return nil
}

func (s *MejaService) SimpanArea(kid, id int64, req dto.AreaRequest) error {
	nama := strings.TrimSpace(req.Nama)
	if nama == "" {
		return types.BadRequest("Nama area wajib diisi")
	}
	if id == 0 {
		if _, err := s.db.Exec("INSERT INTO area (klien_id,nama,urutan) VALUES (?,?,?)", kid, nama, req.Urutan); err != nil {
			return err
		}
	} else {
		res, err := s.db.Exec("UPDATE area SET nama=?, urutan=? WHERE id=? AND klien_id=? AND aktif=1", nama, req.Urutan, id, kid)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			if cekArea(s.db, kid, &id) != nil {
				return types.NotFound("Area")
			}
		}
	}
	s.hub.Publish(kid, "meja")
	return nil
}

// HapusArea hanya untuk area kosong, supaya meja tidak kehilangan areanya diam-diam.
func (s *MejaService) HapusArea(kid, id int64) error {
	var n int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM meja WHERE klien_id=? AND area_id=? AND aktif=1", kid, id).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return types.BadRequest(fmt.Sprintf("Area masih berisi %d meja. Pindahkan atau hapus mejanya dulu", n))
	}
	res, err := s.db.Exec("UPDATE area SET aktif=0 WHERE id=? AND klien_id=? AND aktif=1", id, kid)
	if err != nil {
		return err
	}
	if c, _ := res.RowsAffected(); c == 0 {
		return types.NotFound("Area")
	}
	s.hub.Publish(kid, "meja")
	return nil
}
