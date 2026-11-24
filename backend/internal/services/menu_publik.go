package services

import (
	"warkop/internal/dto"
	"warkop/internal/types"
	"warkop/internal/utils"
)

// KlienDariKode mencari klien aktif untuk link menu publik /m/<kode>.
func (s *MenuService) KlienDariKode(kode string) (int64, error) {
	var kid int64
	if s.db.QueryRow("SELECT id FROM klien WHERE kode=? AND aktif=1", kode).Scan(&kid) != nil || !utils.KlienAktif(kid) {
		return 0, types.NotFound("Menu")
	}
	return kid, nil
}

// Publik menyusun menu per tenant untuk pelanggan. Status habis dihitung sama seperti di kasir.
func (s *MenuService) Publik(kid int64) ([]dto.TenantPublik, error) {
	kats, err := s.List(kid)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.Query("SELECT id,nama,buka FROM tenant WHERE klien_id=? AND aktif=1 ORDER BY id", kid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []dto.TenantPublik{}
	idx := map[int64]int{}
	for rows.Next() {
		var id int64
		t := dto.TenantPublik{Menu: []dto.MenuPublikItem{}}
		if err := rows.Scan(&id, &t.Nama, &t.Buka); err != nil {
			return nil, err
		}
		idx[id] = len(out)
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	lain := -1 // menu tanpa tenant aktif dikumpulkan di "Lainnya"
	for _, k := range kats {
		for _, m := range k.Menu {
			item := dto.MenuPublikItem{
				Nama:     m.Nama,
				Kategori: k.Nama,
				Harga:    m.Harga,
				Habis:    !m.Tersedia || (m.TrackStok && m.Stok <= 0) || (m.PorsiBahan != nil && *m.PorsiBahan == 0),
				Varian:   []dto.VarianPublik{},
			}
			for _, g := range m.Varian {
				v := dto.VarianPublik{Nama: g.Nama, Wajib: g.Wajib, GantiHarga: g.GantiHarga, Opsi: []dto.OpsiPublik{}}
				for _, o := range g.Opsi {
					v.Opsi = append(v.Opsi, dto.OpsiPublik{Nama: o.Nama, Tambahan: o.Tambahan})
				}
				item.Varian = append(item.Varian, v)
			}
			i, ada := idx[m.TenantID]
			if !ada {
				if lain < 0 {
					lain = len(out)
					out = append(out, dto.TenantPublik{Nama: "Lainnya", Buka: true, Menu: []dto.MenuPublikItem{}})
				}
				i = lain
			}
			out[i].Menu = append(out[i].Menu, item)
		}
	}

	tampil := out[:0]
	for _, t := range out {
		if len(t.Menu) > 0 {
			tampil = append(tampil, t)
		}
	}
	return tampil, nil
}
