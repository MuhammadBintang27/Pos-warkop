package controllers

import (
	"warkop/internal/dto"
	"warkop/internal/middleware"
	"warkop/internal/services"
	"warkop/internal/types"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type PesananController struct {
	svc *services.PesananService
}

func NewPesananController(svc *services.PesananService) *PesananController {
	return &PesananController{svc: svc}
}

func (h *PesananController) Create(c *fiber.Ctx) error {
	var req dto.CreatePesananRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	p, err := h.svc.Create(middleware.CurrentUser(c), req)
	if err != nil {
		return handleErr(c, err, "CreatePesanan error")
	}
	return utils.SendSuccess(c, fiber.StatusCreated, "Pesanan berhasil dibuat", p)
}

func (h *PesananController) Get(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	p, err := h.svc.Get(kid(c), id)
	if err != nil {
		return handleErr(c, err, "GetPesanan error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Pesanan berhasil diambil", p)
}

func (h *PesananController) AddItem(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.AddItemRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	p, err := h.svc.AddItem(kid(c), id, req)
	if err != nil {
		return handleErr(c, err, "AddItem error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Item berhasil ditambahkan", p)
}

func (h *PesananController) UpdateItem(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	itemID, err := utils.ParamID(c, "itemId")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.UpdateItemRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	p, err := h.svc.UpdateItem(kid(c), id, itemID, req)
	if err != nil {
		return handleErr(c, err, "UpdateItem error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Item berhasil diperbarui", p)
}

func (h *PesananController) DeleteItem(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	itemID, err := utils.ParamID(c, "itemId")
	if err != nil {
		return handleErr(c, err, "")
	}
	p, err := h.svc.DeleteItem(kid(c), id, itemID)
	if err != nil {
		return handleErr(c, err, "DeleteItem error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Item berhasil dihapus", p)
}

func (h *PesananController) Kirim(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.KirimRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return handleErr(c, types.BadRequest("Data kirim tidak valid"), "")
		}
	}
	res, err := h.svc.Kirim(middleware.CurrentUser(c), id, cetakServer(c), req.ItemIDs)
	if err != nil {
		return handleErr(c, err, "Kirim error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Pesanan dikirim ke dapur", res)
}

func (h *PesananController) SetDiskon(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.DiskonRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	p, err := h.svc.SetDiskon(kid(c), id, req)
	if err != nil {
		return handleErr(c, err, "SetDiskon error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Diskon berhasil diatur", p)
}

func (h *PesananController) Batal(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.BatalRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	res, err := h.svc.Batal(middleware.CurrentUser(c), id, req, cetakServer(c))
	if err != nil {
		return handleErr(c, err, "Batal error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Pesanan dibatalkan", res)
}

func (h *PesananController) BatalItem(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	itemID, err := utils.ParamID(c, "itemId")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.BatalItemRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	res, err := h.svc.BatalItem(middleware.CurrentUser(c), id, itemID, req, cetakServer(c))
	if err != nil {
		return handleErr(c, err, "BatalItem error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Item dibatalkan", res)
}

func (h *PesananController) Pindah(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.PindahRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	p, err := h.svc.Pindah(kid(c), id, req)
	if err != nil {
		return handleErr(c, err, "Pindah error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Pesanan dipindahkan", p)
}

func (h *PesananController) Gabung(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.GabungRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	p, err := h.svc.Gabung(kid(c), id, req)
	if err != nil {
		return handleErr(c, err, "Gabung error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Pesanan digabung", p)
}

func (h *PesananController) Pisah(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.PisahRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	p, err := h.svc.Pisah(middleware.CurrentUser(c), id, req)
	if err != nil {
		return handleErr(c, err, "Pisah error")
	}
	return utils.SendSuccess(c, fiber.StatusCreated, "Pesanan dipisah", p)
}

// cetakServer false bila aksi dikirim ulang dari antrean offline HP kasir: tiketnya sudah dicetak HP, jangan dobel.
func cetakServer(c *fiber.Ctx) bool { return c.Get("X-Offline") != "1" }
