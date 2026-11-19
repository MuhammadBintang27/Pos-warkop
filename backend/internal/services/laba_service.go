package services

import (
	"database/sql"
	"math"
	"time"

	"warkop/internal/dto"
)

type LabaService struct {
	db *sql.DB
}

func NewLabaService(db *sql.DB) *LabaService {
	return &LabaService{db: db}
}

// bagian FROM bersama: item bill lunas milik klien pada hari kerja (tanggal shift dibuka) dalam rentang.
const itemLunasSQL = `FROM pesanan_item i
	JOIN pesanan p ON p.id=i.pesanan_id AND p.klien_id=? AND p.status='lunas'
	JOIN shift sh ON sh.id=p.shift_id AND sh.buka_at>=? AND sh.buka_at<?
	LEFT JOIN tenant t ON t.id=i.tenant_id`

const milikSendiriSQL = "COALESCE(t.milik_sendiri,1)=1"

func bulat(f float64) int { return int(math.Round(f)) }

func hitungLaba(r *dto.LabaRugiResponse) {
	r.Pendapatan = r.PenjualanSendiri + r.KomisiTenant
	r.LabaKotor = r.Pendapatan - r.HPP - r.Susut
	r.LabaBersih = r.LabaKotor - r.TotalPengeluaran - r.Gaji
}

// LabaRugi: pendapatan = penjualan tenant milik sendiri + komisi tenant lain; HPP dari salinan biaya bahan saat dipesan;
// gaji dihitung dari slip yang dibayar dalam rentang. shortcut: berbasis kas untuk gaji (tanggal dibayar), bukan periode kerja.
func (s *LabaService) LabaRugi(kid int64, dari, sampai string) (*dto.LabaRugiResponse, error) {
	from, to, err := parseRange(dari, sampai, time.Now())
	if err != nil {
		return nil, err
	}
	out := &dto.LabaRugiResponse{Dari: from.Format(dateLayout), Sampai: to.AddDate(0, 0, -1).Format(dateLayout), Pengeluaran: []dto.PengeluaranKategori{}}

	rows, err := s.db.Query(`SELECT COALESCE(t.milik_sendiri,1), COALESCE(t.komisi_persen,100), SUM(i.harga*i.qty),
		COALESCE(SUM(ROUND(p.diskon*(i.harga*i.qty)/ps.sub)),0)
		`+itemLunasSQL+`
		JOIN (SELECT pesanan_id, SUM(harga*qty) AS sub FROM pesanan_item WHERE klien_id=? GROUP BY pesanan_id) ps ON ps.pesanan_id=p.id AND ps.sub>0
		GROUP BY i.tenant_id`, kid, from, to, kid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var sendiri bool
		var persen float64
		var kotor, diskon int
		if err := rows.Scan(&sendiri, &persen, &kotor, &diskon); err != nil {
			rows.Close()
			return nil, err
		}
		komisi, _ := hitungSetoran(kotor, diskon, persen)
		if sendiri {
			out.PenjualanSendiri += komisi
		} else {
			out.KomisiTenant += komisi
		}
	}
	rows.Close()

	var hpp, modal, susut float64
	if err := s.db.QueryRow(`SELECT COALESCE(SUM(i.qty*ib.jumlah*ib.biaya),0) `+itemLunasSQL+`
		JOIN item_bahan ib ON ib.item_id=i.id WHERE `+milikSendiriSQL, kid, from, to).Scan(&hpp); err != nil {
		return nil, err
	}
	if err := s.db.QueryRow(`SELECT COALESCE(SUM(i.qty*i.modal),0) `+itemLunasSQL+` WHERE `+milikSendiriSQL, kid, from, to).Scan(&modal); err != nil {
		return nil, err
	}
	out.HPP = bulat(hpp + modal)

	// biaya NULL = log stok awal saat bahan dibuat, bukan selisih
	if err := s.db.QueryRow(`SELECT COALESCE(-SUM(jumlah*biaya),0) FROM bahan_log
		WHERE klien_id=? AND jenis='opname' AND biaya IS NOT NULL AND at>=? AND at<?`, kid, from, to).Scan(&susut); err != nil {
		return nil, err
	}
	out.Susut = bulat(susut)

	if err := s.db.QueryRow(`SELECT COALESCE(SUM(i.harga*i.qty),0) `+itemLunasSQL+`
		WHERE `+milikSendiriSQL+` AND i.modal=0 AND NOT EXISTS (SELECT 1 FROM item_bahan ib WHERE ib.item_id=i.id)`, kid, from, to).Scan(&out.OmzetTanpaResep); err != nil {
		return nil, err
	}

	rows, err = s.db.Query(`SELECT e.kategori, SUM(e.jumlah) FROM pengeluaran e JOIN shift sh ON sh.id=e.shift_id
		WHERE e.klien_id=? AND sh.buka_at>=? AND sh.buka_at<? GROUP BY e.kategori ORDER BY 2 DESC`, kid, from, to)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var k dto.PengeluaranKategori
		if err := rows.Scan(&k.Kategori, &k.Total); err != nil {
			rows.Close()
			return nil, err
		}
		k.DalamHPP = k.Kategori == "bahan"
		// refund sudah mengurangi pendapatan karena bill-nya batal; jangan dikurangkan dua kali
		if !k.DalamHPP && k.Kategori != "refund" {
			out.TotalPengeluaran += k.Total
		}
		out.Pengeluaran = append(out.Pengeluaran, k)
	}
	rows.Close()

	if err := s.db.QueryRow(`SELECT COALESCE(SUM(pokok+tunjangan+bonus-potongan),0) FROM gaji_slip
		WHERE klien_id=? AND status='dibayar' AND dibayar_at>=? AND dibayar_at<?`, kid, from, to).Scan(&out.Gaji); err != nil {
		return nil, err
	}
	hitungLaba(out)
	return out, nil
}

// PerMenu: penjualan dan HPP nyata per menu pada rentang, dari salinan biaya saat dipesan. Menu tanpa resep ditandai.
func (s *LabaService) PerMenu(kid int64, dari, sampai string) ([]dto.MarginMenu, error) {
	from, to, err := parseRange(dari, sampai, time.Now())
	if err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`SELECT i.menu_id, MAX(i.nama), SUM(i.qty), SUM(i.harga*i.qty),
		COALESCE(SUM(i.qty*(i.modal+(SELECT COALESCE(SUM(ib.jumlah*ib.biaya),0) FROM item_bahan ib WHERE ib.item_id=i.id))),0),
		MAX(i.modal>0 OR EXISTS(SELECT 1 FROM item_bahan ib WHERE ib.item_id=i.id))
		`+itemLunasSQL+` WHERE `+milikSendiriSQL+` GROUP BY i.menu_id ORDER BY 4 DESC`, kid, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []dto.MarginMenu{}
	for rows.Next() {
		var m dto.MarginMenu
		var hpp float64
		if err := rows.Scan(&m.MenuID, &m.Nama, &m.Terjual, &m.Penjualan, &hpp, &m.AdaResep); err != nil {
			return nil, err
		}
		m.HPP = bulat(hpp)
		if m.AdaResep {
			m.Laba = m.Penjualan - m.HPP
			if m.Penjualan > 0 {
				m.MarginPersen = math.Round(float64(m.Laba)/float64(m.Penjualan)*1000) / 10
			}
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
