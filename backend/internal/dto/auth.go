package dto

import "warkop/internal/types"

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	ID       int64      `json:"id"`
	KlienID  int64      `json:"klienId"`
	Nama     string     `json:"nama"`
	Username string     `json:"username"`
	Role     types.Role `json:"role"`
	Aktif    bool       `json:"aktif"`
	TenantID int64      `json:"tenantId"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type CreateUserRequest struct {
	Nama     string `json:"nama" validate:"required,max=100"`
	Username string `json:"username" validate:"required,max=50"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,role"`
	TenantID int64  `json:"tenantId"` // wajib bila role tenant
}

type UpdateUserRequest struct {
	Nama     string `json:"nama" validate:"required,max=100"`
	Role     string `json:"role" validate:"required,role"`
	Aktif    *bool  `json:"aktif" validate:"required"`
	Password string `json:"password" validate:"omitempty,min=6"`
	TenantID int64  `json:"tenantId"` // wajib bila role tenant
}
