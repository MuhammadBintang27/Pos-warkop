package controllers

import (
	"warkop/internal/dto"
	"warkop/internal/middleware"
	"warkop/internal/services"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type CetakController struct {
	svc *services.CetakService
}

func NewCetakController(svc *services.CetakService) *CetakController {
	return &CetakController{svc: svc}
}

func (h *CetakController) ListPrinter(c *fiber.Ctx) error {
	list, err := h.svc.ListPrinter(kid(c))
	if err != nil {
		return handleErr(c, err, "ListPrinter error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Printer berhasil diambil", list)
}

func (h *CetakController) CreatePrinter(c *fiber.Ctx) error {
	var req dto.PrinterRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.CreatePrinter(kid(c), req); err != nil {
		return handleErr(c, err, "CreatePrinter error")
	}
	return utils.SendSuccess(c, fiber.StatusCreated, "Printer berhasil ditambahkan")
}

func (h *CetakController) UpdatePrinter(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.PrinterRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.UpdatePrinter(kid(c), id, req); err != nil {
		return handleErr(c, err, "UpdatePrinter error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Printer berhasil diperbarui")
}

func (h *CetakController) DeletePrinter(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	if err := h.svc.DeletePrinter(kid(c), id); err != nil {
		return handleErr(c, err, "DeletePrinter error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Printer berhasil dihapus")
}

func (h *CetakController) TesPrinter(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	if err := h.svc.Tes(kid(c), id); err != nil {
		return handleErr(c, err, "TesPrinter error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Tiket percobaan masuk antrean")
}

func (h *CetakController) BuatToken(c *fiber.Ctx) error {
	token, err := h.svc.BuatToken(kid(c))
	if err != nil {
		return handleErr(c, err, "BuatToken error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Token stasiun dibuat. Simpan sekarang, tidak akan ditampilkan lagi", dto.TokenStasiun{Token: token})
}

func (h *CetakController) Riwayat(c *fiber.Ctx) error {
	list, err := h.svc.Riwayat(kid(c))
	if err != nil {
		return handleErr(c, err, "RiwayatCetak error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Riwayat cetak berhasil diambil", list)
}

func (h *CetakController) Ulang(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	if err := h.svc.Ulang(kid(c), id); err != nil {
		return handleErr(c, err, "UlangCetak error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Tiket dikirim ulang ke antrean")
}

func (h *CetakController) Status(c *fiber.Ctx) error {
	st, err := h.svc.Status(kid(c))
	if err != nil {
		return handleErr(c, err, "StatusCetak error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Status cetak berhasil diambil", st)
}

func (h *CetakController) StasiunJobs(c *fiber.Ctx) error {
	jobs, err := h.svc.AmbilJob(middleware.StasiunKlien(c))
	if err != nil {
		return handleErr(c, err, "StasiunJobs error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Antrean cetak", jobs)
}

func (h *CetakController) StasiunSelesai(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	if err := h.svc.Selesai(middleware.StasiunKlien(c), id); err != nil {
		return handleErr(c, err, "StasiunSelesai error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Ditandai tercetak")
}

func (h *CetakController) StasiunGagal(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.GagalRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.Gagal(middleware.StasiunKlien(c), id, req.Galat); err != nil {
		return handleErr(c, err, "StasiunGagal error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Kegagalan dicatat")
}

func (h *CetakController) CetakStruk(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	antre, err := h.svc.CetakStruk(kid(c), id)
	if err != nil {
		return handleErr(c, err, "CetakStruk error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Struk diproses", dto.CetakHasil{Antre: antre})
}

func (h *CetakController) CetakClosing(c *fiber.Ctx) error {
	var req dto.ClosingPayload
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	antre, err := h.svc.CetakClosing(kid(c), req)
	if err != nil {
		return handleErr(c, err, "CetakClosing error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Bill closing diproses", dto.CetakHasil{Antre: antre})
}

func (h *CetakController) CetakRekapTenant(c *fiber.Ctx) error {
	var req dto.RekapTenantPayload
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	antre, err := h.svc.CetakRekapTenant(kid(c), req)
	if err != nil {
		return handleErr(c, err, "CetakRekapTenant error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Rekap diproses", dto.CetakHasil{Antre: antre})
}

func (h *CetakController) CetakShift(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	antre, err := h.svc.CetakShift(kid(c), id)
	if err != nil {
		return handleErr(c, err, "CetakShift error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Laporan shift diproses", dto.CetakHasil{Antre: antre})
}
