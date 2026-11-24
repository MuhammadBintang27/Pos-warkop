package services

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"warkop/internal/dto"
	"warkop/internal/models"
	"warkop/internal/types"
	"warkop/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

var reNamaPengguna = regexp.MustCompile(`^[a-z0-9._-]{3,30}$`)

type UserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) List(kid int64) ([]dto.UserResponse, error) {
	rows, err := s.db.Query("SELECT id,nama,username,role,aktif,COALESCE(tenant_id,0) FROM users WHERE klien_id=? AND dihapus IS NULL ORDER BY nama", kid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []dto.UserResponse{}
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Nama, &u.Username, &u.Role, &u.Aktif, &u.TenantID); err != nil {
			return nil, err
		}
		out = append(out, toUserResponse(u))
	}
	return out, rows.Err()
}

// Create membentuk username "<klien>-<nama>"; awalan selalu dari klien pembuatnya, bukan dari input.
func (s *UserService) Create(kid int64, req dto.CreateUserRequest) error {
	nama := strings.ToLower(strings.TrimSpace(req.Username))
	if !reNamaPengguna.MatchString(nama) {
		return types.BadRequest("Username 3-30 karakter: huruf kecil, angka, titik, garis bawah atau strip, tanpa spasi")
	}
	tenantID, err := s.tenantAkun(kid, req.Role, req.TenantID)
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT INTO users (klien_id,nama,username,password_hash,role,tenant_id) VALUES (?,?,?,?,?,?)",
		kid, req.Nama, fmt.Sprintf("%d-%s", kid, nama), string(hash), req.Role, tenantID)
	if utils.MySQLErrNo(err) == utils.MySQLDuplicate {
		return types.BadRequest("Username sudah dipakai")
	}
	return err
}

func (s *UserService) Update(kid, actorID, id int64, req dto.UpdateUserRequest) error {
	if id == actorID && (!*req.Aktif || types.Role(req.Role) != types.RoleOwner) {
		return types.BadRequest("Tidak bisa menonaktifkan atau menurunkan role akun sendiri")
	}
	tenantID, err := s.tenantAkun(kid, req.Role, req.TenantID)
	if err != nil {
		return err
	}
	var roleLama string
	var aktifLama bool
	var tenantLama int64
	if s.db.QueryRow("SELECT role,aktif,COALESCE(tenant_id,0) FROM users WHERE id=? AND klien_id=? AND dihapus IS NULL", id, kid).Scan(&roleLama, &aktifLama, &tenantLama) != nil {
		return types.NotFound("User")
	}
	if _, err := s.db.Exec("UPDATE users SET nama=?, role=?, aktif=?, tenant_id=? WHERE id=? AND klien_id=?", req.Nama, req.Role, *req.Aktif, tenantID, id, kid); err != nil {
		return err
	}
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if _, err := s.db.Exec("UPDATE users SET password_hash=? WHERE id=? AND klien_id=?", string(hash), id, kid); err != nil {
			return err
		}
	}
	// ganti nama saja tidak perlu memaksa login ulang
	if req.Password != "" || roleLama != req.Role || aktifLama != *req.Aktif || tenantLama != tenantID.Int64 {
		return cabutToken(s.db, id)
	}
	return nil
}

// Delete menghapus lunak: akun nonaktif dan disembunyikan, riwayat transaksinya tetap.
// Username diberi akhiran "~<id>" supaya username yang sama bisa dipakai akun baru.
func (s *UserService) Delete(kid, actorID, id int64) error {
	if id == actorID {
		return types.BadRequest("Tidak bisa menghapus akun sendiri")
	}
	res, err := s.db.Exec("UPDATE users SET aktif=0, dihapus=NOW(), username=CONCAT(username,'~',id) WHERE id=? AND klien_id=? AND dihapus IS NULL", id, kid)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return types.NotFound("User")
	}
	return cabutToken(s.db, id)
}

// cabutToken membuat semua token user yang sudah terbit tidak berlaku lagi.
func cabutToken(db *sql.DB, uid int64) error {
	sejak := time.Now().Truncate(time.Second) // iat JWT berpresisi detik
	if _, err := db.Exec("UPDATE users SET token_sejak=? WHERE id=?", sejak, uid); err != nil {
		return err
	}
	utils.CabutToken(uid, sejak)
	return nil
}

// tenantAkun: akun tenant wajib terhubung ke tenant aktif milik klien yang sama; role lain selalu tanpa tenant.
func (s *UserService) tenantAkun(kid int64, role string, tenantID int64) (sql.NullInt64, error) {
	if types.Role(role) != types.RoleTenant {
		return sql.NullInt64{}, nil
	}
	var ok int
	if tenantID <= 0 || s.db.QueryRow("SELECT 1 FROM tenant WHERE id=? AND klien_id=? AND aktif=1", tenantID, kid).Scan(&ok) != nil {
		return sql.NullInt64{}, types.BadRequest("Pilih tenant untuk akun tenant")
	}
	return sql.NullInt64{Int64: tenantID, Valid: true}, nil
}
