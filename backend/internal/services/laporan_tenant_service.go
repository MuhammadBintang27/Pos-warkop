package services

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
	"time"

	"warkop/internal/dto"
	"warkop/internal/types"
	"warkop/internal/utils"
)

type LaporanTenantService struct {
	db *sql.DB
}

type kunciHarian struct {
	tenantID int64
	tanggal  string
}

type jumlahJual struct {
	kotor, diskon, transaksi int
}

type closingRow struct {
	id                       int64
	transaksi, kotor, diskon int
	persen                   float64
	komisi, setor            int
	disetorAt, dibuatAt      *time.Time
	dibuatOleh, disetorOleh  string
}

const (
	statusBelum   = "belum"
	statusClosing = "closing"
	statusDisetor = "disetor"
	statusBerubah = "berubah"

	// Penjualan tenant per hari kerja (tanggal shift dibuka) dari bill lunas.
	// Diskon bill dibagi proporsional menurut nilai item tiap tenant.
	penjualanTenantHarianSQL = `SELECT COALESCE(i.tenant_id,0), DATE(sh.buka_at), SUM(i.harga*i.qty),
		COALESCE(SUM(ROUND(p.diskon*(i.harga*i.qty)/ps.sub)),0), COUNT(DISTINCT p.id)
		FROM pesanan_item i
		JOIN pesanan p ON p.id=i.pesanan_id AND p.klien_id=? AND p.status='lunas'
		JOIN shift sh ON sh.id=p.shift_id AND sh.buka_at>=? AND sh.buka_at<?
		JOIN (SELECT pesanan_id, SUM(harga*qty) AS sub FROM pesanan_item WHERE klien_id=? GROUP BY pesanan_id) ps ON ps.pesanan_id=p.id AND ps.sub>0
		WHERE i.closing_id IS NULL
		GROUP BY i.tenant_id, DATE(sh.buka_at)`

	// Hari kerja sebuah pembatalan: tanggal shift yang sedang berjalan saat itu, atau tanggal biasa bila tidak ada shift.
	hariKerjaBatalSQL = `COALESCE((SELECT DATE(s.buka_at) FROM shift s
		WHERE s.klien_id=b.klien_id AND s.buka_at<=b.at AND (s.tutup_at IS NULL OR s.tutup_at>=b.at) ORDER BY s.id DESC LIMIT 1), DATE(b.at))`
)

func NewLaporanTenantService(db *sql.DB) *LaporanTenantService {
	return &LaporanTenantService{db: db}
}

// statusPenutupan: "berubah" berarti ada transaksi yang masuk atau berubah setelah closing, sehingga angka closing sudah usang.
func statusPenutupan(ada, disetor bool, snapKotor, snapDiskon, kotor, diskon int) string {
	switch {
	case !ada:
		return statusBelum
	case snapKotor != kotor || snapDiskon != diskon:
		return statusBerubah
	case disetor:
		return statusDisetor
	}
	return statusClosing
}

func penjualanHarian(q querier, kid int64, from, to time.Time) (map[kunciHarian]jumlahJual, error) {
	rows, err := q.Query(penjualanTenantHarianSQL, kid, from, to, kid)
	if err != nil {
		return nil, err
	}
	out := map[kunciHarian]jumlahJual{}
	for rows.Next() {
		var id int64
		var tgl time.Time
		var j jumlahJual
		if err := rows.Scan(&id, &tgl, &j.kotor, &j.diskon, &j.transaksi); err != nil {
			rows.Close()
			return nil, err
		}
		out[kunciHarian{id, tgl.Format(dateLayout)}] = j
	}
	rows.Close()

	// bill belum dibayar yang sudah ditalangi pada closing tanggal itu: memakai nilai yang dibekukan saat closing
	rows, err = q.Query(`SELECT c.tenant_id, c.tanggal, SUM(cb.kotor), SUM(cb.diskon), COUNT(*)
		FROM closing_tenant_bill cb JOIN closing_tenant c ON c.id=cb.closing_id
		WHERE c.klien_id=? AND c.tanggal>=? AND c.tanggal<? GROUP BY c.tenant_id, c.tanggal`, kid, from.Format(dateLayout), to.Format(dateLayout))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var tgl time.Time
		var t jumlahJual
		if err := rows.Scan(&id, &tgl, &t.kotor, &t.diskon, &t.transaksi); err != nil {
			return nil, err
		}
		k := kunciHarian{id, tgl.Format(dateLayout)}
		j := out[k]
		j.kotor, j.diskon, j.transaksi = j.kotor+t.kotor, j.diskon+t.diskon, j.transaksi+t.transaksi
		out[k] = j
	}
	return out, rows.Err()
}

func (s *LaporanTenantService) Harian(kid int64, q dto.HarianQuery) (*dto.HarianResponse, error) {
	from, to, err := parseRange(q.Dari, q.Sampai, time.Now())
	if err != nil {
		return nil, err
	}
	jual, err := penjualanHarian(s.db, kid, from, to)
	if err != nil {
		return nil, err
	}

	type infoTenant struct {
		nama    string
		persen  float64
		sendiri bool
	}
	tenants := map[int64]infoTenant{}
	rows, err := s.db.Query("SELECT id,nama,komisi_persen,milik_sendiri FROM tenant WHERE klien_id=?", kid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int64
		var t infoTenant
		if err := rows.Scan(&id, &t.nama, &t.persen, &t.sendiri); err != nil {
			rows.Close()
			return nil, err
		}
		tenants[id] = t
	}
	rows.Close()

	closed := map[kunciHarian]closingRow{}
	rows, err = s.db.Query(`SELECT c.id,c.tenant_id,c.tanggal,c.transaksi,c.kotor,c.diskon,c.komisi_persen,c.komisi,c.setor,
		c.disetor_at,c.dibuat_at,uc.nama,COALESCE(us.nama,'')
		FROM closing_tenant c JOIN users uc ON uc.id=c.dibuat_oleh LEFT JOIN users us ON us.id=c.disetor_oleh
		WHERE c.klien_id=? AND c.tanggal>=? AND c.tanggal<?`, kid, from.Format(dateLayout), to.Format(dateLayout))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var c closingRow
		var tenantID int64
		var tgl time.Time
		if err := rows.Scan(&c.id, &tenantID, &tgl, &c.transaksi, &c.kotor, &c.diskon, &c.persen, &c.komisi, &c.setor,
			&c.disetorAt, &c.dibuatAt, &c.dibuatOleh, &c.disetorOleh); err != nil {
			rows.Close()
			return nil, err
		}
		closed[kunciHarian{tenantID, tgl.Format(dateLayout)}] = c
	}
	rows.Close()

	keys := map[kunciHarian]bool{}
	for k := range jual {
		keys[k] = true
	}
	for k := range closed {
		keys[k] = true
	}

	out := &dto.HarianResponse{Dari: from.Format(dateLayout), Sampai: to.AddDate(0, 0, -1).Format(dateLayout), Baris: []dto.HarianTenant{}}
	for k := range keys {
		t, ok := tenants[k.tenantID]
		if !ok {
			continue
		}
		j := jual[k]
		c, ada := closed[k]
		b := dto.HarianTenant{
			TenantID: k.tenantID, Nama: t.nama, Tanggal: k.tanggal, MilikSendiri: t.sendiri,
			Transaksi: j.transaksi, Kotor: j.kotor, Diskon: j.diskon, KomisiPersen: t.persen,
			Status: statusPenutupan(ada, c.disetorAt != nil, c.kotor, c.diskon, j.kotor, j.diskon),
		}
		if ada { // angka yang dibekukan saat closing yang dipakai untuk setoran
			b.ClosingID, b.DisetorAt, b.DisetorOleh = c.id, c.disetorAt, c.disetorOleh
			b.DibuatAt, b.DibuatOleh = c.dibuatAt, c.dibuatOleh
			b.Transaksi, b.Kotor, b.Diskon, b.KomisiPersen, b.Komisi, b.Setor = c.transaksi, c.kotor, c.diskon, c.persen, c.komisi, c.setor
		} else {
			b.Komisi, b.Setor = hitungSetoran(b.Kotor, b.Diskon, b.KomisiPersen)
		}
		out.Baris = append(out.Baris, b)
	}
	out.Tertunda = []dto.TertundaTenant{}
	rows, err = s.db.Query(`SELECT COALESCE(i.tenant_id,0), COUNT(DISTINCT p.id), SUM(i.harga*i.qty)
		FROM pesanan_item i JOIN pesanan p ON p.id=i.pesanan_id AND p.klien_id=? AND p.status='open'
		WHERE i.dikirim=1 AND i.closing_id IS NULL GROUP BY i.tenant_id`, kid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var t dto.TertundaTenant
		if err := rows.Scan(&t.TenantID, &t.Bill, &t.Nilai); err != nil {
			return nil, err
		}
		out.Tertunda = append(out.Tertunda, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// tenant yang hanya punya bill belum dibayar tetap mendapat baris di tanggal terakhir, supaya bisa di-closing dengan talangan
	ada := map[kunciHarian]bool{}
	for _, b := range out.Baris {
		ada[kunciHarian{b.TenantID, b.Tanggal}] = true
	}
	for _, t := range out.Tertunda {
		info, ok := tenants[t.TenantID]
		k := kunciHarian{t.TenantID, out.Sampai}
		if !ok || ada[k] {
			continue
		}
		ada[k] = true
		out.Baris = append(out.Baris, dto.HarianTenant{TenantID: t.TenantID, Nama: info.nama, Tanggal: out.Sampai, MilikSendiri: info.sendiri, KomisiPersen: info.persen, Status: statusBelum})
	}
	sort.Slice(out.Baris, func(a, b int) bool {
		if out.Baris[a].Tanggal != out.Baris[b].Tanggal {
			return out.Baris[a].Tanggal > out.Baris[b].Tanggal
		}
		return out.Baris[a].TenantID < out.Baris[b].TenantID
	})
	return out, nil
}

type talanganBill struct {
	pesananID     int64
	kotor, diskon int
}

// pilihTalangan memeriksa bill terbuka yang dipilih dan menghitung bagian tenant (item terkirim yang belum ditalangi) beserta bagian diskonnya.
func pilihTalangan(tx *sql.Tx, kid, tenantID int64, ids []int64) ([]talanganBill, error) {
	var out []talanganBill
	seen := map[int64]bool{}
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		var diskon int
		if tx.QueryRow("SELECT diskon FROM pesanan WHERE id=? AND klien_id=? AND status='open' FOR UPDATE", id, kid).Scan(&diskon) != nil {
			return nil, types.BadRequest(fmt.Sprintf("Bill #%d tidak ditemukan atau sudah dibayar", id))
		}
		var kotor, sub int
		if err := tx.QueryRow(`SELECT COALESCE(SUM(CASE WHEN tenant_id=? AND dikirim=1 AND closing_id IS NULL THEN harga*qty END),0), COALESCE(SUM(harga*qty),0)
			FROM pesanan_item WHERE klien_id=? AND pesanan_id=?`, tenantID, kid, id).Scan(&kotor, &sub); err != nil {
			return nil, err
		}
		if kotor == 0 {
			return nil, types.BadRequest(fmt.Sprintf("Bill #%d tidak punya item tenant ini yang sudah dikirim", id))
		}
		d := 0
		if sub > 0 {
			d = int(math.Round(float64(diskon) * float64(kotor) / float64(sub)))
		}
		out = append(out, talanganBill{id, kotor, d})
	}
	return out, nil
}

// BelumBayar mengembalikan bill terbuka yang bisa ditalangi pada closing tenant: item tenant yang sudah dikirim dan belum pernah ditalangi.
func (s *LaporanTenantService) BelumBayar(kid, tenantID int64) ([]dto.BelumBayarBill, error) {
	rows, err := s.db.Query(`SELECT p.id, COALESCE(m.nama,''), p.atas_nama, p.dibuat_at, p.diskon,
		(SELECT COALESCE(SUM(harga*qty),0) FROM pesanan_item WHERE pesanan_id=p.id)
		FROM pesanan p LEFT JOIN meja m ON m.id=p.meja_id
		WHERE p.klien_id=? AND p.status='open' AND EXISTS
			(SELECT 1 FROM pesanan_item i WHERE i.pesanan_id=p.id AND i.tenant_id=? AND i.dikirim=1 AND i.closing_id IS NULL)
		ORDER BY p.id`, kid, tenantID)
	if err != nil {
		return nil, err
	}
	type kepala struct {
		bill          dto.BelumBayarBill
		diskon, total int
	}
	var daftar []*kepala
	byID := map[int64]*kepala{}
	for rows.Next() {
		k := &kepala{}
		if err := rows.Scan(&k.bill.PesananID, &k.bill.MejaNama, &k.bill.AtasNama, &k.bill.DibuatAt, &k.diskon, &k.total); err != nil {
			rows.Close()
			return nil, err
		}
		k.bill.Items = []dto.BelumBayarItem{}
		daftar, byID[k.bill.PesananID] = append(daftar, k), k
	}
	rows.Close()

	out := []dto.BelumBayarBill{}
	if len(daftar) == 0 {
		return out, nil
	}
	rows, err = s.db.Query(`SELECT i.pesanan_id,i.nama,i.opsi,i.qty,i.harga FROM pesanan_item i JOIN pesanan p ON p.id=i.pesanan_id AND p.status='open'
		WHERE i.klien_id=? AND i.tenant_id=? AND i.dikirim=1 AND i.closing_id IS NULL ORDER BY i.id`, kid, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var pid int64
		var it dto.BelumBayarItem
		if err := rows.Scan(&pid, &it.Nama, &it.Opsi, &it.Qty, &it.Harga); err != nil {
			return nil, err
		}
		if k := byID[pid]; k != nil {
			k.bill.Items = append(k.bill.Items, it)
			k.bill.Kotor += it.Harga * it.Qty
		}
	}
	for _, k := range daftar {
		if k.total > 0 {
			k.bill.Diskon = int(math.Round(float64(k.diskon) * float64(k.bill.Kotor) / float64(k.total)))
		}
		out = append(out, k.bill)
	}
	return out, rows.Err()
}

// Closing membekukan penjualan (bill yang sudah lunas) satu tenant pada satu tanggal.
// Tenant tidak terikat shift kasir, jadi closing boleh kapan saja; transaksi yang masuk sesudahnya menandai status "berubah".
// Bill yang belum dibayar tidak menghalangi: nilainya baru masuk saat bill itu dibayar, atau ikut bila dipilih untuk ditalangi.
func (s *LaporanTenantService) Closing(actor *utils.JWTPayload, req dto.ClosingRequest) error {
	kid := actor.KlienID
	from, to, err := parseRange(req.Tanggal, req.Tanggal, time.Now())
	if err != nil {
		return err
	}
	if from.After(time.Now()) {
		return types.BadRequest("Closing hanya bisa untuk hari ini atau sebelumnya")
	}
	return utils.WithTx(s.db, func(tx *sql.Tx) error {
		var persen float64
		if tx.QueryRow("SELECT komisi_persen FROM tenant WHERE id=? AND klien_id=? FOR UPDATE", req.TenantID, kid).Scan(&persen) != nil {
			return types.NotFound("Tenant")
		}
		jual, err := penjualanHarian(tx, kid, from, to)
		if err != nil {
			return err
		}
		j := jual[kunciHarian{req.TenantID, from.Format(dateLayout)}]
		talangan, err := pilihTalangan(tx, kid, req.TenantID, req.PesananIDs)
		if err != nil {
			return err
		}
		for _, t := range talangan {
			j.kotor, j.diskon, j.transaksi = j.kotor+t.kotor, j.diskon+t.diskon, j.transaksi+1
		}
		if j.kotor == 0 {
			return types.BadRequest("Tidak ada penjualan tenant ini pada tanggal tersebut")
		}
		komisi, setor := hitungSetoran(j.kotor, j.diskon, persen)
		res, err := tx.Exec(`INSERT INTO closing_tenant (klien_id,tenant_id,tanggal,transaksi,kotor,diskon,komisi_persen,komisi,setor,dibuat_oleh,dibuat_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?)`, kid, req.TenantID, from.Format(dateLayout), j.transaksi, j.kotor, j.diskon, persen, komisi, setor, actor.UserID, time.Now())
		if utils.MySQLErrNo(err) == utils.MySQLDuplicate {
			return types.BadRequest("Tanggal ini sudah di-closing untuk tenant tersebut")
		}
		if err != nil {
			return err
		}
		closingID, err := res.LastInsertId()
		if err != nil {
			return err
		}
		for _, t := range talangan {
			if _, err := tx.Exec("INSERT INTO closing_tenant_bill (klien_id,closing_id,pesanan_id,kotor,diskon) VALUES (?,?,?,?,?)", kid, closingID, t.pesananID, t.kotor, t.diskon); err != nil {
				return err
			}
			if _, err := tx.Exec("UPDATE pesanan_item SET closing_id=? WHERE klien_id=? AND pesanan_id=? AND tenant_id=? AND dikirim=1 AND closing_id IS NULL", closingID, kid, t.pesananID, req.TenantID); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *LaporanTenantService) adaClosing(kid, id int64) bool {
	var x int
	return s.db.QueryRow("SELECT 1 FROM closing_tenant WHERE id=? AND klien_id=?", id, kid).Scan(&x) == nil
}

func (s *LaporanTenantService) Setor(actor *utils.JWTPayload, id int64) error {
	kid := actor.KlienID
	var sendiri bool
	if s.db.QueryRow("SELECT t.milik_sendiri FROM closing_tenant c JOIN tenant t ON t.id=c.tenant_id WHERE c.id=? AND c.klien_id=?", id, kid).Scan(&sendiri) != nil {
		return types.NotFound("Closing")
	}
	if sendiri {
		return types.BadRequest("Tenant milik sendiri tidak ada setoran")
	}
	res, err := s.db.Exec("UPDATE closing_tenant SET disetor_at=?, disetor_oleh=? WHERE id=? AND klien_id=? AND disetor_at IS NULL", time.Now(), actor.UserID, id, kid)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		if !s.adaClosing(kid, id) {
			return types.NotFound("Closing")
		}
		return types.BadRequest("Sudah ditandai disetor")
	}
	return nil
}

// BatalClosing menghapus closing yang belum disetor dan melepaskan bill talangannya kembali menjadi tertunda.
func (s *LaporanTenantService) BatalClosing(kid, id int64) error {
	return utils.WithTx(s.db, func(tx *sql.Tx) error {
		var disetor sql.NullTime
		if tx.QueryRow("SELECT disetor_at FROM closing_tenant WHERE id=? AND klien_id=? FOR UPDATE", id, kid).Scan(&disetor) != nil {
			return types.NotFound("Closing")
		}
		if disetor.Valid {
			return types.BadRequest("Sudah disetor, closing tidak bisa dibatalkan")
		}
		if _, err := tx.Exec("UPDATE pesanan_item SET closing_id=NULL WHERE klien_id=? AND closing_id=?", kid, id); err != nil {
			return err
		}
		_, err := tx.Exec("DELETE FROM closing_tenant WHERE id=? AND klien_id=?", id, kid)
		return err
	})
}

// Detail: penjualan per menu (qty dan nilai) beserta pembatalannya untuk satu tenant.
func (s *LaporanTenantService) Detail(kid int64, q dto.DetailTenantQuery) (*dto.DetailTenantResponse, error) {
	from, to, err := parseRange(q.Dari, q.Sampai, time.Now())
	if err != nil {
		return nil, err
	}
	out := &dto.DetailTenantResponse{Dari: from.Format(dateLayout), Sampai: to.AddDate(0, 0, -1).Format(dateLayout),
		Menu: []dto.DetailMenuTenant{}, Batal: []dto.BatalTenant{}, Talangan: []dto.TalanganBill{}}
	if s.db.QueryRow("SELECT nama FROM tenant WHERE id=? AND klien_id=?", q.TenantID, kid).Scan(&out.Tenant) != nil {
		return nil, types.NotFound("Tenant")
	}

	menu := map[string]*dto.DetailMenuTenant{}
	ambil := func(nama string) *dto.DetailMenuTenant {
		if m, ok := menu[nama]; ok {
			return m
		}
		m := &dto.DetailMenuTenant{Nama: nama}
		menu[nama] = m
		return m
	}

	rows, err := s.db.Query(`SELECT i.nama, SUM(i.qty), SUM(i.qty*i.harga) FROM pesanan_item i
		JOIN pesanan p ON p.id=i.pesanan_id AND p.klien_id=? AND p.status='lunas'
		JOIN shift sh ON sh.id=p.shift_id AND sh.buka_at>=? AND sh.buka_at<?
		WHERE i.tenant_id=? AND i.closing_id IS NULL GROUP BY i.nama`, kid, from, to, q.TenantID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var nama string
		var qty, nilai int
		if err := rows.Scan(&nama, &qty, &nilai); err != nil {
			rows.Close()
			return nil, err
		}
		m := ambil(nama)
		m.Qty, m.Nilai = qty, nilai
		out.Qty, out.Nilai = out.Qty+qty, out.Nilai+nilai
	}
	rows.Close()

	// item bill belum dibayar yang ikut closing (ditalangi) pada rentang ini
	rows, err = s.db.Query(`SELECT i.nama, SUM(i.qty), SUM(i.qty*i.harga) FROM pesanan_item i JOIN closing_tenant c ON c.id=i.closing_id
		WHERE c.klien_id=? AND c.tenant_id=? AND c.tanggal>=? AND c.tanggal<? GROUP BY i.nama`, kid, q.TenantID, from.Format(dateLayout), to.Format(dateLayout))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var nama string
		var qty, nilai int
		if err := rows.Scan(&nama, &qty, &nilai); err != nil {
			rows.Close()
			return nil, err
		}
		m := ambil(nama)
		m.Qty, m.Nilai = m.Qty+qty, m.Nilai+nilai
		out.Qty, out.Nilai = out.Qty+qty, out.Nilai+nilai
	}
	rows.Close()

	rows, err = s.db.Query(`SELECT cb.pesanan_id, COALESCE(m.nama,''), p.atas_nama, cb.kotor
		FROM closing_tenant_bill cb JOIN closing_tenant c ON c.id=cb.closing_id JOIN pesanan p ON p.id=cb.pesanan_id LEFT JOIN meja m ON m.id=p.meja_id
		WHERE c.klien_id=? AND c.tenant_id=? AND c.tanggal>=? AND c.tanggal<? ORDER BY cb.id`, kid, q.TenantID, from.Format(dateLayout), to.Format(dateLayout))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var t dto.TalanganBill
		if err := rows.Scan(&t.PesananID, &t.MejaNama, &t.AtasNama, &t.Kotor); err != nil {
			rows.Close()
			return nil, err
		}
		out.Talangan = append(out.Talangan, t)
	}
	rows.Close()

	rows, err = s.db.Query(`SELECT b.nama, SUM(b.qty), SUM(b.qty*b.harga) FROM pembatalan b
		WHERE b.klien_id=? AND b.tenant_id=? AND `+hariKerjaBatalSQL+` >= ? AND `+hariKerjaBatalSQL+` < ? GROUP BY b.nama`,
		kid, q.TenantID, from.Format(dateLayout), to.Format(dateLayout))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var nama string
		var qty, nilai int
		if err := rows.Scan(&nama, &qty, &nilai); err != nil {
			rows.Close()
			return nil, err
		}
		m := ambil(nama)
		m.BatalQty, m.BatalNilai = qty, nilai
		out.BatalQty, out.BatalNilai = out.BatalQty+qty, out.BatalNilai+nilai
	}
	rows.Close()

	for _, m := range menu {
		out.Menu = append(out.Menu, *m)
	}
	sort.Slice(out.Menu, func(a, b int) bool {
		if out.Menu[a].Qty != out.Menu[b].Qty {
			return out.Menu[a].Qty > out.Menu[b].Qty
		}
		return out.Menu[a].Nama < out.Menu[b].Nama
	})

	rows, err = s.db.Query(`SELECT b.pesanan_id,b.nama,b.opsi,b.qty,b.qty*b.harga,b.alasan,b.sudah_dikirim,b.sudah_dibuat,u.nama,b.at
		FROM pembatalan b JOIN users u ON u.id=b.oleh
		WHERE b.klien_id=? AND b.tenant_id=? AND `+hariKerjaBatalSQL+` >= ? AND `+hariKerjaBatalSQL+` < ? ORDER BY b.at DESC LIMIT 500`,
		kid, q.TenantID, from.Format(dateLayout), to.Format(dateLayout))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var b dto.BatalTenant
		if err := rows.Scan(&b.PesananID, &b.Nama, &b.Opsi, &b.Qty, &b.Nilai, &b.Alasan, &b.SudahDikirim, &b.SudahDibuat, &b.Oleh, &b.At); err != nil {
			return nil, err
		}
		out.Batal = append(out.Batal, b)
	}
	return out, rows.Err()
}
