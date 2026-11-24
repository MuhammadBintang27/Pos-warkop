package services

import (
	"bytes"
	"encoding/base64"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"
)

// lebarTitik: lebar area cetak printer thermal 203 dpi dalam titik.
var lebarTitik = map[int]int{58: 384, 80: 576}

// logoRaster mengubah logo klien (data URL) menjadi perintah gambar ESC/POS hitam putih, paling lebar separuh kertas.
// Format yang tidak bisa dibaca Go (webp, svg) dilewati: struk tetap tercetak tanpa logo.
func logoRaster(dataURL string, lebarMM int) []byte {
	i := strings.Index(dataURL, ";base64,")
	if i < 0 {
		return nil
	}
	raw, err := base64.StdEncoding.DecodeString(dataURL[i+len(";base64,"):])
	if err != nil {
		return nil
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil
	}
	titik := lebarTitik[lebarMM]
	if titik == 0 {
		titik = 384
	}
	return rasterHitamPutih(img, titik/2, titik/3)
}

// rasterHitamPutih mengecilkan gambar (tidak pernah membesarkan) lalu memakai ambang 50%: thermal hanya bisa hitam atau putih.
// Bagian transparan dianggap kertas putih.
func rasterHitamPutih(img image.Image, maksW, maksH int) []byte {
	b := img.Bounds()
	sw, sh := b.Dx(), b.Dy()
	if sw == 0 || sh == 0 {
		return nil
	}
	skala := min(float64(maksW)/float64(sw), float64(maksH)/float64(sh), 1)
	w, h := max(int(float64(sw)*skala), 1), max(int(float64(sh)*skala), 1)
	bw := (w + 7) / 8
	data := make([]byte, bw*h)
	for y := range h {
		y0 := y * sh / h
		y1 := max((y+1)*sh/h, y0+1)
		for x := range w {
			x0 := x * sw / w
			x1 := max((x+1)*sw/w, x0+1)
			var jumlah, n uint64
			for sy := y0; sy < y1; sy++ {
				for sx := x0; sx < x1; sx++ {
					r, g, bl, a := img.At(b.Min.X+sx, b.Min.Y+sy).RGBA() // sudah dikali alpha
					jumlah += uint64((299*r+587*g+114*bl)/1000 + (0xffff - a))
					n++
				}
			}
			if jumlah/n < 0x8000 {
				data[y*bw+x/8] |= 0x80 >> (x % 8)
			}
		}
	}
	out := []byte{gs, 0x76, 0x30, 0, byte(bw), byte(bw >> 8), byte(h), byte(h >> 8)}
	return append(out, data...)
}
