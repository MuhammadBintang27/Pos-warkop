package dto

import "time"

type AnalitikResponse struct {
	Dari             string            `json:"dari"`
	Sampai           string            `json:"sampai"`
	SebelumnyaDari   string            `json:"sebelumnyaDari"`
	SebelumnyaSampai string            `json:"sebelumnyaSampai"`
	Ringkasan        RingkasanPeriode  `json:"ringkasan"`
	Sebelumnya       RingkasanPeriode  `json:"sebelumnya"`
	Tren             []TrenHari        `json:"tren"`
	JamRamai         []JamRamai        `json:"jamRamai"`
	Kategori         []KategoriTotal   `json:"kategori"`
	Karyawan         []KinerjaKaryawan `json:"karyawan"`
	Kasbon           []KasbonUmur      `json:"kasbon"`
	Stok             []StokNilai       `json:"stok"`
}

// RingkasanPeriode: laba memakai rumus yang sama dengan Laba Rugi (pendapatan = penjualan sendiri + komisi tenant).
type RingkasanPeriode struct {
	Omzet      int `json:"omzet"`
	Transaksi  int `json:"transaksi"`
	RataRata   int `json:"rataRata"`
	Pendapatan int `json:"pendapatan"`
	HPP        int `json:"hpp"`
	LabaKotor  int `json:"labaKotor"`
	LabaBersih int `json:"labaBersih"`
}

type TrenHari struct {
	Tanggal    string `json:"tanggal"`
	Omzet      int    `json:"omzet"`
	Transaksi  int    `json:"transaksi"`
	Pendapatan int    `json:"pendapatan"`
	HPP        int    `json:"hpp"`   // termasuk bahan susut
	Biaya      int    `json:"biaya"` // pengeluaran operasional + gaji
	LabaKotor  int    `json:"labaKotor"`
	LabaBersih int    `json:"labaBersih"`
}

type JamRamai struct {
	Hari  int `json:"hari"` // 0 = Senin
	Jam   int `json:"jam"`
	Bill  int `json:"bill"`
	Omzet int `json:"omzet"`
}

type KategoriTotal struct {
	Nama  string `json:"nama"`
	Qty   int    `json:"qty"`
	Total int    `json:"total"` // sebelum diskon bill
}

type KinerjaKaryawan struct {
	UserID       int64  `json:"userId"`
	Nama         string `json:"nama"`
	Role         string `json:"role"`
	BillDibuat   int    `json:"billDibuat"`
	NilaiDibuat  int    `json:"nilaiDibuat"`
	BillDibayar  int    `json:"billDibayar"`
	NilaiDibayar int    `json:"nilaiDibayar"`
	Diskon       int    `json:"diskon"` // diskon pada bill yang pembayarannya dia terima
	BatalQty     int    `json:"batalQty"`
	BatalNilai   int    `json:"batalNilai"`
	Shift        int    `json:"shift"`
	SelisihKas   int    `json:"selisihKas"` // kas dihitung - kas sistem, total shift yang dia buka
}

type KasbonUmur struct {
	ID       int64      `json:"id"`
	Nama     string     `json:"nama"`
	Telp     string     `json:"telp"`
	Sisa     int        `json:"sisa"`
	Tertua   *time.Time `json:"tertua"` // kasbon tertua yang belum tertutup cicilan (FIFO)
	UmurHari int        `json:"umurHari"`
}

type StokNilai struct {
	ID         int64  `json:"id"`
	Nama       string `json:"nama"`
	Satuan     string `json:"satuan"`
	Stok       int    `json:"stok"`
	Nilai      int    `json:"nilai"`
	SusutQty   int    `json:"susutQty"`
	SusutNilai int    `json:"susutNilai"`
}
