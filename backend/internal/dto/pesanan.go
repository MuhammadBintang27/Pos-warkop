package dto

import (
	"time"

	"warkop/internal/types"
)

type CreatePesananRequest struct {
	MejaID   *int64 `json:"mejaId"`
	AtasNama string `json:"atasNama" validate:"max=100"`
}

type AddItemRequest struct {
	MenuID  int64   `json:"menuId" validate:"required,gt=0"`
	Qty     int     `json:"qty" validate:"required,min=1,max=999"`
	OpsiIDs []int64 `json:"opsiIds"`
	Catatan string  `json:"catatan" validate:"max=255"`
}

type UpdateItemRequest struct {
	Qty     int    `json:"qty" validate:"required,min=1,max=999"`
	Catatan string `json:"catatan" validate:"max=255"`
}

type DiskonRequest struct {
	Diskon int `json:"diskon" validate:"gte=0"`
}

type PindahRequest struct {
	MejaID int64 `json:"mejaId" validate:"required,gt=0"`
}

type GabungRequest struct {
	KePesananID int64 `json:"kePesananId" validate:"required,gt=0"`
}

type PisahItem struct {
	ItemID int64 `json:"itemId" validate:"required,gt=0"`
	Qty    int   `json:"qty" validate:"required,min=1"`
}

type PisahRequest struct {
	Items []PisahItem `json:"items" validate:"required,min=1,dive"`
}

type PembayaranInput struct {
	Metode      types.Metode `json:"metode" validate:"required,metode"`
	Jumlah      int          `json:"jumlah" validate:"required,gt=0"`
	Diterima    int          `json:"diterima" validate:"gte=0"`
	PelangganID *int64       `json:"pelangganId"`
}

type BayarRequest struct {
	Pembayaran []PembayaranInput `json:"pembayaran" validate:"required,min=1,dive"`
}

type BatalItemRequest struct {
	Qty         int    `json:"qty" validate:"required,min=1"`
	Alasan      string `json:"alasan" validate:"required,max=100"`
	SudahDibuat bool   `json:"sudahDibuat"`
}

// KirimRequest: kosong = kirim semua item yang belum dikirim. ItemIDs diisi antrean offline HP kasir.
type KirimRequest struct {
	ItemIDs []int64 `json:"itemIds"`
}

type BatalRequest struct {
	Alasan      string `json:"alasan" validate:"max=100"`
	SudahDibuat bool   `json:"sudahDibuat"`
	// ItemIDs: item yang dilihat kasir saat membatalkan offline; bill yang punya item lain ditolak
	ItemIDs []int64 `json:"itemIds"`
}

type PembatalanResponse struct {
	Nama         string    `json:"nama"`
	Opsi         string    `json:"opsi"`
	Qty          int       `json:"qty"`
	Harga        int       `json:"harga"`
	Alasan       string    `json:"alasan"`
	SudahDikirim bool      `json:"sudahDikirim"`
	SudahDibuat  bool      `json:"sudahDibuat"`
	Oleh         string    `json:"oleh"`
	At           time.Time `json:"at"`
}

type ItemResponse struct {
	TenantID   int64  `json:"tenantId"`
	TenantNama string `json:"tenantNama"`
	ID         int64  `json:"id"`
	MenuID     int64  `json:"menuId"`
	Nama       string `json:"nama"`
	Harga      int    `json:"harga"`
	Qty        int    `json:"qty"`
	Opsi       string `json:"opsi"`
	Catatan    string `json:"catatan"`
	Dikirim    bool   `json:"dikirim"`
	// diisi saat tenant menekan Proses (mulai dibuat)
	DiprosesAt *time.Time `json:"diprosesAt"`
	// diisi saat tenant menekan Selesai (sudah diantar ke meja)
	SiapAt *time.Time `json:"siapAt"`
}

type PembayaranResponse struct {
	Metode    types.Metode `json:"metode"`
	Jumlah    int          `json:"jumlah"`
	Diterima  int          `json:"diterima"`
	Kembalian int          `json:"kembalian"`
	Pelanggan string       `json:"pelanggan,omitempty"`
}

type PesananResponse struct {
	ID         int64                `json:"id"`
	MejaID     *int64               `json:"mejaId"`
	MejaNama   string               `json:"mejaNama"`
	AtasNama   string               `json:"atasNama"`
	Status     types.PesananStatus  `json:"status"`
	Diskon     int                  `json:"diskon"`
	Subtotal   int                  `json:"subtotal"`
	Total      int                  `json:"total"`
	DibuatAt   time.Time            `json:"dibuatAt"`
	LunasAt    *time.Time           `json:"lunasAt"`
	Kasir      string               `json:"kasir,omitempty"`
	Items      []ItemResponse       `json:"items"`
	Pembayaran []PembayaranResponse `json:"pembayaran"`
	Pembatalan []PembatalanResponse `json:"pembatalan"`
}

type BatalResponse struct {
	Pesanan *PesananResponse `json:"pesanan"`
	Tiket   *KirimResponse   `json:"tiket"`
}

type KirimResponse struct {
	MejaNama string         `json:"mejaNama"`
	AtasNama string         `json:"atasNama"`
	Oleh     string         `json:"oleh"`
	Items    []ItemResponse `json:"items"`
	// Tambahan: tenantId -> true bila tenant itu sudah pernah menerima item dari bill ini (judul "PESANAN TAMBAHAN").
	Tambahan map[int64]bool `json:"tambahan"`
	// Antrean: tenantId -> true bila tiketnya sudah diantrekan ke printer tenant; klien tidak perlu mencetaknya lagi.
	Antrean map[int64]bool `json:"antrean"`
}

type KoreksiRequest struct {
	Alasan      string `json:"alasan" validate:"required,max=100"`
	SudahDibuat bool   `json:"sudahDibuat"` // refund: stok tidak dikembalikan bila makanan sudah dibuat
}
