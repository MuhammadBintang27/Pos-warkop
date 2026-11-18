package dto

import "time"

type KaryawanRequest struct {
	Nama    string `json:"nama" validate:"required,max=100"`
	Jabatan string `json:"jabatan" validate:"max=50"`
	Jenis   string `json:"jenis" validate:"required,oneof=bulanan harian"`
	Gaji    int    `json:"gaji" validate:"gte=0"` // per bulan, atau per hari kerja bila harian
}

type KaryawanResponse struct {
	ID       int64  `json:"id"`
	Nama     string `json:"nama"`
	Jabatan  string `json:"jabatan"`
	Jenis    string `json:"jenis"`
	Gaji     int    `json:"gaji"`
	SaldoBon int    `json:"saldoBon"`
}

type BonRequest struct {
	Jumlah     int    `json:"jumlah" validate:"required,gt=0"`
	Tanggal    string `json:"tanggal" validate:"omitempty,len=10"`
	Keterangan string `json:"keterangan" validate:"max=255"`
}

type BonResponse struct {
	ID         int64  `json:"id"`
	Tanggal    string `json:"tanggal"`
	Jumlah     int    `json:"jumlah"` // positif = bon diambil, negatif = dipotong dari gaji
	Keterangan string `json:"keterangan"`
	SlipID     *int64 `json:"slipId"`
}

type SlipRequest struct {
	KaryawanID   int64  `json:"karyawanId" validate:"required,gt=0"`
	PeriodeAwal  string `json:"periodeAwal" validate:"required,len=10"`
	PeriodeAkhir string `json:"periodeAkhir" validate:"required,len=10"`
	Hari         int    `json:"hari" validate:"gte=0,lte=31"` // hari kerja, wajib untuk karyawan harian
	Tunjangan    int    `json:"tunjangan" validate:"gte=0"`
	Bonus        int    `json:"bonus" validate:"gte=0"`
	Potongan     int    `json:"potongan" validate:"gte=0"`
	PotonganBon  int    `json:"potonganBon" validate:"gte=0"`
	Catatan      string `json:"catatan" validate:"max=255"`
}

type SlipQuery struct {
	Dari       string `json:"dari" query:"dari"`
	Sampai     string `json:"sampai" query:"sampai"`
	KaryawanID int64  `json:"karyawanId" query:"karyawanId"`
}

type SlipResponse struct {
	ID           int64      `json:"id"`
	KaryawanID   int64      `json:"karyawanId"`
	Karyawan     string     `json:"karyawan"`
	Jabatan      string     `json:"jabatan"`
	Jenis        string     `json:"jenis"`
	PeriodeAwal  string     `json:"periodeAwal"`
	PeriodeAkhir string     `json:"periodeAkhir"`
	Hari         int        `json:"hari"`
	Pokok        int        `json:"pokok"`
	Tunjangan    int        `json:"tunjangan"`
	Bonus        int        `json:"bonus"`
	Potongan     int        `json:"potongan"`
	PotonganBon  int        `json:"potonganBon"`
	Netto        int        `json:"netto"`
	Catatan      string     `json:"catatan"`
	Status       string     `json:"status"`
	DibuatAt     time.Time  `json:"dibuatAt"`
	DibayarAt    *time.Time `json:"dibayarAt"`
	DibayarOleh  string     `json:"dibayarOleh"`
}
