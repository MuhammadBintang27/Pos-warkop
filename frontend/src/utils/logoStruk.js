// Logo hitam putih seperti hasil printer thermal; aturannya sama dengan rasterHitamPutih di backend (escpos_logo.go):
// piksel jadi tinta bila kanal tergelapnya < 50%, bagian transparan dianggap kertas putih.
const cache = new Map()

export function logoStruk(url) {
  if (!url) return Promise.resolve('')
  if (!cache.has(url)) cache.set(url, buat(url).catch(() => url)) // gagal dibaca: pakai logo asli
  return cache.get(url)
}

async function buat(url) {
  const img = new Image()
  img.src = url
  await img.decode()
  const c = document.createElement('canvas')
  c.width = img.naturalWidth
  c.height = img.naturalHeight
  const ctx = c.getContext('2d')
  ctx.drawImage(img, 0, 0)
  const d = ctx.getImageData(0, 0, c.width, c.height)
  const px = d.data
  for (let i = 0; i < px.length; i += 4) {
    const a = px[i + 3] / 255
    const v = Math.min(px[i], px[i + 1], px[i + 2]) * a + 255 * (1 - a) < 128 ? 0 : 255
    px[i] = px[i + 1] = px[i + 2] = v
    px[i + 3] = 255
  }
  ctx.putImageData(d, 0, 0)
  return c.toDataURL('image/png')
}
