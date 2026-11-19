package services

import (
	"time"

	"warkop/internal/dto"
)

const shiftPerHalaman = 20

// RiwayatShift: daftar shift berhalaman (server) dengan angka ringkas per shift; filter kasir dan "hanya selisih".
func (s *ShiftService) Riwayat(kid int64, q dto.ShiftQuery) (*dto.ShiftListResponse, error) {
	from, to, err := parseRange(q.Dari, q.Sampai, time.Now())
	if err != nil {
		return nil, err
	}
	where := " FROM shift s JOIN users u ON u.id=s.kasir_id LEFT JOIN users ut ON ut.id=s.tutup_oleh WHERE s.klien_id=? AND s.buka_at>=? AND s.buka_at<?"
	args := []any{kid, from, to}
	if q.Kasir > 0 {
		where += " AND s.kasir_id=?"
		args = append(args, q.Kasir)
	}
	if q.Selisih {
		where += " AND s.kas_hitung IS NOT NULL AND s.kas_hitung<>s.kas_sistem"
	}
	out := &dto.ShiftListResponse{Items: []dto.ShiftRingkas{}}
	if err := s.db.QueryRow("SELECT COUNT(*), COALESCE(SUM(s.kas_hitung-s.kas_sistem),0)"+where, args...).Scan(&out.Total, &out.TotalSelisih); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`SELECT s.id,u.nama,COALESCE(ut.nama,''),s.buka_at,s.tutup_at,s.kas_awal,s.kas_sistem,s.kas_hitung,
		(SELECT COUNT(*) FROM pesanan p WHERE p.shift_id=s.id AND p.status='lunas'),
		(SELECT COALESCE(SUM(b.jumlah),0) FROM pembayaran b JOIN pesanan p ON p.id=b.pesanan_id AND p.status='lunas' WHERE b.shift_id=s.id),
		(SELECT COALESCE(SUM(b.jumlah),0) FROM pembayaran b WHERE b.shift_id=s.id AND b.metode='tunai')
			+ (SELECT COALESCE(SUM(k.jumlah),0) FROM kasbon_bayar k WHERE k.shift_id=s.id AND k.metode='tunai'),
		(SELECT COALESCE(SUM(e.jumlah),0) FROM pengeluaran e WHERE e.shift_id=s.id)`+where+
		" ORDER BY s.buka_at DESC, s.id DESC LIMIT ? OFFSET ?", append(args, shiftPerHalaman, q.Page*shiftPerHalaman)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var r dto.ShiftRingkas
		if err := rows.Scan(&r.ID, &r.Kasir, &r.TutupOleh, &r.BukaAt, &r.TutupAt, &r.KasAwal, &r.KasSistem, &r.KasHitung,
			&r.Bill, &r.Penjualan, &r.Tunai, &r.Keluar); err != nil {
			return nil, err
		}
		r.Selisih = selisihKas(r.KasSistem, r.KasHitung)
		out.Items = append(out.Items, r)
	}
	return out, rows.Err()
}

func selisihKas(sistem, hitung *int) *int {
	if sistem == nil || hitung == nil {
		return nil
	}
	d := *hitung - *sistem
	return &d
}

// detailShift dipakai halaman detail dan cetak laporan shift.
func detailShift(q querier, kid, id int64) (*dto.ShiftDetail, error) {
	sh, err := loadShift(q, kid, id)
	if err != nil {
		return nil, err
	}
	out := &dto.ShiftDetail{ShiftResponse: *sh, PerMetode: []dto.MetodeTotal{}, Pengeluaran: []dto.PengeluaranResponse{}, Kasbon: []dto.KasbonBayarShift{}}
	if err := q.QueryRow(`SELECT COUNT(*), COALESCE(SUM(p.diskon),0) FROM pesanan p WHERE p.shift_id=? AND p.klien_id=? AND p.status='lunas'`, id, kid).
		Scan(&out.Bill, &out.Diskon); err != nil {
		return nil, err
	}
	if err := q.QueryRow(`SELECT COUNT(*) FROM pesanan p WHERE p.shift_id=? AND p.klien_id=? AND p.status='batal'`, id, kid).Scan(&out.BillBatal); err != nil {
		return nil, err
	}

	rows, err := q.Query(`SELECT b.metode, SUM(b.jumlah) FROM pembayaran b JOIN pesanan p ON p.id=b.pesanan_id AND p.status='lunas'
		WHERE b.shift_id=? AND b.klien_id=? GROUP BY b.metode ORDER BY 2 DESC`, id, kid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var m dto.MetodeTotal
		if err := rows.Scan(&m.Metode, &m.Total); err != nil {
			rows.Close()
			return nil, err
		}
		out.Penjualan += m.Total
		out.PerMetode = append(out.PerMetode, m)
	}
	rows.Close()

	rows, err = q.Query(`SELECT e.id,e.keterangan,e.jumlah,e.kategori,e.at,COALESCE(u.nama,'') FROM pengeluaran e
		LEFT JOIN users u ON u.id=e.dicatat_oleh WHERE e.shift_id=? AND e.klien_id=? ORDER BY e.at`, id, kid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var e dto.PengeluaranResponse
		if err := rows.Scan(&e.ID, &e.Keterangan, &e.Jumlah, &e.Kategori, &e.At, &e.Oleh); err != nil {
			rows.Close()
			return nil, err
		}
		out.Keluar += e.Jumlah
		out.Pengeluaran = append(out.Pengeluaran, e)
	}
	rows.Close()

	rows, err = q.Query(`SELECT pl.nama, k.jumlah, k.metode, k.at FROM kasbon_bayar k JOIN pelanggan pl ON pl.id=k.pelanggan_id
		WHERE k.shift_id=? AND k.klien_id=? ORDER BY k.at`, id, kid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var k dto.KasbonBayarShift
		if err := rows.Scan(&k.Pelanggan, &k.Jumlah, &k.Metode, &k.At); err != nil {
			return nil, err
		}
		if k.Metode == "tunai" {
			out.KasbonTunai += k.Jumlah
		}
		out.Kasbon = append(out.Kasbon, k)
	}
	for _, m := range out.PerMetode {
		if m.Metode == "tunai" {
			out.PenjualanTunai = m.Total
		}
	}
	return out, rows.Err()
}

func (s *ShiftService) Detail(kid, id int64) (*dto.ShiftDetail, error) {
	return detailShift(s.db, kid, id)
}
