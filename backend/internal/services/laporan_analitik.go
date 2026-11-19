package services

import (
	"sort"
	"time"

	"warkop/internal/dto"
	"warkop/internal/types"
)

// Analitik dihitung per tab supaya halaman hanya menjalankan query yang sedang dilihat:
// ringkasan (tren + periode lalu), waktu (jam ramai + kategori), karyawan, kasbon (umur kasbon + stok).
// Hari = tanggal shift dibuka, sama dengan Laporan.
func (s *LaporanService) Analitik(kid int64, dari, sampai, bagian string) (*dto.AnalitikResponse, error) {
	now := time.Now()
	from, to, err := parseRange(dari, sampai, now)
	if err != nil {
		return nil, err
	}
	hari := int(to.Sub(from).Hours()/24 + 0.5)
	prevFrom := from.AddDate(0, 0, -hari)
	out := &dto.AnalitikResponse{
		Dari: from.Format(dateLayout), Sampai: to.AddDate(0, 0, -1).Format(dateLayout),
		SebelumnyaDari: prevFrom.Format(dateLayout), SebelumnyaSampai: from.AddDate(0, 0, -1).Format(dateLayout),
	}
	type langkah = func(*dto.AnalitikResponse, int64, time.Time, time.Time) error
	var jalan []langkah
	switch bagian {
	case "ringkasan":
		if out.Tren, err = s.tren(kid, from, to); err != nil {
			return nil, err
		}
		out.Ringkasan = ringkas(out.Tren)
		prev, err := s.tren(kid, prevFrom, from)
		if err != nil {
			return nil, err
		}
		out.Sebelumnya = ringkas(prev)
	case "waktu":
		jalan = []langkah{s.jamRamai, s.kategori, s.perTenant}
	case "karyawan":
		jalan = []langkah{s.karyawan}
	case "kasbon":
		jalan = []langkah{s.kasbonUmur, s.stokNilai}
	default:
		return nil, types.BadRequest("Bagian laporan tidak dikenal")
	}
	for _, f := range jalan {
		if err := f(out, kid, from, to); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func ringkas(tren []dto.TrenHari) dto.RingkasanPeriode {
	var r dto.RingkasanPeriode
	for _, t := range tren {
		r.Omzet += t.Omzet
		r.Transaksi += t.Transaksi
		r.Pendapatan += t.Pendapatan
		r.HPP += t.HPP
		r.LabaKotor += t.LabaKotor
		r.LabaBersih += t.LabaBersih
	}
	if r.Transaksi > 0 {
		r.RataRata = r.Omzet / r.Transaksi
	}
	return r
}

// perTanggal menjumlahkan kolom angka hasil query yang kolom pertamanya DATE(...).
func (s *LaporanService) perTanggal(q string, args ...any) (map[string][]float64, error) {
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	out := map[string][]float64{}
	for rows.Next() {
		var tgl time.Time
		vals := make([]float64, len(cols)-1)
		dst := []any{&tgl}
		for i := range vals {
			dst = append(dst, &vals[i])
		}
		if err := rows.Scan(dst...); err != nil {
			return nil, err
		}
		k := tgl.Format(dateLayout)
		if ada := out[k]; ada != nil {
			for i := range vals {
				ada[i] += vals[i]
			}
		} else {
			out[k] = vals
		}
	}
	return out, rows.Err()
}

func (s *LaporanService) tren(kid int64, from, to time.Time) ([]dto.TrenHari, error) {
	omzet, err := s.perTanggal(`SELECT DATE(sh.buka_at), SUM(b.jumlah) FROM pembayaran b JOIN pesanan p ON p.id=b.pesanan_id
		JOIN shift sh ON sh.id=b.shift_id WHERE p.klien_id=? AND p.status='lunas' AND sh.buka_at>=? AND sh.buka_at<? GROUP BY 1`, kid, from, to)
	if err != nil {
		return nil, err
	}
	trx, err := s.perTanggal(`SELECT DATE(sh.buka_at), COUNT(*) FROM pesanan p JOIN shift sh ON sh.id=p.shift_id
		WHERE p.klien_id=? AND p.status='lunas' AND sh.buka_at>=? AND sh.buka_at<? GROUP BY 1`, kid, from, to)
	if err != nil {
		return nil, err
	}
	// pendapatan per hari per tenant: milik sendiri penuh, tenant lain hanya komisinya (rumus Laba Rugi)
	rows, err := s.db.Query(`SELECT DATE(sh.buka_at), COALESCE(t.komisi_persen,100), COALESCE(t.milik_sendiri,1), SUM(i.harga*i.qty),
		COALESCE(SUM(ROUND(p.diskon*(i.harga*i.qty)/ps.sub)),0) `+itemLunasSQL+`
		JOIN (SELECT pesanan_id, SUM(harga*qty) AS sub FROM pesanan_item WHERE klien_id=? GROUP BY pesanan_id) ps ON ps.pesanan_id=p.id AND ps.sub>0
		GROUP BY DATE(sh.buka_at), i.tenant_id`, kid, from, to, kid)
	if err != nil {
		return nil, err
	}
	pendapatan := map[string]int{}
	for rows.Next() {
		var tgl time.Time
		var persen float64
		var sendiri bool
		var kotor, diskon int
		if err := rows.Scan(&tgl, &persen, &sendiri, &kotor, &diskon); err != nil {
			rows.Close()
			return nil, err
		}
		komisi, _ := hitungSetoran(kotor, diskon, persen) // sama persis dengan LabaRugi, termasuk tenant milik sendiri
		pendapatan[tgl.Format(dateLayout)] += komisi
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	hpp, err := s.perTanggal(`SELECT DATE(sh.buka_at), SUM(i.qty*(i.modal+(SELECT COALESCE(SUM(ib.jumlah*ib.biaya),0) FROM item_bahan ib WHERE ib.item_id=i.id))) `+
		itemLunasSQL+` WHERE `+milikSendiriSQL+` GROUP BY 1`, kid, from, to)
	if err != nil {
		return nil, err
	}
	susut, err := s.perTanggal(`SELECT DATE(at), -SUM(jumlah*biaya) FROM bahan_log
		WHERE klien_id=? AND jenis='opname' AND biaya IS NOT NULL AND at>=? AND at<? GROUP BY 1`, kid, from, to)
	if err != nil {
		return nil, err
	}
	biaya, err := s.perTanggal(`SELECT DATE(sh.buka_at), SUM(e.jumlah) FROM pengeluaran e JOIN shift sh ON sh.id=e.shift_id
		WHERE e.klien_id=? AND e.kategori NOT IN ('bahan','refund') AND sh.buka_at>=? AND sh.buka_at<? GROUP BY 1`, kid, from, to)
	if err != nil {
		return nil, err
	}
	gaji, err := s.perTanggal(`SELECT DATE(dibayar_at), SUM(pokok+tunjangan+bonus-potongan) FROM gaji_slip
		WHERE klien_id=? AND status='dibayar' AND dibayar_at>=? AND dibayar_at<? GROUP BY 1`, kid, from, to)
	if err != nil {
		return nil, err
	}

	nilai := func(m map[string][]float64, k string) float64 {
		if v := m[k]; v != nil {
			return v[0]
		}
		return 0
	}
	out := []dto.TrenHari{}
	for d := from; d.Before(to); d = d.AddDate(0, 0, 1) {
		k := d.Format(dateLayout)
		t := dto.TrenHari{
			Tanggal: k, Omzet: bulat(nilai(omzet, k)), Transaksi: bulat(nilai(trx, k)), Pendapatan: pendapatan[k],
			HPP:   bulat(nilai(hpp, k) + nilai(susut, k)),
			Biaya: bulat(nilai(biaya, k) + nilai(gaji, k)),
		}
		t.LabaKotor = t.Pendapatan - t.HPP
		t.LabaBersih = t.LabaKotor - t.Biaya
		out = append(out, t)
	}
	return out, nil
}

func (s *LaporanService) jamRamai(out *dto.AnalitikResponse, kid int64, from, to time.Time) error {
	rows, err := s.db.Query(`SELECT WEEKDAY(p.lunas_at), HOUR(p.lunas_at), COUNT(*),
		COALESCE(SUM((SELECT SUM(b.jumlah) FROM pembayaran b WHERE b.pesanan_id=p.id)),0)
		FROM pesanan p JOIN shift sh ON sh.id=p.shift_id AND sh.buka_at>=? AND sh.buka_at<?
		WHERE p.klien_id=? AND p.status='lunas' GROUP BY 1,2`, from, to, kid)
	if err != nil {
		return err
	}
	defer rows.Close()
	out.JamRamai = []dto.JamRamai{}
	for rows.Next() {
		var j dto.JamRamai
		if err := rows.Scan(&j.Hari, &j.Jam, &j.Bill, &j.Omzet); err != nil {
			return err
		}
		out.JamRamai = append(out.JamRamai, j)
	}
	return rows.Err()
}

func (s *LaporanService) kategori(out *dto.AnalitikResponse, kid int64, from, to time.Time) error {
	rows, err := s.db.Query(`SELECT COALESCE(k.nama,'Tanpa kategori'), SUM(i.qty), SUM(i.harga*i.qty) `+itemLunasSQL+`
		LEFT JOIN menu mn ON mn.id=i.menu_id LEFT JOIN kategori k ON k.id=mn.kategori_id
		GROUP BY k.id, k.nama ORDER BY 3 DESC`, kid, from, to)
	if err != nil {
		return err
	}
	defer rows.Close()
	out.Kategori = []dto.KategoriTotal{}
	for rows.Next() {
		var k dto.KategoriTotal
		if err := rows.Scan(&k.Nama, &k.Qty, &k.Total); err != nil {
			return err
		}
		out.Kategori = append(out.Kategori, k)
	}
	return rows.Err()
}

func (s *LaporanService) perTenant(out *dto.AnalitikResponse, kid int64, from, to time.Time) error {
	rows, err := s.db.Query(`SELECT COALESCE(t.nama,'Tanpa tenant'), COUNT(DISTINCT i.pesanan_id), SUM(i.qty), SUM(i.harga*i.qty) `+itemLunasSQL+`
		GROUP BY i.tenant_id, t.nama ORDER BY 4 DESC`, kid, from, to)
	if err != nil {
		return err
	}
	defer rows.Close()
	out.Tenant = []dto.TenantTotal{}
	for rows.Next() {
		var t dto.TenantTotal
		if err := rows.Scan(&t.Nama, &t.Bill, &t.Qty, &t.Total); err != nil {
			return err
		}
		out.Tenant = append(out.Tenant, t)
	}
	return rows.Err()
}

// perUser menjalankan query yang kolom pertamanya user id, sisanya angka.
func (s *LaporanService) perUser(q string, args ...any) (map[int64][]int, error) {
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	out := map[int64][]int{}
	for rows.Next() {
		var id int64
		vals := make([]float64, len(cols)-1)
		dst := []any{&id}
		for i := range vals {
			dst = append(dst, &vals[i])
		}
		if err := rows.Scan(dst...); err != nil {
			return nil, err
		}
		n := make([]int, len(vals))
		for i, v := range vals {
			n[i] = bulat(v)
		}
		out[id] = n
	}
	return out, rows.Err()
}

func (s *LaporanService) karyawan(out *dto.AnalitikResponse, kid int64, from, to time.Time) error {
	lunasDalam := `FROM pesanan p JOIN shift sh ON sh.id=p.shift_id AND sh.buka_at>=? AND sh.buka_at<? WHERE p.klien_id=? AND p.status='lunas'`
	dibuat, err := s.perUser(`SELECT p.dibuat_oleh, COUNT(*), COALESCE(SUM((SELECT SUM(harga*qty) FROM pesanan_item WHERE pesanan_id=p.id) - p.diskon),0) `+
		lunasDalam+` GROUP BY 1`, from, to, kid)
	if err != nil {
		return err
	}
	dibayar, err := s.perUser(`SELECT b.diterima_oleh, COUNT(DISTINCT b.pesanan_id), SUM(b.jumlah) FROM pembayaran b
		JOIN pesanan p ON p.id=b.pesanan_id AND p.status='lunas' JOIN shift sh ON sh.id=b.shift_id
		WHERE b.klien_id=? AND b.diterima_oleh IS NOT NULL AND sh.buka_at>=? AND sh.buka_at<? GROUP BY 1`, kid, from, to)
	if err != nil {
		return err
	}
	diskon, err := s.perUser(`SELECT k, SUM(d) FROM (SELECT (SELECT b.diterima_oleh FROM pembayaran b WHERE b.pesanan_id=p.id LIMIT 1) AS k, p.diskon AS d `+
		lunasDalam+` AND p.diskon>0) x WHERE k IS NOT NULL GROUP BY k`, from, to, kid)
	if err != nil {
		return err
	}
	batal, err := s.perUser(`SELECT oleh, SUM(qty), SUM(harga*qty) FROM pembatalan WHERE klien_id=? AND at>=? AND at<? GROUP BY 1`, kid, from, to)
	if err != nil {
		return err
	}
	shift, err := s.perUser(`SELECT kasir_id, COUNT(*), COALESCE(SUM(kas_hitung-kas_sistem),0) FROM shift
		WHERE klien_id=? AND tutup_at IS NOT NULL AND buka_at>=? AND buka_at<? GROUP BY 1`, kid, from, to)
	if err != nil {
		return err
	}

	rows, err := s.db.Query("SELECT id,nama,role FROM users WHERE klien_id=? AND role<>'tenant' ORDER BY nama", kid)
	if err != nil {
		return err
	}
	defer rows.Close()
	ambil := func(m map[int64][]int, id int64, i int) int {
		if v := m[id]; v != nil {
			return v[i]
		}
		return 0
	}
	out.Karyawan = []dto.KinerjaKaryawan{}
	for rows.Next() {
		var k dto.KinerjaKaryawan
		if err := rows.Scan(&k.UserID, &k.Nama, &k.Role); err != nil {
			return err
		}
		id := k.UserID
		k.BillDibuat, k.NilaiDibuat = ambil(dibuat, id, 0), ambil(dibuat, id, 1)
		k.BillDibayar, k.NilaiDibayar = ambil(dibayar, id, 0), ambil(dibayar, id, 1)
		k.Diskon = ambil(diskon, id, 0)
		k.BatalQty, k.BatalNilai = ambil(batal, id, 0), ambil(batal, id, 1)
		k.Shift, k.SelisihKas = ambil(shift, id, 0), ambil(shift, id, 1)
		if k.BillDibuat+k.BillDibayar+k.BatalQty+k.Shift == 0 {
			continue // tidak aktif pada rentang ini
		}
		out.Karyawan = append(out.Karyawan, k)
	}
	return rows.Err()
}

type kasbonEntri struct {
	jumlah int
	at     time.Time
}

// umurKasbon: cicilan menutup kasbon tertua dulu (FIFO); kembalikan sisa dan tanggal kasbon tertua yang belum tertutup.
func umurKasbon(entri []kasbonEntri, dibayar int) (sisa int, tertua *time.Time) {
	for _, e := range entri {
		if dibayar >= e.jumlah {
			dibayar -= e.jumlah
			continue
		}
		if tertua == nil {
			at := e.at
			tertua = &at
		}
		sisa += e.jumlah - dibayar
		dibayar = 0
	}
	return sisa, tertua
}

// kasbonUmur tidak bergantung rentang tanggal: posisi utang per hari ini.
func (s *LaporanService) kasbonUmur(out *dto.AnalitikResponse, kid int64, _, _ time.Time) error {
	rows, err := s.db.Query(`SELECT pelanggan_id, jumlah, at FROM pembayaran
		WHERE klien_id=? AND metode='kasbon' AND pelanggan_id IS NOT NULL ORDER BY at, id`, kid)
	if err != nil {
		return err
	}
	entri := map[int64][]kasbonEntri{}
	for rows.Next() {
		var id int64
		var e kasbonEntri
		if err := rows.Scan(&id, &e.jumlah, &e.at); err != nil {
			rows.Close()
			return err
		}
		entri[id] = append(entri[id], e)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	bayar, err := s.perUser("SELECT pelanggan_id, SUM(jumlah) FROM kasbon_bayar WHERE klien_id=? GROUP BY 1", kid)
	if err != nil {
		return err
	}
	rows, err = s.db.Query("SELECT id,nama,telp FROM pelanggan WHERE klien_id=?", kid)
	if err != nil {
		return err
	}
	defer rows.Close()
	out.Kasbon = []dto.KasbonUmur{}
	now := time.Now()
	for rows.Next() {
		var k dto.KasbonUmur
		if err := rows.Scan(&k.ID, &k.Nama, &k.Telp); err != nil {
			return err
		}
		dibayar := 0
		if v := bayar[k.ID]; v != nil {
			dibayar = v[0]
		}
		k.Sisa, k.Tertua = umurKasbon(entri[k.ID], dibayar)
		if k.Sisa <= 0 {
			continue
		}
		if k.Tertua != nil {
			k.UmurHari = int(now.Sub(*k.Tertua).Hours() / 24)
		}
		out.Kasbon = append(out.Kasbon, k)
	}
	sort.Slice(out.Kasbon, func(a, b int) bool { return out.Kasbon[a].UmurHari > out.Kasbon[b].UmurHari })
	return rows.Err()
}

func (s *LaporanService) stokNilai(out *dto.AnalitikResponse, kid int64, from, to time.Time) error {
	susut, err := s.perUser(`SELECT bahan_id, -SUM(jumlah), -SUM(jumlah*biaya) FROM bahan_log
		WHERE klien_id=? AND jenis='opname' AND biaya IS NOT NULL AND at>=? AND at<? GROUP BY 1`, kid, from, to)
	if err != nil {
		return err
	}
	rows, err := s.db.Query("SELECT id,nama,satuan,stok,harga_rata FROM bahan WHERE klien_id=? AND aktif=1", kid)
	if err != nil {
		return err
	}
	defer rows.Close()
	out.Stok = []dto.StokNilai{}
	for rows.Next() {
		var b dto.StokNilai
		var rata float64
		if err := rows.Scan(&b.ID, &b.Nama, &b.Satuan, &b.Stok, &rata); err != nil {
			return err
		}
		b.Nilai = bulat(float64(max(b.Stok, 0)) * rata)
		if v := susut[b.ID]; v != nil {
			b.SusutQty, b.SusutNilai = v[0], v[1]
		}
		out.Stok = append(out.Stok, b)
	}
	sort.Slice(out.Stok, func(a, b int) bool { return out.Stok[a].Nilai > out.Stok[b].Nilai })
	return rows.Err()
}
