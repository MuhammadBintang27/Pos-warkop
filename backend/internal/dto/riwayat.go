package dto

import (
	"time"

	"warkop/internal/types"
)

type RiwayatQuery struct {
	Dari   string `json:"dari" query:"dari"`
	Sampai string `json:"sampai" query:"sampai"`
	Status string `json:"status" query:"status" validate:"omitempty,oneof=open lunas batal"`
	Q      string `json:"q" query:"q" validate:"max=100"`
	Oleh   int64  `json:"oleh" query:"oleh" validate:"gte=0"`
	Tenant int64  `json:"tenant" query:"tenant" validate:"gte=0"` // hanya bill yang memuat item tenant ini
	Page   int    `json:"page" query:"page" validate:"gte=0"`
	Limit  int    `json:"limit" query:"limit" validate:"gte=0,max=100"`
}

type RiwayatItem struct {
	ID          int64               `json:"id"`
	MejaNama    string              `json:"mejaNama"`
	AtasNama    string              `json:"atasNama"`
	Status      types.PesananStatus `json:"status"`
	JumlahItem  int                 `json:"jumlahItem"`
	Total       int                 `json:"total"`
	NilaiTenant int                 `json:"nilaiTenant"` // nilai item tenant yang difilter, sebelum diskon
	Metode      string              `json:"metode"`
	Pembuat     string              `json:"pembuat"`
	Kasir       string              `json:"kasir"`
	DibuatAt    time.Time           `json:"dibuatAt"`
	LunasAt     *time.Time          `json:"lunasAt"`
}

type RiwayatResponse struct {
	Items []RiwayatItem `json:"items"`
	Total int           `json:"total"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
	Lunas int           `json:"lunas"`
	Omzet int           `json:"omzet"`
}
