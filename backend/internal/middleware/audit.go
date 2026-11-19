package middleware

import (
	"encoding/json"
	"time"

	"warkop/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// Audit mencatat aksi sensitif yang berhasil ke audit_log. Isi permintaan disimpan apa adanya kecuali password.
func Audit(aksi string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		u := CurrentUser(c)
		if err != nil || u == nil || u.KlienID == 0 || c.Response().StatusCode() >= 400 {
			return err
		}
		detail := c.Method() + " " + c.Path()
		if body := c.Body(); len(body) > 0 {
			var m map[string]any
			if json.Unmarshal(body, &m) == nil {
				delete(m, "password")
				body, _ = json.Marshal(m)
			}
			detail += " " + string(body)
		}
		if r := []rune(detail); len(r) > 1000 {
			detail = string(r[:1000])
		}
		// gagal mencatat audit tidak boleh membatalkan aksi yang sudah tersimpan
		if _, e := config.DB.Exec("INSERT INTO audit_log (klien_id,user_id,aksi,detail,at) VALUES (?,?,?,?,?)",
			u.KlienID, u.UserID, aksi, detail, time.Now()); e != nil {
			log.Error().Err(e).Str("aksi", aksi).Msg("audit gagal dicatat")
		}
		return nil
	}
}
