<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { get, put } from '@/api/http'
import { onEvent } from '@/composables/useEvents'
import { useArea } from '@/composables/useArea'
import { useRun } from '@/composables/useRun'
import { useAuthStore } from '@/stores/auth'
import { useKasirStore } from '@/stores/kasir'
import { rp } from '@/utils/format'

// Posisi dalam satuan 1% lebar kanvas, sehingga denah ikut membesar/mengecil seperti gambar.
const W = 9
const H = 7
const KOLOM = 9
const MAX_Y = 400

const router = useRouter()
const auth = useAuthStore()
const kasir = useKasirStore()
const run = useRun()

const denah = ref({ area: [], meja: [] })
// tiap area punya denah sendiri: hanya meja area yang dipilih yang tampil dan diatur
const { tabs, aktif, pilih, cocok } = useArea(denah, { kunci: 'denah.area' })
const meja = computed(() => denah.value.meja.filter(cocok))
const posisi = reactive({})
const edit = ref(false)
const kanvas = ref()
let seret = null

const clamp = (v, min, max) => Math.min(max, Math.max(min, v))
const snap = (v) => Math.round(v * 2) / 2
const auto = (i) => ({ x: 1 + (i % KOLOM) * (W + 1.8), y: 1 + Math.floor(i / KOLOM) * (H + 1.5) })
const total = (m) => m.pesanan.reduce((s, p) => s + p.total, 0)
const tinggi = computed(() => Math.max(52, ...meja.value.map((m) => (posisi[m.id]?.y ?? 0) + H + 2)))

function sinkron() {
  meja.value.forEach((m, i) => {
    if (edit.value && posisi[m.id]) return
    posisi[m.id] = { x: m.posX ?? auto(i).x, y: m.posY ?? auto(i).y }
  })
}

async function muat() {
  denah.value = await get('/meja')
  sinkron()
}

function buka(m) {
  if (edit.value) return
  run(async () => {
    await kasir.bukaMeja(m)
    router.push('/')
  })
}

function mulaiSeret(e, m) {
  if (!edit.value) return
  const r = kanvas.value.getBoundingClientRect()
  const p = posisi[m.id]
  seret = {
    id: m.id,
    r,
    dx: e.clientX - (r.left + (p.x / 100) * r.width),
    dy: e.clientY - (r.top + (p.y / 100) * r.width),
  }
  e.currentTarget.setPointerCapture(e.pointerId)
}

function geser(e) {
  if (!seret) return
  const { r, dx, dy, id } = seret
  posisi[id] = {
    x: snap(clamp(((e.clientX - dx - r.left) / r.width) * 100, 0, 100 - W)),
    y: snap(clamp(((e.clientY - dy - r.top) / r.width) * 100, 0, MAX_Y - H)),
  }
}

const selesaiSeret = () => (seret = null)
const susun = () => meja.value.forEach((m, i) => (posisi[m.id] = auto(i)))

async function simpan() {
  const body = { meja: meja.value.map((m) => ({ id: m.id, posX: posisi[m.id].x, posY: posisi[m.id].y })) }
  if (await run(() => put('/meja/posisi', body), 'Denah disimpan')) {
    edit.value = false
    await run(muat)
  }
}

function batalEdit() {
  edit.value = false
  run(muat)
}

watch(aktif, sinkron)

const offs = []
onMounted(() => {
  run(muat)
  const diam = () => muat().catch(() => {})
  offs.push(onEvent('meja', diam), onEvent('pesanan', diam))
})
onBeforeUnmount(() => offs.forEach((off) => off()))
</script>

<template>
  <div class="stack gap">
    <PageHeader
      icon="pi pi-map"
      title="Denah Meja"
      :subtitle="edit ? 'Seret meja untuk mengatur letaknya' : 'Ketuk meja untuk mulai memesan'"
    >
      <template v-if="auth.isOwner">
        <button v-if="!edit" class="pill" @click="edit = true"><i class="pi pi-pencil" /> Atur denah</button>
        <template v-else>
          <button class="pill" @click="susun"><i class="pi pi-th-large" /> Susun otomatis</button>
          <button class="pill" @click="batalEdit">Batal</button>
          <button class="pill" @click="simpan"><i class="pi pi-check" /> Simpan</button>
        </template>
      </template>
    </PageHeader>

    <SelectButton
      v-if="tabs.length"
      :model-value="aktif"
      :options="tabs"
      option-label="nama"
      option-value="id"
      :allow-empty="false"
      :disabled="edit"
      class="area"
      @update:model-value="pilih = $event"
    />

    <div class="legend">
      <span><i class="sw" /> Kosong</span>
      <span><i class="sw terisi" /> Terisi</span>
    </div>

    <section class="panel scroll">
      <div class="wrap">
        <div ref="kanvas" class="kanvas" :class="{ edit }" :style="{ '--h': tinggi }">
          <button
            v-for="m in meja"
            :key="m.id"
            class="meja"
            :class="{ terisi: m.pesanan.length }"
            :style="{ '--x': posisi[m.id]?.x, '--y': posisi[m.id]?.y }"
            @click="buka(m)"
            @pointerdown="mulaiSeret($event, m)"
            @pointermove="geser"
            @pointerup="selesaiSeret"
            @pointercancel="selesaiSeret"
          >
            <strong>{{ m.nama }}</strong>
            <small>{{ m.pesanan.length ? rp(total(m)) : 'Kosong' }}</small>
            <span v-if="m.pesanan.length > 1" class="bill">{{ m.pesanan.length }} bill</span>
          </button>
        </div>
      </div>
      <p v-if="!meja.length" class="none">Belum ada meja{{ tabs.length ? ' di area ini' : '' }}. Tambahkan di Master Data › Meja.</p>
    </section>
  </div>
</template>

<style scoped>
.gap {
  gap: 14px;
}
.area {
  align-self: flex-start;
  flex-wrap: wrap;
}
.legend {
  display: flex;
  gap: 16px;
  color: var(--p-text-muted-color);
}
.legend span {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
.sw {
  width: 16px;
  height: 16px;
  border-radius: 5px;
  border: 1.5px solid var(--p-content-border-color);
  background: var(--p-content-background);
}
.sw.terisi {
  border-color: var(--p-primary-color);
  background: color-mix(in srgb, var(--p-primary-color) 22%, var(--p-content-background));
}
.scroll {
  padding: 12px;
  overflow-x: auto;
}
.wrap {
  container-type: inline-size;
  min-width: 860px;
}
.kanvas {
  position: relative;
  height: calc(var(--h) * 1cqw);
  border-radius: 12px;
  background: color-mix(in srgb, var(--p-primary-color) 4%, var(--p-content-background));
}
.kanvas.edit {
  background-image: radial-gradient(circle, var(--p-content-border-color) 1px, transparent 1px);
  background-size: 2cqw 2cqw;
}
.meja {
  position: absolute;
  left: calc(var(--x) * 1cqw);
  top: calc(var(--y) * 1cqw);
  width: 9cqw;
  height: 7cqw;
  padding: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  border: 1.5px solid var(--p-content-border-color);
  border-radius: 12px;
  background: var(--p-content-background);
  color: var(--p-text-color);
  font: inherit;
  cursor: pointer;
}
.meja strong {
  font-size: max(12px, 1.5cqw);
}
.meja small {
  font-size: max(10px, 1.15cqw);
  color: var(--p-text-muted-color);
}
.meja.terisi {
  border-color: var(--p-primary-color);
  background: color-mix(in srgb, var(--p-primary-color) 18%, var(--p-content-background));
}
.meja.terisi small {
  color: var(--accent-text);
  font-weight: 700;
}
.bill {
  position: absolute;
  top: -8px;
  right: -6px;
  padding: 1px 7px;
  border-radius: 999px;
  font-size: 0.7rem;
  font-weight: 700;
  background: var(--p-primary-color);
  color: var(--p-primary-contrast-color);
}
.kanvas.edit .meja {
  cursor: grab;
  border-style: dashed;
  touch-action: none;
  user-select: none;
}
.none {
  margin: 8px;
  color: var(--p-text-muted-color);
}
</style>
