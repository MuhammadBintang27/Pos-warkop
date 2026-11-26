package middleware

import (
	"math/rand/v2"
	"time"

	"warkop/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// Idempoten menjawab permintaan ber-header X-Idem yang pernah berhasil dengan jawaban pertamanya, tanpa memproses ulang.
// HP kasir mengulang aksi dari antrean offline yang bisa saja sudah sampai ke server sebelum sinyal putus.
// shortcut: dua permintaan dengan kunci sama yang datang bersamaan masih sama-sama diproses; HP mengirim antrean berurutan.
func Idempoten(c *fiber.Ctx) error {
	kunci, u := c.Get("X-Idem"), CurrentUser(c)
	if kunci == "" || len(kunci) > 64 || u == nil {
		return c.Next()
	}
	var status int
	var body string
	if config.DB.QueryRow("SELECT status,body FROM idem WHERE klien_id=? AND kunci=?", u.KlienID, kunci).Scan(&status, &body) == nil {
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
		// tanda jawaban ulangan: HP tahu permintaan pertamanya sudah diproses (dan tiketnya sudah dicetak stasiun)
		c.Set("X-Idem-Ulang", "1")
		return c.Status(status).SendString(body)
	}
	if err := c.Next(); err != nil {
		return err
	}
	// yang gagal tidak disimpan, supaya boleh dicoba lagi setelah masalahnya dibereskan
	if st := c.Response().StatusCode(); st < 400 {
		if _, e := config.DB.Exec("INSERT IGNORE INTO idem (klien_id,kunci,status,body,at) VALUES (?,?,?,?,?)",
			u.KlienID, kunci, st, string(c.Response().Body()), time.Now()); e != nil {
			log.Error().Err(e).Msg("idem gagal disimpan")
		}
		if rand.IntN(500) == 0 { // antrean offline paling lama berumur beberapa hari
			config.DB.Exec("DELETE FROM idem WHERE at < ?", time.Now().AddDate(0, 0, -7))
		}
	}
	return nil
}
