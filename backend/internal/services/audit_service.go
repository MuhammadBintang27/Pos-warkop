package services

import (
	"database/sql"
	"time"

	"warkop/internal/dto"
)

type AuditService struct {
	db *sql.DB
}

func NewAuditService(db *sql.DB) *AuditService {
	return &AuditService{db: db}
}

const auditPerHalaman = 50

func (s *AuditService) List(kid int64, q dto.AuditQuery) (*dto.AuditResponse, error) {
	from, to, err := parseRange(q.Dari, q.Sampai, time.Now())
	if err != nil {
		return nil, err
	}
	where := " FROM audit_log a LEFT JOIN users u ON u.id=a.user_id WHERE a.klien_id=? AND a.at>=? AND a.at<?"
	args := []any{kid, from, to}
	if q.Oleh > 0 {
		where += " AND a.user_id=?"
		args = append(args, q.Oleh)
	}
	out := &dto.AuditResponse{Items: []dto.AuditItem{}}
	if err := s.db.QueryRow("SELECT COUNT(*)"+where, args...).Scan(&out.Total); err != nil {
		return nil, err
	}
	rows, err := s.db.Query("SELECT a.id,COALESCE(u.nama,'?'),a.aksi,a.detail,a.at"+where+" ORDER BY a.id DESC LIMIT ? OFFSET ?",
		append(args, auditPerHalaman, q.Page*auditPerHalaman)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var a dto.AuditItem
		if err := rows.Scan(&a.ID, &a.Nama, &a.Aksi, &a.Detail, &a.At); err != nil {
			return nil, err
		}
		out.Items = append(out.Items, a)
	}
	return out, rows.Err()
}
