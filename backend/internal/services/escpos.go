package services

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"warkop/internal/dto"
)

const (
	esc = 0x1b
	gs  = 0x1d
)

var karakterPerBaris = map[int]int{58: 32, 80: 48}

// Printer thermal hanya mengerti ASCII sederhana: aksen dibuang, karakter lain diganti "?".
var hilangkanAksen = strings.NewReplacer(
	"à", "a", "á", "a", "â", "a", "ã", "a", "ä", "a", "å", "a",
	"è", "e", "é", "e", "ê", "e", "ë", "e",
	"ì", "i", "í", "i", "î", "i", "ï", "i",
	"ò", "o", "ó", "o", "ô", "o", "õ", "o", "ö", "o",
	"ù", "u", "ú", "u", "û", "u", "ü", "u",
	"ñ", "n", "ç", "c",
	"À", "A", "Á", "A", "Â", "A", "Ã", "A", "Ä", "A", "Å", "A",
	"È", "E", "É", "E", "Ê", "E", "Ë", "E",
	"Ì", "I", "Í", "I", "Î", "I", "Ï", "I",
	"Ò", "O", "Ó", "O", "Ô", "O", "Õ", "O", "Ö", "O",
	"Ù", "U", "Ú", "U", "Û", "U", "Ü", "U",
	"Ñ", "N", "Ç", "C",
)

func ascii(s string) string {
	var b strings.Builder
	for _, r := range hilangkanAksen.Replace(s) {
		if r >= 0x20 && r <= 0x7e {
			b.WriteRune(r)
		} else {
			b.WriteByte('?')
		}
	}
	return b.String()
}

func bungkus(teks string, lebar int, indent string) []string {
	var baris []string
	sekarang := indent
	for _, kata := range strings.Fields(ascii(teks)) {
		kandidat := indent + kata
		if strings.TrimSpace(sekarang) != "" {
			kandidat = sekarang + " " + kata
		}
		if len(kandidat) > lebar && strings.TrimSpace(sekarang) != "" {
			baris = append(baris, sekarang)
			sekarang = indent + kata
		} else {
			sekarang = kandidat
		}
	}
	if strings.TrimSpace(sekarang) != "" {
		baris = append(baris, sekarang)
	}
	if len(baris) == 0 {
		return []string{indent}
	}
	return baris
}

// escposTiket mengubah satu tiket dapur menjadi perintah ESC/POS yang langsung bisa dikirim ke printer thermal.
func escposTiket(p dto.TiketPayload, lebarMM int) []byte {
	lebar := karakterPerBaris[lebarMM]
	if lebar == 0 {
		lebar = 32
	}
	var b bytes.Buffer
	perintah := func(x ...byte) { b.Write(x) }
	baris := func(t string) {
		b.WriteString(ascii(t))
		b.WriteByte('\n')
	}
	rata := func(n byte) { perintah(esc, 0x61, n) } // 0 kiri, 1 tengah
	tebal := func(aktif bool) {
		v := byte(0)
		if aktif {
			v = 1
		}
		perintah(esc, 0x45, v)
	}
	ukuran := func(n byte) { perintah(gs, 0x21, n) } // 0 normal, 0x11 dua kali

	batal := p.Judul == "BATAL"
	meja := p.MejaNama
	if meja == "" {
		meja = "BAWA PULANG"
	}

	perintah(esc, 0x40)
	rata(1)
	tebal(true)
	ukuran(0x11)
	for _, l := range bungkus(p.Judul, lebar/2, "") {
		baris(l)
	}
	for _, l := range bungkus(meja, lebar/2, "") {
		baris(l)
	}
	ukuran(0)
	if p.AtasNama != "" {
		baris(p.AtasNama)
	}
	baris(p.Tenant)
	tebal(false)
	baris(p.Waktu.In(time.Local).Format("02/01/06 15.04"))
	if p.Oleh != "" {
		label := "Dikirim"
		if batal {
			label = "Dibatalkan"
		}
		baris(label + " oleh: " + p.Oleh)
	}
	rata(0)
	baris(strings.Repeat("-", lebar))

	for _, i := range p.Items {
		tebal(true)
		for _, l := range bungkus(strconv.Itoa(i.Qty)+"x "+i.Nama, lebar, "") {
			baris(l)
		}
		tebal(false)
		if i.Opsi != "" {
			for _, l := range bungkus(i.Opsi, lebar, "  ") {
				baris(l)
			}
		}
		if i.Catatan != "" {
			tebal(true)
			for _, l := range bungkus("* "+i.Catatan, lebar, "  ") {
				baris(l)
			}
			tebal(false)
		}
	}

	baris(strings.Repeat("-", lebar))
	if batal {
		rata(1)
		tebal(true)
		baris("JANGAN DIBUAT / HENTIKAN")
		tebal(false)
	}
	perintah(esc, 0x64, 4)      // dorong kertas supaya teks terakhir lolos pisau
	perintah(gs, 0x56, 0x42, 0) // potong sebagian
	return b.Bytes()
}
