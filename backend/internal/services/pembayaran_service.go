package services

import (
	"database/sql"
	"fmt"
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
