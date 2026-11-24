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

// huruf judul ala buku menu; dimuat hanya di halaman ini
function muatFont() {
  if (document.getElementById('font-menu')) return
  const l = Object.assign(document.createElement('link'), {
    id: 'font-menu',
    rel: 'stylesheet',
    href: 'https://fonts.googleapis.com/css2?family=Playfair+Display:wght@700;800&display=swap',
  })
  document.head.appendChild(l)
}

onMounted(async () => {
  muatFont()
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
const jumlahMenu = computed(() => data.value?.tenant.reduce((s, t) => s + t.menu.length, 0) ?? 0)

const labelOpsi = (g, o) => (g.gantiHarga ? `${o.nama} ${rp(o.tambahan)}` : o.tambahan ? `${o.nama} +${rp(o.tambahan)}` : o.nama)
const lompat = (i) => document.getElementById(`tenant-${i}`)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
</script>

<template>
  <div class="mp">
    <p v-if="galat" class="kosong">{{ galat }}</p>
    <p v-else-if="!data" class="kosong"><i class="pi pi-spin pi-spinner" /> Memuat menu…</p>

    <template v-else>
      <header class="hero">
        <div class="hero-isi">
          <img v-if="data.usaha.logo" :src="data.usaha.logo" alt="" class="logo" />
          <span class="eyebrow">Daftar Menu</span>
          <h1>{{ data.usaha.nama }}</h1>
          <p v-if="data.usaha.alamat" class="alamat"><i class="pi pi-map-marker" /> {{ data.usaha.alamat }}</p>
          <p class="ringkas">{{ data.tenant.length }} dapur · {{ jumlahMenu }} pilihan menu</p>
        </div>
      </header>

      <div class="tempel">
        <div class="lebar tempel-isi">
          <IconField class="cari">
            <InputIcon class="pi pi-search" />
            <InputText v-model="cari" placeholder="Cari menu, misalnya kopi atau mie…" fluid />
          </IconField>
          <nav v-if="!cari" class="chips">
            <button v-for="(t, i) in tenant" :key="t.nama" class="chip-t" @click="lompat(i)">{{ t.nama }}</button>
          </nav>
        </div>
      </div>

      <main class="lebar">
        <p v-if="!tenant.length" class="kosong">Tidak ada menu yang cocok dengan "{{ cari }}".</p>

        <section v-for="(t, i) in tenant" :id="`tenant-${i}`" :key="t.nama" class="tenant">
          <div class="judul">
            <h2>{{ t.nama }}</h2>
            <span class="garis" />
            <span v-if="!t.buka" class="label tutup">Sedang tutup</span>
            <span class="jumlah">{{ t.menu.length }} menu</span>
          </div>

          <div class="grid">
            <article v-for="(m, j) in t.menu" :key="j" class="kartu" :class="{ redup: m.habis || !t.buka }">
              <span v-if="m.habis" class="pita">Habis</span>
              <span class="kat">{{ m.kategori }}</span>
              <h3>{{ m.nama }}</h3>
              <div v-if="m.varian.length" class="varian">
                <span v-for="g in m.varian" :key="g.nama" class="opsi">
                  <b>{{ g.nama }}</b> {{ g.opsi.map((o) => labelOpsi(g, o)).join(' · ') }}
                </span>
              </div>
              <div class="harga">{{ hargaTampil(m) }}</div>
            </article>
          </div>
        </section>

        <footer class="kaki">
          <p>Harga dapat berubah sewaktu-waktu.<br />Pesan langsung ke kasir atau pelayan kami.</p>
        </footer>
      </main>
    </template>
  </div>
</template>

<style scoped>
.mp {
  --display: 'Playfair Display', Georgia, 'Times New Roman', serif;
  min-height: 100dvh;
  background: color-mix(in srgb, var(--p-primary-color) 5%, var(--p-content-background));
  color: var(--p-text-color);
}
.lebar {
  width: 100%;
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 24px;
}
.kosong {
  padding: 64px 16px;
  text-align: center;
  color: var(--p-text-muted-color);
}

/* ---------- hero ---------- */
.hero {
  position: relative;
  overflow: hidden;
  padding: 56px 24px 72px;
  text-align: center;
  color: var(--p-primary-contrast-color);
  background:
    radial-gradient(rgb(255 255 255 / 0.13) 1.2px, transparent 1.2px) 0 0 / 18px 18px,
    linear-gradient(135deg, var(--p-primary-color), color-mix(in srgb, var(--p-primary-color) 62%, #2a1608));
}
.hero::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: -1px;
  height: 36px;
  background: color-mix(in srgb, var(--p-primary-color) 5%, var(--p-content-background));
  border-radius: 50% 50% 0 0 / 100% 100% 0 0;
}
.hero-isi {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}
.logo {
  width: 96px;
  height: 96px;
  padding: 8px;
  margin-bottom: 6px;
  object-fit: contain;
  border-radius: 24px;
  background: #fff;
  box-shadow: 0 10px 30px rgb(0 0 0 / 0.25);
}
.eyebrow {
  font-size: 0.8rem;
  font-weight: 700;
  letter-spacing: 0.3em;
  text-transform: uppercase;
  opacity: 0.85;
}
.hero h1 {
  margin: 0;
  font-family: var(--display);
  font-size: clamp(2.2rem, 6vw, 3.6rem);
  font-weight: 800;
  line-height: 1.05;
  letter-spacing: 0.01em;
  overflow-wrap: anywhere;
  text-shadow: 0 3px 16px rgb(0 0 0 / 0.2);
}
.alamat,
.ringkas {
  margin: 0;
  opacity: 0.92;
}
.ringkas {
  margin-top: 6px;
  padding: 6px 16px;
  border-radius: 999px;
  font-size: 0.9rem;
  font-weight: 600;
  background: rgb(255 255 255 / 0.18);
}

/* ---------- bilah cari ---------- */
.tempel {
  position: sticky;
  top: 0;
  z-index: 10;
  padding: 12px 0;
  background: color-mix(in srgb, var(--p-content-background) 82%, transparent);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--p-content-border-color);
}
.tempel-isi {
  display: flex;
  align-items: center;
  gap: 14px;
}
.cari {
  flex: 0 1 380px;
}
.chips {
  flex: 1;
  display: flex;
  gap: 8px;
  overflow-x: auto;
  scrollbar-width: none;
}
.chip-t {
  flex: none;
  padding: 8px 16px;
  border: 1px solid color-mix(in srgb, var(--p-primary-color) 35%, transparent);
  border-radius: 999px;
  font: inherit;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  background: var(--p-content-background);
  color: var(--accent-text, var(--p-primary-color));
  transition: background 0.15s, color 0.15s;
}
.chip-t:hover {
  background: var(--p-primary-color);
  color: var(--p-primary-contrast-color);
}

/* ---------- tenant ---------- */
.tenant {
  scroll-margin-top: 90px;
  padding-top: 36px;
}
.judul {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 18px;
}
.judul h2 {
  margin: 0;
  font-family: var(--display);
  font-size: clamp(1.5rem, 3vw, 2rem);
  font-weight: 800;
}
.garis {
  flex: 1;
  height: 2px;
  background: linear-gradient(90deg, var(--p-primary-color), transparent);
  opacity: 0.6;
}
.jumlah {
  font-size: 0.85rem;
  color: var(--p-text-muted-color);
  white-space: nowrap;
}
.label {
  padding: 4px 10px;
  border-radius: 8px;
  font-size: 0.75rem;
  font-weight: 700;
  white-space: nowrap;
}
.tutup {
  background: var(--p-red-500, #ef4444);
  color: #fff;
}

/* ---------- kartu menu ---------- */
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(230px, 1fr));
  gap: 18px;
}
.kartu {
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 18px;
  border: 1px solid var(--p-content-border-color);
  border-radius: 20px;
  background: var(--p-content-background);
  box-shadow: 0 2px 10px rgb(0 0 0 / 0.05);
  transition: transform 0.18s ease, box-shadow 0.18s ease;
}
.kartu::before {
  content: '';
  position: absolute;
  inset: 0 0 auto;
  height: 4px;
  background: linear-gradient(90deg, var(--p-primary-color), color-mix(in srgb, var(--p-primary-color) 40%, transparent));
}
.kartu:hover {
  transform: translateY(-4px);
  box-shadow: 0 14px 30px color-mix(in srgb, var(--p-primary-color) 22%, transparent);
}
.kartu.redup {
  filter: grayscale(0.85);
  opacity: 0.6;
}
.pita {
  position: absolute;
  top: 16px;
  right: -34px;
  width: 130px;
  padding: 4px 0;
  transform: rotate(45deg);
  text-align: center;
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  background: var(--p-red-500, #ef4444);
  color: #fff;
}
.kat {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--accent-text, var(--p-primary-color));
}
.kartu h3 {
  margin: 0;
  font-size: 1.05rem;
  line-height: 1.3;
}
.varian {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.opsi {
  font-size: 0.82rem;
  color: var(--p-text-muted-color);
}
.opsi b {
  color: var(--p-text-color);
  font-weight: 600;
}
.harga {
  margin-top: auto;
  padding-top: 10px;
  border-top: 1px dashed var(--p-content-border-color);
  font-family: var(--display);
  font-size: 1.35rem;
  font-weight: 800;
  color: var(--accent-text, var(--p-primary-color));
}

.kaki {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 48px 0 40px;
  text-align: center;
  color: var(--p-text-muted-color);
}
.kaki p {
  margin: 0;
  font-size: 0.9rem;
}

/* ---------- HP ---------- */
@media (max-width: 640px) {
  .lebar {
    padding: 0 14px;
  }
  .hero {
    padding: 40px 16px 60px;
  }
  .tempel-isi {
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
  }
  .cari {
    flex: none;
  }
  .grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
  }
  .kartu {
    padding: 14px;
    border-radius: 16px;
    gap: 8px;
  }
  .kartu h3 {
    font-size: 0.95rem;
  }
  .harga {
    font-size: 1.1rem;
  }
  .judul .jumlah {
    display: none;
  }
}
</style>
