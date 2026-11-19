package dto

import "time"

type AuditQuery struct {
	Dari   string `query:"dari"`
	Sampai string `query:"sampai"`
	Oleh   int64  `query:"oleh" validate:"gte=0"`
	Page   int    `query:"page" validate:"gte=0"`
}

type AuditItem struct {
	ID     int64     `json:"id"`
	Nama   string    `json:"nama"`
	Aksi   string    `json:"aksi"`
	Detail string    `json:"detail"`
	At     time.Time `json:"at"`
}

type AuditResponse struct {
	Items []AuditItem `json:"items"`
	Total int         `json:"total"`
}
