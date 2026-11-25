package services

import (
	"database/sql"

	"warkop/internal/dto"
	"warkop/internal/types"
	"warkop/internal/utils"
)

type MenuService struct {
	db  *sql.DB
	hub *EventHub
}

func NewMenuService(db *sql.DB, hub *EventHub) *MenuService {
	return &MenuService{db: db, hub: hub}
}

// muatResep mengambil resep dasar, tambahan resep per opsi, dan sisa porsi dari stok bahan untuk seluruh menu klien.
func (s *MenuService) muatResep(kid int64) (resep, opsi map[int64][]dto.ResepItem, porsi map[int64]int, err error) {
	resep, opsi, porsi = map[int64][]dto.ResepItem{}, map[int64][]dto.ResepItem{}, map[int64]int{}
	baca := func(q string, isi func(a, b int64, n int)) error {
		rows, err := s.db.Query(q, kid)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var a, b int64
			var n int
			if err := rows.Scan(&a, &b, &n); err != nil {
				return err
			}
			isi(a, b, n)
		}
		return rows.Err()
	}
	if err = baca("SELECT menu_id,bahan_id,jumlah FROM resep WHERE klien_id=? ORDER BY id", func(m, b int64, n int) {
		resep[m] = append(resep[m], dto.ResepItem{BahanID: b, Jumlah: n})
	}); err != nil {
		return
	}
	if err = baca("SELECT opsi_id,bahan_id,jumlah FROM resep_opsi WHERE klien_id=? ORDER BY id", func(o, b int64, n int) {
		opsi[o] = append(opsi[o], dto.ResepItem{BahanID: b, Jumlah: n})
	}); err != nil {
		return
	}
	err = baca(`SELECT r.menu_id, 0, MIN(FLOOR(b.stok / r.jumlah)) FROM resep r JOIN bahan b ON b.id=r.bahan_id
		WHERE r.klien_id=? AND r.jumlah>0 GROUP BY r.menu_id`, func(m, _ int64, n int) { porsi[m] = n })
	return
}

func (s *MenuService) List(kid int64) ([]*dto.KategoriResponse, error) {
	resep, resepOpsi, porsi, err := s.muatResep(kid)
	if err != nil {
		return nil, err
	}
	kats := []*dto.KategoriResponse{}
	katByID := map[int64]*dto.KategoriResponse{}
	rows, err := s.db.Query("SELECT id,nama,urutan FROM kategori WHERE klien_id=? ORDER BY urutan,id", kid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		k := &dto.KategoriResponse{Menu: []*dto.MenuResponse{}}
		if err := rows.Scan(&k.ID, &k.Nama, &k.Urutan); err != nil {
			rows.Close()
			return nil, err
		}
		kats, katByID[k.ID] = append(kats, k), k
	}
	rows.Close()

	menuByID := map[int64]*dto.MenuResponse{}
	rows, err = s.db.Query("SELECT id,kategori_id,nama,harga,harga_modal,komisi,tersedia,track_stok,stok,COALESCE(tenant_id,0) FROM menu WHERE klien_id=? AND aktif=1 ORDER BY nama", kid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		m := &dto.MenuResponse{Varian: []*dto.GrupResponse{}, Resep: []dto.ResepItem{}}
		if err := rows.Scan(&m.ID, &m.KategoriID, &m.Nama, &m.Harga, &m.HargaModal, &m.Komisi, &m.Tersedia, &m.TrackStok, &m.Stok, &m.TenantID); err != nil {
			rows.Close()
			return nil, err
		}
		if r := resep[m.ID]; r != nil {
			m.Resep = r
		}
		if p, ada := porsi[m.ID]; ada {
			m.PorsiBahan = &p
		}
		menuByID[m.ID] = m
		if k := katByID[m.KategoriID]; k != nil {
			k.Menu = append(k.Menu, m)
		}
	}
	rows.Close()

	rows, err = s.db.Query(`SELECT i.menu_id, SUM(i.qty) FROM pesanan_item i JOIN pesanan p ON p.id=i.pesanan_id
		WHERE p.klien_id=? AND p.status='lunas' AND p.lunas_at >= NOW() - INTERVAL 30 DAY GROUP BY i.menu_id`, kid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int64
		var n int
		if err := rows.Scan(&id, &n); err != nil {
			rows.Close()
			return nil, err
		}
		if m := menuByID[id]; m != nil {
			m.Laku = n
		}
	}
	rows.Close()

	grupByID := map[int64]*dto.GrupResponse{}
	rows, err = s.db.Query(`SELECT g.id,g.menu_id,g.nama,g.wajib,g.multi,g.ganti_harga FROM varian_grup g
		JOIN menu m ON m.id=g.menu_id WHERE g.klien_id=? AND m.aktif=1 ORDER BY g.id`, kid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		g, menuID := &dto.GrupResponse{Opsi: []dto.OpsiResponse{}}, int64(0)
		if err := rows.Scan(&g.ID, &menuID, &g.Nama, &g.Wajib, &g.Multi, &g.GantiHarga); err != nil {
			rows.Close()
			return nil, err
		}
		grupByID[g.ID] = g
		if m := menuByID[menuID]; m != nil {
			m.Varian = append(m.Varian, g)
		}
	}
	rows.Close()

	rows, err = s.db.Query("SELECT id,grup_id,nama,tambahan FROM varian_opsi WHERE klien_id=? ORDER BY id", kid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		o, grupID := dto.OpsiResponse{Bahan: []dto.ResepItem{}}, int64(0)
		if err := rows.Scan(&o.ID, &grupID, &o.Nama, &o.Tambahan); err != nil {
			return nil, err
		}
		if b := resepOpsi[o.ID]; b != nil {
			o.Bahan = b
		}
		if g := grupByID[grupID]; g != nil {
			g.Opsi = append(g.Opsi, o)
		}
	}
	return kats, rows.Err()
}

func (s *MenuService) CreateKategori(kid int64, req dto.KategoriRequest) error {
	_, err := s.db.Exec("INSERT INTO kategori (klien_id,nama,urutan) VALUES (?,?,?)", kid, req.Nama, req.Urutan)
	s.hub.Publish(kid, "menu")
	return err
}

func (s *MenuService) UpdateKategori(kid, id int64, req dto.KategoriRequest) error {
	if _, err := s.db.Exec("UPDATE kategori SET nama=?, urutan=? WHERE id=? AND klien_id=?", req.Nama, req.Urutan, id, kid); err != nil {
		return err
	}
	s.hub.Publish(kid, "menu")
	return nil
}

func (s *MenuService) DeleteKategori(kid, id int64) error {
	// menu yang di-soft-delete tetap mengacu ke kategori, jadi FK menolak penghapusan: disengaja.
	res, err := s.db.Exec("DELETE FROM kategori WHERE id=? AND klien_id=?", id, kid)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return types.NotFound("Kategori")
	}
	s.hub.Publish(kid, "menu")
	return nil
}

func (s *MenuService) Create(kid int64, req dto.MenuRequest) (int64, error) {
	return s.save(kid, 0, req)
}

func (s *MenuService) Update(kid, id int64, req dto.MenuRequest) error {
	_, err := s.save(kid, id, req)
	return err
}

// isiResep menyimpan satu takaran bahan; bahan harus milik klien yang sama. tabel dan kolom hanya diisi konstanta dari kode ini.
func isiResep(tx *sql.Tx, tabel, kolom string, kid, refID int64, r dto.ResepItem) error {
	res, err := tx.Exec("INSERT INTO "+tabel+" (klien_id,"+kolom+",bahan_id,jumlah) SELECT ?,?,id,? FROM bahan WHERE id=? AND klien_id=? AND aktif=1",
		kid, refID, r.Jumlah, r.BahanID, kid)
	if utils.MySQLErrNo(err) == utils.MySQLDuplicate {
		return types.BadRequest("Bahan yang sama tidak boleh dua kali dalam satu resep")
	}
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return types.BadRequest("Bahan tidak ditemukan")
	}
	return nil
}

func cekGantiHarga(varian []dto.GrupRequest) error {
	n := 0
	for _, g := range varian {
		if !g.GantiHarga {
			continue
		}
		if g.Multi {
			return types.BadRequest("Grup penentu harga hanya boleh satu pilihan")
		}
		n++
	}
	if n > 1 {
		return types.BadRequest("Hanya satu grup varian yang boleh menentukan harga")
	}
	return nil
}

func (s *MenuService) save(kid, id int64, req dto.MenuRequest) (int64, error) {
	if err := cekGantiHarga(req.Varian); err != nil {
		return 0, err
	}
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		var ada int
		if tx.QueryRow("SELECT 1 FROM tenant WHERE id=? AND klien_id=? AND aktif=1", req.TenantID, kid).Scan(&ada) != nil {
			return types.BadRequest("Tenant tidak ditemukan")
		}
		if tx.QueryRow("SELECT 1 FROM kategori WHERE id=? AND klien_id=?", req.KategoriID, kid).Scan(&ada) != nil {
			return types.BadRequest("Kategori tidak ditemukan")
		}
		if id == 0 {
			res, err := tx.Exec("INSERT INTO menu (klien_id,kategori_id,nama,harga,harga_modal,komisi,tersedia,track_stok,stok,tenant_id) VALUES (?,?,?,?,?,?,?,?,?,?)",
				kid, req.KategoriID, req.Nama, req.Harga, req.HargaModal, req.Komisi, req.Tersedia, req.TrackStok, req.Stok, req.TenantID)
			if err != nil {
				return err
			}
			if id, err = res.LastInsertId(); err != nil {
				return err
			}
		} else {
			var aktif bool
			if tx.QueryRow("SELECT aktif FROM menu WHERE id=? AND klien_id=? FOR UPDATE", id, kid).Scan(&aktif) != nil || !aktif {
				return types.NotFound("Menu")
			}
			if _, err := tx.Exec("UPDATE menu SET kategori_id=?,nama=?,harga=?,harga_modal=?,komisi=?,tersedia=?,track_stok=?,stok=?,tenant_id=? WHERE id=? AND klien_id=?",
				req.KategoriID, req.Nama, req.Harga, req.HargaModal, req.Komisi, req.Tersedia, req.TrackStok, req.Stok, req.TenantID, id, kid); err != nil {
				return err
			}
			// item pesanan menyalin nama dan harga opsi, jadi varian aman diganti total
			if _, err := tx.Exec("DELETE FROM varian_grup WHERE klien_id=? AND menu_id=?", kid, id); err != nil {
				return err
			}
		}
		if _, err := tx.Exec("DELETE FROM resep WHERE klien_id=? AND menu_id=?", kid, id); err != nil {
			return err
		}
		for _, r := range req.Resep {
			if r.Jumlah <= 0 {
				return types.BadRequest("Takaran resep harus lebih dari 0")
			}
			if err := isiResep(tx, "resep", "menu_id", kid, id, r); err != nil {
				return err
			}
		}
		for _, g := range req.Varian {
			res, err := tx.Exec("INSERT INTO varian_grup (klien_id,menu_id,nama,wajib,multi,ganti_harga) VALUES (?,?,?,?,?,?)",
				kid, id, g.Nama, g.Wajib, g.Multi, g.GantiHarga)
			if err != nil {
				return err
			}
			grupID, _ := res.LastInsertId()
			for _, o := range g.Opsi {
				ores, err := tx.Exec("INSERT INTO varian_opsi (klien_id,grup_id,nama,tambahan) VALUES (?,?,?,?)", kid, grupID, o.Nama, o.Tambahan)
				if err != nil {
					return err
				}
				opsiID, _ := ores.LastInsertId()
				for _, b := range o.Bahan {
					if b.Jumlah == 0 {
						return types.BadRequest("Takaran bahan pada opsi tidak boleh 0")
					}
					if err := isiResep(tx, "resep_opsi", "opsi_id", kid, opsiID, b); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		if utils.MySQLErrNo(err) == 1452 {
			return 0, types.BadRequest("Kategori tidak ditemukan")
		}
		return 0, err
	}
	s.hub.Publish(kid, "menu")
	return id, nil
}

func (s *MenuService) Delete(kid, id int64) error {
	res, err := s.db.Exec("UPDATE menu SET aktif=0 WHERE id=? AND klien_id=? AND aktif=1", id, kid)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return types.NotFound("Menu")
	}
	s.hub.Publish(kid, "menu")
	return nil
}

func (s *MenuService) SetStok(kid, id int64, req dto.StokRequest) error {
	res, err := s.db.Exec("UPDATE menu SET tersedia=?, stok=COALESCE(?,stok), track_stok=COALESCE(?,track_stok) WHERE id=? AND klien_id=? AND aktif=1",
		req.Tersedia, req.Stok, req.TrackStok, id, kid)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return types.NotFound("Menu")
	}
	s.hub.Publish(kid, "menu")
	return nil
}
