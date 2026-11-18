package services

import (
	"database/sql"
	"strings"
	"time"

	"warkop/internal/dto"
	"warkop/internal/types"
	"warkop/internal/utils"
)

type KaryawanService struct {
	db *sql.DB
}

func NewKaryawanService(db *sql.DB) *KaryawanService {
	return &KaryawanService{db: db}
}

const saldoBonSQL = `COALESCE((SELECT SUM(jumlah) FROM karyawan_bon WHERE klien_id=k.klien_id AND karyawan_id=k.id),0)`

func (s *KaryawanService) List(kid int64) ([]dto.KaryawanResponse, error) {
	rows, err := s.db.Query(`SELECT k.id,k.nama,k.jabatan,k.jenis,k.gaji,`+saldoBonSQL+` FROM karyawan k WHERE k.klien_id=? AND k.aktif=1 ORDER BY k.nama`, kid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []dto.KaryawanResponse{}
	for rows.Next() {
		var k dto.KaryawanResponse
		if err := rows.Scan(&k.ID, &k.Nama, &k.Jabatan, &k.Jenis, &k.Gaji, &k.SaldoBon); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

func (s *KaryawanService) Create(kid int64, req dto.KaryawanRequest) error {
	_, err := s.db.Exec("INSERT INTO karyawan (klien_id,nama,jabatan,jenis,gaji) VALUES (?,?,?,?,?)",
		kid, strings.TrimSpace(req.Nama), strings.TrimSpace(req.Jabatan), req.Jenis, req.Gaji)
	return err
}

func (s *KaryawanService) Update(kid, id int64, req dto.KaryawanRequest) error {
	if _, err := s.db.Exec("UPDATE karyawan SET nama=?, jabatan=?, jenis=?, gaji=? WHERE id=? AND klien_id=? AND aktif=1",
		strings.TrimSpace(req.Nama), strings.TrimSpace(req.Jabatan), req.Jenis, req.Gaji, id, kid); err != nil {
		return err
	}
	return s.pastikan(kid, id)
}

func (s *KaryawanService) pastikan(kid, id int64) error {
	var ada int
	if s.db.QueryRow("SELECT 1 FROM karyawan WHERE id=? AND klien_id=? AND aktif=1", id, kid).Scan(&ada) != nil {
		return types.NotFound("Karyawan")
	}
	return nil
}

// Delete menonaktifkan karyawan; slip dan bon lamanya tetap tersimpan. Ditolak bila masih punya bon.
func (s *KaryawanService) Delete(kid, id int64) error {
	var saldo int
	if err := s.db.QueryRow("SELECT COALESCE(SUM(jumlah),0) FROM karyawan_bon WHERE klien_id=? AND karyawan_id=?", kid, id).Scan(&saldo); err != nil {
		return err
	}
	if saldo > 0 {
		return types.BadRequest("Karyawan masih punya bon yang belum lunas")
	}
	res, err := s.db.Exec("UPDATE karyawan SET aktif=0 WHERE id=? AND klien_id=? AND aktif=1", id, kid)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return types.NotFound("Karyawan")
	}
	return nil
}

func (s *KaryawanService) ListBon(kid, id int64) ([]dto.BonResponse, error) {
	rows, err := s.db.Query("SELECT id,tanggal,jumlah,keterangan,slip_id FROM karyawan_bon WHERE klien_id=? AND karyawan_id=? ORDER BY tanggal DESC, id DESC LIMIT 200", kid, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []dto.BonResponse{}
	for rows.Next() {
		var b dto.BonResponse
		var tgl time.Time
		if err := rows.Scan(&b.ID, &tgl, &b.Jumlah, &b.Keterangan, &b.SlipID); err != nil {
			return nil, err
		}
		b.Tanggal = tgl.Format(dateLayout)
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *KaryawanService) TambahBon(kid, id int64, req dto.BonRequest) error {
	if err := s.pastikan(kid, id); err != nil {
		return err
	}
	tgl := time.Now().Format(dateLayout)
	if req.Tanggal != "" {
		if _, err := time.Parse(dateLayout, req.Tanggal); err != nil {
			return types.BadRequest("Format tanggal harus YYYY-MM-DD")
		}
		tgl = req.Tanggal
	}
	_, err := s.db.Exec("INSERT INTO karyawan_bon (klien_id,karyawan_id,tanggal,jumlah,keterangan) VALUES (?,?,?,?,?)",
		kid, id, tgl, req.Jumlah, strings.TrimSpace(req.Keterangan))
	return err
}

// hitungSlip: gaji pokok bulanan tetap, harian dikali hari kerja. Netto tidak boleh negatif.
func hitungSlip(jenis string, gaji, hari, tunjangan, bonus, potongan, potonganBon int) (pokok, netto int, err error) {
	if jenis == "harian" {
		if hari <= 0 {
			return 0, 0, types.BadRequest("Isi jumlah hari kerja untuk karyawan harian")
		}
		pokok = gaji * hari
	} else {
		pokok = gaji
	}
	netto = pokok + tunjangan + bonus - potongan - potonganBon
	if netto < 0 {
		return 0, 0, types.BadRequest("Potongan melebihi gaji, netto tidak boleh negatif")
	}
	return pokok, netto, nil
}

func (s *KaryawanService) BuatSlip(kid, oleh int64, req dto.SlipRequest) error {
	awal, err1 := time.Parse(dateLayout, req.PeriodeAwal)
	akhir, err2 := time.Parse(dateLayout, req.PeriodeAkhir)
	if err1 != nil || err2 != nil || akhir.Before(awal) {
		return types.BadRequest("Periode tidak valid")
	}
	return utils.WithTx(s.db, func(tx *sql.Tx) error {
		var jenis string
		var gaji int
		if tx.QueryRow("SELECT jenis,gaji FROM karyawan WHERE id=? AND klien_id=? AND aktif=1 FOR UPDATE", req.KaryawanID, kid).Scan(&jenis, &gaji) != nil {
			return types.NotFound("Karyawan")
		}
		var saldo int
		if err := tx.QueryRow("SELECT COALESCE(SUM(jumlah),0) FROM karyawan_bon WHERE klien_id=? AND karyawan_id=?", kid, req.KaryawanID).Scan(&saldo); err != nil {
			return err
		}
		if req.PotonganBon > saldo {
			return types.BadRequest("Potongan bon melebihi sisa bon karyawan")
		}
		pokok, netto, err := hitungSlip(jenis, gaji, req.Hari, req.Tunjangan, req.Bonus, req.Potongan, req.PotonganBon)
		if err != nil {
			return err
		}
		var ada int
		if tx.QueryRow("SELECT 1 FROM gaji_slip WHERE klien_id=? AND karyawan_id=? AND periode_awal<=? AND periode_akhir>=? LIMIT 1",
			kid, req.KaryawanID, req.PeriodeAkhir, req.PeriodeAwal).Scan(&ada) == nil {
			return types.BadRequest("Sudah ada slip gaji karyawan ini pada periode yang beririsan")
		}
		_, err = tx.Exec(`INSERT INTO gaji_slip (klien_id,karyawan_id,periode_awal,periode_akhir,hari,pokok,tunjangan,bonus,potongan,potongan_bon,netto,catatan,dibuat_oleh,dibuat_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, kid, req.KaryawanID, req.PeriodeAwal, req.PeriodeAkhir, req.Hari, pokok,
			req.Tunjangan, req.Bonus, req.Potongan, req.PotonganBon, netto, strings.TrimSpace(req.Catatan), oleh, time.Now())
		return err
	})
}

func (s *KaryawanService) ListSlip(kid int64, q dto.SlipQuery) ([]dto.SlipResponse, error) {
	from, to, err := parseRange(q.Dari, q.Sampai, time.Now())
	if err != nil {
		return nil, err
	}
	where, args := "g.klien_id=? AND g.periode_akhir>=? AND g.periode_akhir<?", []any{kid, from, to}
	if q.KaryawanID > 0 {
		where += " AND g.karyawan_id=?"
		args = append(args, q.KaryawanID)
	}
	rows, err := s.db.Query(`SELECT g.id,g.karyawan_id,k.nama,k.jabatan,k.jenis,g.periode_awal,g.periode_akhir,g.hari,g.pokok,g.tunjangan,g.bonus,
		g.potongan,g.potongan_bon,g.netto,g.catatan,g.status,g.dibuat_at,g.dibayar_at,COALESCE(u.nama,'')
		FROM gaji_slip g JOIN karyawan k ON k.id=g.karyawan_id LEFT JOIN users u ON u.id=g.dibayar_oleh
		WHERE `+where+` ORDER BY g.periode_akhir DESC, g.id DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []dto.SlipResponse{}
	for rows.Next() {
		var g dto.SlipResponse
		var awal, akhir time.Time
		if err := rows.Scan(&g.ID, &g.KaryawanID, &g.Karyawan, &g.Jabatan, &g.Jenis, &awal, &akhir, &g.Hari, &g.Pokok, &g.Tunjangan, &g.Bonus,
			&g.Potongan, &g.PotonganBon, &g.Netto, &g.Catatan, &g.Status, &g.DibuatAt, &g.DibayarAt, &g.DibayarOleh); err != nil {
			return nil, err
		}
		g.PeriodeAwal, g.PeriodeAkhir = awal.Format(dateLayout), akhir.Format(dateLayout)
		out = append(out, g)
	}
	return out, rows.Err()
}

func (s *KaryawanService) HapusSlip(kid, id int64) error {
	res, err := s.db.Exec("DELETE FROM gaji_slip WHERE id=? AND klien_id=? AND status='draft'", id, kid)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		var ada int
		if s.db.QueryRow("SELECT 1 FROM gaji_slip WHERE id=? AND klien_id=?", id, kid).Scan(&ada) != nil {
			return types.NotFound("Slip gaji")
		}
		return types.BadRequest("Slip yang sudah dibayar tidak bisa dihapus")
	}
	return nil
}

// BayarSlip menandai slip dibayar dan mencatat potongan bonnya sebagai cicilan.
func (s *KaryawanService) BayarSlip(kid, oleh, id int64) error {
	return utils.WithTx(s.db, func(tx *sql.Tx) error {
		var karyawanID int64
		var potBon int
		var status string
		if tx.QueryRow("SELECT karyawan_id,potongan_bon,status FROM gaji_slip WHERE id=? AND klien_id=? FOR UPDATE", id, kid).Scan(&karyawanID, &potBon, &status) != nil {
			return types.NotFound("Slip gaji")
		}
		if status != "draft" {
			return types.BadRequest("Slip ini sudah dibayar")
		}
		if potBon > 0 {
			var saldo int
			if err := tx.QueryRow("SELECT COALESCE(SUM(jumlah),0) FROM karyawan_bon WHERE klien_id=? AND karyawan_id=? FOR UPDATE", kid, karyawanID).Scan(&saldo); err != nil {
				return err
			}
			if potBon > saldo {
				return types.BadRequest("Sisa bon sudah berkurang, hapus slip ini dan buat ulang")
			}
			if _, err := tx.Exec("INSERT INTO karyawan_bon (klien_id,karyawan_id,tanggal,jumlah,keterangan,slip_id) VALUES (?,?,?,?,?,?)",
				kid, karyawanID, time.Now().Format(dateLayout), -potBon, "Potongan gaji", id); err != nil {
				return err
			}
		}
		_, err := tx.Exec("UPDATE gaji_slip SET status='dibayar', dibayar_at=?, dibayar_oleh=? WHERE id=? AND klien_id=?", time.Now(), oleh, id, kid)
		return err
	})
}
