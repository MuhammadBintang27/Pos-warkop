<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import { onEvent } from '@/composables/useEvents'
import { THEMES } from '@/composables/useAppTheme'
import { useMobile } from '@/composables/useMobile'
import { useAuthStore } from '@/stores/auth'
import { useKasirStore } from '@/stores/kasir'
import { usePengaturanStore } from '@/stores/pengaturan'
import { useRun } from '@/composables/useRun'
import { load, save } from '@/utils/storage'
import { useConfirm } from 'primevue/useconfirm'
import { aksiTerkait, cetakTunda, cetakUlang, cobaLagi, lewati, offline, tutupGagalCetak, tutupTunda } from '@/api/offline'
import { konfirmasi } from '@/utils/confirm'

const NAV = [
  { label: 'Klien', icon: 'pi pi-building', to: '/admin', roles: ['superadmin'] },
  { label: 'Pengguna', icon: 'pi pi-users', to: '/admin/pengguna', roles: ['superadmin'] },
  { label: 'Pesanan Masuk', icon: 'pi pi-inbox', to: '/tenant-saya', roles: ['tenant'] },
  { label: 'Rekapan', icon: 'pi pi-chart-bar', to: '/tenant-saya/rekap', roles: ['tenant'] },
  { label: 'Kasir', icon: 'pi pi-shopping-cart', to: '/', roles: ['pelayan', 'kasir', 'owner'] },
  { label: 'Denah Meja', icon: 'pi pi-map', to: '/denah', roles: ['pelayan', 'kasir', 'owner'] },
  { label: 'Belum Tercetak', icon: 'pi pi-print', to: '/belum-tercetak', roles: ['pelayan', 'kasir', 'owner'], badge: 'cetak' },
  { label: 'Riwayat', icon: 'pi pi-history', to: '/riwayat', roles: ['pelayan', 'kasir', 'owner'] },
  { label: 'Ketersediaan', icon: 'pi pi-power-off', to: '/ketersediaan', roles: ['kasir', 'owner'] },
  { label: 'Stok Bahan', icon: 'pi pi-box', to: '/bahan', roles: ['kasir', 'owner'] },
  // owner membukanya dari grup Laporan › Tenant
  { label: 'Closing Tenant', icon: 'pi pi-lock', to: '/laporan/tenant', roles: ['kasir'] },
  {
    label: 'Master Data',
    icon: 'pi pi-database',
    roles: ['kasir'],
    children: [
      { label: 'Menu', to: '/menu' },
      { label: 'Tenant', to: '/tenant' },
      { label: 'Printer', to: '/printer' },
      { label: 'Meja', to: '/meja' },
    ],
  },
  {
    label: 'Laporan',
    icon: 'pi pi-chart-line',
    roles: ['owner'],
    children: [
      { label: 'Ringkasan & Tren', to: '/laporan' },
      { label: 'Kategori, Tenant & Jam', to: '/laporan/waktu' },
      { label: 'Kinerja Karyawan', to: '/laporan/karyawan' },
      { label: 'Kasbon & Stok', to: '/laporan/kasbon' },
      { label: 'Riwayat Shift', to: '/laporan/shift' },
      { label: 'Tenant', to: '/laporan/tenant' },
      { label: 'Riwayat Pengeluaran', to: '/pengeluaran' },
      { label: 'Laba Rugi', to: '/laba-rugi' },
    ],
  },
  {
    label: 'Master Data',
    icon: 'pi pi-database',
    roles: ['owner'],
    children: [
      { label: 'Menu', to: '/menu' },
      { label: 'Tenant', to: '/tenant' },
      { label: 'Printer', to: '/printer' },
      { label: 'Meja', to: '/meja' },
      { label: 'Karyawan dan Gaji', to: '/karyawan' },
      { label: 'Pengguna', to: '/pengguna' },
    ],
  },
  { label: 'Pengaturan', icon: 'pi pi-cog', roles: ['owner'], children: [{ label: 'Tampilan', to: '/pengaturan' }, { label: 'Log Aktivitas', to: '/audit' }] },
  { label: 'Panduan', icon: 'pi pi-question-circle', to: '/panduan', roles: ['pelayan', 'kasir', 'owner', 'tenant'] },
]

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const peng = usePengaturanStore()
const run = useRun()
const mobile = useMobile()

const q = ref('')
const hidden = ref(false)
const drawer = ref(false)
const expanded = reactive({})
const userMenu = ref()
const online = ref(navigator.onLine)
const topTutup = ref(load('topTutup', false))
watch(topTutup, (v) => save('topTutup', v))

const nav = computed(() => {
  const term = q.value.trim().toLowerCase()
  return NAV.filter((n) => !n.roles || n.roles.includes(auth.role))
    .map((n) => {
      if (!term) return n
      if (n.label.toLowerCase().includes(term)) return n
      const children = n.children?.filter((c) => c.label.toLowerCase().includes(term))
      return children?.length ? { ...n, children } : null
    })
    .filter(Boolean)
})

const showSide = computed(() => (mobile.value ? drawer.value : !hidden.value))
const inRoute = (n) => n.children?.some((c) => c.to === route.path)
const isOpen = (n) => expanded[n.label] ?? inRoute(n)
const brand = computed(() => (auth.role === 'superadmin' ? 'PSD POS' : peng.settings.namaUsaha))
const initial = computed(() => (brand.value || 'W').trim()[0].toUpperCase())
const userItems = [{ label: 'Keluar', icon: 'pi pi-sign-out', command: logout }]

function toggleSide() {
  if (mobile.value) drawer.value = !drawer.value
  else hidden.value = !hidden.value
}

function setTheme(theme) {
  run(() => peng.save({ ...peng.settings, theme }))
}

const jamTerakhir = computed(() => new Date(offline.terakhir).toLocaleString('id-ID', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' }))
const confirm = useConfirm()
function konfirmasiLewati() {
  const n = aksiTerkait()
  const ikut = n > 1 ? ` Bill ini dibuat saat offline, jadi ${n - 1} aksi lanjutannya (tambah item, kirim, bayar, dll.) ikut dibuang.` : ''
  konfirmasi(confirm, {
    header: n > 1 ? `Lewati ${n} aksi?` : 'Lewati aksi ini?',
    message: `"${offline.galat}" tidak akan pernah tercatat di server.${ikut} Lewati hanya bila sudah dipastikan, misalnya bill itu sudah diurus dari perangkat lain.`,
    label: 'Lewati',
    accept: lewati,
  })
}

function logout() {
  // antrean offline hanya bisa dikirim dengan akun yang membuatnya
  if (offline.antrean.length) {
    return toast.add({ severity: 'warn', summary: `Masih ada ${offline.antrean.length} aksi offline belum terkirim`, detail: 'Sambungkan internet dulu sebelum keluar', life: 6000 })
  }
  auth.logout()
  router.push('/login')
}

const setOnline = () => (online.value = navigator.onLine)

// Tiket yang belum tercetak dipantau di semua halaman: badge di menu, dan toast sekali untuk tiap tiket yang baru macet.
const kasir = useKasirStore()
const toast = useToast()
const pantauCetak = ['pelayan', 'kasir', 'owner'].includes(auth.role)
const dikenal = new Set()
const muatCetak = () => kasir.loadCetak().catch(() => {})
let timerCetak, offCetak

watch(
  () => kasir.cetak.daftar,
  (daftar = []) => {
    const baru = daftar.filter((j) => !dikenal.has(j.id))
    baru.forEach((j) => dikenal.add(j.id))
    if (!baru.length || route.path === '/belum-tercetak') return
    const j = baru[0]
    toast.add({
      group: 'cetak',
      severity: 'warn',
      summary: baru.length === 1 ? `Tiket ${j.mejaNama || 'tanpa meja'} (${j.printer}) belum tercetak` : `${baru.length} tiket belum tercetak`,
      detail: [j.galat || (!kasir.cetak.stasiunAktif && 'Stasiun cetak tidak terhubung'), 'Ketuk untuk melihat'].filter(Boolean).join(' — '),
      life: 8000,
    })
  },
)

function bukaBelumTercetak(msg) {
  toast.remove(msg)
  router.push('/belum-tercetak')
}

// versi APK Android yang terpasang (aplikasi 1.1 belum punya versiNama, jadi pakai nomor build-nya)
const hp = window.StasiunAndroid
const versiApk = hp?.versiNama ? hp.versiNama() : hp?.versi ? `build ${hp.versi()}` : ''

watch(() => route.path, () => (drawer.value = false))

// Tarik ke bawah dari atas halaman untuk memuat ulang: halaman digulir di .content, jadi WebView tidak punya bawaannya.
const konten = ref()
const tarik = ref(0)
const BATAS_TARIK = 70
let awalTarik = null

// daftar di dalam halaman yang sedang tergulir harus kembali ke atas dulu, supaya gulir biasa tidak memicu muat ulang
function diPuncak(el) {
  for (; el && el !== konten.value; el = el.parentElement) if (el.scrollTop > 0) return false
  return konten.value.scrollTop === 0
}
function tarikMulai(e) {
  awalTarik = e.touches.length === 1 && diPuncak(e.target) ? e.touches[0].clientY : null
}
function tarikGerak(e) {
  if (awalTarik !== null) tarik.value = Math.max(0, Math.min(BATAS_TARIK * 1.4, (e.touches[0].clientY - awalTarik) / 2))
}
function tarikSelesai() {
  if (tarik.value >= BATAS_TARIK) return location.reload()
  tarik.value = 0
  awalTarik = null
}

onMounted(() => {
  if (pantauCetak) {
    muatCetak()
    offCetak = onEvent('cetak', muatCetak)
    timerCetak = setInterval(muatCetak, 10000) // status stasiun & tiket macet berubah tanpa event
  }
  window.addEventListener('online', setOnline)
  window.addEventListener('offline', setOnline)
})
onBeforeUnmount(() => {
  offCetak?.()
  clearInterval(timerCetak)
  window.removeEventListener('online', setOnline)
  window.removeEventListener('offline', setOnline)
})
</script>

<template>
  <div class="shell">
    <Toast group="cetak" position="top-center">
      <template #message="{ message }">
        <button class="toast-cetak" @click="bukaBelumTercetak(message)">
          <i class="pi pi-print" />
          <span><strong>{{ message.summary }}</strong><small>{{ message.detail }}</small></span>
        </button>
      </template>
    </Toast>
    <div v-if="mobile && drawer" class="scrim" @click="drawer = false" />

    <aside class="side" :class="{ open: showSide, floating: mobile }">
      <div class="logo">
        <img v-if="peng.settings.logo && auth.role !== 'superadmin'" :src="peng.settings.logo" alt="" class="mark img" />
        <div v-else class="mark">{{ initial }}</div>
        <strong>{{ brand }}</strong>
      </div>

      <IconField class="search">
        <InputIcon class="pi pi-search" />
        <InputText v-model="q" placeholder="Cari menu…" fluid />
      </IconField>

      <div class="caption">MENU</div>
      <nav class="nav">
        <template v-for="n in nav" :key="n.label">
          <RouterLink v-if="n.to" :to="n.to" class="item" active-class="active" exact-active-class="active">
            <i :class="n.icon" /> <span>{{ n.label }}</span>
            <span v-if="n.badge && kasir.cetak.macet" class="badge">{{ kasir.cetak.macet }}</span>
          </RouterLink>
          <template v-else>
            <button class="item" :class="{ parent: inRoute(n) }" @click="expanded[n.label] = !isOpen(n)">
              <i :class="n.icon" /> <span>{{ n.label }}</span>
              <i class="pi chev" :class="isOpen(n) ? 'pi-chevron-down' : 'pi-chevron-right'" />
            </button>
            <div v-show="isOpen(n) || q" class="children">
              <RouterLink v-for="c in n.children" :key="c.to" :to="c.to" class="sub" exact-active-class="active">
                <i class="dot" /> {{ c.label }}
              </RouterLink>
            </div>
          </template>
        </template>
      </nav>

      <div class="status">
        <i class="dot" :class="online ? 'on' : 'off'" /> {{ online ? 'Online' : 'Offline' }}
        <span v-if="versiApk" class="versi">· v{{ versiApk }}</span>
      </div>
    </aside>

    <div class="main">
      <button v-if="topTutup" class="top-mini" aria-label="Tampilkan header" title="Tampilkan header" @click="topTutup = false">
        <i class="pi pi-angle-down" />
      </button>
      <header v-else class="top">
        <Button icon="pi pi-bars" variant="outlined" aria-label="Menu" @click="toggleSide" />

        <div class="right">
          <Button icon="pi pi-angle-up" text severity="secondary" aria-label="Sembunyikan header" title="Sembunyikan header" @click="topTutup = true" />
          <template v-if="auth.isOwner && !mobile">
            <i class="pi pi-palette" />
            <Select
              :model-value="peng.settings.theme"
              :options="THEMES"
              option-label="label"
              option-value="id"
              class="theme"
              @update:model-value="setTheme"
            />
          </template>
          <button class="user" @click="userMenu.toggle($event)">
            <Avatar :label="(auth.user?.nama || '?')[0].toUpperCase()" shape="circle" />
            <span class="who">
              <strong>{{ auth.user?.nama }}</strong>
              <small>{{ auth.role }}</small>
            </span>
            <i class="pi pi-chevron-down" />
          </button>
          <Menu ref="userMenu" :model="userItems" popup />
        </div>
      </header>

      <div v-if="offline.lokal || offline.galat || offline.antrean.length || offline.gagalCetak.length || offline.tundaCetak.length" class="offline-bar" role="status">
        <div v-if="offline.lokal && !offline.antrean.length" class="baris">
          <i class="pi pi-wifi" />
          <span>Server tidak terjangkau. Data dari simpanan HP{{ offline.terakhir ? `, terakhir diperbarui ${jamTerakhir}` : '' }}</span>
        </div>
        <div v-if="offline.galat" class="baris galat">
          <i class="pi pi-exclamation-triangle" />
          <span>Aksi offline ditolak server: {{ offline.galat }}</span>
          <Button label="Coba lagi" size="small" @click="cobaLagi" />
          <Button label="Lewati" size="small" severity="danger" text @click="konfirmasiLewati" />
        </div>
        <div v-else-if="offline.antrean.length" class="baris">
          <i class="pi" :class="offline.sinkron ? 'pi-spin pi-sync' : 'pi-cloud-upload'" />
          <span>Mode offline: {{ offline.antrean.length }} aksi menunggu dikirim ke server</span>
        </div>
        <div v-for="(t, i) in offline.tundaCetak" :key="`t${i}`" class="baris">
          <i class="pi pi-hourglass" />
          <span>Tiket {{ t.label }} menunggu kepastian server, mungkin sudah tercetak lewat stasiun. Tanyakan ke dapur bila internet belum kembali.</span>
          <Button label="Cetak sekarang" size="small" @click="cetakTunda(i)" />
          <Button icon="pi pi-times" size="small" text severity="secondary" aria-label="Sudah tercetak, tutup" @click="tutupTunda(i)" />
        </div>
        <div v-for="(g, i) in offline.gagalCetak" :key="i" class="baris galat">
          <i class="pi pi-print" />
          <span>{{ g.label }} gagal dicetak: {{ g.galat }}</span>
          <Button label="Cetak ulang" size="small" @click="cetakUlang(i)" />
          <Button icon="pi pi-times" size="small" text severity="secondary" aria-label="Tutup" @click="tutupGagalCetak(i)" />
        </div>
      </div>
      <div v-if="tarik" class="tarik" :class="{ siap: tarik >= BATAS_TARIK }" :style="{ transform: `translate(-50%, ${tarik}px)` }">
        <i class="pi pi-refresh" :style="{ rotate: `${tarik * 4}deg` }" />
      </div>
      <main
        ref="konten"
        class="content"
        @touchstart.passive="tarikMulai"
        @touchmove.passive="tarikGerak"
        @touchend="tarikSelesai"
        @touchcancel="tarikSelesai"
      >
        <RouterView />
      </main>
    </div>
  </div>
</template>

<style scoped>
.shell {
  display: flex;
  height: 100dvh;
  overflow: hidden;
}

.side {
  width: 0;
  flex: none;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: color-mix(in srgb, var(--p-primary-color) 7%, var(--p-content-background));
  border-right: 1px solid var(--p-content-border-color);
  transition: width 0.2s ease;
}
.side.open {
  width: 260px;
}
.side.floating {
  position: fixed;
  inset: 0 auto 0 0;
  z-index: 50;
  transition: none;
  box-shadow: 8px 0 24px rgb(0 0 0 / 0.18);
}
.scrim {
  position: fixed;
  inset: 0;
  background: rgb(0 0 0 / 0.4);
  z-index: 40;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 64px;
  padding: 0 18px;
  flex: none;
  white-space: nowrap;
  border-bottom: 1px solid var(--p-content-border-color);
}
.mark.img {
  object-fit: contain;
  background: transparent;
}
.mark {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  font-weight: 800;
  background: var(--p-primary-color);
  color: var(--p-primary-contrast-color);
}
.search {
  margin: 16px 16px 4px;
  flex: none;
}
.caption {
  padding: 16px 24px 6px;
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  color: var(--accent-text);
}

.nav {
  flex: 1;
  overflow-y: auto;
  padding: 0 12px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 11px 12px;
  border: 0;
  border-radius: 10px;
  background: none;
  font: inherit;
  font-weight: 600;
  text-align: left;
  color: var(--p-text-color);
  cursor: pointer;
  white-space: nowrap;
}
.item:hover {
  background: color-mix(in srgb, var(--p-primary-color) 12%, transparent);
}
.item.active {
  background: var(--p-primary-color);
  color: var(--p-primary-contrast-color);
}
.item.parent {
  color: var(--accent-text);
}
.badge {
  margin-left: auto;
  min-width: 22px;
  padding: 1px 7px;
  border-radius: 999px;
  font-size: 0.78rem;
  font-weight: 700;
  text-align: center;
  background: var(--p-red-600, #dc2626);
  color: #fff;
}
.toast-cetak {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  width: 100%;
  padding: 0;
  border: 0;
  background: none;
  font: inherit;
  color: inherit;
  text-align: left;
  cursor: pointer;
}
.toast-cetak .pi {
  font-size: 1.3rem;
  margin-top: 2px;
}
.toast-cetak span {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.chev {
  margin-left: auto;
  font-size: 0.8rem;
}
.children {
  margin: 2px 0 6px 22px;
  padding-left: 8px;
  border-left: 2px solid color-mix(in srgb, var(--p-primary-color) 25%, transparent);
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.sub {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 9px 12px;
  border-radius: 10px;
  font-weight: 600;
  font-size: 0.95rem;
  white-space: nowrap;
}
.sub:hover {
  background: color-mix(in srgb, var(--p-primary-color) 12%, transparent);
}
.sub.active {
  background: var(--p-primary-color);
  color: var(--p-primary-contrast-color);
}
.dot {
  flex-shrink: 0; /* label panjang tidak boleh menggencet titiknya */
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
  opacity: 0.7;
}
.status {
  flex: none;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 20px;
  border-top: 1px solid var(--p-content-border-color);
  font-size: 0.9rem;
  font-weight: 600;
}
.status .dot {
  width: 9px;
  height: 9px;
  opacity: 1;
}
.status .on {
  background: var(--p-green-500, #22c55e);
}
.status .off {
  background: var(--p-red-500, #ef4444);
}
.status .versi {
  font-weight: 400;
  color: var(--p-text-muted-color);
}

.main {
  position: relative;
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.top {
  height: 64px;
  flex: none;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 16px;
  background: color-mix(in srgb, var(--p-primary-color) 9%, var(--p-content-background));
  border-bottom: 1px solid var(--p-content-border-color);
}
/* tab melayang, cukup besar untuk jari di HP, tidak memakan baris sendiri */
.top-mini {
  position: absolute;
  top: 0;
  left: 50%;
  z-index: 30;
  width: 72px;
  height: 32px;
  transform: translateX(-50%);
  display: grid;
  place-items: center;
  padding: 0;
  border: 0;
  border-radius: 0 0 14px 14px;
  font-size: 1.1rem;
  cursor: pointer;
  background: var(--p-primary-color);
  color: var(--p-primary-contrast-color);
  box-shadow: 0 3px 10px rgb(0 0 0 / 0.25);
  opacity: 0.9;
}
.top-mini:hover {
  opacity: 1;
}
.right {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 10px;
}
.theme {
  min-width: 140px;
}
.user {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 10px;
  border: 1px solid var(--p-content-border-color);
  border-radius: 12px;
  background: var(--p-content-background);
  font: inherit;
  color: inherit;
  cursor: pointer;
}
.who {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  line-height: 1.2;
}
.who small {
  color: var(--p-text-muted-color);
  text-transform: capitalize;
}
.content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}
.tarik {
  position: absolute;
  top: 0;
  left: 50%;
  z-index: 30;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  background: var(--p-content-background);
  color: var(--p-text-muted-color);
  box-shadow: 0 4px 12px rgb(0 0 0 / 0.18);
}
.tarik.siap {
  color: var(--p-primary-color);
}
.offline-bar {
  flex: none;
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 8px 16px;
  font-size: 0.875rem;
  background: color-mix(in srgb, var(--p-yellow-500) 16%, var(--p-content-background));
  border-bottom: 1px solid var(--p-content-border-color);
}
.offline-bar .baris {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}
.offline-bar .baris span {
  flex: 1;
  min-width: 12rem;
}
.offline-bar .galat {
  color: var(--p-red-600);
}

@media (max-width: 600px) {
  .who {
    display: none;
  }
}
</style>
