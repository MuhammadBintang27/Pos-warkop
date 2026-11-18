package dto

import "time"

type PelangganRequest struct {
	Nama string `json:"nama" validate:"required,max=100"`
	Telp string `json:"telp" validate:"max=30"`
}

type PelangganResponse struct {
	ID        int64  `json:"id"`
	Nama      string `json:"nama"`
	Telp      string `json:"telp"`
	SisaUtang int    `json:"sisaUtang"`
}

type LunasiRequest struct {
	Jumlah int    `json:"jumlah" validate:"required,gt=0"`
	Metode string `json:"metode" validate:"required,oneof=tunai qris transfer"`
}

type BukaShiftRequest struct {
	KasAwal int `json:"kasAwal" validate:"gte=0"`
}

type TutupShiftRequest struct {
	KasHitung int `json:"kasHitung" validate:"gte=0"`
}

type ShiftResponse struct {
	TutupOleh string     `json:"tutupOleh"`
	ID        int64      `json:"id"`
	Kasir     string     `json:"kasir"`
	BukaAt    time.Time  `json:"bukaAt"`
	TutupAt   *time.Time `json:"tutupAt"`
	KasAwal   int        `json:"kasAwal"`
	KasSistem *int       `json:"kasSistem"`
	KasHitung *int       `json:"kasHitung"`
	Selisih   *int       `json:"selisih"`
}

type PengeluaranRequest struct {
	Keterangan string `json:"keterangan" validate:"required,max=255"`
	Jumlah     int    `json:"jumlah" validate:"required,gt=0"`
	Kategori   string `json:"kategori" validate:"omitempty,oneof=bahan operasional sewa transport perawatan lainnya"`
}

type PengeluaranResponse struct {
	Oleh       string    `json:"oleh"`
	ID         int64     `json:"id"`
	Keterangan string    `json:"keterangan"`
	Jumlah     int       `json:"jumlah"`
	Kategori   string    `json:"kategori"`
	At         time.Time `json:"at"`
}

type MetodeTotal struct {
	Metode string `json:"metode"`
	Total  int    `json:"total"`
}

type HariTotal struct {
	Tanggal string `json:"tanggal"`
	Total   int    `json:"total"`
}

type MenuTerlaris struct {
	Nama  string `json:"nama"`
	Qty   int    `json:"qty"`
	Total int    `json:"total"`
}

type LaporanResponse struct {
	Dari      string          `json:"dari"`
	Sampai    string          `json:"sampai"`
	Omzet     int             `json:"omzet"`
	Diskon    int             `json:"diskon"`
	Transaksi int             `json:"transaksi"`
	Keluar    int             `json:"pengeluaran"`
	PerMetode []MetodeTotal   `json:"perMetode"`
	PerHari   []HariTotal     `json:"perHari"`
	Terlaris  []MenuTerlaris  `json:"terlaris"`
	Shift     []ShiftResponse `json:"shift"`
}
