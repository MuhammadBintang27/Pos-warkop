package controllers

import (
	"warkop/internal/dto"
	"warkop/internal/middleware"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// Endpoint /tenant-saya hanya untuk role tenant; tenant_id selalu dari token, bukan dari input.

func (h *TenantController) PesananMasuk(c *fiber.Ctx) error {
	u := middleware.CurrentUser(c)
	res, err := h.svc.PesananMasuk(u.KlienID, u.TenantID)
	if err != nil {
		return handleErr(c, err, "PesananMasuk error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Pesanan masuk berhasil diambil", res)
}

func (h *TenantController) SetSiap(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.SiapRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	u := middleware.CurrentUser(c)
	if err := h.svc.SetSiap(u.KlienID, u.TenantID, id, *req.Siap); err != nil {
		return handleErr(c, err, "SetSiap error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Status item diperbarui")
}
