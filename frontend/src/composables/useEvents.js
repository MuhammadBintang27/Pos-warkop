const NAMES = ['pesanan', 'meja', 'menu', 'pengaturan', 'shift', 'cetak', 'dapur']
const handlers = Object.fromEntries(NAMES.map((n) => [n, new Set()]))
let source = null
let timer = null

// semua halaman yang terbuka memuat ulang datanya: event selama sambungan putus terlewat
const muatUlangSemua = () => NAMES.forEach((n) => handlers[n].forEach((fn) => fn()))
// antrean offline baru terkirim: data di server sudah berubah, sementara halaman masih menampilkan versi HP
window.addEventListener('offline:tersinkron', muatUlangSemua)

export function connectEvents(token) {
  disconnectEvents()
  let pertama = true
  const buka = () => {
    const es = (source = new EventSource(`/api/events?token=${encodeURIComponent(token)}`))
    NAMES.forEach((n) => es.addEventListener(n, () => handlers[n].forEach((fn) => fn())))
    es.onopen = () => {
      if (!pertama) muatUlangSemua()
      pertama = false
    }
    // EventSource berhenti mencoba bila server menjawab galat (mis. 502 saat deploy): sambung ulang sendiri
    es.onerror = () => {
      if (es.readyState !== EventSource.CLOSED || source !== es) return
      clearTimeout(timer)
      timer = setTimeout(buka, 5000)
    }
  }
  buka()
}

export function disconnectEvents() {
  clearTimeout(timer)
  source?.close()
  source = null
}

export function onEvent(name, fn) {
  handlers[name].add(fn)
  return () => handlers[name].delete(fn)
}
