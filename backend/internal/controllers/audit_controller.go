package controllers

import (
	"warkop/internal/dto"
	"warkop/internal/middleware"
	"warkop/internal/services"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type AuditController struct {
	svc *services.AuditService
}

func NewAuditController(svc *services.AuditService) *AuditController {
	return &AuditController{svc: svc}
}

func (h *AuditController) List(c *fiber.Ctx) error {
	var q dto.AuditQuery
	if err := middleware.ParseQuery(c, &q); err != nil {
		return err
	}
	res, err := h.svc.List(middleware.CurrentUser(c).KlienID, q)
	if err != nil {
		return handleErr(c, err, "Audit error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Log aktivitas berhasil diambil", res)
}
