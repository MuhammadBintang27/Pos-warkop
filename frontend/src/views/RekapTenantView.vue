<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { get } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { hari, labelKomisi, rp, ymd } from '@/utils/format'

const STATUS = {
  belum: { label: 'Belum closing', severity: 'secondary' },
  closing: { label: 'Closing, belum disetor', severity: 'info' },
  disetor: { label: 'Disetor', severity: 'success' },
  berubah: { label: 'Berubah setelah closing', severity: 'danger' },
}

const run = useRun()
const rentang = ref([hari(-6), hari()])
const harian = ref({ baris: [], tertunda: [] })
const detail = ref(null)

const tanggalIndo = (s) => s.split('-').reverse().join('/')
const total = computed(() => ({
  kotor: harian.value.baris.reduce((s, b) => s + b.kotor, 0),
  komisi: harian.value.baris.reduce((s, b) => s + b.komisi, 0),
  setor: harian.value.baris.reduce((s, b) => s + b.setor, 0),
}))
const tertunda = computed(() => harian.value.tertunda?.[0] ?? null)

async function muat() {
  const [dari, sampai] = rentang.value
  if (!dari) return
  const params = { dari: ymd(dari), sampai: ymd(sampai || dari) }
  ;[harian.value, detail.value] = await Promise.all([
    get('/tenant-saya/laporan/harian', { params }),
    get('/tenant-saya/laporan/detail', { params }),
  ])
}

watch(rentang, (r) => r[1] !== null && run(muat))
onMounted(() => run(muat))
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-chart-bar" title="Rekapan" :subtitle="detail?.tenant">
      <DatePicker v-model="rentang" selection-mode="range" date-format="dd/mm/yy" :manual-input="false" show-icon placeholder="Pilih tanggal" />
    </PageHeader>

    <div class="stats">
      <StatCard icon="pi pi-wallet" label="Penjualan" :value="rp(total.kotor)" />
      <StatCard icon="pi pi-percentage" label="Komisi pemilik tempat" :value="rp(total.komisi)" />
      <StatCard icon="pi pi-send" label="Bagian Anda" :value="rp(total.setor)" />
    </div>
    <p v-if="tertunda" class="redup">
      <i class="pi pi-clock" /> {{ tertunda.bill }} bill belum dibayar pelanggan ({{ rp(tertunda.nilai) }}), belum masuk penjualan.
    </p>

    <section class="panel box">
      <h3>Per hari</h3>
      <DataTable paginator :rows="10" :rows-per-page-options="[10, 25, 50]" :always-show-paginator="false" :value="harian.baris" size="small" empty-message="Belum ada penjualan di rentang ini">
        <Column header="Tanggal"><template #body="{ data: b }">{{ tanggalIndo(b.tanggal) }}</template></Column>
        <Column field="transaksi" header="Bill" />
        <Column header="Penjualan"><template #body="{ data: b }">{{ rp(b.kotor) }}</template></Column>
        <Column header="Diskon"><template #body="{ data: b }">{{ rp(b.diskon) }}</template></Column>
        <Column header="Komisi"><template #body="{ data: b }">{{ rp(b.komisi) }} <small class="redup">({{ labelKomisi(b) }})</small></template></Column>
        <Column header="Bagian Anda"><template #body="{ data: b }">{{ rp(b.setor) }}</template></Column>
        <Column header="Status">
          <template #body="{ data: b }"><Tag :value="STATUS[b.status]?.label ?? b.status" :severity="STATUS[b.status]?.severity" /></template>
        </Column>
      </DataTable>
    </section>

    <section v-if="detail" class="panel box">
      <h3>Per menu</h3>
      <DataTable paginator :rows="10" :rows-per-page-options="[10, 25, 50]" :always-show-paginator="false" :value="detail.menu" size="small" empty-message="Belum ada menu terjual">
        <Column field="nama" header="Menu" />
        <Column field="qty" header="Terjual" />
        <Column header="Nilai"><template #body="{ data: m }">{{ rp(m.nilai) }}</template></Column>
        <Column header="Dibatalkan"><template #body="{ data: m }">{{ m.batalQty || '—' }}</template></Column>
      </DataTable>
    </section>
  </div>
</template>

<style scoped>
.gap {
  gap: 14px;
}
.box {
  padding: 14px;
}
.box h3 {
  margin: 0 0 10px;
}
.stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
}
.redup {
  color: var(--p-text-muted-color);
}
</style>
