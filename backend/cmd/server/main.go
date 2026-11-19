package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

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
	go func() {
		log.Info().Msgf("Server is running on http://localhost:%s", cfg.Port)
		if err := a.Fiber.Listen(":" + cfg.Port); err != nil {
			log.Fatal().Err(err).Msg("Server crashed")
		}
	}()

	// saat redeploy, Docker kirim SIGTERM: selesaikan request yang sedang jalan dulu (SSE diputus setelah batas waktu)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Info().Msg("Shutting down")
	if err := a.Fiber.ShutdownWithTimeout(8 * time.Second); err != nil {
		log.Error().Err(err).Msg("Shutdown error")
	}
	config.DB.Close()
}
