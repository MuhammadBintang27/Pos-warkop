package app

import (
	"database/sql"
	"path/filepath"
	"strings"
	"time"

	"warkop/internal/config"
	"warkop/internal/controllers"
	"warkop/internal/middleware"
	"warkop/internal/routes"
	"warkop/internal/services"
	"warkop/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

type App struct {
	Fiber *fiber.App
	DB    *sql.DB
}

func New(db *sql.DB) *App {
	cfg := config.Get()

	fcfg := fiber.Config{ErrorHandler: errorHandler}
	if cfg.IsProduction() {
		fcfg.ProxyHeader = fiber.HeaderXForwardedFor // di belakang Caddy/nginx
	}
	f := fiber.New(fcfg)

	f.Use(middleware.RequestLogger)
	if !cfg.IsProduction() {
		f.Use(cors.New(cors.Config{AllowOrigins: "*"}))
	}
	// CSP dimatikan karena PrimeVue menyuntikkan style inline saat runtime.
	f.Use(helmet.New(helmet.Config{ContentSecurityPolicy: ""}))

	f.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	loginLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: 5 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return utils.SendError(c, fiber.StatusTooManyRequests, "Terlalu banyak percobaan login. Coba lagi nanti.")
		},
	})

	hub := services.NewEventHub()
	klienSvc := services.NewKlienService(db)
	if err := klienSvc.MuatAkses(); err != nil {
		panic(err)
	}
	authSvc := services.NewAuthService(db)
	pengaturanSvc := services.NewPengaturanService(db, hub)
	cetakSvc := services.NewCetakService(db, hub, pengaturanSvc)
	pesananSvc := services.NewPesananService(db, hub, cetakSvc)
	kasirCtrl := controllers.NewKasirController(
		services.NewPembayaranService(db, hub),
		services.NewShiftService(db, hub),
		services.NewKasbonService(db),
		services.NewLaporanService(db),
	)

	api := f.Group("/api")
	routes.RegisterAuthRoutes(api, controllers.NewAuthController(authSvc), loginLimiter)
	routes.RegisterPengaturanRoutes(api, controllers.NewPengaturanController(pengaturanSvc))
	routes.RegisterTenantRoutes(api, controllers.NewTenantController(services.NewTenantService(db, hub)))
	routes.RegisterUserRoutes(api, controllers.NewUserController(services.NewUserService(db)))
	routes.RegisterMenuRoutes(api, controllers.NewMenuController(services.NewMenuService(db, hub)))
	routes.RegisterMejaRoutes(api, controllers.NewMejaController(services.NewMejaService(db, hub)))
	routes.RegisterPesananRoutes(api, controllers.NewPesananController(pesananSvc), kasirCtrl)
	routes.RegisterKasirRoutes(api, kasirCtrl)
	routes.RegisterLaporanTenantRoutes(api, controllers.NewLaporanTenantController(services.NewLaporanTenantService(db)))
	routes.RegisterCetakRoutes(api, controllers.NewCetakController(cetakSvc), middleware.Stasiun(cetakSvc.VerifyToken))
	routes.RegisterRiwayatRoutes(api, controllers.NewRiwayatController(services.NewRiwayatService(db)))
	routes.RegisterBahanRoutes(api, controllers.NewBahanController(services.NewBahanService(db, hub)))
	routes.RegisterKaryawanRoutes(api, controllers.NewKaryawanController(services.NewKaryawanService(db), services.NewLabaService(db)))
	routes.RegisterAdminRoutes(api, controllers.NewKlienController(klienSvc, services.NewUserService(db)))
	routes.RegisterEventRoutes(api, controllers.NewEventController(hub))
	api.Use(func(c *fiber.Ctx) error {
		return utils.SendError(c, fiber.StatusNotFound, "Endpoint tidak ditemukan")
	})

	f.Static("/", cfg.StaticDir)
	f.Use(func(c *fiber.Ctx) error {
		if strings.HasPrefix(c.Path(), "/api") {
			return c.Next()
		}
		return c.SendFile(filepath.Join(cfg.StaticDir, "index.html"))
	})

	return &App{Fiber: f, DB: db}
}

func errorHandler(c *fiber.Ctx, err error) error {
	if c.Response().StatusCode() != fiber.StatusOK || len(c.Response().Body()) > 0 {
		return nil
	}
	if he, ok := utils.AsHTTPError(err); ok {
		return utils.SendError(c, he.Status, he.Message)
	}
	if fe, ok := err.(*fiber.Error); ok {
		return utils.SendError(c, fe.Code, fe.Message)
	}
	return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error")
}
