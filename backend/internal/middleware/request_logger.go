package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func RequestLogger(c *fiber.Ctx) error {
	start := time.Now()
	err := c.Next()

	status := c.Response().StatusCode()
	evt := log.Info()
	switch {
	case status >= 500:
		evt = log.Error()
	case status >= 400:
		evt = log.Warn()
	}
	evt.Str("method", c.Method()).Str("path", c.Path()).Int("status", status).
		Dur("latency", time.Since(start)).Str("ip", c.IP()).Msg("request")
	return err
}
