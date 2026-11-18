package dto

type PengaturanRequest struct {
	NamaUsaha   string `json:"namaUsaha" validate:"required,max=100"`
	Alamat      string `json:"alamat" validate:"max=255"`
	FooterStruk string `json:"footerStruk" validate:"max=255"`
	Theme       string `json:"theme" validate:"required,oneof=aura lara material nora"`
	Primary     string `json:"primary" validate:"required,primary_color"`
	Mode        string `json:"mode" validate:"required,oneof=light dark auto"`
	Logo        string `json:"logo" validate:"max=300000"` // data URL gambar, per klien
	// PrinterKasirID: printer yang mencetak struk dan bill closing; kosong berarti lewat dialog cetak browser.
	PrinterKasirID *int64 `json:"printerKasirId" validate:"omitempty,gt=0"`
}
