package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"warkop/internal/dto"
	"warkop/internal/types"
	"warkop/internal/utils"
)

const msgGunakanBatalItem = "Item yang sudah dikirim ke dapur harus dibatalkan lewat pembatalan item (dengan alasan)"

type PesananService struct {
	db    *sql.DB
	hub   *EventHub
	cetak *CetakService
	push  *Pusher
}

type querier interface {
	Query(q string, args ...any) (*sql.Rows, error)
	QueryRow(q string, args ...any) *sql.Row
}

type pesananRow struct {
	ID       int64
	MejaID   sql.NullInt64
	AtasNama string
	Diskon   int
}

type opsiInfo struct {
	nama   string
	plus   int
	grupID int64
}

type grupInfo struct {
	nama                string
	wajib, multi, ganti bool
	dipilih             int
}

type pilihanOpsi struct {
	nilai int
	ganti bool
}

type itemBaru struct {
	menuID, tenantID    int64
	nama, opsi, catatan string
	harga, qty, modal   int
	komisi              int          // salinan menu.komisi per porsi
	bahan               []bahanPakai // takaran per porsi, disalin ke baris baru
}

func NewPesananService(db *sql.DB, hub *EventHub, cetak *CetakService, push *Pusher) *PesananService {
	return &PesananService{db: db, hub: hub, cetak: cetak, push: push}
}

func loadPesanan(q querier, kid, id int64) (*dto.PesananResponse, error) {
	p := &dto.PesananResponse{Items: []dto.ItemResponse{}, Pembayaran: []dto.PembayaranResponse{}, Pembatalan: []dto.PembatalanResponse{}}
	err := q.QueryRow(`SELECT p.id,p.meja_id,COALESCE(m.nama,''),p.atas_nama,p.status,p.diskon,p.dibuat_at,p.lunas_at
		FROM pesanan p LEFT JOIN meja m ON m.id=p.meja_id WHERE p.id=? AND p.klien_id=?`, id, kid).
		Scan(&p.ID, &p.MejaID, &p.MejaNama, &p.AtasNama, &p.Status, &p.Diskon, &p.DibuatAt, &p.LunasAt)
	if err == sql.ErrNoRows {
		return nil, types.NotFound("Pesanan")
	}
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(`SELECT i.id,i.menu_id,i.nama,i.harga,i.qty,i.opsi,i.catatan,i.dikirim,COALESCE(i.tenant_id,0),COALESCE(t.nama,''),i.diproses_at,i.siap_at
		FROM pesanan_item i LEFT JOIN tenant t ON t.id=i.tenant_id WHERE i.klien_id=? AND i.pesanan_id=? ORDER BY i.id`, kid, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var i dto.ItemResponse
		if err := rows.Scan(&i.ID, &i.MenuID, &i.Nama, &i.Harga, &i.Qty, &i.Opsi, &i.Catatan, &i.Dikirim, &i.TenantID, &i.TenantNama, &i.DiprosesAt, &i.SiapAt); err != nil {
			rows.Close()
			return nil, err
		}
		p.Items = append(p.Items, i)
		p.Subtotal += i.Harga * i.Qty
	}
	rows.Close()
	p.Total = p.Subtotal - p.Diskon

	rows, err = q.Query(`SELECT b.metode,b.jumlah,b.diterima,b.kembalian,COALESCE(c.nama,'') FROM pembayaran b
		LEFT JOIN pelanggan c ON c.id=b.pelanggan_id WHERE b.klien_id=? AND b.pesanan_id=? ORDER BY b.id`, kid, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var b dto.PembayaranResponse
		if err := rows.Scan(&b.Metode, &b.Jumlah, &b.Diterima, &b.Kembalian, &b.Pelanggan); err != nil {
			rows.Close()
			return nil, err
		}
		p.Pembayaran = append(p.Pembayaran, b)
	}
	rows.Close()

	rows, err = q.Query(`SELECT b.nama,b.opsi,b.qty,b.harga,b.alasan,b.sudah_dikirim,b.sudah_dibuat,u.nama,b.at
		FROM pembatalan b JOIN users u ON u.id=b.oleh WHERE b.klien_id=? AND b.pesanan_id=? ORDER BY b.id`, kid, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var b dto.PembatalanResponse
		if err := rows.Scan(&b.Nama, &b.Opsi, &b.Qty, &b.Harga, &b.Alasan, &b.SudahDikirim, &b.SudahDibuat, &b.Oleh, &b.At); err != nil {
			rows.Close()
			return nil, err
		}
		p.Pembatalan = append(p.Pembatalan, b)
	}
	rows.Close()

	if p.Status == types.StatusLunas {
		err = q.QueryRow(`SELECT u.nama FROM pembayaran b JOIN users u ON u.id=b.diterima_oleh
			WHERE b.klien_id=? AND b.pesanan_id=? LIMIT 1`, kid, id).Scan(&p.Kasir)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
	}
	return p, nil
}

// lockOpen mengunci baris pesanan dan memastikan masih open, supaya dua perangkat tidak saling menimpa.
func lockOpen(tx *sql.Tx, kid, id int64) (pesananRow, error) {
	var p pesananRow
	var status types.PesananStatus
	err := tx.QueryRow("SELECT id,meja_id,atas_nama,status,diskon FROM pesanan WHERE id=? AND klien_id=? FOR UPDATE", id, kid).
		Scan(&p.ID, &p.MejaID, &p.AtasNama, &status, &p.Diskon)
	if err == sql.ErrNoRows {
		return p, types.NotFound("Pesanan")
	}
	if err != nil {
		return p, err
	}
	if status != types.StatusOpen {
		return p, types.BadRequest("Pesanan sudah " + string(status))
	}
	return p, nil
}

func subtotalOf(tx *sql.Tx, kid, id int64) (n int, err error) {
	err = tx.QueryRow("SELECT COALESCE(SUM(harga*qty),0) FROM pesanan_item WHERE klien_id=? AND pesanan_id=?", kid, id).Scan(&n)
	return
}

// adjustStok: delta negatif mengurangi stok. Hanya berlaku untuk menu ber-track_stok.
func adjustStok(tx *sql.Tx, kid, menuID int64, delta int) error {
	var track bool
	var stok int
	var nama string
	if err := tx.QueryRow("SELECT track_stok,stok,nama FROM menu WHERE id=? AND klien_id=? FOR UPDATE", menuID, kid).Scan(&track, &stok, &nama); err != nil {
		return err
	}
	if !track || delta == 0 {
		return nil
	}
	if stok+delta < 0 {
		return types.BadRequest(fmt.Sprintf("Stok %s tidak cukup (sisa %d)", nama, stok))
	}
	_, err := tx.Exec("UPDATE menu SET stok=stok+? WHERE id=? AND klien_id=?", delta, menuID, kid)
	return err
}

func (s *PesananService) result(kid, id int64, events ...string) (*dto.PesananResponse, error) {
	s.hub.Publish(kid, events...)
	return loadPesanan(s.db, kid, id)
}

func (s *PesananService) Create(actor *utils.JWTPayload, req dto.CreatePesananRequest) (*dto.PesananResponse, error) {
	kid := actor.KlienID
	var id int64
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		if req.MejaID != nil {
			var aktif bool
			if tx.QueryRow("SELECT aktif FROM meja WHERE id=? AND klien_id=? FOR UPDATE", *req.MejaID, kid).Scan(&aktif) != nil || !aktif {
				return types.NotFound("Meja")
			}
			// meja sudah punya pesanan: pakai itu, mencegah dobel saat dua orang menekan meja bersamaan
			err := tx.QueryRow("SELECT id FROM pesanan WHERE klien_id=? AND meja_id=? AND status='open' ORDER BY id LIMIT 1", kid, *req.MejaID).Scan(&id)
			if err == nil {
				return nil
			}
			if err != sql.ErrNoRows {
				return err
			}
		}
		res, err := tx.Exec("INSERT INTO pesanan (klien_id,meja_id,atas_nama,dibuat_oleh,dibuat_at) VALUES (?,?,?,?,?)",
			kid, req.MejaID, strings.TrimSpace(req.AtasNama), actor.UserID, time.Now())
		if err != nil {
			return err
		}
		id, err = res.LastInsertId()
		return err
	})
	if err != nil {
		return nil, err
	}
	return s.result(kid, id, "meja")
}

func (s *PesananService) Get(kid, id int64) (*dto.PesananResponse, error) {
	return loadPesanan(s.db, kid, id)
}

func (s *PesananService) AddItem(kid, id int64, req dto.AddItemRequest) (*dto.PesananResponse, error) {
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		if _, err := lockOpen(tx, kid, id); err != nil {
			return err
		}
		var nama string
		var harga, modal, komisi int
		var tersedia, aktif bool
		var tenantID int64
		var tenantBuka bool
		var tenantNama string
		if tx.QueryRow(`SELECT m.nama,m.harga,m.harga_modal,m.komisi,m.tersedia,m.aktif,COALESCE(m.tenant_id,0),COALESCE(t.buka,1),COALESCE(t.nama,'')
			FROM menu m LEFT JOIN tenant t ON t.id=m.tenant_id WHERE m.id=? AND m.klien_id=?`, req.MenuID, kid).
			Scan(&nama, &harga, &modal, &komisi, &tersedia, &aktif, &tenantID, &tenantBuka, &tenantNama) != nil || !aktif {
			return types.NotFound("Menu")
		}
		if !tenantBuka {
			return types.BadRequest(tenantNama + " sedang tutup")
		}
		if !tersedia {
			return types.BadRequest(nama + " sedang habis")
		}
		opsiNama, hargaJadi, err := resolveOpsi(tx, kid, req.MenuID, harga, req.OpsiIDs)
		if err != nil {
			return err
		}
		if err := adjustStok(tx, kid, req.MenuID, -req.Qty); err != nil {
			return err
		}
		bahan, err := resepUntuk(tx, kid, req.MenuID, req.OpsiIDs)
		if err != nil {
			return err
		}
		if len(bahan) > 0 {
			modal = 0 // menu beresep: HPP dari bahan, bukan harga modal
		}
		itemID, err := sisipItem(tx, kid, id, itemBaru{
			menuID: req.MenuID, tenantID: tenantID, nama: nama, harga: hargaJadi, qty: req.Qty, modal: modal, komisi: komisi,
			opsi: opsiNama, catatan: strings.TrimSpace(req.Catatan), bahan: bahan,
		})
		if err != nil {
			return err
		}
		return pakaiBahan(tx, kid, itemID, req.Qty)
	})
	if err != nil {
		return nil, err
	}
	return s.result(kid, id, "pesanan", "meja", "menu")
}

// sisipItem menambah item ke bill: digabung ke baris belum-dikirim yang sama persis, atau dibuat baris baru.
// Baris yang sudah dikirim ke dapur tidak pernah digabung, supaya tambahannya tetap tercetak.
func sisipItem(tx *sql.Tx, kid, pesananID int64, it itemBaru) (int64, error) {
	var itemID int64
	var qtyAda int
	err := tx.QueryRow(`SELECT id,qty FROM pesanan_item
		WHERE klien_id=? AND pesanan_id=? AND menu_id=? AND harga=? AND komisi=? AND opsi=? AND catatan=? AND dikirim=0 LIMIT 1 FOR UPDATE`,
		kid, pesananID, it.menuID, it.harga, it.komisi, it.opsi, it.catatan).Scan(&itemID, &qtyAda)
	switch {
	case err == nil:
		if qtyAda+it.qty > 99 {
			return 0, types.BadRequest("Maksimal 99 porsi per baris pesanan")
		}
		_, err = tx.Exec("UPDATE pesanan_item SET qty=qty+? WHERE id=? AND klien_id=?", it.qty, itemID, kid)
	case err == sql.ErrNoRows:
		var res sql.Result
		res, err = tx.Exec("INSERT INTO pesanan_item (klien_id,pesanan_id,menu_id,nama,harga,qty,modal,komisi,opsi,catatan,tenant_id) VALUES (?,?,?,?,?,?,?,?,?,?,?)",
			kid, pesananID, it.menuID, it.nama, it.harga, it.qty, it.modal, it.komisi, it.opsi, it.catatan, it.tenantID)
		if err == nil {
			if itemID, err = res.LastInsertId(); err == nil {
				err = simpanSnapshot(tx, kid, itemID, it.bahan)
			}
		}
	}
	return itemID, err
}

// hargaAkhir: opsi dari grup penentu harga menggantikan harga dasar; opsi lain ditambahkan ke harga dasar.
func hargaAkhir(dasar int, pilihan []pilihanOpsi) int {
	tambah := 0
	for _, p := range pilihan {
		if p.ganti {
			dasar = p.nilai
		} else {
			tambah += p.nilai
		}
	}
	return dasar + tambah
}

// resolveOpsi memvalidasi pilihan varian terhadap definisi menu, lalu menghitung nama dan harga jadi per satuan.
func resolveOpsi(tx *sql.Tx, kid, menuID int64, dasar int, ids []int64) (string, int, error) {
	groups := map[int64]*grupInfo{}
	rows, err := tx.Query("SELECT id,nama,wajib,multi,ganti_harga FROM varian_grup WHERE klien_id=? AND menu_id=?", kid, menuID)
	if err != nil {
		return "", 0, err
	}
	for rows.Next() {
		var gid int64
		g := &grupInfo{}
		if err := rows.Scan(&gid, &g.nama, &g.wajib, &g.multi, &g.ganti); err != nil {
			rows.Close()
			return "", 0, err
		}
		groups[gid] = g
	}
	rows.Close()

	opts := map[int64]opsiInfo{}
	rows, err = tx.Query(`SELECT op.id,op.nama,op.tambahan,op.grup_id FROM varian_opsi op
		JOIN varian_grup g ON g.id=op.grup_id WHERE g.klien_id=? AND g.menu_id=?`, kid, menuID)
	if err != nil {
		return "", 0, err
	}
	for rows.Next() {
		var oid int64
		var o opsiInfo
		if err := rows.Scan(&oid, &o.nama, &o.plus, &o.grupID); err != nil {
			rows.Close()
			return "", 0, err
		}
		opts[oid] = o
	}
	rows.Close()

	var names []string
	var pilihan []pilihanOpsi
	seen := map[int64]bool{}
	for _, oid := range ids {
		o, found := opts[oid]
		if !found || seen[oid] {
			return "", 0, types.BadRequest("Pilihan varian tidak valid")
		}
		seen[oid] = true
		g := groups[o.grupID]
		if g.dipilih++; g.dipilih > 1 && !g.multi {
			return "", 0, types.BadRequest("Pilih satu saja untuk " + g.nama)
		}
		names, pilihan = append(names, o.nama), append(pilihan, pilihanOpsi{o.plus, g.ganti})
	}
	for _, g := range groups {
		if g.wajib && g.dipilih == 0 {
			return "", 0, types.BadRequest("Pilih " + g.nama + " dulu")
		}
	}
	return strings.Join(names, ", "), hargaAkhir(dasar, pilihan), nil
}

func (s *PesananService) UpdateItem(kid, id, itemID int64, req dto.UpdateItemRequest) (*dto.PesananResponse, error) {
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		if _, err := lockOpen(tx, kid, id); err != nil {
			return err
		}
		var it itemBaru
		var qty int
		var dikirim bool
		if tx.QueryRow(`SELECT menu_id,nama,harga,qty,opsi,catatan,dikirim,COALESCE(tenant_id,0)
			FROM pesanan_item WHERE id=? AND pesanan_id=? AND klien_id=? FOR UPDATE`, itemID, id, kid).
			Scan(&it.menuID, &it.nama, &it.harga, &qty, &it.opsi, &it.catatan, &dikirim, &it.tenantID) != nil {
			return types.NotFound("Item")
		}
		catatan := strings.TrimSpace(req.Catatan)
		if dikirim {
			switch {
			case req.Qty < qty:
				return types.BadRequest(msgGunakanBatalItem)
			case catatan != it.catatan:
				return types.BadRequest("Catatan item yang sudah dikirim tidak bisa diubah. Pesan ulang dengan catatan baru")
			case req.Qty == qty:
				return nil
			}
			// tambahan atas item yang sudah dikirim menjadi baris baru yang belum dikirim, supaya dapur menerima tiket tambahan
			it.qty = req.Qty - qty
			if err := adjustStok(tx, kid, it.menuID, -it.qty); err != nil {
				return err
			}
			var err error
			if it.bahan, err = snapshotItem(tx, kid, itemID); err != nil {
				return err
			}
			baruID, err := sisipItem(tx, kid, id, it)
			if err != nil {
				return err
			}
			return pakaiBahan(tx, kid, baruID, it.qty)
		}
		if err := adjustStok(tx, kid, it.menuID, qty-req.Qty); err != nil {
			return err
		}
		if err := pakaiBahan(tx, kid, itemID, req.Qty-qty); err != nil {
			return err
		}
		_, err := tx.Exec("UPDATE pesanan_item SET qty=?, catatan=? WHERE id=? AND klien_id=?", req.Qty, catatan, itemID, kid)
		return err
	})
	if err != nil {
		return nil, err
	}
	return s.result(kid, id, "pesanan", "meja", "menu")
}

func (s *PesananService) DeleteItem(kid, id, itemID int64) (*dto.PesananResponse, error) {
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		p, err := lockOpen(tx, kid, id)
		if err != nil {
			return err
		}
		var menuID int64
		var qty int
		var dikirim bool
		if tx.QueryRow("SELECT menu_id,qty,dikirim FROM pesanan_item WHERE id=? AND pesanan_id=? AND klien_id=? FOR UPDATE", itemID, id, kid).
			Scan(&menuID, &qty, &dikirim) != nil {
			return types.NotFound("Item")
		}
		if dikirim {
			return types.BadRequest(msgGunakanBatalItem)
		}
		if err := adjustStok(tx, kid, menuID, qty); err != nil {
			return err
		}
		if err := pakaiBahan(tx, kid, itemID, -qty); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM pesanan_item WHERE id=? AND klien_id=?", itemID, kid); err != nil {
			return err
		}
		return clampDiskon(tx, kid, id, p.Diskon)
	})
	if err != nil {
		return nil, err
	}
	return s.result(kid, id, "pesanan", "meja", "menu")
}

// clampDiskon menjaga diskon tidak melebihi subtotal setelah item berkurang.
func clampDiskon(tx *sql.Tx, kid, id int64, diskon int) error {
	sub, err := subtotalOf(tx, kid, id)
	if err != nil || diskon <= sub {
		return err
	}
	_, err = tx.Exec("UPDATE pesanan SET diskon=? WHERE id=? AND klien_id=?", sub, id, kid)
	return err
}

// Kirim menandai item baru sebagai terkirim ke dapur dan mengembalikannya untuk dicetak.
// cetak=false untuk aksi dari antrean offline HP kasir: tiketnya sudah dicetak HP langsung ke printer.
func (s *PesananService) Kirim(actor *utils.JWTPayload, id int64, cetak bool) (*dto.KirimResponse, error) {
	kid := actor.KlienID
	out := &dto.KirimResponse{Items: []dto.ItemResponse{}, Oleh: actor.Nama}
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		if _, err := lockOpen(tx, kid, id); err != nil {
			return err
		}
		p, err := loadPesanan(tx, kid, id)
		if err != nil {
			return err
		}
		out.MejaNama, out.AtasNama = p.MejaNama, p.AtasNama
		sudah := map[int64]bool{}
		for _, i := range p.Items {
			if i.Dikirim {
				sudah[i.TenantID] = true
			} else {
				out.Items = append(out.Items, i)
			}
		}
		out.Tambahan = map[int64]bool{}
		for _, i := range out.Items {
			out.Tambahan[i.TenantID] = sudah[i.TenantID]
		}
		if _, err = tx.Exec("UPDATE pesanan_item SET dikirim=1, dikirim_at=?, dikirim_oleh=? WHERE klien_id=? AND pesanan_id=? AND dikirim=0",
			time.Now(), actor.Nama, kid, id); err != nil {
			return err
		}
		out.Antrean, err = s.antrekan(tx, kid, jenisDapur, out, cetak)
		return err
	})
	if err != nil {
		return nil, err
	}
	s.hub.Publish(kid, "pesanan")
	if len(out.Antrean) > 0 {
		s.hub.Publish(kid, "cetak")
	}
	for tid, items := range itemPerTenant(out.Items) {
		judul := "Pesanan baru"
		if out.Tambahan[tid] {
			judul = "Pesanan tambahan"
		}
		s.push.KeTenant(kid, tid, pesanTiket(judul, out.MejaNama, out.AtasNama, items))
	}
	return out, nil
}

// kabariBatal memberi tahu tenant bahwa item terkirimnya dibatalkan: stasiun cetak dibangunkan bila ada tiket BATAL, HP tenant dapat notifikasi.
func kabariBatal(hub *EventHub, push *Pusher, kid int64, meja, atasNama string, items []dto.ItemResponse, adaAntrean bool) {
	if adaAntrean {
		hub.Publish(kid, "cetak")
	}
	for tid, it := range itemPerTenant(items) {
		push.KeTenant(kid, tid, pesanTiket("Dibatalkan", meja, atasNama, it))
	}
}

func itemPerTenant(items []dto.ItemResponse) map[int64][]dto.ItemResponse {
	out := map[int64][]dto.ItemResponse{}
	for _, it := range items {
		if it.TenantID != 0 {
			out[it.TenantID] = append(out[it.TenantID], it)
		}
	}
	return out
}

// pesanTiket: isi notifikasi tenant, ringkasan yang sama dengan tiket dapur.
func pesanTiket(judul, meja, atasNama string, items []dto.ItemResponse) PesanPush {
	if meja == "" {
		meja = "Tanpa meja"
	}
	judul += " · " + meja
	if atasNama != "" {
		judul += " (" + atasNama + ")"
	}
	isi := make([]string, len(items))
	for i, it := range items {
		isi[i] = fmt.Sprintf("%d× %s", it.Qty, it.Nama)
	}
	return PesanPush{Judul: judul, Isi: strings.Join(isi, ", ")}
}

func (s *PesananService) SetDiskon(kid, id int64, req dto.DiskonRequest) (*dto.PesananResponse, error) {
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		if _, err := lockOpen(tx, kid, id); err != nil {
			return err
		}
		sub, err := subtotalOf(tx, kid, id)
		if err != nil {
			return err
		}
		if req.Diskon > sub {
			return types.BadRequest("Diskon tidak boleh melebihi subtotal")
		}
		_, err = tx.Exec("UPDATE pesanan SET diskon=? WHERE id=? AND klien_id=?", req.Diskon, id, kid)
		return err
	})
	if err != nil {
		return nil, err
	}
	return s.result(kid, id, "pesanan", "meja")
}

func catatBatal(tx *sql.Tx, kid, pesananID int64, it dto.ItemResponse, qty int, alasan string, sudahDibuat bool, oleh int64) error {
	_, err := tx.Exec(`INSERT INTO pembatalan (klien_id,pesanan_id,nama,opsi,harga,qty,alasan,sudah_dikirim,sudah_dibuat,tenant_id,oleh,at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, kid, pesananID, it.Nama, it.Opsi, it.Harga, qty, alasan, it.Dikirim, sudahDibuat, it.TenantID, oleh, time.Now())
	return err
}

// Batal membatalkan seluruh pesanan. Bila sudah ada item terkirim, alasan wajib dan tiket BATAL dikembalikan untuk dapur.
func (s *PesananService) Batal(actor *utils.JWTPayload, id int64, req dto.BatalRequest, cetak bool) (*dto.BatalResponse, error) {
	kid := actor.KlienID
	var tiket *dto.KirimResponse
	alasan := strings.TrimSpace(req.Alasan)
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		if _, err := lockOpen(tx, kid, id); err != nil {
			return err
		}
		p, err := loadPesanan(tx, kid, id)
		if err != nil {
			return err
		}
		var terkirim []dto.ItemResponse
		for _, i := range p.Items {
			if i.Dikirim {
				terkirim = append(terkirim, i)
			}
		}
		if len(terkirim) > 0 && alasan == "" {
			return types.BadRequest("Alasan pembatalan wajib diisi karena sudah ada item yang dikirim ke dapur")
		}
		if alasan == "" {
			alasan = "Dibatalkan sebelum dikirim"
		}
		for _, i := range p.Items {
			sudahDibuat := i.Dikirim && req.SudahDibuat
			if !sudahDibuat {
				if err := adjustStok(tx, kid, i.MenuID, i.Qty); err != nil {
					return err
				}
				if err := pakaiBahan(tx, kid, i.ID, -i.Qty); err != nil {
					return err
				}
			}
			if err := catatBatal(tx, kid, id, i, i.Qty, alasan, sudahDibuat, actor.UserID); err != nil {
				return err
			}
		}
		if len(terkirim) > 0 {
			tiket = &dto.KirimResponse{MejaNama: p.MejaNama, AtasNama: p.AtasNama, Oleh: actor.Nama, Items: terkirim}
			if tiket.Antrean, err = s.antrekan(tx, kid, jenisBatal, tiket, cetak); err != nil {
				return err
			}
		}
		_, err = tx.Exec("UPDATE pesanan SET status='batal' WHERE id=? AND klien_id=?", id, kid)
		return err
	})
	if err != nil {
		return nil, err
	}
	s.hub.Publish(kid, "pesanan", "meja", "menu")
	if tiket != nil {
		kabariBatal(s.hub, s.push, kid, tiket.MejaNama, tiket.AtasNama, tiket.Items, len(tiket.Antrean) > 0)
	}
	p, err := loadPesanan(s.db, kid, id)
	if err != nil {
		return nil, err
	}
	return &dto.BatalResponse{Pesanan: p, Tiket: tiket}, nil
}

// BatalItem membatalkan sebagian atau seluruh jumlah satu item. Stok kembali kecuali makanannya sudah dibuat.
func (s *PesananService) BatalItem(actor *utils.JWTPayload, id, itemID int64, req dto.BatalItemRequest, cetak bool) (*dto.BatalResponse, error) {
	kid := actor.KlienID
	alasan := strings.TrimSpace(req.Alasan)
	if alasan == "" {
		return nil, types.BadRequest("Alasan pembatalan wajib diisi")
	}
	var tiket *dto.KirimResponse
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		p, err := lockOpen(tx, kid, id)
		if err != nil {
			return err
		}
		var it dto.ItemResponse
		if tx.QueryRow(`SELECT menu_id,nama,harga,qty,opsi,dikirim,COALESCE(tenant_id,0),
			COALESCE((SELECT nama FROM tenant WHERE id=pesanan_item.tenant_id),'')
			FROM pesanan_item WHERE id=? AND pesanan_id=? AND klien_id=? FOR UPDATE`, itemID, id, kid).
			Scan(&it.MenuID, &it.Nama, &it.Harga, &it.Qty, &it.Opsi, &it.Dikirim, &it.TenantID, &it.TenantNama) != nil {
			return types.NotFound("Item")
		}
		if req.Qty > it.Qty {
			return types.BadRequest("Jumlah pembatalan melebihi jumlah pesanan")
		}
		sudahDibuat := it.Dikirim && req.SudahDibuat
		if !sudahDibuat {
			if err := adjustStok(tx, kid, it.MenuID, req.Qty); err != nil {
				return err
			}
			if err := pakaiBahan(tx, kid, itemID, -req.Qty); err != nil {
				return err
			}
		}
		if req.Qty == it.Qty {
			_, err = tx.Exec("DELETE FROM pesanan_item WHERE id=? AND klien_id=?", itemID, kid)
		} else {
			_, err = tx.Exec("UPDATE pesanan_item SET qty=qty-? WHERE id=? AND klien_id=?", req.Qty, itemID, kid)
		}
		if err != nil {
			return err
		}
		if err := catatBatal(tx, kid, id, it, req.Qty, alasan, sudahDibuat, actor.UserID); err != nil {
			return err
		}
		if err := clampDiskon(tx, kid, id, p.Diskon); err != nil {
			return err
		}
		if it.Dikirim {
			pl, err := loadPesanan(tx, kid, id)
			if err != nil {
				return err
			}
			tiket = &dto.KirimResponse{MejaNama: pl.MejaNama, AtasNama: pl.AtasNama, Oleh: actor.Nama,
				Items: []dto.ItemResponse{{Nama: it.Nama, Opsi: it.Opsi, Qty: req.Qty, TenantID: it.TenantID, TenantNama: it.TenantNama}}}
			if tiket.Antrean, err = s.antrekan(tx, kid, jenisBatal, tiket, cetak); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.hub.Publish(kid, "pesanan", "meja", "menu")
	if tiket != nil {
		kabariBatal(s.hub, s.push, kid, tiket.MejaNama, tiket.AtasNama, tiket.Items, len(tiket.Antrean) > 0)
	}
	pesanan, err := loadPesanan(s.db, kid, id)
	if err != nil {
		return nil, err
	}
	return &dto.BatalResponse{Pesanan: pesanan, Tiket: tiket}, nil
}

func (s *PesananService) Pindah(kid, id int64, req dto.PindahRequest) (*dto.PesananResponse, error) {
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		if _, err := lockOpen(tx, kid, id); err != nil {
			return err
		}
		var aktif bool
		if tx.QueryRow("SELECT aktif FROM meja WHERE id=? AND klien_id=? FOR UPDATE", req.MejaID, kid).Scan(&aktif) != nil || !aktif {
			return types.NotFound("Meja tujuan")
		}
		var n int
		if err := tx.QueryRow(`SELECT COUNT(*) FROM pesanan p WHERE p.klien_id=? AND p.meja_id=? AND p.status='open' AND p.id<>?
			AND EXISTS (SELECT 1 FROM pesanan_item WHERE pesanan_id=p.id)`, kid, req.MejaID, id).Scan(&n); err != nil {
			return err
		}
		if n > 0 {
			return types.BadRequest("Meja tujuan sudah terisi, gunakan gabung")
		}
		_, err := tx.Exec("UPDATE pesanan SET meja_id=? WHERE id=? AND klien_id=?", req.MejaID, id, kid)
		return err
	})
	if err != nil {
		return nil, err
	}
	return s.result(kid, id, "pesanan", "meja")
}

// Gabung memindahkan semua item pesanan id ke pesanan tujuan, lalu menandai id sebagai 'gabung'.
func (s *PesananService) Gabung(kid, id int64, req dto.GabungRequest) (*dto.PesananResponse, error) {
	if req.KePesananID == id {
		return nil, types.BadRequest("Tidak bisa menggabung ke pesanan yang sama")
	}
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		// kunci berurutan menurut id agar dua gabung bersamaan tidak deadlock
		first, second := id, req.KePesananID
		if second < first {
			first, second = second, first
		}
		if _, err := lockOpen(tx, kid, first); err != nil {
			return err
		}
		if _, err := lockOpen(tx, kid, second); err != nil {
			return err
		}
		var diskon int
		if err := tx.QueryRow("SELECT diskon FROM pesanan WHERE id=? AND klien_id=?", id, kid).Scan(&diskon); err != nil {
			return err
		}
		if _, err := tx.Exec("UPDATE pesanan_item SET pesanan_id=? WHERE pesanan_id=? AND klien_id=?", req.KePesananID, id, kid); err != nil {
			return err
		}
		if _, err := tx.Exec("UPDATE pesanan SET diskon=diskon+? WHERE id=? AND klien_id=?", diskon, req.KePesananID, kid); err != nil {
			return err
		}
		_, err := tx.Exec("UPDATE pesanan SET status='gabung', diskon=0 WHERE id=? AND klien_id=?", id, kid)
		return err
	})
	if err != nil {
		return nil, err
	}
	return s.result(kid, req.KePesananID, "pesanan", "meja")
}

// Pisah (split bill) memindahkan item terpilih ke pesanan baru di meja yang sama.
func (s *PesananService) Pisah(actor *utils.JWTPayload, id int64, req dto.PisahRequest) (*dto.PesananResponse, error) {
	kid := actor.KlienID
	var newID int64
	err := utils.WithTx(s.db, func(tx *sql.Tx) error {
		p, err := lockOpen(tx, kid, id)
		if err != nil {
			return err
		}
		res, err := tx.Exec("INSERT INTO pesanan (klien_id,meja_id,atas_nama,dibuat_oleh,dibuat_at) VALUES (?,?,?,?,?)",
			kid, p.MejaID, p.AtasNama, actor.UserID, time.Now())
		if err != nil {
			return err
		}
		if newID, err = res.LastInsertId(); err != nil {
			return err
		}
		seen := map[int64]bool{}
		for _, x := range req.Items {
			if seen[x.ItemID] {
				return types.BadRequest("Item dobel")
			}
			seen[x.ItemID] = true
			var it dto.ItemResponse
			if tx.QueryRow("SELECT menu_id,nama,harga,qty,opsi,catatan,dikirim,COALESCE(tenant_id,0) FROM pesanan_item WHERE id=? AND pesanan_id=? AND klien_id=? FOR UPDATE", x.ItemID, id, kid).
				Scan(&it.MenuID, &it.Nama, &it.Harga, &it.Qty, &it.Opsi, &it.Catatan, &it.Dikirim, &it.TenantID) != nil {
				return types.NotFound("Item")
			}
			if x.Qty > it.Qty {
				return types.BadRequest("Jumlah pisah melebihi pesanan untuk " + it.Nama)
			}
			if x.Qty == it.Qty {
				_, err = tx.Exec("UPDATE pesanan_item SET pesanan_id=? WHERE id=? AND klien_id=?", newID, x.ItemID, kid)
			} else if _, err = tx.Exec("UPDATE pesanan_item SET qty=qty-? WHERE id=? AND klien_id=?", x.Qty, x.ItemID, kid); err == nil {
				var res sql.Result
				// salin dari baris asal supaya status kirim/siap tenant ikut terbawa
				res, err = tx.Exec(`INSERT INTO pesanan_item (klien_id,pesanan_id,menu_id,nama,harga,qty,modal,komisi,opsi,catatan,dikirim,tenant_id,diproses_at,siap_at,dikirim_at,dikirim_oleh)
					SELECT klien_id,?,menu_id,nama,harga,?,modal,komisi,opsi,catatan,dikirim,tenant_id,diproses_at,siap_at,dikirim_at,dikirim_oleh FROM pesanan_item WHERE id=? AND klien_id=?`,
					newID, x.Qty, x.ItemID, kid)
				if err == nil {
					var baruID int64
					if baruID, err = res.LastInsertId(); err == nil {
						_, err = tx.Exec("INSERT INTO item_bahan (klien_id,item_id,bahan_id,jumlah,biaya) SELECT klien_id,?,bahan_id,jumlah,biaya FROM item_bahan WHERE item_id=? AND klien_id=?", baruID, x.ItemID, kid)
					}
				}
			}
			if err != nil {
				return err
			}
		}
		sub, err := subtotalOf(tx, kid, id)
		if err != nil {
			return err
		}
		if sub == 0 {
			return types.BadRequest("Sisakan minimal satu item di pesanan asal")
		}
		return clampDiskon(tx, kid, id, p.Diskon)
	})
	if err != nil {
		return nil, err
	}
	return s.result(kid, newID, "pesanan", "meja")
}

// antrekan melewatkan tiket printer bila cetak=false (aksi dari antrean offline yang tiketnya sudah dicetak HP kasir).
func (s *PesananService) antrekan(tx *sql.Tx, kid int64, jenis string, t *dto.KirimResponse, cetak bool) (map[int64]bool, error) {
	if !cetak {
		return map[int64]bool{}, nil
	}
	return s.cetak.Antrekan(tx, kid, jenis, t)
}
