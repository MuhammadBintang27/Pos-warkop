package controllers

import (
	"warkop/internal/dto"
	"warkop/internal/middleware"
	"warkop/internal/services"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type KaryawanController struct {
	svc  *services.KaryawanService
	laba *services.LabaService
}

func NewKaryawanController(svc *services.KaryawanService, laba *services.LabaService) *KaryawanController {
	return &KaryawanController{svc: svc, laba: laba}
}

func (h *KaryawanController) List(c *fiber.Ctx) error {
	list, err := h.svc.List(kid(c))
	if err != nil {
		return handleErr(c, err, "ListKaryawan error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Daftar karyawan berhasil diambil", list)
}

func (h *KaryawanController) Create(c *fiber.Ctx) error {
	var req dto.KaryawanRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.Create(kid(c), req); err != nil {
		return handleErr(c, err, "CreateKaryawan error")
	}
	return utils.SendSuccess(c, fiber.StatusCreated, "Karyawan berhasil ditambahkan")
}

func (h *KaryawanController) Update(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.KaryawanRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.Update(kid(c), id, req); err != nil {
		return handleErr(c, err, "UpdateKaryawan error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Karyawan berhasil diperbarui")
}

func (h *KaryawanController) Delete(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	if err := h.svc.Delete(kid(c), id); err != nil {
		return handleErr(c, err, "DeleteKaryawan error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Karyawan berhasil dihapus")
}

func (h *KaryawanController) ListBon(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	list, err := h.svc.ListBon(kid(c), id)
	if err != nil {
		return handleErr(c, err, "ListBon error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Daftar bon berhasil diambil", list)
}

func (h *KaryawanController) TambahBon(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.BonRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.TambahBon(kid(c), id, req); err != nil {
		return handleErr(c, err, "TambahBon error")
	}
	return utils.SendSuccess(c, fiber.StatusCreated, "Bon dicatat")
}

func (h *KaryawanController) ListSlip(c *fiber.Ctx) error {
	var q dto.SlipQuery
	if err := middleware.ParseQuery(c, &q); err != nil {
		return err
	}
	list, err := h.svc.ListSlip(kid(c), q)
	if err != nil {
		return handleErr(c, err, "ListSlip error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Daftar slip gaji berhasil diambil", list)
}

func (h *KaryawanController) BuatSlip(c *fiber.Ctx) error {
	var req dto.SlipRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.BuatSlip(kid(c), middleware.CurrentUser(c).UserID, req); err != nil {
		return handleErr(c, err, "BuatSlip error")
	}
	return utils.SendSuccess(c, fiber.StatusCreated, "Slip gaji dibuat")
}

func (h *KaryawanController) HapusSlip(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	if err := h.svc.HapusSlip(kid(c), id); err != nil {
		return handleErr(c, err, "HapusSlip error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Slip gaji dihapus")
}

func (h *KaryawanController) BayarSlip(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	if err := h.svc.BayarSlip(kid(c), middleware.CurrentUser(c).UserID, id); err != nil {
		return handleErr(c, err, "BayarSlip error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Gaji ditandai dibayar")
}

func (h *KaryawanController) LabaRugi(c *fiber.Ctx) error {
	var q dto.HarianQuery
	if err := middleware.ParseQuery(c, &q); err != nil {
		return err
	}
	res, err := h.laba.LabaRugi(kid(c), q.Dari, q.Sampai)
	if err != nil {
		return handleErr(c, err, "LabaRugi error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Laba rugi berhasil diambil", res)
}

func (h *KaryawanController) MarginMenu(c *fiber.Ctx) error {
	var q dto.HarianQuery
	if err := middleware.ParseQuery(c, &q); err != nil {
		return err
	}
	res, err := h.laba.PerMenu(kid(c), q.Dari, q.Sampai)
	if err != nil {
		return handleErr(c, err, "MarginMenu error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Margin menu berhasil diambil", res)
}
