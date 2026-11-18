package config

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Port           string
	AppEnv         string
	DBDSN          string
	JWTSecret      string
	JWTExpiryHours int
	TZName         string
	StaticDir      string
	// FCMServiceAccount: isi JSON service account Firebase; kosong = push notifikasi tenant mati.
	FCMServiceAccount string
}

var (
	cfg     *Config
	cfgOnce sync.Once
)

func (c *Config) IsProduction() bool { return c.AppEnv == "production" }

func Load() *Config {
	cfgOnce.Do(func() {
		_ = godotenv.Load()
		cfg = &Config{
			Port:              getEnv("PORT", "8080"),
			AppEnv:            getEnv("APP_ENV", "development"),
			DBDSN:             os.Getenv("DB_DSN"),
			JWTSecret:         os.Getenv("JWT_SECRET"),
			JWTExpiryHours:    getEnvInt("JWT_EXPIRY_HOURS", 12),
			TZName:            getEnv("TZ_NAME", "Asia/Jakarta"),
			StaticDir:         getEnv("STATIC_DIR", "../frontend/dist"),
			FCMServiceAccount: os.Getenv("FCM_SERVICE_ACCOUNT"),
		}
		if cfg.DBDSN == "" || len(cfg.JWTSecret) < 16 {
			log.Fatal().Msg("DB_DSN wajib diisi dan JWT_SECRET minimal 16 karakter")
		}
		loc, err := time.LoadLocation(cfg.TZName)
		if err != nil {
			log.Fatal().Err(err).Msg("TZ_NAME tidak valid")
		}
		time.Local = loc
	})
	return cfg
}

func Get() *Config { return Load() }

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return v
	}
	return def
}
