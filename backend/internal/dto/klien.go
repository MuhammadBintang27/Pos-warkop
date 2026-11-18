package dto

import "time"

type KlienRequest struct {
	Kode          string `json:"kode" validate:"required,max=30"`
	Nama          string `json:"nama" validate:"required,max=100"`
	Alamat        string `json:"alamat" validate:"max=255"`
	Bulan         int    `json:"bulan" validate:"gte=0,lte=36"` // langganan awal; 0 = belum ada
	OwnerNama     string `json:"ownerNama" validate:"max=100"`
	OwnerUsername string `json:"ownerUsername" validate:"max=30"`
	OwnerPassword string `json:"ownerPassword" validate:"omitempty,min=6"`
}

type KlienEditRequest struct {
	Nama   string `json:"nama" validate:"required,max=100"`
	Alamat string `json:"alamat" validate:"max=255"`
}

type KlienAktifRequest struct {
	Aktif *bool `json:"aktif" validate:"required"`
}

type KlienLanggananRequest struct {
	Bulan int `json:"bulan" validate:"required,min=1,max=36"`
}

type KlienResponse struct {
	ID              int64      `json:"id"`
	Kode            string     `json:"kode"`
	Nama            string     `json:"nama"`
	Alamat          string     `json:"alamat"`
	Aktif           bool       `json:"aktif"`
	LanggananSampai *time.Time `json:"langgananSampai"`
	Langganan       string     `json:"langganan"` // belum, aktif, segera, berakhir
	Pengguna        int        `json:"pengguna"`
	Transaksi       int        `json:"transaksi"`
	TerakhirAt      *time.Time `json:"terakhirAt"`
}

// KlienKredensial: akun owner yang baru dibuat atau direset. Password kosong bila owner memilihnya sendiri.
type KlienKredensial struct {
	KlienID  int64  `json:"klienId"`
	Username string `json:"username"`
	Password string `json:"password"`
}
