package main

import (
	"os"
	"strings"

	"warkop/internal/config"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) != 3 || len(os.Args[2]) < 8 {
		log.Fatal().Msg("pemakaian: go run ./cmd/superadmin <username> <password, minimal 8 karakter>")
	}
	config.Load()
	if err := config.ConnectDB(); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MySQL")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(os.Args[2]), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to hash password")
	}
	username := strings.ToLower(strings.TrimSpace(os.Args[1]))
	if _, err := config.DB.Exec(`INSERT INTO superadmin (username,password_hash) VALUES (?,?)
		ON DUPLICATE KEY UPDATE password_hash=VALUES(password_hash)`, username, string(hash)); err != nil {
		log.Fatal().Err(err).Msg("Gagal menyimpan superadmin")
	}
	log.Info().Str("username", username).Msg("Superadmin disimpan")
}
