package config

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

//go:embed schema.sql
var schemaSQL string

// KunciShift: baris di tabel pengaturan yang dikunci saat membuka shift supaya tidak pernah ada dua shift terbuka.
const KunciShift = "kunci_shift"

var DB *sql.DB

func ConnectDB() error {
	c := Get()
	dsn, err := mysql.ParseDSN(c.DBDSN)
	if err != nil {
		return fmt.Errorf("DB_DSN tidak valid: %w", err)
	}
	dsn.ParseTime, dsn.Loc = true, time.Local

	db, err := sql.Open("mysql", dsn.FormatDSN())
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err := db.Ping(); err != nil {
		return fmt.Errorf("gagal ping mysql: %w", err)
	}

	// Skema dijalankan apa adanya (CREATE TABLE IF NOT EXISTS), lalu kolom baru ditambahkan lewat tambahKolom.
	// multiStatements hanya dipakai di sini.
	migrate := *dsn
	migrate.MultiStatements = true
	mdb, err := sql.Open("mysql", migrate.FormatDSN())
	if err != nil {
		return err
	}
	defer mdb.Close()
	if _, err := mdb.Exec(schemaSQL); err != nil {
		return fmt.Errorf("migrasi gagal: %w", err)
	}
	if err := tambahKolom(db); err != nil {
		return fmt.Errorf("migrasi gagal: %w", err)
	}

	DB = db
	log.Info().Str("database", dsn.DBName).Msg("MySQL connection established")
	return nil
}
