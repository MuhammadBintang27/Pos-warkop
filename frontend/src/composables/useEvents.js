const NAMES = ['pesanan', 'meja', 'menu', 'pengaturan', 'shift', 'cetak']
const handlers = Object.fromEntries(NAMES.map((n) => [n, new Set()]))
let source = null

export function connectEvents(token) {
  disconnectEvents()
  source = new EventSource(`/api/events?token=${encodeURIComponent(token)}`)
  NAMES.forEach((n) => source.addEventListener(n, () => handlers[n].forEach((fn) => fn())))
}

export function disconnectEvents() {
  source?.close()
  source = null
}

export function onEvent(name, fn) {
  handlers[name].add(fn)
  return () => handlers[name].delete(fn)
}
