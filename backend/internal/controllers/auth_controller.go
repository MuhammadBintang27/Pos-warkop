package controllers

import (
	"warkop/internal/dto"
	"warkop/internal/middleware"
	"warkop/internal/services"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type AuthController struct {
	svc *services.AuthService
}

func NewAuthController(svc *services.AuthService) *AuthController {
	return &AuthController{svc: svc}
}

func (h *AuthController) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	res, err := h.svc.Login(req)
	if err != nil {
		return handleErr(c, err, "Login error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Login berhasil", res)
}
