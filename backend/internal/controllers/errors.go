package controllers

import (
	"warkop/internal/middleware"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// kid: klien milik pengguna yang login; dipakai semua endpoint yang lewat Authenticate.
func kid(c *fiber.Ctx) int64 { return middleware.CurrentUser(c).KlienID }

func handleErr(c *fiber.Ctx, err error, fallback string) error {
	if he, ok := utils.AsHTTPError(err); ok {
		return utils.SendError(c, he.Status, he.Message)
	}
	if utils.MySQLErrNo(err) == utils.MySQLFKRestrict {
		return utils.SendError(c, fiber.StatusBadRequest, "Data masih dipakai, tidak bisa dihapus")
	}
	log.Error().Err(err).Msg(fallback)
	return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error")
}
