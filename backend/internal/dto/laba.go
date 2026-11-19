package dto

type PengeluaranKategori struct {
	Kategori string `json:"kategori"`
	Total    int    `json:"total"`
	DalamHPP bool   `json:"dalamHpp"` // belanja bahan tidak dihitung lagi sebagai pengeluaran karena sudah masuk HPP
}

type LabaRugiResponse struct {
	Dari             string                `json:"dari"`
	Sampai           string                `json:"sampai"`
	PenjualanSendiri int                   `json:"penjualanSendiri"` // penjualan tenant milik sendiri setelah diskon
	KomisiTenant     int                   `json:"komisiTenant"`     // komisi dari tenant lain
	Pendapatan       int                   `json:"pendapatan"`
	HPP              int                   `json:"hpp"`
	Susut            int                   `json:"susut"` // nilai selisih opname; negatif bila stok fisik lebih
	LabaKotor       int                   `json:"labaKotor"`
	OmzetTanpaResep  int                   `json:"omzetTanpaResep"` // penjualan sendiri yang menunya belum punya resep, HPP-nya belum terhitung
	Pengeluaran      []PengeluaranKategori `json:"pengeluaran"`
	TotalPengeluaran int                   `json:"totalPengeluaran"` // tanpa belanja bahan
	Gaji             int                   `json:"gaji"`
	LabaBersih       int                   `json:"labaBersih"`
}

type MarginMenu struct {
	MenuID       int64   `json:"menuId"`
	Nama         string  `json:"nama"`
	Terjual      int     `json:"terjual"`
	Penjualan    int     `json:"penjualan"` // sebelum diskon bill
	HPP          int     `json:"hpp"`
	Laba         int     `json:"laba"`
	MarginPersen float64 `json:"marginPersen"`
	AdaResep     bool    `json:"adaResep"`
}
