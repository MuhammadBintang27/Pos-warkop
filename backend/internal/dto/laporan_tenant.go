package dto

import "time"

type HarianQuery struct {
	Dari   string `json:"dari" query:"dari"`
	Sampai string `json:"sampai" query:"sampai"`
}

type DetailTenantQuery struct {
	TenantID int64  `json:"tenantId" query:"tenantId" validate:"required,gt=0"`
	Dari     string `json:"dari" query:"dari"`
	Sampai   string `json:"sampai" query:"sampai"`
}

type ClosingRequest struct {
	TenantID int64  `json:"tenantId" validate:"required,gt=0"`
	Tanggal  string `json:"tanggal" validate:"required"`
	// PesananIDs: bill belum dibayar yang ikut dihitung dan ditalangi pemilik pada closing ini.
	PesananIDs []int64 `json:"pesananIds" validate:"max=200,dive,gt=0"`
}

type BelumBayarItem struct {
	Nama  string `json:"nama"`
	Opsi  string `json:"opsi"`
	Qty   int    `json:"qty"`
	Harga int    `json:"harga"`
}

// BelumBayarBill: bill terbuka dengan item tenant yang sudah dikirim ke dapur, belum dibayar pelanggan.
type BelumBayarBill struct {
	PesananID int64            `json:"pesananId"`
	MejaNama  string           `json:"mejaNama"`
	AtasNama  string           `json:"atasNama"`
	DibuatAt  time.Time        `json:"dibuatAt"`
	Items     []BelumBayarItem `json:"items"`
	Kotor     int              `json:"kotor"`
	Diskon    int              `json:"diskon"` // bagian diskon bill untuk item tenant ini
}

type TalanganBill struct {
	PesananID int64  `json:"pesananId"`
	MejaNama  string `json:"mejaNama"`
	AtasNama  string `json:"atasNama"`
	Kotor     int    `json:"kotor"`
}

type HarianTenant struct {
	TenantID      int64      `json:"tenantId"`
	Nama          string     `json:"nama"`
	Tanggal       string     `json:"tanggal"`
	Transaksi     int        `json:"transaksi"`
	Kotor         int        `json:"kotor"`
	Diskon        int        `json:"diskon"`
	KomisiPersen  float64    `json:"komisiPersen"`
	KomisiPerMenu bool       `json:"komisiPerMenu"`
	Komisi        int        `json:"komisi"`
	Setor         int        `json:"setor"`
	Status        string     `json:"status"`
	MilikSendiri  bool       `json:"milikSendiri"`
	ClosingID     int64      `json:"closingId"`
	DisetorAt     *time.Time `json:"disetorAt"`
	DisetorOleh   string     `json:"disetorOleh"`
	DibuatAt      *time.Time `json:"dibuatAt"`
	DibuatOleh    string     `json:"dibuatOleh"`
}

// TertundaTenant: bill terbuka (belum dibayar pelanggan) berisi menu tenant, belum masuk penjualan.
type TertundaTenant struct {
	TenantID int64 `json:"tenantId"`
	Bill     int   `json:"bill"`
	Nilai    int   `json:"nilai"`
}

type HarianResponse struct {
	Dari     string           `json:"dari"`
	Sampai   string           `json:"sampai"`
	Baris    []HarianTenant   `json:"baris"`
	Tertunda []TertundaTenant `json:"tertunda"`
}

type DetailMenuTenant struct {
	Nama       string `json:"nama"`
	Qty        int    `json:"qty"`
	Nilai      int    `json:"nilai"`
	BatalQty   int    `json:"batalQty"`
	BatalNilai int    `json:"batalNilai"`
}

type BatalTenant struct {
	PesananID    int64     `json:"pesananId"`
	Nama         string    `json:"nama"`
	Opsi         string    `json:"opsi"`
	Qty          int       `json:"qty"`
	Nilai        int       `json:"nilai"`
	Alasan       string    `json:"alasan"`
	SudahDikirim bool      `json:"sudahDikirim"`
	SudahDibuat  bool      `json:"sudahDibuat"`
	Oleh         string    `json:"oleh"`
	At           time.Time `json:"at"`
}

type DetailTenantResponse struct {
	Tenant     string             `json:"tenant"`
	Dari       string             `json:"dari"`
	Sampai     string             `json:"sampai"`
	Menu       []DetailMenuTenant `json:"menu"`
	Batal      []BatalTenant      `json:"batal"`
	Qty        int                `json:"qty"`
	Nilai      int                `json:"nilai"`
	BatalQty   int                `json:"batalQty"`
	BatalNilai int                `json:"batalNilai"`
	Talangan   []TalanganBill     `json:"talangan"` // bill belum dibayar yang ikut closing pada rentang ini
}

type SetorPeriodeRequest struct {
	TenantID int64  `json:"tenantId" validate:"required,gt=0"`
	Dari     string `json:"dari" validate:"required,len=10"`
	Sampai   string `json:"sampai" validate:"required,len=10"`
}
