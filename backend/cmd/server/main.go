package main

import (
	"warkop/internal/app"
	"warkop/internal/config"

	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.Load()

	if err := config.ConnectDB(); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MySQL")
	}

	a := app.New(config.DB)
	log.Info().Msgf("Server is running on http://localhost:%s", cfg.Port)
	if err := a.Fiber.Listen(":" + cfg.Port); err != nil {
		log.Fatal().Err(err).Msg("Server crashed")
	}
}
