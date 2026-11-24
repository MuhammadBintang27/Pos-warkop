package dto

import "time"

// ItemMasuk: item yang sudah dikirim ke dapur untuk satu tenant, dilihat dari akun tenant.
type ItemMasuk struct {
	ID        int64      `json:"id"`
	PesananID int64      `json:"pesananId"`
	MejaNama  string     `json:"mejaNama"`
	AtasNama  string     `json:"atasNama"`
	Nama      string     `json:"nama"`
	Qty       int        `json:"qty"`
	Opsi      string     `json:"opsi"`
	Catatan   string     `json:"catatan"`
	DikirimAt time.Time  `json:"dikirimAt"`
	Oleh      string     `json:"oleh"`
	Tambahan   bool       `json:"tambahan"` // tenant ini sudah pernah menerima kiriman dari bill yang sama
	DiprosesAt *time.Time `json:"diprosesAt"`
	SiapAt     *time.Time `json:"siapAt"`
}

// BatalMasuk: item terkirim yang dibatalkan (termasuk refund), supaya tenant tidak melihat pesanannya hilang begitu saja.
type BatalMasuk struct {
	PesananID   int64     `json:"pesananId"`
	MejaNama    string    `json:"mejaNama"`
	AtasNama    string    `json:"atasNama"`
	Nama        string    `json:"nama"`
	Opsi        string    `json:"opsi"`
	Qty         int       `json:"qty"`
	Alasan      string    `json:"alasan"`
	SudahDibuat bool      `json:"sudahDibuat"`
	Oleh        string    `json:"oleh"`
	At          time.Time `json:"at"`
}

type PesananMasukResponse struct {
	Tenant string       `json:"tenant"`
	Items  []ItemMasuk  `json:"items"`
	Batal  []BatalMasuk `json:"batal"`
}

type SiapRequest struct {
	Siap *bool `json:"siap" validate:"required"`
}

type ProsesRequest struct {
	Proses *bool `json:"proses" validate:"required"`
}

type PerangkatRequest struct {
	Token string `json:"token" validate:"required,max=255"`
}
