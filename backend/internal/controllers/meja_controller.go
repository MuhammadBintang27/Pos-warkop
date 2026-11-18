package controllers

import (
	"warkop/internal/dto"
	"warkop/internal/middleware"
	"warkop/internal/services"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type MejaController struct {
	svc *services.MejaService
}

func NewMejaController(svc *services.MejaService) *MejaController {
	return &MejaController{svc: svc}
}

func (h *MejaController) Denah(c *fiber.Ctx) error {
	denah, err := h.svc.Denah(kid(c))
	if err != nil {
		return handleErr(c, err, "Denah error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Denah meja berhasil diambil", denah)
}

func (h *MejaController) Create(c *fiber.Ctx) error {
	var req dto.MejaRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.Create(kid(c), req); err != nil {
		return handleErr(c, err, "CreateMeja error")
	}
	return utils.SendSuccess(c, fiber.StatusCreated, "Meja berhasil ditambahkan")
}

func (h *MejaController) Generate(c *fiber.Ctx) error {
	var req dto.GenerateMejaRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	res, err := h.svc.Generate(kid(c), req)
	if err != nil {
		return handleErr(c, err, "GenerateMeja error")
	}
	return utils.SendSuccess(c, fiber.StatusCreated, "Meja berhasil dibuat", res)
}

func (h *MejaController) Posisi(c *fiber.Ctx) error {
	var req dto.PosisiRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.SimpanPosisi(kid(c), req); err != nil {
		return handleErr(c, err, "SimpanPosisi error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Denah berhasil disimpan")
}

func (h *MejaController) Update(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.MejaRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.Update(kid(c), id, req); err != nil {
		return handleErr(c, err, "UpdateMeja error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Meja berhasil diperbarui")
}

func (h *MejaController) Delete(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	if err := h.svc.Delete(kid(c), id); err != nil {
		return handleErr(c, err, "DeleteMeja error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Meja berhasil dihapus")
}
