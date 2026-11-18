package middleware

import (
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
)

const stasiunKlienKey = "stasiun_klien"

// StasiunKlien: klien pemilik token stasiun yang sedang memanggil.
func StasiunKlien(c *fiber.Ctx) int64 {
	kid, _ := c.Locals(stasiunKlienKey).(int64)
	return kid
}

// Stasiun memeriksa token stasiun cetak (header X-Stasiun-Token). Stasiun bukan pengguna, jadi tidak memakai JWT login.
func Stasiun(verify func(token string) (int64, bool)) fiber.Handler {
	return func(c *fiber.Ctx) error {
		kid, ok := verify(c.Get("X-Stasiun-Token"))
		if !ok {
			return utils.SendError(c, fiber.StatusUnauthorized, "Token stasiun tidak valid")
		}
		c.Locals(stasiunKlienKey, kid)
		return c.Next()
	}
}
