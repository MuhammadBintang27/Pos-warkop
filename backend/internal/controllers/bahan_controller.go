package controllers

import (
	"warkop/internal/dto"
	"warkop/internal/middleware"
	"warkop/internal/services"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type BahanController struct {
	svc *services.BahanService
}

func NewBahanController(svc *services.BahanService) *BahanController {
	return &BahanController{svc: svc}
}

func (h *BahanController) List(c *fiber.Ctx) error {
	list, err := h.svc.List(kid(c))
	if err != nil {
		return handleErr(c, err, "ListBahan error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Daftar bahan berhasil diambil", list)
}

func (h *BahanController) Create(c *fiber.Ctx) error {
	var req dto.BahanRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.Create(kid(c), req); err != nil {
		return handleErr(c, err, "CreateBahan error")
	}
	return utils.SendSuccess(c, fiber.StatusCreated, "Bahan berhasil ditambahkan")
}

func (h *BahanController) Update(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.BahanRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.Update(kid(c), id, req); err != nil {
		return handleErr(c, err, "UpdateBahan error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Bahan berhasil diperbarui")
}

func (h *BahanController) Delete(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	if err := h.svc.Delete(kid(c), id); err != nil {
		return handleErr(c, err, "DeleteBahan error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Bahan berhasil dihapus")
}

func (h *BahanController) Restock(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.RestockRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.Restock(kid(c), middleware.CurrentUser(c).UserID, id, req); err != nil {
		return handleErr(c, err, "Restock error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Stok bertambah")
}

func (h *BahanController) Opname(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.OpnameRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.Opname(kid(c), middleware.CurrentUser(c).UserID, id, req); err != nil {
		return handleErr(c, err, "Opname error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Stok disesuaikan")
}

func (h *BahanController) Riwayat(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	list, err := h.svc.Riwayat(kid(c), id)
	if err != nil {
		return handleErr(c, err, "RiwayatBahan error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Riwayat bahan berhasil diambil", list)
}

func (h *BahanController) Pemakaian(c *fiber.Ctx) error {
	var q dto.HarianQuery
	if err := middleware.ParseQuery(c, &q); err != nil {
		return err
	}
	list, err := h.svc.Pemakaian(kid(c), q.Dari, q.Sampai)
	if err != nil {
		return handleErr(c, err, "PemakaianBahan error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Pemakaian bahan berhasil diambil", list)
}
