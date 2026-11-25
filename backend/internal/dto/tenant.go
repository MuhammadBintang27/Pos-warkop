package dto

type TenantRequest struct {
	Nama          string       `json:"nama" validate:"required,max=100"`
	KomisiPersen  float64      `json:"komisiPersen" validate:"gte=0,lte=100"`
	KomisiPerMenu bool         `json:"komisiPerMenu"`
	MilikSendiri  bool         `json:"milikSendiri"`
	PrinterID     *int64       `json:"printerId"`
	Warna         string       `json:"warna" validate:"omitempty,hexcolor,len=7"`
	KomisiMenu    []KomisiMenu `json:"komisiMenu" validate:"max=500,dive"` // tarif Rp per porsi menu tenant ini; kosong = tidak diubah
}

type KomisiMenu struct {
	MenuID int64 `json:"menuId" validate:"required,gt=0"`
	Komisi int   `json:"komisi" validate:"gte=0"`
}

type TenantResponse struct {
	ID            int64   `json:"id"`
	Nama          string  `json:"nama"`
	KomisiPersen  float64 `json:"komisiPersen"`
	KomisiPerMenu bool    `json:"komisiPerMenu"`
	Buka          bool    `json:"buka"`
	MilikSendiri  bool    `json:"milikSendiri"`
	PrinterID     *int64  `json:"printerId"`
	Warna         string  `json:"warna"`
}

type BukaRequest struct {
	Buka bool `json:"buka"`
}

type SetoranTenant struct {
	ID            int64   `json:"id"`
	Nama          string  `json:"nama"`
	KomisiPersen  float64 `json:"komisiPersen"`
	KomisiPerMenu bool    `json:"komisiPerMenu"`
	Kotor         int     `json:"kotor"`
	Diskon        int     `json:"diskon"`
	Komisi        int     `json:"komisi"`
	Setor         int     `json:"setor"`
}

type SetoranResponse struct {
	Dari   string          `json:"dari"`
	Sampai string          `json:"sampai"`
	Tenant []SetoranTenant `json:"tenant"`
}
