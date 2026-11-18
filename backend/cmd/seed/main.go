package main

import (
	"database/sql"
	"os"
	"strconv"

	"warkop/internal/config"
	"warkop/internal/utils"

	"github.com/rs/zerolog/log"
)

type opsi struct {
	Nama     string
	Tambahan int
}

type grup struct {
	Nama  string
	Wajib bool
	Multi bool
	Opsi  []opsi
}

type menu struct {
	Nama   string
	Harga  int
	Lacak  bool
	Stok   int
	Varian []grup
}

type kategori struct {
	Nama string
	Menu []menu
}

type hasil struct {
	kategori, menu, dilewati int
}

var (
	suhu          = varian("Suhu", true, false, pilih("Panas", 0), pilih("Es", 1000))
	gula          = varian("Gula", false, false, pilih("Gula Sedikit", 0), pilih("Tanpa Gula", 0))
	pedas         = varian("Level Pedas", true, false, pilih("Tidak Pedas", 0), pilih("Sedang", 0), pilih("Pedas", 0), pilih("Extra Pedas", 0))
	tambahanMie   = varian("Tambahan", false, true, pilih("Telur", 3000), pilih("Keju", 2000), pilih("Sosis", 2000), pilih("Bakso", 3000))
	tambahanNasi  = varian("Tambahan", false, true, pilih("Telur Ceplok", 3000), pilih("Kerupuk", 1000))
	telur         = varian("Telur", true, false, pilih("Ceplok", 0), pilih("Dadar", 0), pilih("Orak-arik", 0))
	rasaRoti      = varian("Rasa", true, true, pilih("Coklat", 0), pilih("Susu", 0), pilih("Srikaya", 0), pilih("Kacang", 0), pilih("Keju", 2000))
	toppingPisang = varian("Topping", false, true, pilih("Coklat", 0), pilih("Keju", 2000), pilih("Susu", 0))
)

var data = []kategori{
	{"Kopi", []menu{
		item("Kopi Hitam", 4000, suhu, gula),
		item("Kopi Susu", 6000, suhu, gula),
		item("Kopi Jahe", 6000, suhu),
		item("Kopi Susu Gula Aren", 10000, suhu),
		item("Kopi Sachet", 5000, suhu),
	}},
	{"Non-Kopi", []menu{
		item("Teh Manis", 4000, suhu, gula),
		item("Teh Tarik", 7000, suhu),
		item("Jeruk Peras", 6000, suhu),
		item("Susu Jahe", 7000, suhu),
		item("Jahe Merah", 5000, suhu),
		item("Susu Coklat", 7000, suhu),
		item("Milo", 8000, suhu),
	}},
	{"Makanan", []menu{
		item("Indomie Rebus", 8000, tambahanMie),
		item("Indomie Goreng", 8000, tambahanMie),
		item("Nasi Goreng", 13000, pedas, tambahanNasi),
		item("Mie Goreng Jawa", 13000, pedas, tambahanNasi),
		item("Nasi Telur", 10000, telur),
		item("Nasi Ayam Geprek", 15000, pedas),
		item("Roti Bakar", 10000, rasaRoti),
		item("Pisang Bakar", 10000, toppingPisang),
		item("Telur Setengah Matang", 6000),
		item("Bubur Kacang Hijau", 8000),
	}},
	{"Gorengan", []menu{
		stokItem("Pisang Goreng", 1000, 30),
		stokItem("Tahu Isi", 1000, 30),
		stokItem("Tempe Mendoan", 1000, 30),
		stokItem("Bakwan", 1000, 30),
		stokItem("Cireng", 1000, 20),
		stokItem("Risol", 2000, 20),
	}},
	{"Lainnya", []menu{
		stokItem("Air Mineral", 3000, 48),
		stokItem("Teh Botol", 5000, 24),
		stokItem("Kerupuk", 1000, 50),
	}},
}

func pilih(nama string, tambahan int) opsi { return opsi{nama, tambahan} }

func varian(nama string, wajib, multi bool, opsi ...opsi) grup {
	return grup{nama, wajib, multi, opsi}
}

func item(nama string, harga int, varian ...grup) menu {
	return menu{Nama: nama, Harga: harga, Varian: varian}
}

func stokItem(nama string, harga, stok int, varian ...grup) menu {
	return menu{Nama: nama, Harga: harga, Lacak: true, Stok: stok, Varian: varian}
}

func main() {
	config.Load()
	if err := config.ConnectDB(); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MySQL")
	}

	if len(os.Args) != 2 {
		log.Fatal().Msg("pemakaian: go run ./cmd/seed <klien id>")
	}
	kid, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil || kid <= 0 {
		log.Fatal().Msg("klien id tidak valid")
	}

	var h hasil
	err = utils.WithTx(config.DB, func(tx *sql.Tx) error {
		var tenantID int64
		if err := tx.QueryRow("SELECT id FROM tenant WHERE klien_id=? AND aktif=1 ORDER BY id LIMIT 1", kid).Scan(&tenantID); err != nil {
			return err
		}
		for i, k := range data {
			kategoriID, baru, err := cariAtauBuatKategori(tx, kid, k.Nama, i+1)
			if err != nil {
				return err
			}
			if baru {
				h.kategori++
			}
			for _, m := range k.Menu {
				dibuat, err := buatMenu(tx, kid, kategoriID, tenantID, m)
				if err != nil {
					return err
				}
				if dibuat {
					h.menu++
				} else {
					h.dilewati++
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Seed gagal, tidak ada data yang masuk")
	}
	log.Info().Int("kategoriBaru", h.kategori).Int("menuBaru", h.menu).Int("menuDilewati", h.dilewati).Msg("Seed menu selesai")
}

func cariAtauBuatKategori(tx *sql.Tx, kid int64, nama string, urutan int) (id int64, baru bool, err error) {
	err = tx.QueryRow("SELECT id FROM kategori WHERE klien_id=? AND nama=?", kid, nama).Scan(&id)
	if err != sql.ErrNoRows {
		return id, false, err
	}
	res, err := tx.Exec("INSERT INTO kategori (klien_id,nama,urutan) VALUES (?,?,?)", kid, nama, urutan)
	if err != nil {
		return 0, false, err
	}
	id, err = res.LastInsertId()
	return id, true, err
}

func buatMenu(tx *sql.Tx, kid, kategoriID, tenantID int64, m menu) (bool, error) {
	var ada int
	err := tx.QueryRow("SELECT 1 FROM menu WHERE klien_id=? AND kategori_id=? AND nama=? AND aktif=1", kid, kategoriID, m.Nama).Scan(&ada)
	if err == nil {
		return false, nil
	}
	if err != sql.ErrNoRows {
		return false, err
	}

	res, err := tx.Exec("INSERT INTO menu (klien_id,kategori_id,nama,harga,track_stok,stok,tenant_id) VALUES (?,?,?,?,?,?,?)",
		kid, kategoriID, m.Nama, m.Harga, m.Lacak, m.Stok, tenantID)
	if err != nil {
		return false, err
	}
	menuID, err := res.LastInsertId()
	if err != nil {
		return false, err
	}
	for _, g := range m.Varian {
		res, err := tx.Exec("INSERT INTO varian_grup (klien_id,menu_id,nama,wajib,multi) VALUES (?,?,?,?,?)", kid, menuID, g.Nama, g.Wajib, g.Multi)
		if err != nil {
			return false, err
		}
		grupID, err := res.LastInsertId()
		if err != nil {
			return false, err
		}
		for _, o := range g.Opsi {
			if _, err := tx.Exec("INSERT INTO varian_opsi (klien_id,grup_id,nama,tambahan) VALUES (?,?,?,?)", kid, grupID, o.Nama, o.Tambahan); err != nil {
				return false, err
			}
		}
	}
	return true, nil
}
