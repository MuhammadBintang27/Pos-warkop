package controllers

import (
	"warkop/internal/dto"
	"warkop/internal/middleware"
	"warkop/internal/services"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type TenantController struct {
	svc *services.TenantService
}

func NewTenantController(svc *services.TenantService) *TenantController {
	return &TenantController{svc: svc}
}

func (h *TenantController) List(c *fiber.Ctx) error {
	list, err := h.svc.List(kid(c))
	if err != nil {
		return handleErr(c, err, "ListTenant error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Tenant berhasil diambil", list)
}

func (h *TenantController) Create(c *fiber.Ctx) error {
	var req dto.TenantRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.Create(kid(c), req); err != nil {
		return handleErr(c, err, "CreateTenant error")
	}
	return utils.SendSuccess(c, fiber.StatusCreated, "Tenant berhasil ditambahkan")
}

func (h *TenantController) Update(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.TenantRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.Update(kid(c), id, req); err != nil {
		return handleErr(c, err, "UpdateTenant error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Tenant berhasil diperbarui")
}

func (h *TenantController) Buka(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.BukaRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.SetBuka(kid(c), id, req.Buka); err != nil {
		return handleErr(c, err, "BukaTenant error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Status tenant diperbarui")
}

func (h *TenantController) Delete(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	if err := h.svc.Delete(kid(c), id); err != nil {
		return handleErr(c, err, "DeleteTenant error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Tenant berhasil dihapus")
}
