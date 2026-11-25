package dto

import (
	"encoding/json"
	"time"
)

type PrinterRequest struct {
	Nama    string `json:"nama" validate:"required,max=100"`
	IP      string `json:"ip" validate:"required,max=100"`
	Port    int    `json:"port" validate:"required,min=1,max=65535"`
	LebarMM int    `json:"lebarMm" validate:"oneof=58 80"`
}

type PrinterResponse struct {
	ID      int64  `json:"id"`
	Nama    string `json:"nama"`
	IP      string `json:"ip"`
	Port    int    `json:"port"`
	LebarMM int    `json:"lebarMm"`
	Antri   int    `json:"antri"`
}

type TiketItem struct {
	Qty     int    `json:"qty"`
	Nama    string `json:"nama"`
	Opsi    string `json:"opsi"`
	Catatan string `json:"catatan"`
}

// TiketPayload: isi satu tiket untuk satu tenant. Stasiun cetak yang memformatnya menjadi perintah printer.
type TiketPayload struct {
	Judul    string      `json:"judul"`
	MejaNama string      `json:"mejaNama"`
	AtasNama string      `json:"atasNama"`
	Oleh     string      `json:"oleh"`
	Tenant   string      `json:"tenant"`
	Waktu    time.Time   `json:"waktu"`
	Items    []TiketItem `json:"items"`
}

type CetakJob struct {
	ID        int64      `json:"id"`
	Printer   string     `json:"printer"`
	Jenis     string     `json:"jenis"`
	Judul     string     `json:"judul"`
	Tenant    string     `json:"tenant"`
	MejaNama  string     `json:"mejaNama"`
	Status    string     `json:"status"`
	Percobaan int        `json:"percobaan"`
	Galat     string     `json:"galat"`
	DibuatAt  time.Time  `json:"dibuatAt"`
	SelesaiAt *time.Time `json:"selesaiAt"`
}

type CetakStatus struct {
	Antri        int          `json:"antri"`
	Macet        int          `json:"macet"`
	StasiunAktif bool         `json:"stasiunAktif"`
	Daftar       []CetakMacet `json:"daftar"`
}

// CetakMacet: tiket yang sudah lewat batas macet tapi belum tercetak, untuk peringatan di layar kasir/pelayan.
type CetakMacet struct {
	ID        int64           `json:"id"`
	Jenis     string          `json:"jenis"`
	Printer   string          `json:"printer"`
	Judul     string          `json:"judul"`
	Tenant    string          `json:"tenant"`
	MejaNama  string          `json:"mejaNama"`
	Percobaan int             `json:"percobaan"`
	Galat     string          `json:"galat"`
	DibuatAt  time.Time       `json:"dibuatAt"`
	Payload   json.RawMessage `json:"payload"` // isi tiket, untuk cetak cadangan lewat browser
}

type TokenStasiun struct {
	Token string `json:"token"`
}

type StasiunJob struct {
	ID      int64           `json:"id"`
	Jenis   string          `json:"jenis"`
	IP      string          `json:"ip"`
	Port    int             `json:"port"`
	LebarMM int             `json:"lebarMm"`
	Payload json.RawMessage `json:"payload"`
	Data    []byte          `json:"data"` // perintah printer siap kirim (base64 di JSON)
}

type GagalRequest struct {
	Galat string `json:"galat" validate:"max=255"`
}
