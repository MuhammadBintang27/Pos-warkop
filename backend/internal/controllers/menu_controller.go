package controllers

import (
	"warkop/internal/dto"
	"warkop/internal/middleware"
	"warkop/internal/services"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type MenuController struct {
	svc *services.MenuService
}

func NewMenuController(svc *services.MenuService) *MenuController {
	return &MenuController{svc: svc}
}

func (h *MenuController) List(c *fiber.Ctx) error {
	kategori, err := h.svc.List(kid(c))
	if err != nil {
		return handleErr(c, err, "ListMenu error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Menu berhasil diambil", fiber.Map{"kategori": kategori})
}

func (h *MenuController) CreateKategori(c *fiber.Ctx) error {
	var req dto.KategoriRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.CreateKategori(kid(c), req); err != nil {
		return handleErr(c, err, "CreateKategori error")
	}
	return utils.SendSuccess(c, fiber.StatusCreated, "Kategori berhasil ditambahkan")
}

func (h *MenuController) UpdateKategori(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.KategoriRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.UpdateKategori(kid(c), id, req); err != nil {
		return handleErr(c, err, "UpdateKategori error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Kategori berhasil diperbarui")
}

func (h *MenuController) DeleteKategori(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	if err := h.svc.DeleteKategori(kid(c), id); err != nil {
		return handleErr(c, err, "DeleteKategori error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Kategori berhasil dihapus")
}

func (h *MenuController) Create(c *fiber.Ctx) error {
	var req dto.MenuRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	id, err := h.svc.Create(kid(c), req)
	if err != nil {
		return handleErr(c, err, "CreateMenu error")
	}
	return utils.SendSuccess(c, fiber.StatusCreated, "Menu berhasil ditambahkan", fiber.Map{"id": id})
}

func (h *MenuController) Update(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.MenuRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.Update(kid(c), id, req); err != nil {
		return handleErr(c, err, "UpdateMenu error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Menu berhasil diperbarui")
}

func (h *MenuController) Delete(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	if err := h.svc.Delete(kid(c), id); err != nil {
		return handleErr(c, err, "DeleteMenu error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Menu berhasil dihapus")
}

func (h *MenuController) SetStok(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.StokRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.SetStok(kid(c), id, req); err != nil {
		return handleErr(c, err, "SetStok error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Stok berhasil diperbarui")
}
