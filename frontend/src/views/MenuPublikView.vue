<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { errMsg, get } from '@/api/http'
import { applyTheme, DEFAULT_SETTINGS } from '@/composables/useAppTheme'
import { hargaTampil, rp } from '@/utils/format'

const route = useRoute()
const data = ref(null)
const galat = ref('')
const cari = ref('')

onMounted(async () => {
  try {
    data.value = await get(`/publik/menu/${encodeURIComponent(route.params.kode)}`)
    const u = data.value.usaha
    document.title = `Menu ${u.nama}`
    // persist false: tema pelanggan tidak menimpa tema tersimpan bila perangkat ini juga dipakai staf
    applyTheme({ ...DEFAULT_SETTINGS, theme: u.theme, primary: u.primary, mode: u.mode }, { persist: false })
  } catch (e) {
    galat.value = e.response?.status === 404 ? 'Menu tidak ditemukan. Periksa lagi link atau QR-nya.' : errMsg(e)
  }
})

const tenant = computed(() => {
  const q = cari.value.trim().toLowerCase()
  if (!data.value) return []
  if (!q) return data.value.tenant
  return data.value.tenant
    .map((t) => ({ ...t, menu: t.menu.filter((m) => m.nama.toLowerCase().includes(q) || m.kategori.toLowerCase().includes(q)) }))
    .filter((t) => t.menu.length)
})

const labelOpsi = (g, o) => (g.gantiHarga ? `${o.nama} ${rp(o.tambahan)}` : o.tambahan ? `${o.nama} +${rp(o.tambahan)}` : o.nama)
const lompat = (i) => document.getElementById(`tenant-${i}`)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
</script>

<template>
  <div class="mp">
    <p v-if="galat" class="kosong">{{ galat }}</p>
    <p v-else-if="!data" class="kosong"><i class="pi pi-spin pi-spinner" /> Memuat menu…</p>

    <template v-else>
      <header class="kepala">
        <img v-if="data.usaha.logo" :src="data.usaha.logo" alt="" class="logo" />
        <div>
          <h1>{{ data.usaha.nama }}</h1>
          <small v-if="data.usaha.alamat">{{ data.usaha.alamat }}</small>
        </div>
      </header>

      <div class="tempel">
        <IconField>
          <InputIcon class="pi pi-search" />
          <InputText v-model="cari" placeholder="Cari menu…" fluid />
        </IconField>
        <nav v-if="!cari" class="chips">
          <button v-for="(t, i) in tenant" :key="t.nama" class="chip-t" @click="lompat(i)">{{ t.nama }}</button>
        </nav>
      </div>

      <p v-if="!tenant.length" class="kosong">Tidak ada menu yang cocok.</p>

      <section v-for="(t, i) in tenant" :id="`tenant-${i}`" :key="t.nama" class="tenant">
        <h2>
          {{ t.nama }}
          <span v-if="!t.buka" class="tutup">Tutup</span>
        </h2>
        <ul>
          <li v-for="(m, j) in t.menu" :key="j" :class="{ redup: m.habis || !t.buka }">
            <div class="isi">
              <strong>{{ m.nama }}</strong>
              <small class="kat">{{ m.kategori }}</small>
              <small v-for="g in m.varian" :key="g.nama" class="varian">{{ g.nama }}: {{ g.opsi.map((o) => labelOpsi(g, o)).join(', ') }}</small>
            </div>
            <div class="harga">
              <b>{{ hargaTampil(m) }}</b>
              <span v-if="m.habis" class="habis">Habis</span>
            </div>
          </li>
        </ul>
      </section>

      <footer class="kaki">Harga dapat berubah sewaktu-waktu. Pesan langsung ke kasir atau pelayan.</footer>
    </template>
  </div>
</template>

<style scoped>
.mp {
  max-width: 720px;
  min-height: 100dvh;
  margin: 0 auto;
  padding: 0 16px 32px;
  background: var(--p-content-background);
  color: var(--p-text-color);
}
.kosong {
  padding: 48px 0;
  text-align: center;
  color: var(--p-text-muted-color);
}
.kepala {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 24px 0 12px;
}
.logo {
  width: 56px;
  height: 56px;
  object-fit: contain;
  border-radius: 12px;
}
.kepala h1 {
  margin: 0;
  font-size: 1.5rem;
}
.kepala small {
  color: var(--p-text-muted-color);
}
.tempel {
  position: sticky;
  top: 0;
  z-index: 5;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 10px 0;
  background: var(--p-content-background);
}
.chips {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  scrollbar-width: none;
}
.chip-t {
  flex: none;
  padding: 7px 14px;
  border: 1px solid var(--p-content-border-color);
  border-radius: 999px;
  font: inherit;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  background: var(--p-content-background);
  color: var(--accent-text, var(--p-primary-color));
}
.tenant {
  scroll-margin-top: 120px;
  margin-top: 18px;
}
.tenant h2 {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0 0 4px;
  padding-bottom: 8px;
  font-size: 1.15rem;
  border-bottom: 2px solid var(--p-primary-color);
}
.tutup,
.habis {
  padding: 2px 8px;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 700;
  background: var(--p-red-500, #ef4444);
  color: #fff;
}
ul {
  list-style: none;
  margin: 0;
  padding: 0;
}
li {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid var(--p-content-border-color);
}
li.redup {
  opacity: 0.5;
}
.isi {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}
.kat,
.varian {
  color: var(--p-text-muted-color);
}
.harga {
  flex: none;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
  color: var(--accent-text, var(--p-primary-color));
}
.kaki {
  margin-top: 28px;
  text-align: center;
  font-size: 0.85rem;
  color: var(--p-text-muted-color);
}
</style>
