package main

import (
	"os"

	"warkop/internal/config"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) != 3 || len(os.Args[2]) < 6 {
		log.Fatal().Msg("pemakaian: go run ./cmd/resetpw <username lengkap, mis. 1-owner> <password, minimal 6 karakter>")
	}
	username, password := os.Args[1], os.Args[2]

	config.Load()
	if err := config.ConnectDB(); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MySQL")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to hash password")
	}
	res, err := config.DB.Exec("UPDATE users SET password_hash=?, aktif=1 WHERE username=?", string(hash), username)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to update password")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		log.Fatal().Str("username", username).Msg("User tidak ditemukan. Username lengkap berawalan klien, mis. 1-owner")
	}
	log.Info().Str("username", username).Msg("Password diganti dan akun diaktifkan")
}
