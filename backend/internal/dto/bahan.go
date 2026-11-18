package dto

import "time"

type BahanRequest struct {
	Nama    string `json:"nama" validate:"required,max=100"`
	Satuan  string `json:"satuan" validate:"required,max=10"`
	Stok    int    `json:"stok" validate:"gte=0"` // hanya dipakai saat membuat; selanjutnya lewat restock atau opname
	Minimum int    `json:"minimum" validate:"gte=0"`
	// HargaRata: rupiah per satuan. Diisi manual untuk stok awal; restock berharga menghitung ulang rata-ratanya.
	HargaRata float64 `json:"hargaRata" validate:"gte=0"`
}

type RestockRequest struct {
	Jumlah    int    `json:"jumlah" validate:"required,gt=0"`
	HargaBeli int    `json:"hargaBeli" validate:"gte=0"` // total harga pembelian, opsional
	Catatan   string `json:"catatan" validate:"max=255"`
}

type OpnameRequest struct {
	StokAktual int    `json:"stokAktual" validate:"gte=0"`
	Catatan    string `json:"catatan" validate:"max=255"`
}

type BahanResponse struct {
	ID        int64   `json:"id"`
	Nama      string  `json:"nama"`
	Satuan    string  `json:"satuan"`
	Stok      int     `json:"stok"`
	Minimum   int     `json:"minimum"`
	HargaRata float64 `json:"hargaRata"`
	Status    string  `json:"status"` // ok, menipis, habis
}

type BahanLogResponse struct {
	ID        int64     `json:"id"`
	Jenis     string    `json:"jenis"`
	Jumlah    int       `json:"jumlah"` // perubahan stok: negatif bila berkurang
	StokAkhir int       `json:"stokAkhir"`
	HargaBeli *int      `json:"hargaBeli"`
	Catatan   string    `json:"catatan"`
	Oleh      string    `json:"oleh"`
	At        time.Time `json:"at"`
}

type BahanPemakaian struct {
	ID           int64  `json:"id"`
	Nama         string `json:"nama"`
	Satuan       string `json:"satuan"`
	Dipakai      int    `json:"dipakai"`
	Restock      int    `json:"restock"`
	SelisihOpnam int    `json:"selisihOpname"` // jumlah koreksi dari opname; negatif berarti stok fisik lebih sedikit dari sistem
	StokSekarang int    `json:"stokSekarang"`
}

// ResepItem: takaran satu bahan. Pada resep dasar harus positif; pada opsi varian boleh negatif untuk mengurangi takaran.
type ResepItem struct {
	BahanID int64 `json:"bahanId" validate:"required,gt=0"`
	Jumlah  int   `json:"jumlah" validate:"required"`
}
