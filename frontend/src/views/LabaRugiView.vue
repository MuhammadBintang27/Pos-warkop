<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { get } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { hari, rp, ymd } from '@/utils/format'
import { labelKategori } from '@/utils/kategori'

const MODE = [
  { label: 'Ringkasan', value: 'ringkasan' },
  { label: 'Margin menu', value: 'margin' },
  { label: 'Pengeluaran', value: 'pengeluaran' },
]
const PRESET = [
  { label: 'Hari ini', range: () => [hari(), hari()] },
  { label: '7 hari', range: () => [hari(-6), hari()] },
  { label: 'Bulan ini', range: () => [new Date(new Date().getFullYear(), new Date().getMonth(), 1), hari()] },
]

const run = useRun()
const mode = ref('ringkasan')
const rentang = ref(PRESET[2].range())
const data = ref(null)
const margin = ref([])

const totalSemua = computed(() => (data.value?.pengeluaran ?? []).reduce((s, k) => s + k.total, 0))
const persen = (n) => (totalSemua.value ? Math.round((n / totalSemua.value) * 100) : 0)
const margins = computed(() => margin.value.map((m) => ({ ...m, marginTampil: m.adaResep ? m.marginPersen : null })))

async function muat() {
  const [dari, sampai] = rentang.value
  if (!dari) return
  const q = `dari=${ymd(dari)}&sampai=${ymd(sampai || dari)}`
  await run(async () => {
    ;[data.value, margin.value] = await Promise.all([get(`/laba-rugi?${q}`), get(`/margin-menu?${q}`)])
  })
}

watch(rentang, (r) => r[1] !== null && muat())
onMounted(muat)
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-chart-bar" title="Laba Rugi" :subtitle="data ? `${data.dari} s/d ${data.sampai}` : ''">
      <button v-for="p in PRESET" :key="p.label" class="pill" @click="rentang = p.range()">{{ p.label }}</button>
      <DatePicker v-model="rentang" selection-mode="range" date-format="dd/mm/yy" :manual-input="false" show-icon placeholder="Pilih tanggal" />
    </PageHeader>

    <template v-if="data">
      <div class="stats">
        <StatCard icon="pi pi-wallet" label="Pendapatan" :value="rp(data.pendapatan)" />
        <StatCard icon="pi pi-box" label="HPP" :value="rp(data.hpp)" />
        <StatCard icon="pi pi-chart-line" label="Laba kotor" :value="rp(data.labaKotor)" />
        <StatCard icon="pi pi-check-circle" label="Laba bersih" :value="rp(data.labaBersih)" />
      </div>

      <Message v-if="data.omzetTanpaResep > 0" severity="warn" size="small">
        Penjualan {{ rp(data.omzetTanpaResep) }} berasal dari menu yang belum punya resep atau harga modal, jadi HPP-nya belum terhitung dan laba kotor bisa terlalu tinggi. Isi resep atau harga modalnya di menu Menu.
      </Message>

      <SelectButton v-model="mode" :options="MODE" option-label="label" option-value="value" :allow-empty="false" />

      <section v-if="mode === 'ringkasan'" class="panel box">
        <div class="baris"><span>Penjualan sendiri (setelah diskon)</span><span>{{ rp(data.penjualanSendiri) }}</span></div>
        <div class="baris"><span>Komisi tenant</span><span>{{ rp(data.komisiTenant) }}</span></div>
        <div class="baris total"><span>Pendapatan</span><span>{{ rp(data.pendapatan) }}</span></div>
        <div class="baris"><span>HPP (bahan terpakai dan harga modal)</span><span>-{{ rp(data.hpp) }}</span></div>
        <div v-if="data.susut" class="baris"><span>Bahan susut (selisih opname)</span><span>{{ data.susut > 0 ? '-' : '+' }}{{ rp(Math.abs(data.susut)) }}</span></div>
        <div class="baris total"><span>Laba kotor</span><span>{{ rp(data.labaKotor) }}</span></div>
        <template v-for="k in data.pengeluaran.filter((x) => !x.dalamHpp)" :key="k.kategori">
          <div class="baris"><span>{{ labelKategori(k.kategori) }}</span><span>-{{ rp(k.total) }}</span></div>
        </template>
        <div class="baris"><span>Gaji karyawan (yang dibayar)</span><span>-{{ rp(data.gaji) }}</span></div>
        <div class="baris total akhir"><span>Laba bersih</span><span>{{ rp(data.labaBersih) }}</span></div>
        <small class="hint">
          Belanja bahan tidak dikurangkan lagi karena sudah masuk HPP lewat bahan yang terpakai. HPP memakai harga rata-rata bahan saat dipesan, atau harga modal untuk menu tanpa resep. Bahan susut adalah stok yang hilang saat opname, dinilai dengan harga rata-rata saat itu.
          Gaji dihitung dari slip yang ditandai dibayar pada rentang ini.
        </small>
      </section>

      <section v-else-if="mode === 'margin'" class="panel box">
        <DataTable :value="margins" size="small" paginator :rows="15" empty-message="Belum ada penjualan pada rentang ini">
          <Column field="nama" header="Menu" sortable />
          <Column field="terjual" header="Terjual" sortable />
          <Column header="Penjualan" field="penjualan" sortable><template #body="{ data: d }">{{ rp(d.penjualan) }}</template></Column>
          <Column header="HPP"><template #body="{ data: d }">{{ d.adaResep ? rp(d.hpp) : '—' }}</template></Column>
          <Column header="Laba" field="laba" sortable><template #body="{ data: d }">{{ d.adaResep ? rp(d.laba) : '—' }}</template></Column>
          <Column header="Margin" field="marginTampil" sortable>
            <template #body="{ data: d }">
              <Tag v-if="!d.adaResep" value="Tanpa resep" severity="secondary" />
              <Tag v-else :value="`${d.marginPersen}%`" :severity="d.marginPersen >= 50 ? 'success' : d.marginPersen >= 30 ? 'warn' : 'danger'" />
            </template>
          </Column>
        </DataTable>
        <small class="hint">Penjualan sebelum diskon bill. Margin di bawah 30% ditandai merah: periksa harga jual atau takaran resepnya.</small>
      </section>

      <section v-else class="panel box">
        <p v-if="!data.pengeluaran.length" class="hint">Belum ada pengeluaran pada rentang ini.</p>
        <div v-for="k in data.pengeluaran" :key="k.kategori" class="kat">
          <div class="baris">
            <span>{{ labelKategori(k.kategori) }} <Tag v-if="k.dalamHpp" value="masuk HPP" severity="secondary" /></span>
            <span>{{ rp(k.total) }} · {{ persen(k.total) }}%</span>
          </div>
          <div class="bar"><div :style="{ width: persen(k.total) + '%' }" /></div>
        </div>
      </section>
    </template>
  </div>
</template>

<style scoped>
.gap {
  gap: 14px;
}
.stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
}
.box {
  padding: 14px;
}
.baris {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 7px 0;
  border-bottom: 1px solid var(--p-content-border-color);
}
.baris.total {
  font-weight: 700;
}
.baris.akhir {
  border-bottom: 0;
  font-size: 1.1em;
}
.kat {
  margin-bottom: 8px;
}
.kat .baris {
  border-bottom: 0;
  padding-bottom: 3px;
}
.bar {
  height: 8px;
  border-radius: 6px;
  background: var(--p-content-border-color);
  overflow: hidden;
}
.bar div {
  height: 100%;
  background: var(--p-primary-color);
}
.hint {
  display: block;
  margin-top: 10px;
  color: var(--p-text-muted-color);
}
</style>
