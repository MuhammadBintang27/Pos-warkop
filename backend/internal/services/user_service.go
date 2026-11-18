package services

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

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
	rows, err := s.db.Query("SELECT id,nama,username,role,aktif FROM users WHERE klien_id=? ORDER BY nama", kid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []dto.UserResponse{}
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Nama, &u.Username, &u.Role, &u.Aktif); err != nil {
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
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT INTO users (klien_id,nama,username,password_hash,role) VALUES (?,?,?,?,?)",
		kid, req.Nama, fmt.Sprintf("%d-%s", kid, nama), string(hash), req.Role)
	if utils.MySQLErrNo(err) == utils.MySQLDuplicate {
		return types.BadRequest("Username sudah dipakai")
	}
	return err
}

func (s *UserService) Update(kid, actorID, id int64, req dto.UpdateUserRequest) error {
	if id == actorID && (!*req.Aktif || types.Role(req.Role) != types.RoleOwner) {
		return types.BadRequest("Tidak bisa menonaktifkan atau menurunkan role akun sendiri")
	}
	res, err := s.db.Exec("UPDATE users SET nama=?, role=?, aktif=? WHERE id=? AND klien_id=?", req.Nama, req.Role, *req.Aktif, id, kid)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		var exists int
		if s.db.QueryRow("SELECT 1 FROM users WHERE id=? AND klien_id=?", id, kid).Scan(&exists) != nil {
			return types.NotFound("User")
		}
	}
	if req.Password == "" {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("UPDATE users SET password_hash=? WHERE id=? AND klien_id=?", string(hash), id, kid)
	return err
}
