<script setup>
import { computed, ref } from 'vue'
import CetakMacetDialog from '@/components/kasir/CetakMacetDialog.vue'
import { useKasirStore } from '@/stores/kasir'

// status cetak dimuat dan dipantau AppLayout untuk semua halaman; halaman ini hanya menampilkannya per meja
const store = useKasirStore()
const TANPA_MEJA = 'Tanpa meja'

const meja = computed(() => {
  const g = new Map()
  for (const j of store.cetak.daftar ?? []) {
    const k = j.mejaNama || TANPA_MEJA
    if (!g.has(k)) g.set(k, [])
    g.get(k).push(j)
  }
  // "Tanpa meja" (bawa pulang, struk, laporan) selalu di akhir
  return [...g].map(([nama, tiket]) => ({ nama, tiket })).sort((a, b) => (a.nama === TANPA_MEJA) - (b.nama === TANPA_MEJA))
})

const dipilih = ref('')
const buka = ref(false)
const tiketDipilih = computed(() => meja.value.find((m) => m.nama === dipilih.value)?.tiket ?? [])
const printer = (m) => [...new Set(m.tiket.map((j) => j.printer))].join(', ')

function pilih(m) {
  dipilih.value = m.nama
  buka.value = true
}

// stasiun di HP ini (aplikasi Android): bisa dinyalakan ulang bila owner sudah menyimpan tokennya
const hp = window.StasiunAndroid
const hpPunyaToken = () => JSON.parse(hp.status()).adaToken
function nyalakanHp() {
  hp.simpan(location.origin, '', true)
  setTimeout(() => store.loadCetak().catch(() => {}), 4000)
}
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-print" title="Belum Tercetak" subtitle="Ketuk meja untuk mencetak lewat browser atau mengabaikan tiketnya" />

    <p v-if="!store.cetak.stasiunAktif" class="info warn">
      <i class="pi pi-exclamation-triangle" /> Stasiun cetak tidak terhubung. Pastikan HP/PC stasiun menyala dan ada internet.
      <template v-if="hp && !hpPunyaToken()"> Minta owner mengaktifkan stasiun dari halaman Printer.</template>
      <button v-if="hp && hpPunyaToken()" class="pill warn" @click="nyalakanHp"><i class="pi pi-power-off" /> Nyalakan stasiun di HP ini</button>
    </p>
    <p v-else-if="meja.length" class="info">Stasiun mencoba lagi tiap 15 detik. Tiket keluar sendiri begitu masalah printernya beres.</p>

    <div v-if="meja.length" class="grid">
      <button v-for="m in meja" :key="m.nama" class="meja" :class="{ lain: m.nama === TANPA_MEJA }" @click="pilih(m)">
        <span class="badge">{{ m.tiket.length }}</span>
        <strong>{{ m.nama }}</strong>
        <small>{{ printer(m) }}</small>
      </button>
    </div>
    <section v-else class="panel kosong"><i class="pi pi-check-circle" /> Semua tiket sudah tercetak.</section>

    <small v-if="store.cetak.macet > (store.cetak.daftar?.length ?? 0)" class="redup">
      Ditampilkan {{ store.cetak.daftar.length }} dari {{ store.cetak.macet }} tiket. Sisanya muncul setelah yang ini beres.
    </small>

    <CetakMacetDialog v-model:visible="buka" :judul="dipilih" :daftar="tiketDipilih" />
  </div>
</template>

<style scoped>
.gap {
  gap: 14px;
}
.info {
  margin: 0;
  color: var(--p-text-muted-color);
}
.info.warn {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  color: var(--p-red-600, #dc2626);
  font-weight: 600;
}
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 12px;
}
.meja {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  min-height: 96px;
  padding: 14px 10px;
  border: 2px solid var(--p-red-400, #f87171);
  border-radius: 14px;
  background: color-mix(in srgb, var(--p-red-500, #ef4444) 8%, var(--p-content-background));
  font: inherit;
  color: inherit;
  cursor: pointer;
}
.meja.lain {
  border-style: dashed;
}
.meja strong {
  font-size: 1.15rem;
  font-weight: 800;
}
.meja small,
.redup {
  color: var(--p-text-muted-color);
  text-align: center;
}
.badge {
  position: absolute;
  top: -8px;
  right: -6px;
  min-width: 24px;
  padding: 2px 7px;
  border-radius: 999px;
  font-size: 0.8rem;
  font-weight: 700;
  background: var(--p-red-600, #dc2626);
  color: #fff;
}
.kosong {
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--p-text-muted-color);
}
.kosong i {
  color: var(--p-green-500, #22c55e);
  font-size: 1.3rem;
}
</style>
