<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { THEMES } from '@/composables/useAppTheme'
import { useMobile } from '@/composables/useMobile'
import { useAuthStore } from '@/stores/auth'
import { usePengaturanStore } from '@/stores/pengaturan'
import { useRun } from '@/composables/useRun'

const NAV = [
  { label: 'Klien', icon: 'pi pi-building', to: '/admin', roles: ['superadmin'] },
  { label: 'Pengguna', icon: 'pi pi-users', to: '/admin/pengguna', roles: ['superadmin'] },
  { label: 'Pesanan Masuk', icon: 'pi pi-inbox', to: '/tenant-saya', roles: ['tenant'] },
  { label: 'Rekapan', icon: 'pi pi-chart-bar', to: '/tenant-saya/rekap', roles: ['tenant'] },
  { label: 'Kasir', icon: 'pi pi-shopping-cart', to: '/', roles: ['pelayan', 'kasir', 'owner'] },
  { label: 'Denah Meja', icon: 'pi pi-map', to: '/denah', roles: ['pelayan', 'kasir', 'owner'] },
  { label: 'Riwayat', icon: 'pi pi-history', to: '/riwayat', roles: ['pelayan', 'kasir', 'owner'] },
  { label: 'Ketersediaan', icon: 'pi pi-power-off', to: '/ketersediaan', roles: ['kasir', 'owner'] },
  { label: 'Stok Bahan', icon: 'pi pi-box', to: '/bahan', roles: ['kasir', 'owner'] },
  {
    label: 'Laporan',
    icon: 'pi pi-chart-line',
    roles: ['owner'],
    children: [
      { label: 'Ringkasan & Tren', to: '/laporan' },
      { label: 'Kategori, Tenant & Jam', to: '/laporan/waktu' },
      { label: 'Kinerja Karyawan', to: '/laporan/karyawan' },
      { label: 'Kasbon & Stok', to: '/laporan/kasbon' },
      { label: 'Tenant', to: '/laporan/tenant' },
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
const brand = computed(() => (auth.role === 'superadmin' ? 'POS Warkop' : peng.settings.namaUsaha))
const initial = computed(() => (brand.value || 'W').trim()[0].toUpperCase())
const userItems = [{ label: 'Keluar', icon: 'pi pi-sign-out', command: logout }]

function toggleSide() {
  if (mobile.value) drawer.value = !drawer.value
  else hidden.value = !hidden.value
}

function setTheme(theme) {
  run(() => peng.save({ ...peng.settings, theme }))
}

function logout() {
  auth.logout()
  router.push('/login')
}

const setOnline = () => (online.value = navigator.onLine)

watch(() => route.path, () => (drawer.value = false))

onMounted(() => {
  window.addEventListener('online', setOnline)
  window.addEventListener('offline', setOnline)
})
onBeforeUnmount(() => {
  window.removeEventListener('online', setOnline)
  window.removeEventListener('offline', setOnline)
})
</script>

<template>
  <div class="shell">
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

      <div class="status"><i class="dot" :class="online ? 'on' : 'off'" /> {{ online ? 'Online' : 'Offline' }}</div>
    </aside>

    <div class="main">
      <header class="top">
        <Button icon="pi pi-bars" variant="outlined" aria-label="Menu" @click="toggleSide" />

        <div class="right">
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

      <main class="content"><RouterView /></main>
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

.main {
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

@media (max-width: 600px) {
  .who {
    display: none;
  }
}
</style>
