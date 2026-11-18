<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { get, post } from '@/api/http'
import { onEvent } from '@/composables/useEvents'
import { useRun } from '@/composables/useRun'
import { hari, tgl, ymd } from '@/utils/format'

const run = useRun()
const data = ref({ tenant: '', items: [] })

// kartu yang baru masuk (3 menit terakhir) diberi sorotan; now diperbarui supaya sorotannya hilang sendiri
const BARU_MS = 3 * 60 * 1000
const now = ref(Date.now())
const baru = (b) => now.value - new Date(b.at).getTime() < BARU_MS
let detak

// satu kartu per kali Kirim, sama seperti satu tiket dapur di printer
function kelompok(items) {
  const map = new Map()
  for (const it of items) {
    const key = `${it.pesananId}|${it.dikirimAt}`
    if (!map.has(key)) map.set(key, { key, id: it.pesananId, meja: it.mejaNama, atasNama: it.atasNama, at: it.dikirimAt, oleh: it.oleh, tambahan: it.tambahan, items: [] })
    map.get(key).items.push(it)
  }
  return [...map.values()]
}
// kartu panjang ditutup dulu: antrean menampilkan 3 menu, selesai 2 menu, sisanya lewat "Lihat semua"
const BATAS = { antrean: 3, selesai: 2 }
const terbuka = reactive({})
const tampil = (b, jenis) => (terbuka[b.key] ? b.items : b.items.slice(0, BATAS[jenis]))
const sisa = (b, jenis) => b.items.length - BATAS[jenis]

const antrean = computed(() => kelompok(data.value.items.filter((i) => !i.siapAt)))
const siap = computed(() => kelompok(data.value.items.filter((i) => i.siapAt)))

// riwayat selesai per hari, dipaginasi per kartu
const PER_HALAMAN = 12
const tanggal = ref(hari())
const first = ref(0)
const siapHalaman = computed(() => siap.value.slice(first.value, first.value + PER_HALAMAN))
const hariIni = computed(() => ymd(tanggal.value) === ymd(new Date()))

// Saat halaman terbuka, notifikasi Android tidak muncul; bunyikan chime yang sama untuk kartu antrean yang baru tiba.
const suara = new Audio('/pesanan.wav')
let dikenal = null
async function muat() {
  data.value = await get('/tenant-saya/pesanan', { params: { tanggal: ymd(tanggal.value) } })
  const kunci = new Set(antrean.value.map((b) => b.key))
  if (dikenal && [...kunci].some((k) => !dikenal.has(k))) suara.play().catch(() => {})
  dikenal = kunci
}
const tandai = (it, s) => run(async () => {
  await post(`/tenant-saya/item/${it.id}/siap`, { siap: s })
  await muat()
})

const offs = []
onMounted(() => {
  detak = setInterval(() => (now.value = Date.now()), 15000)
  run(muat)
  offs.push(onEvent('pesanan', muat), onEvent('dapur', muat))
})
watch(tanggal, (t) => {
  if (!t) return (tanggal.value = hari())
  first.value = 0
  run(muat)
})
onBeforeUnmount(() => {
  clearInterval(detak)
  offs.forEach((off) => off())
})
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-inbox" title="Pesanan Masuk" :subtitle="data.tenant" />

    <h3 class="judul">Antrean <Tag :value="String(antrean.reduce((n, b) => n + b.items.length, 0))" /></h3>
    <p v-if="!antrean.length" class="kosong">Belum ada pesanan yang perlu dibuat.</p>
    <div class="grid">
      <section v-for="b in antrean" :key="b.key" class="panel kartu antre" :class="{ tambah: b.tambahan, baru: baru(b) }">
        <div class="jenis">
          <span v-if="baru(b)" class="lencana">BARU</span>
          {{ b.tambahan ? 'Pesanan tambahan' : 'Pesanan baru' }}
        </div>
        <header>
          <strong class="meja">{{ b.meja || 'Tanpa meja' }}</strong>
          <small class="redup kanan">#{{ b.id }}</small>
        </header>
        <div class="info">
          <span v-if="b.atasNama"><i class="pi pi-user" /> {{ b.atasNama }}</span>
          <span><i class="pi pi-send" /> Dikirim {{ tgl(b.at) }}<template v-if="b.oleh"> oleh <b>{{ b.oleh }}</b></template></span>
        </div>
        <div v-for="it in tampil(b, 'antrean')" :key="it.id" class="baris">
          <div>
            <b>{{ it.qty }}×</b> {{ it.nama }}
            <small v-if="it.opsi" class="redup"> · {{ it.opsi }}</small>
            <div v-if="it.catatan" class="catatan"><i class="pi pi-comment" /> {{ it.catatan }}</div>
          </div>
          <Button label="Selesai" icon="pi pi-check" size="small" @click="tandai(it, true)" />
        </div>
        <button v-if="sisa(b, 'antrean') > 0" class="lihat" @click="terbuka[b.key] = !terbuka[b.key]">
          <i :class="terbuka[b.key] ? 'pi pi-chevron-up' : 'pi pi-chevron-down'" />
          {{ terbuka[b.key] ? 'Sembunyikan' : `Lihat semua (${sisa(b, 'antrean')} lagi)` }}
        </button>
      </section>
    </div>

    <div class="judul selesai-judul">
      <h3>Selesai <Tag :value="String(siap.reduce((n, b) => n + b.items.length, 0))" severity="secondary" /></h3>
      <DatePicker v-model="tanggal" date-format="dd/mm/yy" :max-date="new Date()" :manual-input="false" show-icon class="kanan" />
    </div>
    <p v-if="!siap.length" class="kosong">{{ hariIni ? 'Belum ada pesanan yang selesai hari ini.' : 'Tidak ada pesanan selesai di tanggal ini.' }}</p>
    <template v-else>
      <div class="grid">
        <section v-for="b in siapHalaman" :key="b.key" class="panel kartu selesai">
          <header>
            <strong class="meja">{{ b.meja || 'Tanpa meja' }}</strong>
            <small class="redup kanan">#{{ b.id }}</small>
          </header>
          <div class="info">
            <span v-if="b.atasNama"><i class="pi pi-user" /> {{ b.atasNama }}</span>
            <span><i class="pi pi-send" /> Dikirim {{ tgl(b.at) }}<template v-if="b.oleh"> oleh <b>{{ b.oleh }}</b></template></span>
          </div>
          <div v-for="it in tampil(b, 'selesai')" :key="it.id" class="baris">
            <div>
              <b>{{ it.qty }}×</b> {{ it.nama }}
              <div class="selesai-at"><i class="pi pi-check-circle" /> Selesai {{ tgl(it.siapAt) }}</div>
            </div>
            <Button label="Batal" icon="pi pi-undo" size="small" text severity="secondary" @click="tandai(it, false)" />
          </div>
          <button v-if="sisa(b, 'selesai') > 0" class="lihat" @click="terbuka[b.key] = !terbuka[b.key]">
            <i :class="terbuka[b.key] ? 'pi pi-chevron-up' : 'pi pi-chevron-down'" />
            {{ terbuka[b.key] ? 'Sembunyikan' : `Lihat semua (${sisa(b, 'selesai')} lagi)` }}
          </button>
        </section>
      </div>
      <Paginator v-if="siap.length > PER_HALAMAN" v-model:first="first" :rows="PER_HALAMAN" :total-records="siap.length" />
    </template>
  </div>
</template>

<style scoped>
.gap {
  gap: 14px;
}
.judul {
  margin: 4px 0 0;
  display: flex;
  align-items: center;
  gap: 8px;
}
.selesai-judul {
  margin-top: 10px;
  flex-wrap: wrap;
}
.selesai-judul h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}
.kosong,
.redup {
  color: var(--p-text-muted-color);
}
.grid {
  display: grid;
  align-items: start; /* kartu setinggi isinya sendiri, tidak ikut kartu terpanjang */
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
}
.kartu {
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.kartu header {
  display: flex;
  align-items: baseline;
  gap: 8px;
  flex-wrap: wrap;
}
.kanan {
  margin-left: auto;
}
/* antrean: berwarna dan tegas, supaya beda jelas dengan yang sudah selesai */
.antre {
  --warna: var(--p-primary-color);
  border: 2px solid var(--warna);
  border-top-width: 6px;
  background: color-mix(in srgb, var(--warna) 10%, var(--p-content-background));
}
.antre.tambah {
  --warna: var(--p-primary-700);
}
.antre.baru {
  animation: denyut 1.6s ease-in-out infinite;
}
@keyframes denyut {
  0%,
  100% {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--warna) 55%, transparent);
  }
  50% {
    box-shadow: 0 0 0 8px color-mix(in srgb, var(--warna) 0%, transparent);
  }
}
@media (prefers-reduced-motion: reduce) {
  .antre.baru {
    animation: none;
    box-shadow: 0 0 0 4px color-mix(in srgb, var(--warna) 40%, transparent);
  }
}
.lencana {
  display: inline-block;
  margin-right: 6px;
  padding: 1px 8px;
  border-radius: 999px;
  background: var(--warna);
  color: var(--p-primary-contrast-color);
  letter-spacing: 0.08em;
}
.jenis {
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--accent-text);
}
.meja {
  font-size: 1.2rem;
}
.info {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 14px;
  font-size: 0.9rem;
  color: var(--p-text-muted-color);
}
.baris {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding-top: 8px;
  border-top: 1px solid var(--p-content-border-color);
}
.catatan {
  margin-top: 2px;
  font-size: 0.9rem;
  color: var(--accent-text);
}
/* selesai: warna tema yang pudar dan tenang, kontras dengan antrean yang tegas */
.selesai {
  border-left: 5px solid var(--p-primary-200);
  background: color-mix(in srgb, var(--p-primary-color) 4%, var(--p-content-background));
  opacity: 0.85;
}
.lihat {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 6px;
  border: 0;
  border-top: 1px dashed var(--p-content-border-color);
  background: none;
  font: inherit;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--accent-text);
  cursor: pointer;
}
.selesai-at {
  margin-top: 2px;
  font-size: 0.85rem;
  color: var(--accent-text);
}
</style>
