-- Satu database dipakai semua klien (warkop). Setiap tabel memuat klien_id; tidak ada ALTER, bila skema berubah buat ulang database.
CREATE TABLE IF NOT EXISTS klien (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  kode VARCHAR(30) NOT NULL UNIQUE,  -- awalan username: <id>-nama
  nama VARCHAR(100) NOT NULL,
  alamat VARCHAR(255) NOT NULL DEFAULT '',
  aktif TINYINT(1) NOT NULL DEFAULT 1,  -- akses dimatikan manual oleh superadmin
  langganan_sampai DATE NULL            -- NULL: belum ada langganan tercatat
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Pemilik SaaS: tidak terikat klien mana pun, hanya mengelola daftar klien.
CREATE TABLE IF NOT EXISTS superadmin (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(50) NOT NULL UNIQUE,
  password_hash VARCHAR(100) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS users (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  nama VARCHAR(100) NOT NULL,
  username VARCHAR(50) NOT NULL UNIQUE,
  password_hash VARCHAR(100) NOT NULL,
  role ENUM('pelayan','kasir','owner','tenant') NOT NULL,
  tenant_id BIGINT NULL,        -- hanya untuk role tenant: tenant yang pesanan dan rekapannya boleh dilihat
  aktif TINYINT(1) NOT NULL DEFAULT 1,
  token_sejak DATETIME NULL,    -- token login yang terbit sebelum ini ditolak (nonaktif, ganti role, ganti password)
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS kategori (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  nama VARCHAR(100) NOT NULL,
  urutan INT NOT NULL DEFAULT 0,
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS tenant (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  nama VARCHAR(100) NOT NULL,
  komisi_persen DECIMAL(5,2) NOT NULL DEFAULT 0,  -- bagian pemilik tempat dari penjualan tenant
  buka TINYINT(1) NOT NULL DEFAULT 1,             -- tenant tutup: semua menunya tidak bisa dipesan
  milik_sendiri TINYINT(1) NOT NULL DEFAULT 0,    -- milik pemilik tempat: komisi 100%, tidak ada setoran
  printer_id BIGINT NULL,                         -- printer tiket dapur tenant ini
  aktif TINYINT(1) NOT NULL DEFAULT 1,
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS menu (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  kategori_id BIGINT NOT NULL,
  nama VARCHAR(100) NOT NULL,
  harga INT NOT NULL,
  harga_modal INT NOT NULL DEFAULT 0,  -- HPP per porsi untuk menu tanpa resep (barang jadi: rokok, kerupuk)
  tersedia TINYINT(1) NOT NULL DEFAULT 1,
  track_stok TINYINT(1) NOT NULL DEFAULT 0,
  stok INT NOT NULL DEFAULT 0,
  aktif TINYINT(1) NOT NULL DEFAULT 1,
  tenant_id BIGINT NULL,
  FOREIGN KEY (kategori_id) REFERENCES kategori(id),
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS varian_grup (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  menu_id BIGINT NOT NULL,
  nama VARCHAR(50) NOT NULL,
  wajib TINYINT(1) NOT NULL DEFAULT 0,
  multi TINYINT(1) NOT NULL DEFAULT 0,
  ganti_harga TINYINT(1) NOT NULL DEFAULT 0,  -- bila 1, nilai opsi = harga lengkap, bukan tambahan
  FOREIGN KEY (menu_id) REFERENCES menu(id) ON DELETE CASCADE,
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS varian_opsi (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  grup_id BIGINT NOT NULL,
  nama VARCHAR(50) NOT NULL,
  tambahan INT NOT NULL DEFAULT 0,
  FOREIGN KEY (grup_id) REFERENCES varian_grup(id) ON DELETE CASCADE,
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS meja (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  nama VARCHAR(50) NOT NULL,
  aktif TINYINT(1) NOT NULL DEFAULT 1,
  pos_x DECIMAL(6,2) NULL,
  pos_y DECIMAL(6,2) NULL,
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS shift (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  kasir_id BIGINT NOT NULL,
  buka_at DATETIME NOT NULL,
  tutup_at DATETIME NULL,
  kas_awal INT NOT NULL,
  kas_sistem INT NULL,
  kas_hitung INT NULL,
  tutup_oleh BIGINT NULL,
  FOREIGN KEY (kasir_id) REFERENCES users(id),
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS pelanggan (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  nama VARCHAR(100) NOT NULL,
  telp VARCHAR(30) NOT NULL DEFAULT '',
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS pesanan (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  meja_id BIGINT NULL,
  atas_nama VARCHAR(100) NOT NULL DEFAULT '',
  status ENUM('open','lunas','batal','gabung') NOT NULL DEFAULT 'open',
  diskon INT NOT NULL DEFAULT 0,
  shift_id BIGINT NULL,
  dibuat_oleh BIGINT NOT NULL,
  dibuat_at DATETIME NOT NULL,
  lunas_at DATETIME NULL,
  INDEX (status),
  INDEX (lunas_at),
  FOREIGN KEY (meja_id) REFERENCES meja(id),
  FOREIGN KEY (shift_id) REFERENCES shift(id),
  FOREIGN KEY (dibuat_oleh) REFERENCES users(id),
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS pesanan_item (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  pesanan_id BIGINT NOT NULL,
  menu_id BIGINT NOT NULL,
  nama VARCHAR(100) NOT NULL,
  harga INT NOT NULL,            -- harga satuan termasuk tambahan varian, disalin saat dipesan
  qty INT NOT NULL,
  modal INT NOT NULL DEFAULT 0,  -- salinan menu.harga_modal saat dipesan, hanya untuk menu tanpa resep
  opsi VARCHAR(255) NOT NULL DEFAULT '',
  catatan VARCHAR(255) NOT NULL DEFAULT '',
  dikirim TINYINT(1) NOT NULL DEFAULT 0,
  tenant_id BIGINT NULL,        -- disalin dari menu saat dipesan
  closing_id BIGINT NULL,       -- terisi bila sudah ditalangi di closing tenant sebelum bill dibayar; tidak dihitung lagi saat lunas
  siap_at DATETIME NULL,        -- ditandai siap oleh akun tenant; hanya catatan tenant
  dikirim_at DATETIME NULL,     -- satu kali Kirim = satu tiket; dipakai kartu pesanan masuk tenant
  dikirim_oleh VARCHAR(100) NOT NULL DEFAULT '',
  INDEX (pesanan_id),
  INDEX (closing_id),
  FOREIGN KEY (pesanan_id) REFERENCES pesanan(id),
  FOREIGN KEY (menu_id) REFERENCES menu(id),
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS pembayaran (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  pesanan_id BIGINT NOT NULL,
  metode ENUM('tunai','qris','transfer','kasbon') NOT NULL,
  jumlah INT NOT NULL,
  diterima INT NOT NULL,
  kembalian INT NOT NULL DEFAULT 0,
  pelanggan_id BIGINT NULL,
  diterima_oleh BIGINT NULL,
  shift_id BIGINT NOT NULL,
  at DATETIME NOT NULL,
  INDEX (pesanan_id),
  INDEX (at),
  FOREIGN KEY (pesanan_id) REFERENCES pesanan(id),
  FOREIGN KEY (pelanggan_id) REFERENCES pelanggan(id),
  FOREIGN KEY (shift_id) REFERENCES shift(id),
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS kasbon_bayar (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  pelanggan_id BIGINT NOT NULL,
  jumlah INT NOT NULL,
  metode ENUM('tunai','qris','transfer') NOT NULL,
  shift_id BIGINT NOT NULL,
  at DATETIME NOT NULL,
  FOREIGN KEY (pelanggan_id) REFERENCES pelanggan(id),
  FOREIGN KEY (shift_id) REFERENCES shift(id),
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS pengeluaran (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  shift_id BIGINT NOT NULL,
  keterangan VARCHAR(255) NOT NULL,
  jumlah INT NOT NULL,
  kategori VARCHAR(20) NOT NULL DEFAULT 'lainnya',
  dicatat_oleh BIGINT NULL,
  at DATETIME NOT NULL,
  INDEX (at),
  FOREIGN KEY (shift_id) REFERENCES shift(id),
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS pengaturan (
  klien_id BIGINT NOT NULL,
  kunci VARCHAR(50) NOT NULL,
  nilai TEXT NOT NULL,
  PRIMARY KEY (klien_id, kunci),
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS pembatalan (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  pesanan_id BIGINT NOT NULL,
  nama VARCHAR(100) NOT NULL,
  opsi VARCHAR(255) NOT NULL DEFAULT '',
  harga INT NOT NULL,
  qty INT NOT NULL,
  alasan VARCHAR(100) NOT NULL,
  sudah_dikirim TINYINT(1) NOT NULL,
  sudah_dibuat TINYINT(1) NOT NULL DEFAULT 0,
  tenant_id BIGINT NULL,
  oleh BIGINT NOT NULL,
  at DATETIME NOT NULL,
  INDEX (pesanan_id),
  FOREIGN KEY (pesanan_id) REFERENCES pesanan(id),
  FOREIGN KEY (oleh) REFERENCES users(id),
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Closing (tutup bill) tenant per tanggal: angka dibekukan saat closing supaya setoran tidak berubah diam-diam.
CREATE TABLE IF NOT EXISTS closing_tenant (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  tenant_id BIGINT NOT NULL,
  tanggal DATE NOT NULL,
  transaksi INT NOT NULL,
  kotor INT NOT NULL,
  diskon INT NOT NULL,
  komisi_persen DECIMAL(5,2) NOT NULL,
  komisi INT NOT NULL,
  setor INT NOT NULL,
  dibuat_oleh BIGINT NOT NULL,
  dibuat_at DATETIME NOT NULL,
  disetor_at DATETIME NULL,
  disetor_oleh BIGINT NULL,
  UNIQUE KEY uq_closing (klien_id, tenant_id, tanggal),
  FOREIGN KEY (tenant_id) REFERENCES tenant(id),
  FOREIGN KEY (dibuat_oleh) REFERENCES users(id),
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Printer thermal di jaringan warkop; dicetak oleh stasiun cetak (aplikasi desktop di PC kasir).
CREATE TABLE IF NOT EXISTS printer (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  nama VARCHAR(100) NOT NULL,
  ip VARCHAR(100) NOT NULL,
  port INT NOT NULL DEFAULT 9100,
  lebar_mm INT NOT NULL DEFAULT 58,
  aktif TINYINT(1) NOT NULL DEFAULT 1,
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Antrean tugas cetak: disimpan di server supaya tiket tidak hilang bila printer atau stasiun sedang mati.
CREATE TABLE IF NOT EXISTS cetak_job (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  printer_id BIGINT NOT NULL,
  jenis VARCHAR(10) NOT NULL,
  payload TEXT NOT NULL,
  data MEDIUMBLOB NULL,
  status ENUM('antri','selesai') NOT NULL DEFAULT 'antri',
  percobaan INT NOT NULL DEFAULT 0,
  galat VARCHAR(255) NOT NULL DEFAULT '',
  dibuat_at DATETIME NOT NULL,
  selesai_at DATETIME NULL,
  INDEX (status, dibuat_at),
  FOREIGN KEY (printer_id) REFERENCES printer(id),
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Stok bahan baku. Jumlah selalu bilangan bulat dalam satuan terkecil (gram, ml, pcs).
CREATE TABLE IF NOT EXISTS bahan (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  nama VARCHAR(100) NOT NULL,
  satuan VARCHAR(10) NOT NULL,
  stok INT NOT NULL DEFAULT 0,
  minimum INT NOT NULL DEFAULT 0,  -- di bawah atau sama dengan ini ditandai menipis; 0 = tanpa peringatan
  harga_rata DECIMAL(14,4) NOT NULL DEFAULT 0,  -- rupiah per satuan, rata-rata bergerak dari restock berharga
  aktif TINYINT(1) NOT NULL DEFAULT 1,
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS resep (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  menu_id BIGINT NOT NULL,
  bahan_id BIGINT NOT NULL,
  jumlah INT NOT NULL,
  UNIQUE KEY uq_resep (menu_id, bahan_id),
  FOREIGN KEY (menu_id) REFERENCES menu(id) ON DELETE CASCADE,
  FOREIGN KEY (bahan_id) REFERENCES bahan(id),
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Tambahan (atau pengurangan, bila negatif) takaran bahan untuk satu opsi varian.
CREATE TABLE IF NOT EXISTS resep_opsi (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  opsi_id BIGINT NOT NULL,
  bahan_id BIGINT NOT NULL,
  jumlah INT NOT NULL,
  UNIQUE KEY uq_resep_opsi (opsi_id, bahan_id),
  FOREIGN KEY (opsi_id) REFERENCES varian_opsi(id) ON DELETE CASCADE,
  FOREIGN KEY (bahan_id) REFERENCES bahan(id),
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Takaran per porsi yang disalin saat item dipesan, supaya pembatalan mengembalikan jumlah yang sama walau resep berubah.
CREATE TABLE IF NOT EXISTS item_bahan (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  item_id BIGINT NOT NULL,
  bahan_id BIGINT NOT NULL,
  jumlah INT NOT NULL,
  biaya DECIMAL(14,4) NOT NULL DEFAULT 0,  -- harga rata-rata per satuan saat dipesan, dasar HPP
  INDEX (item_id),
  FOREIGN KEY (item_id) REFERENCES pesanan_item(id) ON DELETE CASCADE,
  FOREIGN KEY (bahan_id) REFERENCES bahan(id),
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS bahan_log (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  bahan_id BIGINT NOT NULL,
  jenis ENUM('restock','opname','pakai','kembali') NOT NULL,
  jumlah INT NOT NULL,             -- perubahan stok, negatif bila berkurang
  stok_akhir INT NOT NULL,
  harga_beli INT NULL,
  biaya DECIMAL(14,4) NULL,        -- harga rata per satuan saat opname, dasar nilai susut di laba rugi
  catatan VARCHAR(255) NOT NULL DEFAULT '',
  oleh BIGINT NULL,
  at DATETIME NOT NULL,
  INDEX (bahan_id, at),
  FOREIGN KEY (bahan_id) REFERENCES bahan(id),
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Karyawan tidak selalu punya akun login, jadi dicatat terpisah (user_id hanya bila keduanya sama orang).
CREATE TABLE IF NOT EXISTS karyawan (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  nama VARCHAR(100) NOT NULL,
  jabatan VARCHAR(50) NOT NULL DEFAULT '',
  jenis ENUM('bulanan','harian') NOT NULL DEFAULT 'bulanan',
  gaji INT NOT NULL DEFAULT 0,  -- per bulan, atau per hari kerja bila harian
  user_id BIGINT NULL,
  aktif TINYINT(1) NOT NULL DEFAULT 1,
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Bon karyawan: jumlah positif = pinjaman diambil, negatif = dipotong dari gaji.
CREATE TABLE IF NOT EXISTS karyawan_bon (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  karyawan_id BIGINT NOT NULL,
  tanggal DATE NOT NULL,
  jumlah INT NOT NULL,
  keterangan VARCHAR(255) NOT NULL DEFAULT '',
  slip_id BIGINT NULL,
  INDEX (karyawan_id),
  FOREIGN KEY (karyawan_id) REFERENCES karyawan(id),
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS gaji_slip (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  karyawan_id BIGINT NOT NULL,
  periode_awal DATE NOT NULL,
  periode_akhir DATE NOT NULL,
  hari INT NOT NULL DEFAULT 0,          -- hari kerja, untuk karyawan harian
  pokok INT NOT NULL,
  tunjangan INT NOT NULL DEFAULT 0,
  bonus INT NOT NULL DEFAULT 0,
  potongan INT NOT NULL DEFAULT 0,      -- potongan lain (mis. terlambat)
  potongan_bon INT NOT NULL DEFAULT 0,  -- cicilan bon yang dipotong
  netto INT NOT NULL,
  catatan VARCHAR(255) NOT NULL DEFAULT '',
  status ENUM('draft','dibayar') NOT NULL DEFAULT 'draft',
  dibuat_oleh BIGINT NOT NULL,
  dibuat_at DATETIME NOT NULL,
  dibayar_at DATETIME NULL,
  dibayar_oleh BIGINT NULL,
  INDEX (karyawan_id),
  INDEX (dibayar_at),
  FOREIGN KEY (karyawan_id) REFERENCES karyawan(id),
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Bill belum dibayar yang ikut closing tenant (pemilik menalangi). Nilainya dibekukan di sini supaya status closing tidak berubah
-- bila diskon bill diubah sesudahnya.
CREATE TABLE IF NOT EXISTS closing_tenant_bill (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  closing_id BIGINT NOT NULL,
  pesanan_id BIGINT NOT NULL,
  kotor INT NOT NULL,
  diskon INT NOT NULL,
  INDEX (closing_id),
  FOREIGN KEY (closing_id) REFERENCES closing_tenant(id) ON DELETE CASCADE,
  FOREIGN KEY (pesanan_id) REFERENCES pesanan(id),
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- HP akun tenant yang menerima push notifikasi FCM. Satu token = satu HP; dipindah bila HP dipakai akun lain.
CREATE TABLE IF NOT EXISTS perangkat_push (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  tenant_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  token VARCHAR(255) NOT NULL UNIQUE,
  dibuat_at DATETIME NOT NULL,
  INDEX (klien_id, tenant_id),
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS audit_log (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  klien_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  aksi VARCHAR(50) NOT NULL,
  detail VARCHAR(1000) NOT NULL DEFAULT '',  -- method, path, dan isi permintaan tanpa password, dipotong
  at DATETIME NOT NULL,
  INDEX (klien_id, at),
  FOREIGN KEY (klien_id) REFERENCES klien(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
