<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { get } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { usePengaturanStore } from '@/stores/pengaturan'
import { cetakShift } from '@/utils/cetak'
import { unduhExcel } from '@/utils/ekspor'
import { labelKategori } from '@/utils/kategori'
import { hari, rp, tgl, ymd } from '@/utils/format'

const PRESET = [
  { label: '7 hari', range: () => [hari(-6), hari()] },
  { label: 'Bulan ini', range: () => [new Date(new Date().getFullYear(), new Date().getMonth(), 1), hari()] },
  { label: '3 bulan', range: () => [hari(-89), hari()] },
]
const PER_HALAMAN = 20

const run = useRun()
const peng = usePengaturanStore()
const rentang = ref([hari(-29), hari()])
const kasir = ref(null)
const hanyaSelisih = ref(false)
const page = ref(0)
const data = ref({ items: [], total: 0, totalSelisih: 0 })
const users = ref([])
const loading = ref(false)
const detail = ref(null)
const detailBuka = ref(false)

const params = (p = page.value) => {
  const [dari, sampai] = rentang.value
  const q = { dari: ymd(dari), sampai: ymd(sampai || dari), page: p }
  if (kasir.value) q.kasir = kasir.value
  if (hanyaSelisih.value) q.selisih = true
  return q
}

async function muat() {
  if (!rentang.value[0]) return
  loading.value = true
  await run(async () => (data.value = await get('/shift', { params: params() })))
  loading.value = false
}

watch([rentang, kasir, hanyaSelisih], () => {
  if (rentang.value[1] === null) return
  page.value = 0
  muat()
})

async function lihat(s) {
  detail.value = null
  detailBuka.value = true
  await run(async () => (detail.value = await get(`/shift/${s.id}`)))
}

const durasi = (s) => {
  if (!s.tutupAt) return 'berjalan'
  const menit = Math.round((new Date(s.tutupAt) - new Date(s.bukaAt)) / 60000)
  return `${Math.floor(menit / 60)} j ${menit % 60} m`
}
const tagSelisih = (s) => (s.selisih == null ? null : s.selisih === 0 ? 'success' : s.selisih > 0 ? 'warn' : 'danger')
const kasirOpsi = computed(() => users.value.filter((u) => u.role === 'kasir' || u.role === 'owner'))

const cetak = () => run(async () => (await cetakShift(detail.value, peng.settings)) && null, 'Laporan shift dicetak')

// Excel memuat semua halaman pada rentang dan filter yang sama
async function excel() {
  await run(async () => {
    const semua = []
    for (let p = 0; p * PER_HALAMAN < data.value.total; p++) semua.push(...(await get('/shift', { params: params(p) })).items)
    const [dari, sampai] = rentang.value
    await unduhExcel(`shift-${ymd(dari)}_${ymd(sampai || dari)}.xlsx`, [
      {
        nama: 'Shift',
        kolom: [
          { header: 'No.', key: 'id' }, { header: 'Dibuka oleh', key: 'kasir', lebar: 18 }, { header: 'Buka', key: 'buka', lebar: 20 },
          { header: 'Ditutup oleh', key: 'tutupOleh', lebar: 18 }, { header: 'Tutup', key: 'tutup', lebar: 20 }, { header: 'Bill', key: 'bill' },
          { header: 'Penjualan', key: 'penjualan', uang: true, lebar: 16 }, { header: 'Tunai masuk', key: 'tunai', uang: true, lebar: 16 },
          { header: 'Pengeluaran', key: 'keluar', uang: true, lebar: 16 }, { header: 'Kas sistem', key: 'kasSistem', uang: true, lebar: 16 },
          { header: 'Kas dihitung', key: 'kasHitung', uang: true, lebar: 16 }, { header: 'Selisih', key: 'selisih', uang: true, lebar: 14 },
        ],
        baris: semua.map((s) => ({ ...s, buka: tgl(s.bukaAt), tutup: s.tutupAt ? tgl(s.tutupAt) : 'berjalan' })),
      },
    ])
  }, 'Excel diunduh')
}

onMounted(() => {
  muat()
  run(async () => (users.value = await get('/users')))
})
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-clock" title="Riwayat Shift" subtitle="Buka-tutup kasir, penjualan, dan selisih kas per shift">
      <button v-for="p in PRESET" :key="p.label" class="pill" @click="rentang = p.range()">{{ p.label }}</button>
      <DatePicker v-model="rentang" selection-mode="range" date-format="dd/mm/yy" :manual-input="false" show-icon placeholder="Pilih tanggal" />
      <Button label="Excel" icon="pi pi-file-excel" severity="secondary" outlined :disabled="!data.total" @click="excel" />
    </PageHeader>

    <div class="stats">
      <StatCard icon="pi pi-clock" label="Jumlah shift" :value="String(data.total)" />
      <StatCard icon="pi pi-exclamation-circle" label="Total selisih kas" :value="rp(data.totalSelisih)" />
    </div>

    <section class="panel box">
      <div class="filters">
        <Select v-model="kasir" :options="kasirOpsi" option-label="nama" option-value="id" placeholder="Semua kasir" show-clear class="kasir" />
        <label class="cek"><Checkbox v-model="hanyaSelisih" binary input-id="selisih" /><span>Hanya yang ada selisih kas</span></label>
      </div>
      <DataTable
        :value="data.items"
        :loading="loading"
        size="small"
        lazy
        paginator
        :rows="PER_HALAMAN"
        :total-records="data.total"
        :first="page * PER_HALAMAN"
        :always-show-paginator="false"
        scrollable
        empty-message="Tidak ada shift pada rentang ini"
        @page="(e) => ((page = e.page), muat())"
      >
        <Column header="Buka">
          <template #body="{ data: s }">{{ tgl(s.bukaAt) }}<small class="redup"><br />{{ s.kasir }}</small></template>
        </Column>
        <Column header="Tutup">
          <template #body="{ data: s }">
            <template v-if="s.tutupAt">{{ tgl(s.tutupAt) }}<small class="redup"><br />{{ s.tutupOleh }} · {{ durasi(s) }}</small></template>
            <Tag v-else value="Berjalan" severity="info" />
          </template>
        </Column>
        <Column field="bill" header="Bill" />
        <Column header="Penjualan"><template #body="{ data: s }">{{ rp(s.penjualan) }}</template></Column>
        <Column header="Tunai masuk"><template #body="{ data: s }">{{ rp(s.tunai) }}</template></Column>
        <Column header="Pengeluaran"><template #body="{ data: s }">{{ s.keluar ? rp(s.keluar) : '—' }}</template></Column>
        <Column header="Kas sistem / dihitung">
          <template #body="{ data: s }">
            <template v-if="s.kasSistem != null">{{ rp(s.kasSistem) }}<small class="redup"><br />{{ s.kasHitung != null ? rp(s.kasHitung) : '—' }}</small></template>
            <span v-else class="redup">—</span>
          </template>
        </Column>
        <Column header="Selisih">
          <template #body="{ data: s }"><Tag v-if="tagSelisih(s)" :value="rp(s.selisih)" :severity="tagSelisih(s)" /></template>
        </Column>
        <Column>
          <template #body="{ data: s }"><Button icon="pi pi-eye" text severity="secondary" aria-label="Detail" @click="lihat(s)" /></template>
        </Column>
      </DataTable>
    </section>

    <Dialog v-model:visible="detailBuka" modal :header="detail ? `Shift #${detail.id}` : 'Shift'" :style="{ width: '32rem' }" :breakpoints="{ '600px': '96vw' }">
      <div v-if="detail" class="stack">
        <div class="row"><span>Buka</span><span>{{ tgl(detail.bukaAt) }} · {{ detail.kasir }}</span></div>
        <div class="row">
          <span>Tutup</span>
          <span v-if="detail.tutupAt">{{ tgl(detail.tutupAt) }} · {{ detail.tutupOleh }} ({{ durasi(detail) }})</span>
          <Tag v-else value="Masih berjalan" severity="info" />
        </div>
        <hr />
        <div class="row"><span>Bill lunas</span><strong>{{ detail.bill }}</strong></div>
        <div v-if="detail.billBatal" class="row"><span>Bill batal / refund</span><span>{{ detail.billBatal }}</span></div>
        <div v-for="m in detail.perMetode" :key="m.metode" class="row sub"><span class="cap">{{ m.metode }}</span><span>{{ rp(m.total) }}</span></div>
        <div class="row"><span>Penjualan</span><strong>{{ rp(detail.penjualan) }}</strong></div>
        <small v-if="detail.diskon" class="redup">Sudah dipotong diskon {{ rp(detail.diskon) }}</small>
        <hr />
        <strong>Kas laci</strong>
        <div class="row"><span>Kas awal</span><span>{{ rp(detail.kasAwal) }}</span></div>
        <div class="row"><span>+ Penjualan tunai</span><span>{{ rp(detail.penjualanTunai) }}</span></div>
        <div v-if="detail.kasbonTunai" class="row"><span>+ Cicilan kasbon tunai</span><span>{{ rp(detail.kasbonTunai) }}</span></div>
        <div class="row"><span>− Pengeluaran</span><span>{{ rp(detail.keluar) }}</span></div>
        <div v-for="e in detail.pengeluaran" :key="e.id" class="row sub">
          <span>{{ e.keterangan }} <small class="redup">· {{ labelKategori(e.kategori) }} · {{ e.oleh }}</small></span><span>{{ rp(e.jumlah) }}</span>
        </div>
        <div v-if="detail.kasSistem != null" class="row total"><span>Kas sistem</span><span>{{ rp(detail.kasSistem) }}</span></div>
        <div v-if="detail.kasHitung != null" class="row"><span>Kas dihitung</span><span>{{ rp(detail.kasHitung) }}</span></div>
        <div v-if="detail.selisih != null" class="row total">
          <span>Selisih</span><Tag :value="rp(detail.selisih)" :severity="tagSelisih(detail)" />
        </div>
        <template v-if="detail.kasbon.length">
          <hr />
          <strong>Cicilan kasbon</strong>
          <div v-for="(k, i) in detail.kasbon" :key="i" class="row sub">
            <span>{{ k.pelanggan }} <small class="redup">· {{ k.metode }} · {{ tgl(k.at) }}</small></span><span>{{ rp(k.jumlah) }}</span>
          </div>
        </template>
      </div>
      <template #footer>
        <Button label="Cetak laporan shift" icon="pi pi-print" severity="secondary" outlined :disabled="!detail" @click="cetak" />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.gap {
  gap: 14px;
}
.stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 14px;
}
.box {
  padding: 14px;
}
.filters {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 14px;
  margin-bottom: 10px;
}
.kasir {
  min-width: 200px;
}
.cek {
  display: flex;
  align-items: center;
  gap: 8px;
}
.redup {
  color: var(--p-text-muted-color);
}
.row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}
.row.sub {
  padding-left: 12px;
  font-size: 0.92rem;
}
.row.total {
  font-weight: 700;
}
.cap {
  text-transform: capitalize;
}
hr {
  border: 0;
  border-top: 1px dashed var(--p-content-border-color);
  margin: 4px 0;
  width: 100%;
}
</style>
