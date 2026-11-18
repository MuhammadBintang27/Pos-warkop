package services

import (
	"database/sql"
	"net/http"
	"strings"

	"warkop/internal/dto"
	"warkop/internal/models"
	"warkop/internal/types"
	"warkop/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db *sql.DB
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Login(req dto.LoginRequest) (*dto.LoginResponse, error) {
	var u models.User
	err := s.db.QueryRow(`SELECT u.id,u.klien_id,u.nama,u.username,u.password_hash,u.role,u.aktif
		FROM users u JOIN klien k ON k.id=u.klien_id WHERE u.username=?`, strings.ToLower(strings.TrimSpace(req.Username))).
		Scan(&u.ID, &u.KlienID, &u.Nama, &u.Username, &u.PasswordHash, &u.Role, &u.Aktif)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == sql.ErrNoRows {
		return s.loginSuper(req)
	}
	if !u.Aktif || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		return nil, types.NewHTTPError(http.StatusUnauthorized, "Username atau password salah")
	}
	if !utils.KlienAktif(u.KlienID) {
		return nil, types.NewHTTPError(http.StatusForbidden, "Akses usaha ini dinonaktifkan. Hubungi penyedia layanan")
	}
	token, err := utils.GenerateToken(u.ID, u.KlienID, u.Nama, u.Role)
	if err != nil {
		return nil, err
	}
	return &dto.LoginResponse{Token: token, User: toUserResponse(u)}, nil
}

func (s *AuthService) loginSuper(req dto.LoginRequest) (*dto.LoginResponse, error) {
	var id int64
	var username, hash string
	err := s.db.QueryRow("SELECT id,username,password_hash FROM superadmin WHERE username=?", strings.ToLower(strings.TrimSpace(req.Username))).Scan(&id, &username, &hash)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == sql.ErrNoRows || bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		return nil, types.NewHTTPError(http.StatusUnauthorized, "Username atau password salah")
	}
	token, err := utils.GenerateToken(id, 0, username, types.RoleSuper)
	if err != nil {
		return nil, err
	}
	return &dto.LoginResponse{Token: token, User: dto.UserResponse{ID: id, Nama: username, Username: username, Role: types.RoleSuper, Aktif: true}}, nil
}

func toUserResponse(u models.User) dto.UserResponse {
	return dto.UserResponse{ID: u.ID, KlienID: u.KlienID, Nama: u.Nama, Username: u.Username, Role: u.Role, Aktif: u.Aktif}
}
