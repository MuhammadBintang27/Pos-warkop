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
	if tid := tenantSaya(c); tid != 0 {
		res.Baris = saring(res.Baris, func(b dto.HarianTenant) bool { return b.TenantID == tid })
		res.Tertunda = saring(res.Tertunda, func(t dto.TertundaTenant) bool { return t.TenantID == tid })
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Penjualan harian tenant berhasil diambil", res)
}

func (h *LaporanTenantController) Detail(c *fiber.Ctx) error {
	var q dto.DetailTenantQuery
	if tid := tenantSaya(c); tid != 0 {
		c.Request().URI().QueryArgs().Set("tenantId", strconv.FormatInt(tid, 10))
	}
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
	if tid := tenantSaya(c); tid != 0 {
		id, err = tid, nil
	}
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

// tenantSaya: tenant milik akun tenant yang login (0 untuk role lain), supaya laporannya terkunci ke tenant itu.
func tenantSaya(c *fiber.Ctx) int64 {
	if u := middleware.CurrentUser(c); u.Role == types.RoleTenant {
		return u.TenantID
	}
	return 0
}

func saring[T any](xs []T, ok func(T) bool) []T {
	out := []T{}
	for _, x := range xs {
		if ok(x) {
			out = append(out, x)
		}
	}
	return out
}

func (h *LaporanTenantController) SetorPeriode(c *fiber.Ctx) error {
	var req dto.SetorPeriodeRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	n, err := h.svc.SetorPeriode(middleware.CurrentUser(c), req)
	if err != nil {
		return handleErr(c, err, "SetorPeriode error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Ditandai sudah disetor", fiber.Map{"hari": n})
}
