package config

import (
	"database/sql"
	"fmt"
)

// kolomBaru: kolom yang ditambahkan setelah tabelnya sudah ada di DB klien. schema.sql hanya membuat tabel baru,
// jadi kolom baru juga harus didaftarkan di sini; dijalankan setiap start dan dilewati bila kolom sudah ada.
// shortcut: hanya ADD COLUMN; ubah tipe/hapus kolom/index baru masih perlu ALTER manual atau migrasi bernomor.
var kolomBaru = []struct{ tabel, kolom, definisi string }{
	{"menu", "harga_modal", "INT NOT NULL DEFAULT 0 AFTER harga"},
	{"pesanan_item", "modal", "INT NOT NULL DEFAULT 0 AFTER qty"},
	{"bahan_log", "biaya", "DECIMAL(14,4) NULL AFTER harga_beli"},
	{"users", "token_sejak", "DATETIME NULL AFTER aktif"},
	{"meja", "area_id", "BIGINT NULL AFTER nama"},
	{"tenant", "komisi_per_menu", "TINYINT(1) NOT NULL DEFAULT 0 AFTER komisi_persen"},
	{"menu", "komisi", "INT NOT NULL DEFAULT 0 AFTER harga_modal"},
	{"pesanan_item", "komisi", "INT NOT NULL DEFAULT 0 AFTER modal"},
	{"closing_tenant", "komisi_per_menu", "TINYINT(1) NOT NULL DEFAULT 0 AFTER komisi_persen"},
	{"closing_tenant_bill", "komisi", "INT NOT NULL DEFAULT 0 AFTER diskon"},
	{"users", "dihapus", "DATETIME NULL AFTER token_sejak"},
	{"pesanan_item", "diproses_at", "DATETIME NULL AFTER closing_id"},
	{"cetak_job", "diabaikan_oleh", "VARCHAR(100) NULL AFTER selesai_at"},
	{"tenant", "warna", "VARCHAR(7) NOT NULL DEFAULT '' AFTER printer_id"},
}

func tambahKolom(db *sql.DB) error {
	for _, k := range kolomBaru {
		var ada int
		if err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME=? AND COLUMN_NAME=?`, k.tabel, k.kolom).Scan(&ada); err != nil {
			return err
		}
		if ada > 0 {
			continue
		}
		if _, err := db.Exec(fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s", k.tabel, k.kolom, k.definisi)); err != nil {
			return fmt.Errorf("tambah kolom %s.%s: %w", k.tabel, k.kolom, err)
		}
	}
	return nil
}
