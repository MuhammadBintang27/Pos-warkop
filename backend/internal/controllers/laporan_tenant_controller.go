package controllers

import (
	"strconv"

	"warkop/internal/dto"
	"warkop/internal/middleware"
	"warkop/internal/services"
	"warkop/internal/types"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type LaporanTenantController struct {
	svc *services.LaporanTenantService
}

func NewLaporanTenantController(svc *services.LaporanTenantService) *LaporanTenantController {
	return &LaporanTenantController{svc: svc}
}

func (h *LaporanTenantController) Harian(c *fiber.Ctx) error {
	var q dto.HarianQuery
	if err := middleware.ParseQuery(c, &q); err != nil {
		return err
	}
	res, err := h.svc.Harian(kid(c), q)
	if err != nil {
		return handleErr(c, err, "HarianTenant error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Penjualan harian tenant berhasil diambil", res)
}

func (h *LaporanTenantController) Detail(c *fiber.Ctx) error {
	var q dto.DetailTenantQuery
	if err := middleware.ParseQuery(c, &q); err != nil {
		return err
	}
	res, err := h.svc.Detail(kid(c), q)
	if err != nil {
		return handleErr(c, err, "DetailTenant error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Detail penjualan tenant berhasil diambil", res)
}

func (h *LaporanTenantController) BelumBayar(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Query("tenantId"), 10, 64)
	if err != nil || id <= 0 {
		return handleErr(c, types.BadRequest("tenantId tidak valid"), "")
	}
	list, err := h.svc.BelumBayar(kid(c), id)
	if err != nil {
		return handleErr(c, err, "BelumBayar error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Bill belum dibayar berhasil diambil", list)
}

func (h *LaporanTenantController) Closing(c *fiber.Ctx) error {
	var req dto.ClosingRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.Closing(middleware.CurrentUser(c), req); err != nil {
		return handleErr(c, err, "Closing error")
	}
	return utils.SendSuccess(c, fiber.StatusCreated, "Closing berhasil")
}

func (h *LaporanTenantController) Setor(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	if err := h.svc.Setor(middleware.CurrentUser(c), id); err != nil {
		return handleErr(c, err, "Setor error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Ditandai sudah disetor")
}

func (h *LaporanTenantController) BatalClosing(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	if err := h.svc.BatalClosing(kid(c), id); err != nil {
		return handleErr(c, err, "BatalClosing error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Closing dibatalkan")
}
