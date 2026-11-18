package middleware

import (
	"errors"
	"strings"

	"warkop/internal/types"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const UserContextKey = "user"

func Authenticate(c *fiber.Ctx) error {
	token := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
	// EventSource tidak bisa mengirim header, jadi hanya stream events yang boleh pakai query.
	if token == "" && strings.HasSuffix(c.Path(), "/events") {
		token = c.Query("token")
	}
	if token == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized - token tidak ditemukan")
	}
	payload, err := utils.VerifyToken(token)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return utils.SendError(c, fiber.StatusUnauthorized, "Sesi habis, silakan login lagi")
		}
		return utils.SendError(c, fiber.StatusUnauthorized, "Token tidak valid")
	}
	if payload.KlienID != 0 && !utils.KlienAktif(payload.KlienID) {
		return utils.SendError(c, fiber.StatusUnauthorized, "Akses usaha ini dinonaktifkan. Hubungi penyedia layanan")
	}
	c.Locals(UserContextKey, payload)
	return c.Next()
}

// OptionalAuthenticate mengisi pengguna bila token valid, tanpa menolak permintaan tanpa token.
func OptionalAuthenticate(c *fiber.Ctx) error {
	if payload, err := utils.VerifyToken(strings.TrimPrefix(c.Get("Authorization"), "Bearer ")); err == nil {
		c.Locals(UserContextKey, payload)
	}
	return c.Next()
}

func Authorize(roles ...types.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		payload := CurrentUser(c)
		if payload == nil {
			return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized")
		}
		for _, r := range roles {
			if payload.Role == r {
				return c.Next()
			}
		}
		return utils.SendError(c, fiber.StatusForbidden, "Anda tidak punya akses ke resource ini")
	}
}

func CurrentUser(c *fiber.Ctx) *utils.JWTPayload {
	p, _ := c.Locals(UserContextKey).(*utils.JWTPayload)
	return p
}
