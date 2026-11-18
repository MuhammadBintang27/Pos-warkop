package services

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"warkop/internal/dto"
	"warkop/internal/types"
	"warkop/internal/utils"
)

const (
	logPakai    = "pakai"
	logKembali  = "kembali"
	logRestock  = "restock"
	logOpname   = "opname"
	statusOK    = "ok"
	statusTipis = "menipis"
	statusHabis = "habis"
)

type BahanService struct {
	db  *sql.DB
	hub *EventHub
}

func NewBahanService(db *sql.DB, hub *EventHub) *BahanService {
	return &BahanService{db: db, hub: hub}
}

func statusBahan(stok, minimum int) string {
	switch {
	case stok <= 0:
		return statusHabis
	case minimum > 0 && stok <= minimum:
		return statusTipis
	}
	return statusOK
}

func catatLog(tx *sql.Tx, kid, bahanID int64, jenis string, jumlah, stokAkhir int, harga *int, catatan string, oleh *int64) error {
	_, err := tx.Exec(`INSERT INTO bahan_log (klien_id,bahan_id,jenis,jumlah,stok_akhir,harga_beli,catatan,oleh,at) VALUES (?,?,?,?,?,?,?,?,?)`,
		kid, bahanID, jenis, jumlah, stokAkhir, harga, catatan, oleh, time.Now())
	return err
}

func (s *BahanService) List(kid int64) ([]dto.BahanResponse, error) {
	rows, err := s.db.Query("SELECT id,nama,satuan,stok,minimum,harga_rata FROM bahan WHERE klien_id=? AND aktif=1 ORDER BY nama", kid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []dto.BahanResponse{}
	for rows.Next() {
		var b dto.BahanResponse
		if err := rows.Scan(&b.ID, &b.Nama, &b.Satuan, &b.Stok, &b.Minimum, &b.HargaRata); err != nil {
			return nil, err
		}
		b.Status = statusBahan(b.Stok, b.Minimum)
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *BahanService) Create(kid int64, req dto.BahanRequest) error {
	return utils.WithTx(s.db, func(tx *sql.Tx) error {
		res, err := tx.Exec("INSERT INTO bahan (klien_id,nama,satuan,stok,minimum,harga_rata) VALUES (?,?,?,?,?,?)",
			kid, strings.TrimSpace(req.Nama), strings.TrimSpace(req.Satuan), req.Stok, req.Minimum, req.HargaRata)
		if err != nil {
			return err
		}
		if req.Stok == 0 {
			return nil
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		return catatLog(tx, kid, id, logOpname, req.Stok, req.Stok, nil, "Stok awal", nil)
	})
}

func (s *BahanService) Update(kid, id int64, req dto.BahanRequest) error {
	if _, err := s.db.Exec("UPDATE bahan SET nama=?, satuan=?, minimum=?, harga_rata=? WHERE id=? AND klien_id=? AND aktif=1",
		strings.TrimSpace(req.Nama), strings.TrimSpace(req.Satuan), req.Minimum, req.HargaRata, id, kid); err != nil {
		return err
	}
	var ada int
	if s.db.QueryRow("SELECT 1 FROM bahan WHERE id=? AND klien_id=? AND aktif=1", id, kid).Scan(&ada) != nil {
		return types.NotFound("Bahan")
	}
	return nil
}

// Delete menonaktifkan bahan; ditolak selama masih dipakai di resep menu.
func (s *BahanService) Delete(kid, id int64) error {
	var dipakai int
	if err := s.db.QueryRow(`SELECT
		(SELECT COUNT(*) FROM resep r JOIN menu m ON m.id=r.menu_id AND m.aktif=1 WHERE r.klien_id=? AND r.bahan_id=?) +
		(SELECT COUNT(*) FROM resep_opsi WHERE klien_id=? AND bahan_id=?)`, kid, id, kid, id).Scan(&dipakai); err != nil {
		return err
	}
	if dipakai > 0 {
		return types.BadRequest("Bahan masih dipakai di resep menu. Lepaskan dari resep dulu")
	}
	res, err := s.db.Exec("UPDATE bahan SET aktif=0 WHERE id=? AND klien_id=? AND aktif=1", id, kid)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return types.NotFound("Bahan")
	}
	return nil
}

// ubahStok mengunci baris bahan, menerapkan perubahan, dan mencatatnya dalam satu transaksi.
func (s *BahanService) ubahStok(kid, id int64, oleh int64, jenis string, hitung func(stok int) (delta int, err error), harga *int, catatan string) error {
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		var stok int
		if tx.QueryRow("SELECT stok FROM bahan WHERE id=? AND klien_id=? AND aktif=1 FOR UPDATE", id, kid).Scan(&stok) != nil {
			return types.NotFound("Bahan")
		}
		delta, err := hitung(stok)
		if err != nil {
			return err
		}
		if _, err := tx.Exec("UPDATE bahan SET stok=stok+? WHERE id=? AND klien_id=?", delta, id, kid); err != nil {
			return err
		}
		return catatLog(tx, kid, id, jenis, delta, stok+delta, harga, strings.TrimSpace(catatan), &oleh)
	})
	if err == nil {
		s.hub.Publish(kid, "menu") // ketersediaan menu bergantung pada stok bahan
	}
	return err
}

// rataBaru menghitung harga rata-rata bergerak per satuan setelah restock berharga. Stok minus tidak ikut menimbang.
func rataBaru(stok int, rata float64, jumlah, totalHarga int) float64 {
	if totalHarga <= 0 || jumlah <= 0 {
		return rata
	}
	dasar := max(stok, 0)
	return (float64(dasar)*rata + float64(totalHarga)) / float64(dasar+jumlah)
}

func (s *BahanService) Restock(kid, oleh, id int64, req dto.RestockRequest) error {
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		var stok int
		var rata float64
		if tx.QueryRow("SELECT stok,harga_rata FROM bahan WHERE id=? AND klien_id=? AND aktif=1 FOR UPDATE", id, kid).Scan(&stok, &rata) != nil {
			return types.NotFound("Bahan")
		}
		if _, err := tx.Exec("UPDATE bahan SET stok=stok+?, harga_rata=? WHERE id=? AND klien_id=?", req.Jumlah, rataBaru(stok, rata, req.Jumlah, req.HargaBeli), id, kid); err != nil {
			return err
		}
		var harga *int
		if req.HargaBeli > 0 {
			harga = &req.HargaBeli
		}
		return catatLog(tx, kid, id, logRestock, req.Jumlah, stok+req.Jumlah, harga, strings.TrimSpace(req.Catatan), &oleh)
	})
	if err == nil {
		s.hub.Publish(kid, "menu")
	}
	return err
}

func (s *BahanService) Opname(kid, oleh, id int64, req dto.OpnameRequest) error {
	return s.ubahStok(kid, id, oleh, logOpname, func(stok int) (int, error) { return req.StokAktual - stok, nil }, nil, req.Catatan)
}

func (s *BahanService) Riwayat(kid, id int64) ([]dto.BahanLogResponse, error) {
	rows, err := s.db.Query(`SELECT l.id,l.jenis,l.jumlah,l.stok_akhir,l.harga_beli,l.catatan,COALESCE(u.nama,''),l.at
		FROM bahan_log l LEFT JOIN users u ON u.id=l.oleh WHERE l.klien_id=? AND l.bahan_id=? ORDER BY l.id DESC LIMIT 100`, kid, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []dto.BahanLogResponse{}
	for rows.Next() {
		var l dto.BahanLogResponse
		if err := rows.Scan(&l.ID, &l.Jenis, &l.Jumlah, &l.StokAkhir, &l.HargaBeli, &l.Catatan, &l.Oleh, &l.At); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *BahanService) Pemakaian(kid int64, dari, sampai string) ([]dto.BahanPemakaian, error) {
	from, to, err := parseRange(dari, sampai, time.Now())
	if err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`SELECT b.id,b.nama,b.satuan,b.stok,
		COALESCE(SUM(CASE WHEN l.jenis IN ('pakai','kembali') THEN -l.jumlah END),0),
		COALESCE(SUM(CASE WHEN l.jenis='restock' THEN l.jumlah END),0),
		COALESCE(SUM(CASE WHEN l.jenis='opname' THEN l.jumlah END),0)
		FROM bahan b LEFT JOIN bahan_log l ON l.bahan_id=b.id AND l.at>=? AND l.at<?
		WHERE b.klien_id=? AND b.aktif=1 GROUP BY b.id,b.nama,b.satuan,b.stok ORDER BY b.nama`, from, to, kid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []dto.BahanPemakaian{}
	for rows.Next() {
		var p dto.BahanPemakaian
		if err := rows.Scan(&p.ID, &p.Nama, &p.Satuan, &p.StokSekarang, &p.Dipakai, &p.Restock, &p.SelisihOpnam); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ---- dipakai pesanan ----

type bahanPakai struct {
	bahanID int64
	jumlah  int
}

// resepUntuk menghitung takaran per porsi: resep dasar menu ditambah tambahan opsi yang dipilih. Takaran nol atau negatif dibuang.
func resepUntuk(tx *sql.Tx, kid, menuID int64, opsiIDs []int64) ([]bahanPakai, error) {
	total := map[int64]int{}
	baca := func(q string, args ...any) error {
		rows, err := tx.Query(q, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id int64
			var n int
			if err := rows.Scan(&id, &n); err != nil {
				return err
			}
			total[id] += n
		}
		return rows.Err()
	}
	if err := baca("SELECT bahan_id,jumlah FROM resep WHERE klien_id=? AND menu_id=?", kid, menuID); err != nil {
		return nil, err
	}
	if len(opsiIDs) > 0 {
		args := []any{kid}
		for _, id := range opsiIDs {
			args = append(args, id)
		}
		if err := baca("SELECT bahan_id,jumlah FROM resep_opsi WHERE klien_id=? AND opsi_id IN ("+strings.TrimSuffix(strings.Repeat("?,", len(opsiIDs)), ",")+")", args...); err != nil {
			return nil, err
		}
	}
	var out []bahanPakai
	for id, n := range total {
		if n > 0 {
			out = append(out, bahanPakai{id, n})
		}
	}
	sort.Slice(out, func(a, b int) bool { return out[a].bahanID < out[b].bahanID }) // urutan tetap menghindari deadlock antar transaksi
	return out, nil
}

func simpanSnapshot(tx *sql.Tx, kid, itemID int64, bahan []bahanPakai) error {
	for _, b := range bahan {
		if _, err := tx.Exec("INSERT INTO item_bahan (klien_id,item_id,bahan_id,jumlah,biaya) SELECT ?,?,id,?,harga_rata FROM bahan WHERE id=? AND klien_id=?", kid, itemID, b.jumlah, b.bahanID, kid); err != nil {
			return err
		}
	}
	return nil
}

func snapshotItem(tx *sql.Tx, kid, itemID int64) ([]bahanPakai, error) {
	rows, err := tx.Query("SELECT bahan_id,jumlah FROM item_bahan WHERE klien_id=? AND item_id=? ORDER BY bahan_id", kid, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []bahanPakai
	for rows.Next() {
		var b bahanPakai
		if err := rows.Scan(&b.bahanID, &b.jumlah); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// pakaiBahan mengurangi stok bahan sebanyak takaran item (disalin saat dipesan) untuk porsi sebanyak itu.
// Porsi positif memakai bahan dan ditolak bila stok tidak cukup; porsi negatif mengembalikannya.
func pakaiBahan(tx *sql.Tx, kid, itemID int64, porsi int) error {
	if porsi == 0 {
		return nil
	}
	bahan, err := snapshotItem(tx, kid, itemID)
	if err != nil {
		return err
	}
	for _, b := range bahan {
		var stok int
		var nama, satuan string
		if err := tx.QueryRow("SELECT stok,nama,satuan FROM bahan WHERE id=? AND klien_id=? FOR UPDATE", b.bahanID, kid).Scan(&stok, &nama, &satuan); err != nil {
			return err
		}
		delta := -porsi * b.jumlah
		if stok+delta < 0 {
			return types.BadRequest(fmt.Sprintf("Bahan %s tidak cukup (sisa %d %s)", nama, stok, satuan))
		}
		if _, err := tx.Exec("UPDATE bahan SET stok=stok+? WHERE id=? AND klien_id=?", delta, b.bahanID, kid); err != nil {
			return err
		}
		jenis := logPakai
		if delta > 0 {
			jenis = logKembali
		}
		if err := catatLog(tx, kid, b.bahanID, jenis, delta, stok+delta, nil, "", nil); err != nil {
			return err
		}
	}
	return nil
}
