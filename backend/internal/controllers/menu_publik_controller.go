package controllers

import (
	"warkop/internal/dto"
	"warkop/internal/services"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type MenuPublikController struct {
	menu *services.MenuService
	peng *services.PengaturanService
}

func NewMenuPublikController(menu *services.MenuService, peng *services.PengaturanService) *MenuPublikController {
	return &MenuPublikController{menu: menu, peng: peng}
}

// Get: GET /api/publik/menu/:kode, tanpa login.
func (h *MenuPublikController) Get(c *fiber.Ctx) error {
	kid, err := h.menu.KlienDariKode(c.Params("kode"))
	if err != nil {
		return handleErr(c, err, "")
	}
	tenant, err := h.menu.Publik(kid)
	if err != nil {
		return handleErr(c, err, "MenuPublik error")
	}
	p, err := h.peng.Get(kid)
	if err != nil {
		return handleErr(c, err, "MenuPublik pengaturan error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Menu berhasil diambil", dto.MenuPublikResponse{
		Usaha:  dto.UsahaPublik{Nama: p.NamaUsaha, Alamat: p.Alamat, Logo: p.Logo, Theme: p.Theme, Primary: p.Primary, Mode: p.Mode},
		Tenant: tenant,
	})
}
