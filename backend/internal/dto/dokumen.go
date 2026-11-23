package dto

import "time"

// Tag json judul/tenant/mejaNama sengaja sama dengan TiketPayload supaya riwayat cetak bisa menampilkan semua jenis tugas.

type StrukItem struct {
	Qty   int    `json:"qty"`
	Nama  string `json:"nama"`
	Opsi  string `json:"opsi"`
	Harga int    `json:"harga"`
}

type StrukBayar struct {
	Metode    string `json:"metode"`
	Pelanggan string `json:"pelanggan"`
	Diterima  int    `json:"diterima"`
	Kembalian int    `json:"kembalian"`
}

type StrukPayload struct {
	Judul    string       `json:"judul"`
	Tenant   string       `json:"tenant"`
	MejaNama string       `json:"mejaNama"`
	Usaha    string       `json:"usaha"`
	Alamat   string       `json:"alamat"`
	Footer   string       `json:"footer"`
	ID       int64        `json:"id"`
	AtasNama string       `json:"atasNama"`
	Kasir    string       `json:"kasir"`
	Waktu    time.Time    `json:"waktu"`
	Items    []StrukItem  `json:"items"`
	Subtotal int          `json:"subtotal"`
	Diskon   int          `json:"diskon"`
	Total    int          `json:"total"`
	Bayar    []StrukBayar `json:"bayar"`
}

type ClosingMenu struct {
	Qty   int    `json:"qty"`
	Nama  string `json:"nama"`
	Nilai int    `json:"nilai"`
}

// ClosingPayload dikirim dari layar closing yang sedang dilihat owner; nama usaha diambil server dari pengaturan.
type ClosingPayload struct {
	Judul         string        `json:"judul"`
	Tenant        string        `json:"tenant" validate:"required,max=100"`
	MejaNama      string        `json:"mejaNama"`
	Usaha         string        `json:"usaha"`
	Tanggal       string        `json:"tanggal" validate:"required,len=10"`
	MilikSendiri  bool          `json:"milikSendiri"`
	Transaksi     int           `json:"transaksi"`
	Kotor         int           `json:"kotor"`
	Diskon        int           `json:"diskon"`
	KomisiPersen  float64       `json:"komisiPersen"`
	KomisiPerMenu bool          `json:"komisiPerMenu"`
	Komisi        int           `json:"komisi"`
	Setor         int           `json:"setor"`
	BatalQty      int           `json:"batalQty"`
	BatalNilai    int           `json:"batalNilai"`
	TertundaBill  int           `json:"tertundaBill"`
	TertundaNilai int           `json:"tertundaNilai"`
	TalanganBill  int           `json:"talanganBill"` // bill belum dibayar yang ditalangi pemilik pada closing ini
	TalanganNilai int           `json:"talanganNilai"`
	DibuatOleh    string        `json:"dibuatOleh"`
	DibuatAt      *time.Time    `json:"dibuatAt"`
	DisetorOleh   string        `json:"disetorOleh"`
	DisetorAt     *time.Time    `json:"disetorAt"`
	Menu          []ClosingMenu `json:"menu" validate:"max=300,dive"`
}

type CetakHasil struct {
	Antre bool `json:"antre"` // false: printer kasir belum diatur, klien mencetak lewat browser
}

// RekapTenantPayload: satu struk untuk satu kali setor periode (mis. mingguan). Angka per hari dikirim apa adanya
// dari laporan harian (angka beku closing bila sudah di-closing); total dihitung server.
type RekapTenantPayload struct {
	Usaha         string      `json:"usaha"`
	Tenant        string      `json:"tenant" validate:"required,max=100"`
	Dari          string      `json:"dari" validate:"required,len=10"`
	Sampai        string      `json:"sampai" validate:"required,len=10"`
	MilikSendiri  bool        `json:"milikSendiri"`
	KomisiPersen  float64     `json:"komisiPersen"`
	KomisiPerMenu bool        `json:"komisiPerMenu"`
	Hari          []RekapHari `json:"hari" validate:"required,min=1,max=400,dive"`
}

type RekapHari struct {
	Tanggal   string `json:"tanggal" validate:"required,len=10"`
	Transaksi int    `json:"transaksi"`
	Kotor     int    `json:"kotor"`
	Diskon    int    `json:"diskon"`
	Komisi    int    `json:"komisi"`
	Setor     int    `json:"setor"`
	Status    string `json:"status"` // belum | closing | disetor | berubah | sendiri
}
