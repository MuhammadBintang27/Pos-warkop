package controllers

import (
	"warkop/internal/dto"
	"warkop/internal/middleware"
	"warkop/internal/services"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type PengaturanController struct {
	svc *services.PengaturanService
}

func NewPengaturanController(svc *services.PengaturanService) *PengaturanController {
	return &PengaturanController{svc: svc}
}

func klienOpsional(c *fiber.Ctx) int64 {
	if u := middleware.CurrentUser(c); u != nil {
		return u.KlienID
	}
	return 0
}

func (h *PengaturanController) Get(c *fiber.Ctx) error {
	p, err := h.svc.Get(klienOpsional(c))
	if err != nil {
		return handleErr(c, err, "GetPengaturan error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Pengaturan berhasil diambil", p)
}

func (h *PengaturanController) Update(c *fiber.Ctx) error {
	var req dto.PengaturanRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.Update(kid(c), req); err != nil {
		return handleErr(c, err, "UpdatePengaturan error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Pengaturan berhasil disimpan", req)
}
