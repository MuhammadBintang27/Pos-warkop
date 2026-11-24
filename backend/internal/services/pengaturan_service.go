package services

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"time"

	"warkop/internal/dto"
	"warkop/internal/types"
)

type PengaturanService struct {
	db        *sql.DB
	hub       *EventHub
	uploadDir string
}

const kunciTampilan = "tampilan"

var defaultPengaturan = dto.PengaturanRequest{NamaUsaha: "Warkop", Theme: "aura", Primary: "amber", Mode: "light", DialogDapur: true}

func NewPengaturanService(db *sql.DB, hub *EventHub, uploadDir string) *PengaturanService {
	return &PengaturanService{db: db, hub: hub, uploadDir: uploadDir}
}

const MaksLogo = 1 << 20

// SimpanLogo menyimpan file logo ke folder unggahan dan mengembalikan URL-nya. Logo baru dipakai setelah pengaturan disimpan.
// shortcut: logo yang diunggah tapi tidak jadi disimpan tetap tertinggal di folder; bersihkan bila folder mulai besar.
func (s *PengaturanService) SimpanLogo(kid int64, raw []byte) (string, error) {
	if len(raw) > MaksLogo {
		return "", types.BadRequest("Logo maksimal 1 MB")
	}
	_, format, err := image.DecodeConfig(bytes.NewReader(raw))
	ext := map[string]string{"png": "png", "jpeg": "jpg"}[format]
	if err != nil || ext == "" {
		return "", types.BadRequest("Logo harus berupa gambar PNG atau JPG")
	}
	if err := os.MkdirAll(s.uploadDir, 0o755); err != nil {
		return "", err
	}
	nama := fmt.Sprintf("logo-%d-%d.%s", kid, time.Now().UnixNano(), ext)
	if err := os.WriteFile(filepath.Join(s.uploadDir, nama), raw, 0o644); err != nil {
		return "", err
	}
	return "/uploads/" + nama, nil
}

// fileLogo: path di disk untuk URL logo milik klien ini, kosong bila bukan logo unggahan klien tersebut.
func (s *PengaturanService) fileLogo(kid int64, url string) string {
	nama := strings.TrimPrefix(url, "/uploads/")
	if nama == url || nama != filepath.Base(nama) || !strings.HasPrefix(nama, fmt.Sprintf("logo-%d-", kid)) {
		return ""
	}
	return filepath.Join(s.uploadDir, nama)
}

// BacaLogo mengembalikan isi gambar logo klien untuk dicetak; nil bila tidak ada atau tidak terbaca.
func (s *PengaturanService) BacaLogo(kid int64, logo string) []byte {
	if i := strings.Index(logo, ";base64,"); strings.HasPrefix(logo, "data:") && i > 0 { // data lama sebelum unggah file
		raw, _ := base64.StdEncoding.DecodeString(logo[i+len(";base64,"):])
		return raw
	}
	if f := s.fileLogo(kid, logo); f != "" {
		raw, _ := os.ReadFile(f)
		return raw
	}
	return nil
}

func (s *PengaturanService) Get(kid int64) (*dto.PengaturanRequest, error) {
	var raw string
	out := defaultPengaturan
	err := s.db.QueryRow("SELECT nilai FROM pengaturan WHERE klien_id=? AND kunci=?", kid, kunciTampilan).Scan(&raw)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == nil {
		if err := json.Unmarshal([]byte(raw), &out); err != nil {
			return nil, err
		}
	}
	out.Kode = ""
	if kid > 0 && s.db.QueryRow("SELECT kode FROM klien WHERE id=?", kid).Scan(&out.Kode) != nil {
		out.Kode = ""
	}
	return &out, nil
}

func (s *PengaturanService) Update(kid int64, req dto.PengaturanRequest) error {
	if req.PrinterKasirID != nil {
		var ada int
		if s.db.QueryRow("SELECT 1 FROM printer WHERE id=? AND klien_id=? AND aktif=1", *req.PrinterKasirID, kid).Scan(&ada) != nil {
			return types.BadRequest("Printer kasir tidak ditemukan")
		}
	}
	lama, err := s.Get(kid)
	if err != nil {
		return err
	}
	if req.Logo != "" && req.Logo != lama.Logo && s.fileLogo(kid, req.Logo) == "" {
		return types.BadRequest("Logo tidak valid, unggah ulang")
	}
	req.Kode = ""
	raw, err := json.Marshal(req)
	if err != nil {
		return err
	}
	if _, err := s.db.Exec(`INSERT INTO pengaturan (klien_id,kunci,nilai) VALUES (?,?,?)
		ON DUPLICATE KEY UPDATE nilai=VALUES(nilai)`, kid, kunciTampilan, string(raw)); err != nil {
		return err
	}
	if f := s.fileLogo(kid, lama.Logo); f != "" && lama.Logo != req.Logo {
		_ = os.Remove(f) // logo lama tidak dipakai lagi
	}
	s.hub.Publish(kid, "pengaturan")
	return nil
}
