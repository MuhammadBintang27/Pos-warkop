package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"warkop/internal/config"
	"warkop/internal/dto"
	"warkop/internal/types"
	"warkop/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

var reKode = regexp.MustCompile(`^[a-z0-9]{3,30}$`)

const (
	langgananBelum    = "belum"
	langgananAktif    = "aktif"
	langgananSegera   = "segera"
	langgananBerakhir = "berakhir"
	ambangSegera      = 7 * 24 * time.Hour
)

type KlienService struct {
	db *sql.DB
}

func NewKlienService(db *sql.DB) *KlienService {
	return &KlienService{db: db}
}

// statusLangganan: langganan berlaku sampai akhir tanggal "sampai" (inklusif).
func statusLangganan(sampai *time.Time, now time.Time) string {
	if sampai == nil {
		return langgananBelum
	}
	hariIni := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, sampai.Location())
	switch {
	case sampai.Before(hariIni):
		return langgananBerakhir
	case sampai.Sub(hariIni) <= ambangSegera:
		return langgananSegera
	}
	return langgananAktif
}

func passwordAcak() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashPassword(pw string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(h), err
}

// MuatAkses memuat klien yang aksesnya dimatikan ke memori, dipanggil sekali saat server mulai.
func (s *KlienService) MuatAkses() error {
	rows, err := s.db.Query("SELECT id FROM klien WHERE aktif=0")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		utils.SetKlienAktif(id, false)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	cabut, err := s.db.Query("SELECT id,token_sejak FROM users WHERE token_sejak IS NOT NULL")
	if err != nil {
		return err
	}
	defer cabut.Close()
	for cabut.Next() {
		var id int64
		var sejak time.Time
		if err := cabut.Scan(&id, &sejak); err != nil {
			return err
		}
		utils.CabutToken(id, sejak)
	}
	return cabut.Err()
}

// Buat membuat klien beserta tenant bawaan dan akun owner. Nomor klien = nomor terbesar yang ada + 1.
// Password owner diisi superadmin, atau dibuat acak dan dikembalikan sekali bila dikosongkan.
func (s *KlienService) Buat(req dto.KlienRequest) (*dto.KlienKredensial, error) {
	if !reKode.MatchString(req.Kode) {
		return nil, types.BadRequest("Kode klien 3-30 huruf kecil atau angka")
	}
	namaOwner := strings.ToLower(strings.TrimSpace(req.OwnerUsername))
	if namaOwner == "" {
		namaOwner = "owner"
	}
	if !reNamaPengguna.MatchString(namaOwner) {
		return nil, types.BadRequest("Username owner 3-30 karakter: huruf kecil, angka, titik, garis bawah atau strip")
	}
	tampil := ""
	pw := req.OwnerPassword
	if pw == "" {
		var err error
		if pw, err = passwordAcak(); err != nil {
			return nil, err
		}
		tampil = pw
	}
	hash, err := hashPassword(pw)
	if err != nil {
		return nil, err
	}
	ownerNama := strings.TrimSpace(req.OwnerNama)
	if ownerNama == "" {
		ownerNama = "Owner"
	}

	out := &dto.KlienKredensial{Password: tampil}
	err = utils.WithTx(s.db, func(tx *sql.Tx) error {
		var id int64
		if err := tx.QueryRow("SELECT COALESCE(MAX(id),0)+1 FROM klien FOR UPDATE").Scan(&id); err != nil {
			return err
		}
		_, err := tx.Exec(`INSERT INTO klien (id,kode,nama,alamat,langganan_sampai)
			VALUES (?,?,?,?, CASE WHEN ?>0 THEN DATE_ADD(CURDATE(), INTERVAL ? MONTH) END)`,
			id, req.Kode, strings.TrimSpace(req.Nama), strings.TrimSpace(req.Alamat), req.Bulan, req.Bulan)
		if utils.MySQLErrNo(err) == utils.MySQLDuplicate {
			return types.BadRequest("Kode klien sudah dipakai")
		}
		if err != nil {
			return err
		}
		out.KlienID, out.Username = id, fmt.Sprintf("%d-%s", id, namaOwner)
		if _, err = tx.Exec("INSERT INTO users (klien_id,nama,username,password_hash,role) VALUES (?,?,?,?,'owner')", id, ownerNama, out.Username, hash); err != nil {
			return err
		}
		if _, err = tx.Exec("INSERT INTO tenant (klien_id,nama,komisi_persen,milik_sendiri) VALUES (?,'Dapur Utama',100,1)", id); err != nil {
			return err
		}
		_, err = tx.Exec("INSERT INTO pengaturan (klien_id,kunci,nilai) VALUES (?,?,'')", id, config.KunciShift)
		return err
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Pastikan memeriksa klien ada sebelum pengguna dibuat atas namanya.
func (s *KlienService) Pastikan(id int64) error {
	var ada int
	if s.db.QueryRow("SELECT 1 FROM klien WHERE id=?", id).Scan(&ada) != nil {
		return types.NotFound("Klien")
	}
	return nil
}

func (s *KlienService) List() ([]dto.KlienResponse, error) {
	rows, err := s.db.Query(`SELECT k.id,k.kode,k.nama,k.alamat,k.aktif,k.langganan_sampai,
		(SELECT COUNT(*) FROM users u WHERE u.klien_id=k.id AND u.aktif=1),
		(SELECT COUNT(*) FROM pesanan p WHERE p.klien_id=k.id AND p.status='lunas'),
		(SELECT MAX(p.lunas_at) FROM pesanan p WHERE p.klien_id=k.id)
		FROM klien k ORDER BY k.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []dto.KlienResponse{}
	now := time.Now()
	for rows.Next() {
		var k dto.KlienResponse
		if err := rows.Scan(&k.ID, &k.Kode, &k.Nama, &k.Alamat, &k.Aktif, &k.LanggananSampai, &k.Pengguna, &k.Transaksi, &k.TerakhirAt); err != nil {
			return nil, err
		}
		k.Langganan = statusLangganan(k.LanggananSampai, now)
		out = append(out, k)
	}
	return out, rows.Err()
}

func (s *KlienService) Edit(id int64, req dto.KlienEditRequest) error {
	if _, err := s.db.Exec("UPDATE klien SET nama=?, alamat=? WHERE id=?", strings.TrimSpace(req.Nama), strings.TrimSpace(req.Alamat), id); err != nil {
		return err
	}
	return s.Pastikan(id)
}

// SetAktif mematikan atau membuka akses klien: pengguna yang sedang login pun langsung ditolak. Datanya tetap utuh.
func (s *KlienService) SetAktif(id int64, aktif bool) error {
	if _, err := s.db.Exec("UPDATE klien SET aktif=? WHERE id=?", aktif, id); err != nil {
		return err
	}
	if err := s.Pastikan(id); err != nil {
		return err
	}
	utils.SetKlienAktif(id, aktif)
	return nil
}

// Perpanjang menambah langganan dari tanggal berakhir saat ini, atau dari hari ini bila sudah lewat atau belum ada.
func (s *KlienService) Perpanjang(id int64, bulan int) error {
	if _, err := s.db.Exec(`UPDATE klien SET langganan_sampai=DATE_ADD(GREATEST(COALESCE(langganan_sampai,CURDATE()),CURDATE()), INTERVAL ? MONTH) WHERE id=?`, bulan, id); err != nil {
		return err
	}
	return s.Pastikan(id)
}

// ResetOwner mengganti password owner pertama klien dengan yang acak dan mengaktifkannya lagi.
func (s *KlienService) ResetOwner(id int64) (*dto.KlienKredensial, error) {
	var uid int64
	var username string
	if s.db.QueryRow("SELECT id,username FROM users WHERE klien_id=? AND role='owner' AND dihapus IS NULL ORDER BY id LIMIT 1", id).Scan(&uid, &username) != nil {
		return nil, types.NotFound("Owner klien")
	}
	pw, err := passwordAcak()
	if err != nil {
		return nil, err
	}
	hash, err := hashPassword(pw)
	if err != nil {
		return nil, err
	}
	if _, err := s.db.Exec("UPDATE users SET password_hash=?, aktif=1 WHERE id=?", hash, uid); err != nil {
		return nil, err
	}
	if err := cabutToken(s.db, uid); err != nil {
		return nil, err
	}
	return &dto.KlienKredensial{KlienID: id, Username: username, Password: pw}, nil
}
