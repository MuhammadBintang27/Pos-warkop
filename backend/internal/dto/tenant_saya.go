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
	Tambahan  bool       `json:"tambahan"` // tenant ini sudah pernah menerima kiriman dari bill yang sama
	SiapAt    *time.Time `json:"siapAt"`
}

type PesananMasukResponse struct {
	Tenant string      `json:"tenant"`
	Items  []ItemMasuk `json:"items"`
}

type SiapRequest struct {
	Siap *bool `json:"siap" validate:"required"`
}

type PerangkatRequest struct {
	Token string `json:"token" validate:"required,max=255"`
}
