package routes

import (
	"warkop/internal/controllers"
	"warkop/internal/middleware"
	"warkop/internal/types"

	"github.com/gofiber/fiber/v2"
)

var (
	authed = middleware.Authenticate
	kasir  = middleware.Authorize(types.KasirUp...)
	owner  = middleware.Authorize(types.RoleOwner)
	super  = middleware.Authorize(types.RoleSuper)
)

func RegisterAuthRoutes(r fiber.Router, h *controllers.AuthController, loginLimiter fiber.Handler) {
	r.Post("/auth/login", loginLimiter, h.Login)
}

func RegisterPengaturanRoutes(r fiber.Router, h *controllers.PengaturanController) {
	r.Get("/pengaturan", middleware.OptionalAuthenticate, h.Get) // tanpa login: tema bawaan, supaya halaman login tetap bertema
	r.Put("/pengaturan", authed, owner, h.Update)
}

func RegisterTenantRoutes(r fiber.Router, h *controllers.TenantController) {
	r.Get("/tenant", authed, h.List)
	r.Post("/tenant", authed, owner, h.Create)
	r.Put("/tenant/:id", authed, owner, h.Update)
	r.Post("/tenant/:id/buka", authed, kasir, h.Buka)
	r.Delete("/tenant/:id", authed, owner, h.Delete)
}

func RegisterUserRoutes(r fiber.Router, h *controllers.UserController) {
	r.Get("/users", authed, owner, h.List)
	r.Post("/users", authed, owner, h.Create)
	r.Put("/users/:id", authed, owner, h.Update)
}

func RegisterMenuRoutes(r fiber.Router, h *controllers.MenuController) {
	r.Get("/menu", authed, h.List)
	r.Post("/menu", authed, owner, h.Create)
	r.Put("/menu/:id", authed, owner, h.Update)
	r.Delete("/menu/:id", authed, owner, h.Delete)
	r.Post("/menu/:id/stok", authed, kasir, h.SetStok)

	r.Post("/kategori", authed, owner, h.CreateKategori)
	r.Put("/kategori/:id", authed, owner, h.UpdateKategori)
	r.Delete("/kategori/:id", authed, owner, h.DeleteKategori)
}

func RegisterMejaRoutes(r fiber.Router, h *controllers.MejaController) {
	r.Get("/meja", authed, h.Denah)
	r.Post("/meja", authed, owner, h.Create)
	r.Post("/meja/generate", authed, owner, h.Generate)
	r.Put("/meja/posisi", authed, owner, h.Posisi)
	r.Put("/meja/:id", authed, owner, h.Update)
	r.Delete("/meja/:id", authed, owner, h.Delete)
}

func RegisterPesananRoutes(r fiber.Router, h *controllers.PesananController, k *controllers.KasirController) {
	r.Post("/pesanan", authed, h.Create)
	r.Get("/pesanan/:id", authed, h.Get)
	r.Post("/pesanan/:id/item", authed, h.AddItem)
	r.Put("/pesanan/:id/item/:itemId", authed, h.UpdateItem)
	r.Delete("/pesanan/:id/item/:itemId", authed, h.DeleteItem)
	r.Post("/pesanan/:id/kirim", authed, h.Kirim)
	r.Post("/pesanan/:id/pindah", authed, h.Pindah)
	r.Put("/pesanan/:id/diskon", authed, kasir, h.SetDiskon)
	r.Post("/pesanan/:id/batal", authed, kasir, h.Batal)
	r.Post("/pesanan/:id/item/:itemId/batal", authed, kasir, h.BatalItem)
	r.Post("/pesanan/:id/gabung", authed, kasir, h.Gabung)
	r.Post("/pesanan/:id/pisah", authed, kasir, h.Pisah)
	r.Post("/pesanan/:id/bayar", authed, kasir, k.Bayar)
}

func RegisterKasirRoutes(r fiber.Router, h *controllers.KasirController) {
	r.Get("/shift/aktif", authed, kasir, h.ShiftAktif)
	r.Post("/shift/buka", authed, kasir, h.BukaShift)
	r.Post("/shift/tutup", authed, kasir, h.TutupShift)
	r.Get("/pengeluaran", authed, kasir, h.ListPengeluaran)
	r.Post("/pengeluaran", authed, kasir, h.CreatePengeluaran)
	r.Get("/pelanggan", authed, kasir, h.ListPelanggan)
	r.Post("/pelanggan", authed, kasir, h.CreatePelanggan)
	r.Post("/pelanggan/:id/lunasi", authed, kasir, h.Lunasi)
	r.Get("/laporan", authed, owner, h.Laporan)
	r.Get("/laporan/tenant", authed, owner, h.LaporanTenant)
}

func RegisterCetakRoutes(r fiber.Router, h *controllers.CetakController, stasiun fiber.Handler) {
	r.Get("/printer", authed, owner, h.ListPrinter)
	r.Post("/printer", authed, owner, h.CreatePrinter)
	r.Put("/printer/:id", authed, owner, h.UpdatePrinter)
	r.Delete("/printer/:id", authed, owner, h.DeletePrinter)
	r.Post("/printer/:id/tes", authed, owner, h.TesPrinter)
	r.Post("/stasiun/token", authed, owner, h.BuatToken)
	r.Get("/cetak/riwayat", authed, owner, h.Riwayat)
	r.Post("/cetak/:id/ulang", authed, kasir, h.Ulang)
	r.Post("/pesanan/:id/cetak-struk", authed, h.CetakStruk)
	r.Post("/cetak/closing", authed, owner, h.CetakClosing)
	r.Get("/cetak/status", authed, kasir, h.Status)

	r.Get("/stasiun/jobs", stasiun, h.StasiunJobs)
	r.Post("/stasiun/jobs/:id/selesai", stasiun, h.StasiunSelesai)
	r.Post("/stasiun/jobs/:id/gagal", stasiun, h.StasiunGagal)
}

func RegisterLaporanTenantRoutes(r fiber.Router, h *controllers.LaporanTenantController) {
	r.Get("/laporan/tenant/harian", authed, owner, h.Harian)
	r.Get("/laporan/tenant/detail", authed, owner, h.Detail)
	r.Get("/laporan/tenant/belum-bayar", authed, owner, h.BelumBayar)
	r.Post("/closing", authed, owner, h.Closing)
	r.Post("/closing/:id/setor", authed, owner, h.Setor)
	r.Delete("/closing/:id", authed, owner, h.BatalClosing)
}

func RegisterRiwayatRoutes(r fiber.Router, h *controllers.RiwayatController) {
	r.Get("/riwayat", authed, h.List)
}

func RegisterBahanRoutes(r fiber.Router, h *controllers.BahanController) {
	r.Get("/bahan", authed, kasir, h.List)
	r.Get("/bahan/pemakaian", authed, owner, h.Pemakaian)
	r.Post("/bahan", authed, owner, h.Create)
	r.Put("/bahan/:id", authed, owner, h.Update)
	r.Delete("/bahan/:id", authed, owner, h.Delete)
	r.Post("/bahan/:id/restock", authed, kasir, h.Restock)
	r.Post("/bahan/:id/opname", authed, kasir, h.Opname)
	r.Get("/bahan/:id/riwayat", authed, kasir, h.Riwayat)
}

func RegisterKaryawanRoutes(r fiber.Router, h *controllers.KaryawanController) {
	r.Get("/karyawan", authed, owner, h.List)
	r.Post("/karyawan", authed, owner, h.Create)
	r.Put("/karyawan/:id", authed, owner, h.Update)
	r.Delete("/karyawan/:id", authed, owner, h.Delete)
	r.Get("/karyawan/:id/bon", authed, owner, h.ListBon)
	r.Post("/karyawan/:id/bon", authed, owner, h.TambahBon)
	r.Get("/gaji", authed, owner, h.ListSlip)
	r.Post("/gaji", authed, owner, h.BuatSlip)
	r.Delete("/gaji/:id", authed, owner, h.HapusSlip)
	r.Post("/gaji/:id/bayar", authed, owner, h.BayarSlip)
	r.Get("/laba-rugi", authed, owner, h.LabaRugi)
	r.Get("/margin-menu", authed, owner, h.MarginMenu)
}

func RegisterAdminRoutes(r fiber.Router, h *controllers.KlienController) {
	r.Get("/admin/klien", authed, super, h.List)
	r.Post("/admin/klien", authed, super, h.Create)
	r.Put("/admin/klien/:id", authed, super, h.Edit)
	r.Post("/admin/klien/:id/akses", authed, super, h.SetAktif)
	r.Post("/admin/klien/:id/langganan", authed, super, h.Perpanjang)
	r.Post("/admin/klien/:id/reset-owner", authed, super, h.ResetOwner)
	r.Get("/admin/klien/:id/users", authed, super, h.ListUsers)
	r.Post("/admin/klien/:id/users", authed, super, h.CreateUser)
	r.Put("/admin/klien/:id/users/:uid", authed, super, h.UpdateUser)
}

func RegisterEventRoutes(r fiber.Router, h *controllers.EventController) {
	r.Get("/events", authed, h.Stream)
}
