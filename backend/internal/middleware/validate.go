package middleware

import (
	"errors"
	"reflect"
	"strings"
	"sync"

	"warkop/internal/types"
	"warkop/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var (
	v     *validator.Validate
	vOnce sync.Once
)

func Validator() *validator.Validate {
	vOnce.Do(func() {
		v = validator.New(validator.WithRequiredStructEnabled())
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
		v.RegisterValidation("role", func(fl validator.FieldLevel) bool {
			return types.Role(fl.Field().String()).Valid()
		})
		v.RegisterValidation("primary_color", func(fl validator.FieldLevel) bool {
			return types.ValidPrimary(fl.Field().String())
		})
		v.RegisterValidation("metode", func(fl validator.FieldLevel) bool {
			return types.Metode(fl.Field().String()).Valid()
		})
	})
	return v
}

func ParseAndValidate(c *fiber.Ctx, dst interface{}) error {
	if err := c.BodyParser(dst); err != nil {
		_ = utils.SendError(c, fiber.StatusBadRequest, "Format data tidak valid",
			[]types.FieldError{{Field: "body", Message: err.Error()}})
		return fiber.ErrBadRequest
	}
	return validateStruct(c, dst)
}

func ParseQuery(c *fiber.Ctx, dst interface{}) error {
	if err := c.QueryParser(dst); err != nil {
		_ = utils.SendError(c, fiber.StatusBadRequest, "Parameter query tidak valid",
			[]types.FieldError{{Field: "query", Message: err.Error()}})
		return fiber.ErrBadRequest
	}
	return validateStruct(c, dst)
}

func validateStruct(c *fiber.Ctx, dst interface{}) error {
	if err := Validator().Struct(dst); err != nil {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			fields := make([]types.FieldError, 0, len(verrs))
			for _, fe := range verrs {
				fields = append(fields, types.FieldError{Field: fe.Field(), Message: humanMessage(fe)})
			}
			_ = utils.SendError(c, fiber.StatusBadRequest, "Validasi gagal", fields)
			return fiber.ErrBadRequest
		}
		return err
	}
	return nil
}

func humanMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "min", "gte":
		return "minimal " + fe.Param()
	case "max", "lte":
		return "maksimal " + fe.Param()
	case "gt":
		return "harus lebih dari " + fe.Param()
	}
	return "tidak valid"
}
