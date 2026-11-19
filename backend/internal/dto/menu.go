package dto

type KategoriRequest struct {
	Nama   string `json:"nama" validate:"required,max=100"`
	Urutan int    `json:"urutan"`
}

type OpsiRequest struct {
	Nama     string      `json:"nama" validate:"required,max=50"`
	Tambahan int         `json:"tambahan" validate:"gte=0"`
	Bahan    []ResepItem `json:"bahan" validate:"max=50,dive"` // tambahan takaran bahan bila opsi ini dipilih
}

type GrupRequest struct {
	Nama       string        `json:"nama" validate:"required,max=50"`
	Wajib      bool          `json:"wajib"`
	Multi      bool          `json:"multi"`
	GantiHarga bool          `json:"gantiHarga"`
	Opsi       []OpsiRequest `json:"opsi" validate:"required,min=1,dive"`
}

type MenuRequest struct {
	TenantID   int64         `json:"tenantId" validate:"required,gt=0"`
	KategoriID int64         `json:"kategoriId" validate:"required,gt=0"`
	Nama       string        `json:"nama" validate:"required,max=100"`
	Harga      int           `json:"harga" validate:"gte=0"`
	HargaModal int           `json:"hargaModal" validate:"gte=0"`
	Tersedia   bool          `json:"tersedia"`
	TrackStok  bool          `json:"trackStok"`
	Stok       int           `json:"stok" validate:"gte=0"`
	Varian     []GrupRequest `json:"varian" validate:"dive"`
	Resep      []ResepItem   `json:"resep" validate:"max=50,dive"`
}

type StokRequest struct {
	Tersedia bool `json:"tersedia"`
	Stok     *int `json:"stok" validate:"omitempty,gte=0"`
}

type OpsiResponse struct {
	ID       int64       `json:"id"`
	Nama     string      `json:"nama"`
	Tambahan int         `json:"tambahan"`
	Bahan    []ResepItem `json:"bahan"`
}

type GrupResponse struct {
	ID         int64          `json:"id"`
	Nama       string         `json:"nama"`
	Wajib      bool           `json:"wajib"`
	Multi      bool           `json:"multi"`
	GantiHarga bool           `json:"gantiHarga"`
	Opsi       []OpsiResponse `json:"opsi"`
}

type MenuResponse struct {
	TenantID   int64           `json:"tenantId"`
	ID         int64           `json:"id"`
	KategoriID int64           `json:"kategoriId"`
	Nama       string          `json:"nama"`
	Harga      int             `json:"harga"`
	HargaModal int             `json:"hargaModal"`
	Tersedia   bool            `json:"tersedia"`
	TrackStok  bool            `json:"trackStok"`
	Stok       int             `json:"stok"`
	Varian     []*GrupResponse `json:"varian"`
	Resep      []ResepItem     `json:"resep"`
	PorsiBahan *int            `json:"porsiBahan"` // porsi yang masih bisa dibuat dari stok bahan; kosong bila menu tanpa resep
}

type KategoriResponse struct {
	ID     int64           `json:"id"`
	Nama   string          `json:"nama"`
	Urutan int             `json:"urutan"`
	Menu   []*MenuResponse `json:"menu"`
}

type MejaRequest struct {
	Nama string `json:"nama" validate:"required,max=50"`
}

type GenerateMejaRequest struct {
	Awalan string `json:"awalan" validate:"required,max=40"`
	Mulai  int    `json:"mulai" validate:"required,min=1,max=9999"`
	Jumlah int    `json:"jumlah" validate:"required,min=1,max=200"`
}

type GenerateMejaResponse struct {
	Dibuat   int `json:"dibuat"`
	Dilewati int `json:"dilewati"`
}

type PesananRingkas struct {
	ID       int64  `json:"id"`
	AtasNama string `json:"atasNama"`
	Total    int    `json:"total"`
}

type MejaResponse struct {
	ID      int64            `json:"id"`
	Nama    string           `json:"nama"`
	PosX    *float64         `json:"posX"`
	PosY    *float64         `json:"posY"`
	Pesanan []PesananRingkas `json:"pesanan"`
}

type PosisiMeja struct {
	ID   int64   `json:"id" validate:"required,gt=0"`
	PosX float64 `json:"posX" validate:"gte=0,lte=100"`
	PosY float64 `json:"posY" validate:"gte=0,lte=400"`
}

type PosisiRequest struct {
	Meja []PosisiMeja `json:"meja" validate:"required,min=1,max=500,dive"`
}

type DenahResponse struct {
	Meja     []*MejaResponse  `json:"meja"`
	Takeaway []PesananRingkas `json:"takeaway"`
}
