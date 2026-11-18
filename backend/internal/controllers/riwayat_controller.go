package controllers

import (
	"warkop/internal/dto"
	"warkop/internal/middleware"
	"warkop/internal/services"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type RiwayatController struct {
	svc *services.RiwayatService
}

func NewRiwayatController(svc *services.RiwayatService) *RiwayatController {
	return &RiwayatController{svc: svc}
}

func (h *RiwayatController) List(c *fiber.Ctx) error {
	var q dto.RiwayatQuery
	if err := middleware.ParseQuery(c, &q); err != nil {
		return err
	}
	res, err := h.svc.List(middleware.CurrentUser(c), q)
	if err != nil {
		return handleErr(c, err, "Riwayat error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Riwayat berhasil diambil", res)
}
