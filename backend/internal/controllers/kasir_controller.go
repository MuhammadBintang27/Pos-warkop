package controllers

import (
	"warkop/internal/dto"
	"warkop/internal/middleware"
	"warkop/internal/services"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type KasirController struct {
	pembayaran *services.PembayaranService
	shift      *services.ShiftService
	kasbon     *services.KasbonService
	laporan    *services.LaporanService
}

func NewKasirController(
	pembayaran *services.PembayaranService,
	shift *services.ShiftService,
	kasbon *services.KasbonService,
	laporan *services.LaporanService,
) *KasirController {
	return &KasirController{pembayaran: pembayaran, shift: shift, kasbon: kasbon, laporan: laporan}
}

func (h *KasirController) Bayar(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.BayarRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	p, err := h.pembayaran.Bayar(middleware.CurrentUser(c), id, req)
	if err != nil {
		return handleErr(c, err, "Bayar error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Pembayaran berhasil", p)
}

func (h *KasirController) ShiftAktif(c *fiber.Ctx) error {
	sh, err := h.shift.Aktif(kid(c))
	if err != nil {
		return handleErr(c, err, "ShiftAktif error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Shift aktif berhasil diambil", sh)
}

func (h *KasirController) BukaShift(c *fiber.Ctx) error {
	var req dto.BukaShiftRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	sh, err := h.shift.Buka(kid(c), middleware.CurrentUser(c).UserID, req)
	if err != nil {
		return handleErr(c, err, "BukaShift error")
	}
	return utils.SendSuccess(c, fiber.StatusCreated, "Shift dibuka", sh)
}

func (h *KasirController) TutupShift(c *fiber.Ctx) error {
	var req dto.TutupShiftRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	sh, err := h.shift.Tutup(kid(c), middleware.CurrentUser(c).UserID, req)
	if err != nil {
		return handleErr(c, err, "TutupShift error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Shift ditutup", sh)
}

func (h *KasirController) ListPengeluaran(c *fiber.Ctx) error {
	list, err := h.shift.ListPengeluaran(kid(c))
	if err != nil {
		return handleErr(c, err, "ListPengeluaran error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Pengeluaran berhasil diambil", list)
}

func (h *KasirController) CreatePengeluaran(c *fiber.Ctx) error {
	var req dto.PengeluaranRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.shift.CreatePengeluaran(kid(c), middleware.CurrentUser(c).UserID, req); err != nil {
		return handleErr(c, err, "CreatePengeluaran error")
	}
	return utils.SendSuccess(c, fiber.StatusCreated, "Pengeluaran dicatat")
}

func (h *KasirController) ListPelanggan(c *fiber.Ctx) error {
	list, err := h.kasbon.List(kid(c))
	if err != nil {
		return handleErr(c, err, "ListPelanggan error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Pelanggan berhasil diambil", list)
}

func (h *KasirController) CreatePelanggan(c *fiber.Ctx) error {
	var req dto.PelangganRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.kasbon.Create(kid(c), req); err != nil {
		return handleErr(c, err, "CreatePelanggan error")
	}
	return utils.SendSuccess(c, fiber.StatusCreated, "Pelanggan ditambahkan")
}

func (h *KasirController) Lunasi(c *fiber.Ctx) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return handleErr(c, err, "")
	}
	var req dto.LunasiRequest
	if err := middleware.ParseAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.kasbon.Lunasi(kid(c), id, req); err != nil {
		return handleErr(c, err, "Lunasi error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Pelunasan dicatat")
}

func (h *KasirController) LaporanTenant(c *fiber.Ctx) error {
	res, err := h.laporan.Tenant(kid(c), c.Query("dari"), c.Query("sampai"))
	if err != nil {
		return handleErr(c, err, "LaporanTenant error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Setoran tenant berhasil diambil", res)
}

func (h *KasirController) Laporan(c *fiber.Ctx) error {
	res, err := h.laporan.Get(kid(c), c.Query("dari"), c.Query("sampai"))
	if err != nil {
		return handleErr(c, err, "Laporan error")
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Laporan berhasil diambil", res)
}
