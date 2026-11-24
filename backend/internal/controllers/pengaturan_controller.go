package controllers

import (
	"io"

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

func (h *PengaturanController) UnggahLogo(c *fiber.Ctx) error {
	fh, err := c.FormFile("logo")
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "File logo wajib diisi")
	}
	if fh.Size > services.MaksLogo {
		return utils.SendError(c, fiber.StatusBadRequest, "Logo maksimal 1 MB")
	}
	f, err := fh.Open()
	if err != nil {
		return handleErr(c, err, "UnggahLogo error")
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, services.MaksLogo+1))
	if err != nil {
		return handleErr(c, err, "UnggahLogo error")
	}
	url, err := h.svc.SimpanLogo(kid(c), raw)
	if err != nil {
		return handleErr(c, err, "UnggahLogo error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Logo terunggah", fiber.Map{"url": url})
}
