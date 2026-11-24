package controllers

import (
	"warkop/internal/dto"
	"warkop/internal/middleware"
	"warkop/internal/services"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// Endpoint /tenant-saya hanya untuk role tenant; tenant_id selalu dari token, bukan dari input.

func (h *TenantController) PesananMasuk(c *fiber.Ctx) error {
	u := middleware.CurrentUser(c)
	res, err := h.svc.PesananMasuk(u.KlienID, u.TenantID, c.Query("tanggal"))
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

func (h *TenantController) SetProses(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.ProsesRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	u := middleware.CurrentUser(c)
	if err := h.svc.SetProses(u.KlienID, u.TenantID, id, *req.Proses); err != nil {
		return handleErr(c, err, "SetProses error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Status item diperbarui")
}

type PerangkatController struct {
	push *services.Pusher
}

func NewPerangkatController(push *services.Pusher) *PerangkatController {
	return &PerangkatController{push: push}
}

func (h *PerangkatController) Simpan(c *fiber.Ctx) error {
	var req dto.PerangkatRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	u := middleware.CurrentUser(c)
	if err := h.push.SimpanPerangkat(u.KlienID, u.TenantID, u.UserID, req.Token); err != nil {
		return handleErr(c, err, "SimpanPerangkat error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Perangkat terdaftar untuk notifikasi")
}

func (h *PerangkatController) Hapus(c *fiber.Ctx) error {
	var req dto.PerangkatRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.push.HapusPerangkat(kid(c), req.Token); err != nil {
		return handleErr(c, err, "HapusPerangkat error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Perangkat dilepas")
}
