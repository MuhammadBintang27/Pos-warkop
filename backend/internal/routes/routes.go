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
	// tenantRole: akun tenant tanpa tenant_id (data rusak) ditolak di service karena tenant 0 tidak pernah ada.
	tenantRole = middleware.Authorize(types.RoleTenant)
)

func RegisterAuthRoutes(r fiber.Router, h *controllers.AuthController, loginLimiter fiber.Handler) {
	r.Post("/auth/login", loginLimiter, h.Login)
}

func RegisterPengaturanRoutes(r fiber.Router, h *controllers.PengaturanController) {
	r.Get("/pengaturan", middleware.OptionalAuthenticate, h.Get) // tanpa login: tema bawaan, supaya halaman login tetap bertema
	r.Put("/pengaturan", authed, owner, middleware.Audit("Ubah pengaturan"), h.Update)
}

func RegisterTenantRoutes(r fiber.Router, h *controllers.TenantController) {
	r.Get("/tenant", authed, h.List)
	r.Post("/tenant", authed, owner, middleware.Audit("Tambah tenant"), h.Create)
	r.Put("/tenant/:id", authed, owner, middleware.Audit("Ubah tenant"), h.Update)
	r.Post("/tenant/:id/buka", authed, kasir, h.Buka)
	r.Delete("/tenant/:id", authed, owner, middleware.Audit("Hapus tenant"), h.Delete)

	r.Get("/tenant-saya/pesanan", authed, tenantRole, h.PesananMasuk)
	r.Post("/tenant-saya/item/:id/siap", authed, tenantRole, h.SetSiap)
}

func RegisterUserRoutes(r fiber.Router, h *controllers.UserController) {
	r.Get("/users", authed, owner, h.List)
	r.Post("/users", authed, owner, middleware.Audit("Tambah pengguna"), h.Create)
	r.Put("/users/:id", authed, owner, middleware.Audit("Ubah pengguna"), h.Update)
}

func RegisterMenuRoutes(r fiber.Router, h *controllers.MenuController) {
	r.Get("/menu", authed, h.List)
	r.Post("/menu", authed, owner, middleware.Audit("Tambah menu"), h.Create)
	r.Put("/menu/:id", authed, owner, middleware.Audit("Ubah menu"), h.Update)
	r.Delete("/menu/:id", authed, owner, middleware.Audit("Hapus menu"), h.Delete)
	r.Post("/menu/:id/stok", authed, kasir, middleware.Audit("Ubah stok menu"), h.SetStok)

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
	r.Put("/pesanan/:id/diskon", authed, kasir, middleware.Audit("Diskon bill"), h.SetDiskon)
	r.Post("/pesanan/:id/batal", authed, kasir, middleware.Audit("Batal bill"), h.Batal)
	r.Post("/pesanan/:id/item/:itemId/batal", authed, kasir, middleware.Audit("Batal item"), h.BatalItem)
	r.Post("/pesanan/:id/gabung", authed, kasir, h.Gabung)
	r.Post("/pesanan/:id/pisah", authed, kasir, h.Pisah)
	r.Post("/pesanan/:id/bayar", authed, kasir, k.Bayar)
	r.Post("/pesanan/:id/buka-ulang", authed, owner, middleware.Audit("Buka ulang bill"), k.BukaUlang)
	r.Post("/pesanan/:id/refund", authed, owner, middleware.Audit("Refund bill"), k.Refund)
}

func RegisterKasirRoutes(r fiber.Router, h *controllers.KasirController) {
	r.Get("/shift/aktif", authed, kasir, h.ShiftAktif)
	r.Post("/shift/buka", authed, kasir, middleware.Audit("Buka shift"), h.BukaShift)
	r.Post("/shift/tutup", authed, kasir, middleware.Audit("Tutup shift"), h.TutupShift)
	r.Get("/pengeluaran", authed, kasir, h.ListPengeluaran)
	r.Post("/pengeluaran", authed, kasir, middleware.Audit("Catat pengeluaran"), h.CreatePengeluaran)
	r.Get("/pelanggan", authed, kasir, h.ListPelanggan)
	r.Post("/pelanggan", authed, kasir, h.CreatePelanggan)
	r.Post("/pelanggan/:id/lunasi", authed, kasir, middleware.Audit("Pelunasan kasbon"), h.Lunasi)
	r.Get("/laporan", authed, owner, h.Laporan)
	r.Get("/laporan/tenant", authed, owner, h.LaporanTenant)
	r.Get("/laporan/analitik", authed, owner, h.Analitik)
}

func RegisterCetakRoutes(r fiber.Router, h *controllers.CetakController, stasiun fiber.Handler) {
	r.Get("/printer", authed, owner, h.ListPrinter)
	r.Post("/printer", authed, owner, h.CreatePrinter)
	r.Put("/printer/:id", authed, owner, h.UpdatePrinter)
	r.Delete("/printer/:id", authed, owner, h.DeletePrinter)
	r.Post("/printer/:id/tes", authed, owner, h.TesPrinter)
	r.Post("/stasiun/token", authed, owner, h.BuatToken)
	r.Get("/cetak/riwayat", authed, owner, h.Riwayat)
	r.Post("/cetak/:id/ulang", authed, kasir, middleware.Audit("Cetak ulang"), h.Ulang)
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
	r.Post("/closing", authed, owner, middleware.Audit("Closing tenant"), h.Closing)
	r.Post("/closing/:id/setor", authed, owner, middleware.Audit("Setor closing"), h.Setor)
	r.Delete("/closing/:id", authed, owner, middleware.Audit("Batal closing"), h.BatalClosing)

	// Rekapan untuk akun tenant: handler yang sama, tenant dipaksa dari token.
	r.Get("/tenant-saya/laporan/harian", authed, tenantRole, h.Harian)
	r.Get("/tenant-saya/laporan/detail", authed, tenantRole, h.Detail)
	r.Get("/tenant-saya/laporan/belum-bayar", authed, tenantRole, h.BelumBayar)
}

func RegisterRiwayatRoutes(r fiber.Router, h *controllers.RiwayatController) {
	r.Get("/riwayat", authed, h.List)
}

func RegisterBahanRoutes(r fiber.Router, h *controllers.BahanController) {
	r.Get("/bahan", authed, kasir, h.List)
	r.Get("/bahan/pemakaian", authed, owner, h.Pemakaian)
	r.Post("/bahan", authed, owner, h.Create)
	r.Put("/bahan/:id", authed, owner, middleware.Audit("Ubah bahan"), h.Update)
	r.Delete("/bahan/:id", authed, owner, middleware.Audit("Hapus bahan"), h.Delete)
	r.Post("/bahan/:id/restock", authed, kasir, middleware.Audit("Restock bahan"), h.Restock)
	r.Post("/bahan/:id/opname", authed, kasir, middleware.Audit("Opname bahan"), h.Opname)
	r.Get("/bahan/:id/riwayat", authed, kasir, h.Riwayat)
}

func RegisterKaryawanRoutes(r fiber.Router, h *controllers.KaryawanController) {
	r.Get("/karyawan", authed, owner, h.List)
	r.Post("/karyawan", authed, owner, h.Create)
	r.Put("/karyawan/:id", authed, owner, h.Update)
	r.Delete("/karyawan/:id", authed, owner, h.Delete)
	r.Get("/karyawan/:id/bon", authed, owner, h.ListBon)
	r.Post("/karyawan/:id/bon", authed, owner, middleware.Audit("Bon karyawan"), h.TambahBon)
	r.Get("/gaji", authed, owner, h.ListSlip)
	r.Post("/gaji", authed, owner, h.BuatSlip)
	r.Delete("/gaji/:id", authed, owner, h.HapusSlip)
	r.Post("/gaji/:id/bayar", authed, owner, middleware.Audit("Bayar gaji"), h.BayarSlip)
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

func RegisterPerangkatRoutes(r fiber.Router, h *controllers.PerangkatController) {
	r.Post("/tenant-saya/perangkat", authed, tenantRole, h.Simpan)
	r.Post("/tenant-saya/perangkat/hapus", authed, tenantRole, h.Hapus)
}

func RegisterEventRoutes(r fiber.Router, h *controllers.EventController) {
	r.Get("/events", authed, h.Stream)
}

func RegisterAuditRoutes(r fiber.Router, h *controllers.AuditController) {
	r.Get("/audit", authed, owner, h.List)
}
