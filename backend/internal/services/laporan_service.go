package services

import (
	"database/sql"
	"math"
	"time"

	"warkop/internal/dto"
	"warkop/internal/types"
)

type LaporanService struct {
	db *sql.DB
}

const (
	dateLayout   = "2006-01-02"
	maxRangeDays = 366
)

func NewLaporanService(db *sql.DB) *LaporanService {
	return &LaporanService{db: db}
}

// parseRange mengubah tanggal inklusif [dari, sampai] menjadi interval setengah terbuka [from, to).
func parseRange(dari, sampai string, now time.Time) (from, to time.Time, err error) {
	today := now.Format(dateLayout)
	if dari == "" {
		dari = today
	}
	if sampai == "" {
		sampai = today
	}
	d, err1 := time.ParseInLocation(dateLayout, dari, now.Location())
	e, err2 := time.ParseInLocation(dateLayout, sampai, now.Location())
	if err1 != nil || err2 != nil {
		return from, to, types.BadRequest("Format tanggal harus YYYY-MM-DD")
	}
	if e.Before(d) || e.Sub(d) > maxRangeDays*24*time.Hour {
		return from, to, types.BadRequest("Rentang tanggal tidak valid (maksimal 366 hari)")
	}
	return d, e.AddDate(0, 0, 1), nil
}

// hitungSetoran: komisi dihitung dari penjualan setelah diskon; sisanya disetor ke tenant.
func hitungSetoran(kotor, diskon int, persen float64) (komisi, setor int) {
	dasar := kotor - diskon
	komisi = int(math.Round(float64(dasar) * persen / 100))
	return komisi, dasar - komisi
}

// Tenant menghitung setoran per tenant dari bill lunas. Diskon bill dibagi proporsional menurut nilai item tiap tenant.
// ponytail: memakai persentase komisi saat ini, bukan yang berlaku saat transaksi; simpan persen per item bila tarif sering berubah.
func (s *LaporanService) Tenant(kid int64, dari, sampai string) (*dto.SetoranResponse, error) {
	from, to, err := parseRange(dari, sampai, time.Now())
	if err != nil {
		return nil, err
	}

	type jumlah struct{ kotor, diskon int }
	totals := map[int64]jumlah{}
	rows, err := s.db.Query(`SELECT COALESCE(i.tenant_id,0), SUM(i.harga*i.qty), COALESCE(SUM(ROUND(p.diskon*(i.harga*i.qty)/ps.sub)),0)
		FROM pesanan_item i
		JOIN pesanan p ON p.id=i.pesanan_id AND p.klien_id=? AND p.status='lunas'
		JOIN shift sh ON sh.id=p.shift_id AND sh.buka_at>=? AND sh.buka_at<?
		JOIN (SELECT pesanan_id, SUM(harga*qty) AS sub FROM pesanan_item WHERE klien_id=? GROUP BY pesanan_id) ps ON ps.pesanan_id=p.id AND ps.sub>0
		GROUP BY i.tenant_id`, kid, from, to, kid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int64
		var j jumlah
		if err := rows.Scan(&id, &j.kotor, &j.diskon); err != nil {
			rows.Close()
			return nil, err
		}
		totals[id] = j
	}
	rows.Close()

	out := &dto.SetoranResponse{Dari: from.Format(dateLayout), Sampai: to.AddDate(0, 0, -1).Format(dateLayout), Tenant: []dto.SetoranTenant{}}
	rows, err = s.db.Query("SELECT id,nama,komisi_persen,aktif FROM tenant WHERE klien_id=? ORDER BY id", kid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var t dto.SetoranTenant
		var aktif bool
		if err := rows.Scan(&t.ID, &t.Nama, &t.KomisiPersen, &aktif); err != nil {
			return nil, err
		}
		j := totals[t.ID]
		if !aktif && j.kotor == 0 {
			continue
		}
		t.Kotor, t.Diskon = j.kotor, j.diskon
		t.Komisi, t.Setor = hitungSetoran(t.Kotor, t.Diskon, t.KomisiPersen)
		out.Tenant = append(out.Tenant, t)
	}
	return out, rows.Err()
}

// Get: hari penjualan = tanggal shift dibuka (hari kerja), bukan jam di dinding.
// Shift yang melewati tengah malam tetap dihitung satu hari.
func (s *LaporanService) Get(kid int64, dari, sampai string) (*dto.LaporanResponse, error) {
	from, to, err := parseRange(dari, sampai, time.Now())
	if err != nil {
		return nil, err
	}
	out := &dto.LaporanResponse{
		Dari: from.Format(dateLayout), Sampai: to.AddDate(0, 0, -1).Format(dateLayout),
		PerMetode: []dto.MetodeTotal{}, PerHari: []dto.HariTotal{}, Terlaris: []dto.MenuTerlaris{}, Shift: []dto.ShiftResponse{},
	}

	if err := s.db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(p.diskon),0) FROM pesanan p JOIN shift sh ON sh.id=p.shift_id
		WHERE p.klien_id=? AND p.status='lunas' AND sh.buka_at>=? AND sh.buka_at<?`, kid, from, to).Scan(&out.Transaksi, &out.Diskon); err != nil {
		return nil, err
	}
	if err := s.db.QueryRow(`SELECT COALESCE(SUM(e.jumlah),0) FROM pengeluaran e JOIN shift sh ON sh.id=e.shift_id
		WHERE e.klien_id=? AND sh.buka_at>=? AND sh.buka_at<?`, kid, from, to).Scan(&out.Keluar); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(`SELECT b.metode, SUM(b.jumlah) FROM pembayaran b JOIN pesanan p ON p.id=b.pesanan_id
		JOIN shift sh ON sh.id=b.shift_id
		WHERE p.klien_id=? AND p.status='lunas' AND sh.buka_at>=? AND sh.buka_at<? GROUP BY b.metode ORDER BY 2 DESC`, kid, from, to)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var m dto.MetodeTotal
		if err := rows.Scan(&m.Metode, &m.Total); err != nil {
			rows.Close()
			return nil, err
		}
		out.Omzet += m.Total
		out.PerMetode = append(out.PerMetode, m)
	}
	rows.Close()

	rows, err = s.db.Query(`SELECT DATE(sh.buka_at), SUM(b.jumlah) FROM pembayaran b JOIN pesanan p ON p.id=b.pesanan_id
		JOIN shift sh ON sh.id=b.shift_id
		WHERE p.klien_id=? AND p.status='lunas' AND sh.buka_at>=? AND sh.buka_at<? GROUP BY DATE(sh.buka_at) ORDER BY 1`, kid, from, to)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var h dto.HariTotal
		var tgl time.Time
		if err := rows.Scan(&tgl, &h.Total); err != nil {
			rows.Close()
			return nil, err
		}
		h.Tanggal = tgl.Format(dateLayout)
		out.PerHari = append(out.PerHari, h)
	}
	rows.Close()

	rows, err = s.db.Query(`SELECT i.nama, SUM(i.qty), SUM(i.qty*i.harga) FROM pesanan_item i JOIN pesanan p ON p.id=i.pesanan_id
		JOIN shift sh ON sh.id=p.shift_id
		WHERE p.klien_id=? AND p.status='lunas' AND sh.buka_at>=? AND sh.buka_at<? GROUP BY i.nama ORDER BY 2 DESC LIMIT 10`, kid, from, to)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var t dto.MenuTerlaris
		if err := rows.Scan(&t.Nama, &t.Qty, &t.Total); err != nil {
			rows.Close()
			return nil, err
		}
		out.Terlaris = append(out.Terlaris, t)
	}
	rows.Close()

	// hanya 10 terakhir: rentang setahun berarti ratusan shift; daftar lengkap ada di Riwayat Shift
	rows, err = s.db.Query("SELECT id FROM shift WHERE klien_id=? AND buka_at>=? AND buka_at<? ORDER BY id DESC LIMIT 10", kid, from, to)
	if err != nil {
		return nil, err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	for _, id := range ids {
		sh, err := loadShift(s.db, kid, id)
		if err != nil {
			return nil, err
		}
		out.Shift = append(out.Shift, *sh)
	}
	return out, nil
}
