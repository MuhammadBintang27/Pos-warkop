package controllers

import (
	"warkop/internal/dto"
	"warkop/internal/middleware"
	"warkop/internal/services"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type KlienController struct {
	svc   *services.KlienService
	users *services.UserService
}

func NewKlienController(svc *services.KlienService, users *services.UserService) *KlienController {
	return &KlienController{svc: svc, users: users}
}

func (h *KlienController) List(c *fiber.Ctx) error {
	list, err := h.svc.List()
	if err != nil {
		return handleErr(c, err, "ListKlien error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Daftar klien berhasil diambil", list)
}

func (h *KlienController) Create(c *fiber.Ctx) error {
	var req dto.KlienRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	kr, err := h.svc.Buat(req)
	if err != nil {
		return handleErr(c, err, "CreateKlien error")
	}
	return utils.SendSuccess(c, fiber.StatusCreated, "Klien berhasil dibuat", kr)
}

func (h *KlienController) Edit(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.KlienEditRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.Edit(id, req); err != nil {
		return handleErr(c, err, "EditKlien error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Data klien diperbarui")
}

func (h *KlienController) Perpanjang(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.KlienLanggananRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.Perpanjang(id, req.Bulan); err != nil {
		return handleErr(c, err, "PerpanjangKlien error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Langganan diperpanjang")
}

func (h *KlienController) SetAktif(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.KlienAktifRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.SetAktif(id, *req.Aktif); err != nil {
		return handleErr(c, err, "SetAktifKlien error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Status klien diperbarui")
}

func (h *KlienController) ResetOwner(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	kr, err := h.svc.ResetOwner(id)
	if err != nil {
		return handleErr(c, err, "ResetOwner error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Password owner direset", kr)
}

// klienParam membaca :id dan memastikan klien itu ada.
func (h *KlienController) klienParam(c *fiber.Ctx) (int64, error) {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return 0, err
	}
	return id, h.svc.Pastikan(id)
}

func (h *KlienController) ListUsers(c *fiber.Ctx) error {
	id, err := h.klienParam(c)
	if err != nil {
		return handleErr(c, err, "")
	}
	list, err := h.users.List(id)
	if err != nil {
		return handleErr(c, err, "ListUsersKlien error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Daftar pengguna berhasil diambil", list)
}

func (h *KlienController) CreateUser(c *fiber.Ctx) error {
	id, err := h.klienParam(c)
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.CreateUserRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.users.Create(id, req); err != nil {
		return handleErr(c, err, "CreateUserKlien error")
	}
	return utils.SendSuccess(c, fiber.StatusCreated, "Pengguna berhasil ditambahkan")
}

func (h *KlienController) UpdateUser(c *fiber.Ctx) error {
	id, err := h.klienParam(c)
	if err != nil {
		return handleErr(c, err, "")
	}
	uid, err := utils.ParamID(c, "uid")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.UpdateUserRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.users.Update(id, 0, uid, req); err != nil {
		return handleErr(c, err, "UpdateUserKlien error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Pengguna berhasil diperbarui")
}
