package dto

// Menu untuk pelanggan (tanpa login): hanya yang boleh dilihat publik, tanpa harga modal, komisi, stok, atau resep.

type MenuPublikResponse struct {
	Usaha  UsahaPublik    `json:"usaha"`
	Tenant []TenantPublik `json:"tenant"`
}

type UsahaPublik struct {
	Nama    string `json:"nama"`
	Alamat  string `json:"alamat"`
	Logo    string `json:"logo"`
	Theme   string `json:"theme"`
	Primary string `json:"primary"`
	Mode    string `json:"mode"`
}

type TenantPublik struct {
	Nama string           `json:"nama"`
	Buka bool             `json:"buka"`
	Menu []MenuPublikItem `json:"menu"`
}

type MenuPublikItem struct {
	Nama     string         `json:"nama"`
	Kategori string         `json:"kategori"`
	Harga    int            `json:"harga"`
	Habis    bool           `json:"habis"`
	Varian   []VarianPublik `json:"varian"`
}

type VarianPublik struct {
	Nama       string       `json:"nama"`
	Wajib      bool         `json:"wajib"`
	GantiHarga bool         `json:"gantiHarga"` // harga opsi menggantikan harga menu, bukan tambahan
	Opsi       []OpsiPublik `json:"opsi"`
}

type OpsiPublik struct {
	Nama     string `json:"nama"`
	Tambahan int    `json:"tambahan"`
}
