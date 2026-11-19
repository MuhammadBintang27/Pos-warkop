package utils

import (
	"sync"
	"time"
)

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

// tokenSejak: user id -> waktu; token yang terbit sebelum waktu itu sudah dicabut. Dimuat dari users.token_sejak saat mulai.
var tokenSejak sync.Map

func CabutToken(uid int64, sejak time.Time) { tokenSejak.Store(uid, sejak) }

func TokenBerlaku(uid int64, terbit time.Time) bool {
	v, ok := tokenSejak.Load(uid)
	return !ok || !terbit.Before(v.(time.Time))
}
