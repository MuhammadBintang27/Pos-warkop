package utils

import (
	"database/sql"
	"errors"
	"strconv"

	"warkop/internal/types"

	"github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
)

const (
	MySQLDuplicate  uint16 = 1062
	MySQLFKRestrict uint16 = 1451
)

func MySQLErrNo(err error) uint16 {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number
	}
	return 0
}

func WithTx(db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func ParamID(c *fiber.Ctx, name string) (int64, error) {
	id, err := strconv.ParseInt(c.Params(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, types.BadRequest("ID tidak valid")
	}
	return id, nil
}
