import { fileURLToPath, URL } from 'node:url'
import { readFileSync } from 'node:fs'
import { createHash } from 'node:crypto'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import Components from 'unplugin-vue-components/vite'
import { PrimeVueResolver } from '@primevue/auto-import-resolver'

// sw.js dibuat saat build supaya daftar file ber-hash (termasuk halaman lazy) ikut tertanam dan tersimpan semua.
const serviceWorker = () => ({
  name: 'service-worker',
  apply: 'build',
  generateBundle(_, bundle) {
    const hasil = Object.keys(bundle).filter((f) => f !== 'index.html' && !f.endsWith('.map'))
    const files = ['/', '/favicon.svg', '/pesanan.wav', ...hasil.map((f) => '/' + f)]
    const versi = createHash('sha1').update(files.join()).digest('hex').slice(0, 10)
    const src = readFileSync(fileURLToPath(new URL('./src/sw.js', import.meta.url)), 'utf8')
    const source = `const FILES = ${JSON.stringify(files)}\nconst VERSI = '${versi}'\n${src}`
    this.emitFile({ type: 'asset', fileName: 'sw.js', source })
  },
})

export default defineConfig({
  plugins: [vue(), Components({ resolvers: [PrimeVueResolver()] }), serviceWorker()],
  resolve: { alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) } },
  server: { proxy: { '/api': 'http://localhost:8080' } },
  build: { chunkSizeWarningLimit: 800 },
})
