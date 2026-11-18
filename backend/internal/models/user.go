package models

import "warkop/internal/types"

type User struct {
	ID           int64
	KlienID      int64
	Nama         string
	Username     string
	PasswordHash string
	Role         types.Role
	Aktif        bool
}
