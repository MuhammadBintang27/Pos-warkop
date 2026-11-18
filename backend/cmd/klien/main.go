package main

import (
	"os"

	"warkop/internal/config"
	"warkop/internal/dto"
	"warkop/internal/services"

	"github.com/rs/zerolog/log"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal().Msg(`pemakaian: go run ./cmd/klien <kode, huruf kecil/angka> "<nama usaha>"`)
	}
	config.Load()
	if err := config.ConnectDB(); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MySQL")
	}
	kr, err := services.NewKlienService(config.DB).Buat(dto.KlienRequest{Kode: os.Args[1], Nama: os.Args[2]})
	if err != nil {
		log.Fatal().Err(err).Msg("Gagal membuat klien")
	}
	log.Info().Int64("klienId", kr.KlienID).Str("username", kr.Username).Str("password", kr.Password).Msg("Klien dibuat, segera ganti password owner")
}
