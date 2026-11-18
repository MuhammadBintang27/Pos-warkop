package services

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"warkop/internal/dto"
	"warkop/internal/types"
	"warkop/internal/utils"
)

type CetakService struct {
	db   *sql.DB
	hub  *EventHub
	peng *PengaturanService

	mu       sync.Mutex
	terakhir map[int64]time.Time // ponytail: waktu poll stasiun terakhir disimpan di memori; hilang saat server restart dan tidak dibagi antar instance
}

const (
	jenisDapur   = "dapur"
	jenisBatal   = "batal"
	jenisTes     = "tes"
	jenisStruk   = "struk"
	jenisClosing = "closing"

	masaAktifJob        = 6 * time.Hour // tiket yang lebih lama tidak dicetak lagi supaya tidak keluar sebagai tiket basi
	macetSetelah        = 20 * time.Second
	stasiunHidupSetelah = 15 * time.Second
	kunciTokenStasiun   = "token_stasiun"
)

func NewCetakService(db *sql.DB, hub *EventHub, peng *PengaturanService) *CetakService {
	return &CetakService{db: db, hub: hub, peng: peng, terakhir: map[int64]time.Time{}}
}

func batasAktif() time.Time { return time.Now().Add(-masaAktifJob) }

// Antrekan membuat satu tugas cetak per tenant yang punya printer. Mengembalikan tenant mana yang sudah dijadwalkan,
// sehingga klien hanya mencetak sendiri untuk tenant yang belum punya printer.
func (s *CetakService) Antrekan(tx *sql.Tx, kid int64, jenis string, t *dto.KirimResponse) (map[int64]bool, error) {
	antrean := map[int64]bool{}
	var urut []int64
	items := map[int64][]dto.TiketItem{}
	nama := map[int64]string{}
	for _, i := range t.Items {
		if _, ada := items[i.TenantID]; !ada {
			urut = append(urut, i.TenantID)
			nama[i.TenantID] = i.TenantNama
		}
		items[i.TenantID] = append(items[i.TenantID], dto.TiketItem{Qty: i.Qty, Nama: i.Nama, Opsi: i.Opsi, Catatan: i.Catatan})
	}

	for _, id := range urut {
		var printerID int64
		var lebarMM int
		err := tx.QueryRow("SELECT p.id,p.lebar_mm FROM tenant t JOIN printer p ON p.id=t.printer_id AND p.aktif=1 WHERE t.id=? AND t.klien_id=?", id, kid).Scan(&printerID, &lebarMM)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return nil, err
		}
		judul := "BATAL"
		if jenis == jenisDapur {
			judul = "PESANAN BARU"
			if t.Tambahan[id] {
				judul = "PESANAN TAMBAHAN"
			}
		}
		if err := s.simpanJob(tx, kid, printerID, lebarMM, jenis, dto.TiketPayload{
			Judul: judul, MejaNama: t.MejaNama, AtasNama: t.AtasNama, Oleh: t.Oleh, Tenant: nama[id],
			Waktu: time.Now(), Items: items[id],
		}); err != nil {
			return nil, err
		}
		antrean[id] = true
	}
	return antrean, nil
}

// simpanJob menyimpan tiket (untuk riwayat) beserta perintah printer yang sudah siap kirim.
func (s *CetakService) simpanJob(tx *sql.Tx, kid, printerID int64, lebarMM int, jenis string, p dto.TiketPayload) error {
	return s.simpanData(tx, kid, printerID, jenis, p, escposTiket(p, lebarMM))
}

func (s *CetakService) simpanData(tx *sql.Tx, kid, printerID int64, jenis string, payload any, data []byte) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = tx.Exec("INSERT INTO cetak_job (klien_id,printer_id,jenis,payload,data,dibuat_at) VALUES (?,?,?,?,?,?)",
		kid, printerID, jenis, string(raw), data, time.Now())
	return err
}

// printerKasir mengembalikan printer struk klien; ok=false bila belum diatur atau printernya sudah dihapus.
func (s *CetakService) printerKasir(kid int64) (usaha *dto.PengaturanRequest, printerID int64, lebarMM int, ok bool, err error) {
	if usaha, err = s.peng.Get(kid); err != nil || usaha.PrinterKasirID == nil {
		return
	}
	err = s.db.QueryRow("SELECT id,lebar_mm FROM printer WHERE id=? AND klien_id=? AND aktif=1", *usaha.PrinterKasirID, kid).Scan(&printerID, &lebarMM)
	if err == sql.ErrNoRows {
		return usaha, 0, 0, false, nil
	}
	return usaha, printerID, lebarMM, err == nil, err
}

// gabungStruk menyatukan baris yang tercetak sama (nama, varian, harga). Catatan tidak dicetak di struk,
// jadi tidak ikut membedakan; dapur tetap menerima tiket per pengiriman.
func gabungStruk(items []dto.ItemResponse) []dto.StrukItem {
	var out []dto.StrukItem
	idx := map[string]int{}
	for _, i := range items {
		k := i.Nama + "|" + strconv.Itoa(i.Harga) + "|" + i.Opsi
		if n, ada := idx[k]; ada {
			out[n].Qty += i.Qty
			continue
		}
		idx[k] = len(out)
		out = append(out, dto.StrukItem{Qty: i.Qty, Nama: i.Nama, Opsi: i.Opsi, Harga: i.Harga})
	}
	return out
}

// CetakStruk mengantrekan struk pesanan lunas ke printer kasir. Hasil false bila printer kasir belum diatur.
func (s *CetakService) CetakStruk(kid, pesananID int64) (bool, error) {
	usaha, printerID, lebarMM, ok, err := s.printerKasir(kid)
	if err != nil || !ok {
		return false, err
	}
	p, err := loadPesanan(s.db, kid, pesananID)
	if err != nil {
		return false, err
	}
	if p.Status != types.StatusLunas {
		return false, types.BadRequest("Struk hanya untuk pesanan yang sudah lunas")
	}
	payload := dto.StrukPayload{
		Judul: "STRUK", MejaNama: p.MejaNama, Usaha: usaha.NamaUsaha, Alamat: usaha.Alamat, Footer: usaha.FooterStruk,
		ID: p.ID, AtasNama: p.AtasNama, Kasir: p.Kasir, Items: gabungStruk(p.Items),
		Subtotal: p.Subtotal, Diskon: p.Diskon, Total: p.Total,
	}
	payload.Waktu = p.DibuatAt
	if p.LunasAt != nil {
		payload.Waktu = *p.LunasAt
	}
	for _, b := range p.Pembayaran {
		payload.Bayar = append(payload.Bayar, dto.StrukBayar{Metode: string(b.Metode), Pelanggan: b.Pelanggan, Diterima: b.Diterima, Kembalian: b.Kembalian})
	}
	err = utils.WithTx(s.db, func(tx *sql.Tx) error {
		return s.simpanData(tx, kid, printerID, jenisStruk, payload, escposStruk(payload, lebarMM))
	})
	if err != nil {
		return false, err
	}
	s.hub.Publish(kid, "cetak")
	return true, nil
}

// CetakClosing mengantrekan bill closing tenant. Isinya berasal dari layar closing owner; nama usaha diambil dari pengaturan.
// shortcut: angka tidak dihitung ulang di server; owner yang mengirim hanya bisa mencetak angka palsu di printernya sendiri.
func (s *CetakService) CetakClosing(kid int64, p dto.ClosingPayload) (bool, error) {
	usaha, printerID, lebarMM, ok, err := s.printerKasir(kid)
	if err != nil || !ok {
		return false, err
	}
	p.Judul, p.Usaha, p.MejaNama = "BILL CLOSING", usaha.NamaUsaha, ""
	err = utils.WithTx(s.db, func(tx *sql.Tx) error {
		return s.simpanData(tx, kid, printerID, jenisClosing, p, escposClosing(p, lebarMM))
	})
	if err != nil {
		return false, err
	}
	s.hub.Publish(kid, "cetak")
	return true, nil
}

// ---- printer (owner) ----

func (s *CetakService) ListPrinter(kid int64) ([]dto.PrinterResponse, error) {
	rows, err := s.db.Query(`SELECT p.id,p.nama,p.ip,p.port,p.lebar_mm,
		(SELECT COUNT(*) FROM cetak_job j WHERE j.printer_id=p.id AND j.status='antri' AND j.dibuat_at>?)
		FROM printer p WHERE p.klien_id=? AND p.aktif=1 ORDER BY p.id`, batasAktif(), kid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []dto.PrinterResponse{}
	for rows.Next() {
		var p dto.PrinterResponse
		if err := rows.Scan(&p.ID, &p.Nama, &p.IP, &p.Port, &p.LebarMM, &p.Antri); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *CetakService) CreatePrinter(kid int64, req dto.PrinterRequest) error {
	_, err := s.db.Exec("INSERT INTO printer (klien_id,nama,ip,port,lebar_mm) VALUES (?,?,?,?,?)", kid, req.Nama, req.IP, req.Port, req.LebarMM)
	return err
}

func (s *CetakService) UpdatePrinter(kid, id int64, req dto.PrinterRequest) error {
	if _, err := s.db.Exec("UPDATE printer SET nama=?, ip=?, port=?, lebar_mm=? WHERE id=? AND klien_id=? AND aktif=1", req.Nama, req.IP, req.Port, req.LebarMM, id, kid); err != nil {
		return err
	}
	var ada int
	if s.db.QueryRow("SELECT 1 FROM printer WHERE id=? AND klien_id=? AND aktif=1", id, kid).Scan(&ada) != nil {
		return types.NotFound("Printer")
	}
	return nil
}

func (s *CetakService) DeletePrinter(kid, id int64) error {
	res, err := s.db.Exec("UPDATE printer SET aktif=0 WHERE id=? AND klien_id=? AND aktif=1", id, kid)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return types.NotFound("Printer")
	}
	_, err = s.db.Exec("UPDATE tenant SET printer_id=NULL WHERE klien_id=? AND printer_id=?", kid, id)
	return err
}

// Tes mengantrekan tiket percobaan; berhasil bila tercetak, artinya seluruh jalur (server, stasiun, printer) sudah benar.
func (s *CetakService) Tes(kid, id int64) error {
	var nama string
	var lebarMM int
	if s.db.QueryRow("SELECT nama,lebar_mm FROM printer WHERE id=? AND klien_id=? AND aktif=1", id, kid).Scan(&nama, &lebarMM) != nil {
		return types.NotFound("Printer")
	}
	return utils.WithTx(s.db, func(tx *sql.Tx) error {
		return s.simpanJob(tx, kid, id, lebarMM, jenisTes, dto.TiketPayload{
			Judul: "TES CETAK", MejaNama: "-", Tenant: nama, Waktu: time.Now(),
			Items: []dto.TiketItem{{Qty: 1, Nama: "Printer terhubung"}},
		})
	})
}

// ---- token stasiun ----

func hashToken(t string) string {
	h := sha256.Sum256([]byte(t))
	return hex.EncodeToString(h[:])
}

// BuatToken membuat token baru (yang lama tidak berlaku lagi). Token hanya ditampilkan sekali; di database disimpan hash-nya.
func (s *CetakService) BuatToken(kid int64) (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	_, err := s.db.Exec(`INSERT INTO pengaturan (klien_id,kunci,nilai) VALUES (?,?,?) ON DUPLICATE KEY UPDATE nilai=VALUES(nilai)`,
		kid, kunciTokenStasiun, hashToken(token))
	return token, err
}

// VerifyToken mengembalikan klien pemilik token stasiun; token disimpan sebagai hash sehingga pencarian langsung memakai hash.
func (s *CetakService) VerifyToken(token string) (int64, bool) {
	if token == "" {
		return 0, false
	}
	var kid int64
	if s.db.QueryRow("SELECT klien_id FROM pengaturan WHERE kunci=? AND nilai=?", kunciTokenStasiun, hashToken(token)).Scan(&kid) != nil {
		return 0, false
	}
	return kid, true
}

// ---- dipakai stasiun cetak ----

func (s *CetakService) AmbilJob(kid int64) ([]dto.StasiunJob, error) {
	s.mu.Lock()
	s.terakhir[kid] = time.Now()
	s.mu.Unlock()

	rows, err := s.db.Query(`SELECT j.id,j.jenis,p.ip,p.port,p.lebar_mm,j.payload,j.data FROM cetak_job j JOIN printer p ON p.id=j.printer_id
		WHERE j.klien_id=? AND j.status='antri' AND j.dibuat_at>? ORDER BY j.id LIMIT 20`, kid, batasAktif())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []dto.StasiunJob{}
	for rows.Next() {
		var j dto.StasiunJob
		var payload string
		if err := rows.Scan(&j.ID, &j.Jenis, &j.IP, &j.Port, &j.LebarMM, &payload, &j.Data); err != nil {
			return nil, err
		}
		j.Payload = json.RawMessage(payload)
		out = append(out, j)
	}
	return out, rows.Err()
}

func (s *CetakService) Selesai(kid, id int64) error {
	if _, err := s.db.Exec("UPDATE cetak_job SET status='selesai', selesai_at=? WHERE id=? AND klien_id=? AND status='antri'", time.Now(), id, kid); err != nil {
		return err
	}
	s.hub.Publish(kid, "cetak")
	return nil
}

func (s *CetakService) Gagal(kid, id int64, galat string) error {
	_, err := s.db.Exec("UPDATE cetak_job SET percobaan=percobaan+1, galat=? WHERE id=? AND klien_id=? AND status='antri'", galat, id, kid)
	return err
}

// ---- tampilan owner dan kasir ----

func (s *CetakService) Riwayat(kid int64) ([]dto.CetakJob, error) {
	rows, err := s.db.Query(`SELECT j.id,p.nama,j.jenis,j.payload,j.status,j.percobaan,j.galat,j.dibuat_at,j.selesai_at
		FROM cetak_job j JOIN printer p ON p.id=j.printer_id WHERE j.klien_id=? ORDER BY j.id DESC LIMIT 50`, kid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []dto.CetakJob{}
	batas := batasAktif()
	for rows.Next() {
		var j dto.CetakJob
		var payload string
		if err := rows.Scan(&j.ID, &j.Printer, &j.Jenis, &payload, &j.Status, &j.Percobaan, &j.Galat, &j.DibuatAt, &j.SelesaiAt); err != nil {
			return nil, err
		}
		var p dto.TiketPayload
		if json.Unmarshal([]byte(payload), &p) == nil {
			j.Judul, j.Tenant, j.MejaNama = p.Judul, p.Tenant, p.MejaNama
		}
		if j.Status == "antri" && j.DibuatAt.Before(batas) {
			j.Status = "kedaluwarsa"
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

// Ulang membuat salinan tugas cetak sebagai tugas baru (yang lama tetap tercatat).
func (s *CetakService) Ulang(kid, id int64) error {
	res, err := s.db.Exec(`INSERT INTO cetak_job (klien_id,printer_id,jenis,payload,data,dibuat_at) SELECT klien_id,printer_id,jenis,payload,data,? FROM cetak_job WHERE id=? AND klien_id=?`, time.Now(), id, kid)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return types.NotFound("Tugas cetak")
	}
	s.hub.Publish(kid, "cetak")
	return nil
}

func (s *CetakService) Status(kid int64) (*dto.CetakStatus, error) {
	var st dto.CetakStatus
	if err := s.db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(dibuat_at<?),0) FROM cetak_job WHERE klien_id=? AND status='antri' AND dibuat_at>?`,
		kid, time.Now().Add(-macetSetelah), batasAktif()).Scan(&st.Antri, &st.Macet); err != nil {
		return nil, err
	}
	s.mu.Lock()
	st.StasiunAktif = time.Since(s.terakhir[kid]) < stasiunHidupSetelah
	s.mu.Unlock()
	return &st, nil
}
