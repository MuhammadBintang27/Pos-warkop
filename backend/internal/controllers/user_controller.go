package controllers

import (
	"warkop/internal/dto"
	"warkop/internal/middleware"
	"warkop/internal/services"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type UserController struct {
	svc *services.UserService
}

func NewUserController(svc *services.UserService) *UserController {
	return &UserController{svc: svc}
}

func (h *UserController) List(c *fiber.Ctx) error {
	users, err := h.svc.List(kid(c))
	if err != nil {
		return handleErr(c, err, "ListUsers error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Daftar user berhasil diambil", users)
}

func (h *UserController) Create(c *fiber.Ctx) error {
	var req dto.CreateUserRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.Create(kid(c), req); err != nil {
		return handleErr(c, err, "CreateUser error")
	}
	return utils.SendSuccess(c, fiber.StatusCreated, "User berhasil ditambahkan")
}

func (h *UserController) Update(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.UpdateUserRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.Update(kid(c), middleware.CurrentUser(c).UserID, id, req); err != nil {
		return handleErr(c, err, "UpdateUser error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "User berhasil diperbarui")
}
