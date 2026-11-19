package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"warkop/internal/dto"
	"warkop/internal/types"
	"warkop/internal/utils"
)

type PembayaranService struct {
	db  *sql.DB
	hub *EventHub
}

type pembayaranRow struct {
	Metode      types.Metode
	Jumlah      int
	Diterima    int
	Kembalian   int
	PelangganID *int64
}

func NewPembayaranService(db *sql.DB, hub *EventHub) *PembayaranService {
	return &PembayaranService{db: db, hub: hub}
}

// normalizePembayaran memastikan jumlah semua pembayaran sama persis dengan tagihan dan menghitung kembalian tunai.
func normalizePembayaran(total int, in []dto.PembayaranInput) ([]pembayaranRow, error) {
	rows := make([]pembayaranRow, 0, len(in))
	sum := 0
	for _, p := range in {
		r := pembayaranRow{Metode: p.Metode, Jumlah: p.Jumlah, Diterima: p.Jumlah}
		switch p.Metode {
		case types.MetodeTunai:
			if p.Diterima > 0 {
				if p.Diterima < p.Jumlah {
					return nil, types.BadRequest("Uang diterima kurang dari jumlah tunai")
				}
				r.Diterima, r.Kembalian = p.Diterima, p.Diterima-p.Jumlah
			}
		case types.MetodeKasbon:
			if p.PelangganID == nil {
				return nil, types.BadRequest("Pilih pelanggan untuk kasbon")
			}
			r.PelangganID = p.PelangganID
		}
		sum += p.Jumlah
		rows = append(rows, r)
	}
	if sum != total {
		return nil, types.BadRequest(fmt.Sprintf("Total pembayaran (%d) harus sama dengan tagihan (%d)", sum, total))
	}
	return rows, nil
}

func (s *PembayaranService) Bayar(actor *utils.JWTPayload, id int64, req dto.BayarRequest) (*dto.PesananResponse, error) {
	kid := actor.KlienID
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		shiftID, err := activeShiftID(tx, kid)
		if err != nil {
			return err
		}
		if _, err := lockOpen(tx, kid, id); err != nil {
			return err
		}
		p, err := loadPesanan(tx, kid, id)
		if err != nil {
			return err
		}
		if p.Subtotal == 0 {
			return types.BadRequest("Pesanan masih kosong")
		}
		rows, err := normalizePembayaran(p.Total, req.Pembayaran)
		if err != nil {
			return err
		}
		now := time.Now()
		for _, r := range rows {
			if r.PelangganID != nil {
				var exists int
				if tx.QueryRow("SELECT 1 FROM pelanggan WHERE id=? AND klien_id=?", *r.PelangganID, kid).Scan(&exists) != nil {
					return types.NotFound("Pelanggan")
				}
			}
			if _, err := tx.Exec(`INSERT INTO pembayaran (klien_id,pesanan_id,metode,jumlah,diterima,kembalian,pelanggan_id,diterima_oleh,shift_id,at)
				VALUES (?,?,?,?,?,?,?,?,?,?)`, kid, id, r.Metode, r.Jumlah, r.Diterima, r.Kembalian, r.PelangganID, actor.UserID, shiftID, now); err != nil {
				return err
			}
		}
		_, err = tx.Exec("UPDATE pesanan SET status='lunas', lunas_at=?, shift_id=? WHERE id=? AND klien_id=?", now, shiftID, id, kid)
		return err
	})
	if err != nil {
		return nil, err
	}
	s.hub.Publish(kid, "pesanan", "meja")
	return loadPesanan(s.db, kid, id)
}

// lockLunas mengunci bill lunas yang belum masuk closing tenant lain; koreksi setelah closing membuat setoran tenant salah.
// Dapur milik sendiri tidak di-closing, jadi tidak menghalangi.
func lockLunas(tx *sql.Tx, kid, id int64) (*int64, error) {
	var status string
	var mejaID *int64
	err := tx.QueryRow("SELECT status,meja_id FROM pesanan WHERE id=? AND klien_id=? FOR UPDATE", id, kid).Scan(&status, &mejaID)
	if err == sql.ErrNoRows {
		return nil, types.NotFound("Pesanan")
	}
	if err != nil {
		return nil, err
	}
	if status != string(types.StatusLunas) {
		return nil, types.BadRequest("Hanya bill lunas yang bisa dikoreksi")
	}
	var closing bool
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM pesanan_item i JOIN tenant t ON t.id=i.tenant_id AND t.milik_sendiri=0
			WHERE i.pesanan_id=? AND i.klien_id=? AND i.closing_id IS NOT NULL)
		OR EXISTS(SELECT 1 FROM pesanan_item i JOIN tenant t ON t.id=i.tenant_id AND t.milik_sendiri=0
			JOIN pesanan p ON p.id=i.pesanan_id JOIN shift sh ON sh.id=p.shift_id
			JOIN closing_tenant c ON c.klien_id=i.klien_id AND c.tenant_id=i.tenant_id AND c.tanggal=DATE(sh.buka_at)
			WHERE i.pesanan_id=? AND i.klien_id=?)`, id, kid, id, kid).Scan(&closing); err != nil {
		return nil, err
	}
	if closing {
		return nil, types.BadRequest("Bill ini sudah masuk closing tenant. Batalkan closing-nya dulu")
	}
	return mejaID, nil
}

// BukaUlang mengembalikan bill lunas ke open untuk dikoreksi lalu dibayar ulang. Hanya untuk pembayaran di shift
// yang masih buka, supaya kas shift yang sudah ditutup tidak berubah; selebihnya pakai Refund.
func (s *PembayaranService) BukaUlang(actor *utils.JWTPayload, id int64) (*dto.PesananResponse, error) {
	kid := actor.KlienID
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		mejaID, err := lockLunas(tx, kid, id)
		if err != nil {
			return err
		}
		shiftID, err := activeShiftID(tx, kid)
		if _, ok := utils.AsHTTPError(err); ok {
			return types.BadRequest("Tidak ada shift yang buka. Gunakan Refund untuk bill dari shift yang sudah ditutup")
		} else if err != nil {
			return err
		}
		var lain int
		if err := tx.QueryRow("SELECT COUNT(*) FROM pembayaran WHERE pesanan_id=? AND klien_id=? AND shift_id<>?", id, kid, shiftID).Scan(&lain); err != nil {
			return err
		}
		if lain > 0 {
			return types.BadRequest("Bill ini dibayar di shift yang sudah ditutup. Gunakan Refund")
		}
		if mejaID != nil {
			var dipakai int
			if err := tx.QueryRow("SELECT COUNT(*) FROM pesanan WHERE klien_id=? AND meja_id=? AND status='open'", kid, *mejaID).Scan(&dipakai); err != nil {
				return err
			}
			if dipakai > 0 {
				return types.BadRequest("Mejanya sudah dipakai bill lain. Selesaikan atau pindahkan bill itu dulu")
			}
		}
		if _, err := tx.Exec("DELETE FROM pembayaran WHERE pesanan_id=? AND klien_id=?", id, kid); err != nil {
			return err
		}
		_, err = tx.Exec("UPDATE pesanan SET status='open', lunas_at=NULL WHERE id=? AND klien_id=?", id, kid)
		return err
	})
	if err != nil {
		return nil, err
	}
	s.hub.Publish(kid, "pesanan", "meja")
	return loadPesanan(s.db, kid, id)
}

// Refund membatalkan bill lunas. Bagian tunai dicatat sebagai pengeluaran 'refund' di shift aktif (uang keluar dari laci),
// bagian kasbon dihapus dari utang pelanggan, QRIS/transfer dikembalikan di luar sistem.
func (s *PembayaranService) Refund(actor *utils.JWTPayload, id int64, req dto.KoreksiRequest) (*dto.PesananResponse, error) {
	kid := actor.KlienID
	alasan := strings.TrimSpace(req.Alasan)
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		if _, err := lockLunas(tx, kid, id); err != nil {
			return err
		}
		var tunai int
		if err := tx.QueryRow("SELECT COALESCE(SUM(jumlah),0) FROM pembayaran WHERE pesanan_id=? AND klien_id=? AND metode='tunai'", id, kid).Scan(&tunai); err != nil {
			return err
		}
		if tunai > 0 {
			shiftID, err := activeShiftID(tx, kid)
			if _, ok := utils.AsHTTPError(err); ok {
				return types.BadRequest("Buka shift dulu: uang tunai refund diambil dari kas laci")
			} else if err != nil {
				return err
			}
			if _, err := tx.Exec("INSERT INTO pengeluaran (klien_id,shift_id,keterangan,jumlah,kategori,dicatat_oleh,at) VALUES (?,?,?,?,'refund',?,?)",
				kid, shiftID, fmt.Sprintf("Refund bill #%d: %s", id, alasan), tunai, actor.UserID, time.Now()); err != nil {
				return err
			}
		}
		if _, err := tx.Exec("DELETE FROM pembayaran WHERE pesanan_id=? AND klien_id=? AND metode='kasbon'", id, kid); err != nil {
			return err
		}
		p, err := loadPesanan(tx, kid, id)
		if err != nil {
			return err
		}
		for _, i := range p.Items {
			if !req.SudahDibuat {
				if err := adjustStok(tx, kid, i.MenuID, i.Qty); err != nil {
					return err
				}
				if err := pakaiBahan(tx, kid, i.ID, -i.Qty); err != nil {
					return err
				}
			}
			if err := catatBatal(tx, kid, id, i, i.Qty, "Refund: "+alasan, req.SudahDibuat, actor.UserID); err != nil {
				return err
			}
		}
		_, err = tx.Exec("UPDATE pesanan SET status='batal' WHERE id=? AND klien_id=?", id, kid)
		return err
	})
	if err != nil {
		return nil, err
	}
	s.hub.Publish(kid, "pesanan", "meja", "menu", "shift")
	return loadPesanan(s.db, kid, id)
}
