package services

import (
	"database/sql"
	"strings"
	"time"

	"warkop/internal/dto"
	"warkop/internal/types"
	"warkop/internal/utils"
)

type RiwayatService struct {
	db *sql.DB
}

const (
	riwayatLimit = 20
	waktuPesanan = "COALESCE(p.lunas_at, p.dibuat_at)"
	riwayatFrom  = `FROM pesanan p
		LEFT JOIN meja m ON m.id=p.meja_id
		JOIN users u ON u.id=p.dibuat_oleh `
)

func NewRiwayatService(db *sql.DB) *RiwayatService {
	return &RiwayatService{db: db}
}

// riwayatFilter membatasi hasil menurut peran: owner semua, kasir yang dibuat atau dibayar olehnya, pelayan yang dibuatnya.
func riwayatFilter(actor *utils.JWTPayload, q dto.RiwayatQuery, from, to time.Time) (string, []any) {
	where := "WHERE p.klien_id=? AND EXISTS (SELECT 1 FROM pesanan_item WHERE pesanan_id=p.id) AND " + waktuPesanan + ">=? AND " + waktuPesanan + "<?"
	args := []any{actor.KlienID, from, to}

	if q.Status != "" {
		where += " AND p.status=?"
		args = append(args, q.Status)
	}
	if term := strings.TrimSpace(q.Q); term != "" {
		like := "%" + term + "%"
		where += " AND (m.nama LIKE ? OR p.atas_nama LIKE ? OR CAST(p.id AS CHAR)=?)"
		args = append(args, like, like, term)
	}
	switch {
	case actor.Role == types.RoleOwner:
		if q.Oleh > 0 {
			where += " AND p.dibuat_oleh=?"
			args = append(args, q.Oleh)
		}
	case actor.Role == types.RoleKasir:
		where += " AND (p.dibuat_oleh=? OR EXISTS (SELECT 1 FROM pembayaran b WHERE b.pesanan_id=p.id AND b.diterima_oleh=?))"
		args = append(args, actor.UserID, actor.UserID)
	default:
		where += " AND p.dibuat_oleh=?"
		args = append(args, actor.UserID)
	}
	return where, args
}

func (s *RiwayatService) List(actor *utils.JWTPayload, q dto.RiwayatQuery) (*dto.RiwayatResponse, error) {
	from, to, err := parseRange(q.Dari, q.Sampai, time.Now())
	if err != nil {
		return nil, err
	}
	page, limit := max(q.Page, 1), q.Limit
	if limit == 0 {
		limit = riwayatLimit
	}
	where, args := riwayatFilter(actor, q, from, to)
	out := &dto.RiwayatResponse{Items: []dto.RiwayatItem{}, Page: page, Limit: limit}

	if err := s.db.QueryRow("SELECT COUNT(*) "+riwayatFrom+where, args...).Scan(&out.Total); err != nil {
		return nil, err
	}
	if err := s.db.QueryRow(`SELECT COUNT(*), COALESCE(SUM((SELECT SUM(jumlah) FROM pembayaran WHERE pesanan_id=p.id)),0) `+
		riwayatFrom+where+" AND p.status='lunas'", args...).Scan(&out.Lunas, &out.Omzet); err != nil {
		return nil, err
	}

	listArgs := append(append([]any{}, args...), limit, (page-1)*limit)
	rows, err := s.db.Query(`SELECT p.id, COALESCE(m.nama,''), p.atas_nama, p.status, p.dibuat_at, p.lunas_at, u.nama,
		COALESCE((SELECT k.nama FROM pembayaran b JOIN users k ON k.id=b.diterima_oleh WHERE b.pesanan_id=p.id LIMIT 1),''),
		(SELECT COALESCE(SUM(harga*qty),0) FROM pesanan_item WHERE pesanan_id=p.id) - p.diskon,
		(SELECT COALESCE(SUM(qty),0) FROM pesanan_item WHERE pesanan_id=p.id),
		(SELECT GROUP_CONCAT(DISTINCT metode ORDER BY metode) FROM pembayaran WHERE pesanan_id=p.id) `+
		riwayatFrom+where+" ORDER BY "+waktuPesanan+" DESC, p.id DESC LIMIT ? OFFSET ?", listArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var it dto.RiwayatItem
		var metode sql.NullString
		if err := rows.Scan(&it.ID, &it.MejaNama, &it.AtasNama, &it.Status, &it.DibuatAt, &it.LunasAt,
			&it.Pembuat, &it.Kasir, &it.Total, &it.JumlahItem, &metode); err != nil {
			return nil, err
		}
		it.Metode = metode.String
		out.Items = append(out.Items, it)
	}
	return out, rows.Err()
}
