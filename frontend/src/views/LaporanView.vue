<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { get } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { hari, rp, tgl, ymd } from '@/utils/format'

const run = useRun()

const PRESET = [
  { label: 'Hari ini', range: () => [hari(), hari()] },
  { label: '7 hari', range: () => [hari(-6), hari()] },
  { label: 'Bulan ini', range: () => [new Date(new Date().getFullYear(), new Date().getMonth(), 1), hari()] },
]

const rentang = ref([hari(), hari()])
const data = ref(null)
const setoran = ref(null)
const totalSetor = computed(() => setoran.value?.tenant.reduce((s, t) => s + t.setor, 0) ?? 0)

const maxHari = computed(() => Math.max(1, ...(data.value?.perHari.map((h) => h.total) ?? [])))
const maxMetode = computed(() => Math.max(1, ...(data.value?.perMetode.map((m) => m.total) ?? [])))
const rata = computed(() => (data.value?.transaksi ? Math.round(data.value.omzet / data.value.transaksi) : 0))

async function muat() {
  const [dari, sampai] = rentang.value
  if (!dari) return
  const query = `dari=${ymd(dari)}&sampai=${ymd(sampai || dari)}`
  await run(async () => {
    ;[data.value, setoran.value] = await Promise.all([get(`/laporan?${query}`), get(`/laporan/tenant?${query}`)])
  })
}

watch(rentang, (r) => r[1] !== null && muat())
onMounted(muat)
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-chart-line" title="Laporan" :subtitle="data ? `${data.dari} s/d ${data.sampai}` : ''">
      <button v-for="p in PRESET" :key="p.label" class="pill" @click="rentang = p.range()">{{ p.label }}</button>
      <DatePicker v-model="rentang" selection-mode="range" date-format="dd/mm/yy" :manual-input="false" show-icon placeholder="Pilih tanggal" />
    </PageHeader>

    <template v-if="data">
      <div class="stats">
        <StatCard icon="pi pi-wallet" label="Omzet" :value="rp(data.omzet)" />
        <StatCard icon="pi pi-receipt" label="Transaksi" :value="String(data.transaksi)" />
        <StatCard icon="pi pi-calculator" label="Rata-rata / transaksi" :value="rp(rata)" />
        <StatCard icon="pi pi-percentage" label="Diskon" :value="rp(data.diskon)" />
        <StatCard icon="pi pi-arrow-up-right" label="Pengeluaran" :value="rp(data.pengeluaran)" />
      </div>

      <div class="two">
        <section class="panel box">
          <h3>Per metode bayar</h3>
          <p v-if="!data.perMetode.length" class="none">Belum ada transaksi.</p>
          <div v-for="m in data.perMetode" :key="m.metode" class="bar-row">
            <span class="cap">{{ m.metode }}</span>
            <div class="bar"><i :style="{ width: (m.total / maxMetode) * 100 + '%' }" /></div>
            <strong>{{ rp(m.total) }}</strong>
          </div>
        </section>

        <section class="panel box">
          <h3>Per hari</h3>
          <p v-if="!data.perHari.length" class="none">Belum ada transaksi.</p>
          <div v-for="h in data.perHari" :key="h.tanggal" class="bar-row">
            <span>{{ h.tanggal.slice(5) }}</span>
            <div class="bar"><i :style="{ width: (h.total / maxHari) * 100 + '%' }" /></div>
            <strong>{{ rp(h.total) }}</strong>
          </div>
        </section>
      </div>

      <div class="two">
        <section class="panel box">
          <h3>Menu terlaris</h3>
          <DataTable :value="data.terlaris" size="small" empty-message="Belum ada penjualan">
            <Column header="#"><template #body="{ index }">{{ index + 1 }}</template></Column>
            <Column field="nama" header="Menu" />
            <Column field="qty" header="Terjual" />
            <Column header="Nilai"><template #body="{ data: r }">{{ rp(r.total) }}</template></Column>
          </DataTable>
        </section>

        <section class="panel box">
          <h3>Rekap shift</h3>
          <DataTable :value="data.shift" size="small" empty-message="Tidak ada shift">
            <Column field="kasir" header="Dibuka oleh" />
            <Column header="Buka"><template #body="{ data: s }">{{ tgl(s.bukaAt) }}</template></Column>
            <Column header="Ditutup oleh"><template #body="{ data: s }">{{ s.tutupOleh || '—' }}</template></Column>
            <Column header="Kas sistem"><template #body="{ data: s }">{{ s.kasSistem == null ? '—' : rp(s.kasSistem) }}</template></Column>
            <Column header="Selisih">
              <template #body="{ data: s }">
                <Tag v-if="s.tutupAt == null" value="Berjalan" severity="info" />
                <Tag v-else :value="rp(s.selisih)" :severity="s.selisih === 0 ? 'success' : 'danger'" />
              </template>
            </Column>
          </DataTable>
        </section>
      </div>

      <section v-if="setoran && setoran.tenant.length > 1" class="panel box">
        <h3>Setoran tenant</h3>
        <DataTable :value="setoran.tenant" size="small">
          <Column field="nama" header="Tenant" />
          <Column header="Penjualan"><template #body="{ data: t }">{{ rp(t.kotor) }}</template></Column>
          <Column header="Diskon"><template #body="{ data: t }">{{ rp(t.diskon) }}</template></Column>
          <Column header="Komisi">
            <template #body="{ data: t }">{{ rp(t.komisi) }} <small class="none">({{ t.komisiPersen }}%)</small></template>
          </Column>
          <Column header="Setor ke tenant"><template #body="{ data: t }"><strong>{{ rp(t.setor) }}</strong></template></Column>
        </DataTable>
        <p class="total-setor">Total setoran: <strong>{{ rp(totalSetor) }}</strong></p>
        <small class="none">
          Komisi dihitung dari penjualan setelah diskon. Diskon bill dibagi menurut nilai item tiap tenant.
          Bill kasbon sudah dihitung sebagai penjualan.
        </small>
      </section>
    </template>
  </div>
</template>

<style scoped>
.total-setor {
  text-align: right;
  margin: 8px 0;
}
.gap {
  gap: 14px;
}
.stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 14px;
}
.two {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(340px, 1fr));
  gap: 14px;
}
.box {
  padding: 14px;
}
h3 {
  margin: 0 0 12px;
}
.none {
  color: var(--p-text-muted-color);
}
.bar-row {
  display: grid;
  grid-template-columns: 70px 1fr auto;
  align-items: center;
  gap: 10px;
  padding: 5px 0;
}
.bar {
  height: 10px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--p-primary-color) 12%, transparent);
  overflow: hidden;
}
.bar i {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: var(--p-primary-color);
}
.cap {
  text-transform: capitalize;
}
</style>
