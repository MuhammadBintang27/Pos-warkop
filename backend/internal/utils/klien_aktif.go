package utils

import "sync"

// nonaktif: klien yang aksesnya dimatikan superadmin. Disimpan di memori supaya setiap permintaan bisa dicek tanpa query;
// shortcut: hanya benar untuk satu instance server, pakai Redis atau cek DB bila nanti server lebih dari satu.
var nonaktif sync.Map

func SetKlienAktif(id int64, aktif bool) {
	if aktif {
		nonaktif.Delete(id)
		return
	}
	nonaktif.Store(id, struct{}{})
}

func KlienAktif(id int64) bool {
	_, mati := nonaktif.Load(id)
	return !mati
}
