package utils

import (
	"warkop/internal/types"

	"github.com/gofiber/fiber/v2"
)

func SendSuccess(c *fiber.Ctx, status int, message string, data ...interface{}) error {
	res := types.APIResponse{Success: true, Status: status, Message: message}
	if len(data) > 0 && data[0] != nil {
		res.Data = data[0]
	}
	return c.Status(status).JSON(res)
}

func SendError(c *fiber.Ctx, status int, message string, errs ...[]types.FieldError) error {
	res := types.APIResponse{Success: false, Status: status, Message: message}
	if len(errs) > 0 && len(errs[0]) > 0 {
		res.Errors = errs[0]
	}
	return c.Status(status).JSON(res)
}
