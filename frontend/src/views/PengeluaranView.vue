<script setup>
import { onMounted, ref, watch } from 'vue'
import { useConfirm } from 'primevue/useconfirm'
import { del, get, put } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { unduhExcel } from '@/utils/ekspor'
import { hari, rp, tgl, ymd } from '@/utils/format'
import { konfirmasi } from '@/utils/confirm'
import { KATEGORI_PENGELUARAN, labelKategori } from '@/utils/kategori'

const awalBulan = () => new Date(new Date().getFullYear(), new Date().getMonth(), 1)
const PRESET = [
  { label: 'Hari ini', range: () => [hari(), hari()] },
  { label: '7 hari', range: () => [hari(-6), hari()] },
  { label: 'Bulan ini', range: () => [awalBulan(), hari()] },
]

const run = useRun()
const confirm = useConfirm()
const rentang = ref([awalBulan(), hari()])
const kategori = ref(null)
const cari = ref('')
const page = ref(0)
const data = ref({ items: [], total: 0, totalJumlah: 0 })
const loading = ref(false)
const ubah = ref(null) // salinan baris yang sedang diedit

const opsiKategori = [...KATEGORI_PENGELUARAN, { value: 'refund', label: 'Refund bill' }]

const params = (extra) => {
  const [dari, sampai] = rentang.value
  const p = { dari: ymd(dari), sampai: ymd(sampai || dari), ...extra }
  if (kategori.value) p.kategori = kategori.value
  if (cari.value.trim()) p.cari = cari.value.trim()
  return p
}

async function muat() {
  if (!rentang.value[0]) return
  loading.value = true
  await run(async () => (data.value = await get('/pengeluaran/riwayat', { params: params({ page: page.value }) })))
  loading.value = false
}

async function unduh() {
  const [dari, sampai] = rentang.value
  const { items, totalJumlah } = await get('/pengeluaran/riwayat', { params: params({ semua: true }) })
  return unduhExcel(`pengeluaran-${ymd(dari)}_${ymd(sampai || dari)}.xlsx`, [
    {
      nama: 'Pengeluaran',
      kolom: [
        { header: 'Waktu', key: 'waktu', lebar: 22 },
        { header: 'Kategori', key: 'kategori', lebar: 26 },
        { header: 'Keterangan', key: 'keterangan', lebar: 40 },
        { header: 'Jumlah', key: 'jumlah', uang: true },
        { header: 'Dicatat oleh', key: 'oleh', lebar: 18 },
      ],
      baris: [
        ...items.map((p) => ({ waktu: tgl(p.at), kategori: labelKategori(p.kategori), keterangan: p.keterangan, jumlah: p.jumlah, oleh: p.oleh })),
        { keterangan: 'Total', jumlah: totalJumlah },
      ],
    },
  ])
}

async function simpan() {
  const { id, keterangan, jumlah, kategori: k } = ubah.value
  if (await run(() => put(`/pengeluaran/${id}`, { keterangan: keterangan.trim(), jumlah, kategori: k }), 'Pengeluaran diperbarui')) {
    ubah.value = null
    run(muat)
  }
}

function hapus(p) {
  konfirmasi(confirm, {
    header: 'Hapus pengeluaran?',
    message: `"${p.keterangan}" (${rp(p.jumlah)}) dihapus. Laporan dan laba rugi ikut berubah.`,
    accept: async () => (await run(() => del(`/pengeluaran/${p.id}`), 'Pengeluaran dihapus')) && run(muat),
  })
}

function ulang() {
  page.value = 0
  muat()
}

let tunda
watch(cari, () => {
  clearTimeout(tunda)
  tunda = setTimeout(ulang, 350) // tunggu selesai mengetik sebelum minta ke server
})
watch(kategori, ulang)
watch(rentang, (r) => r[1] !== null && ulang())
onMounted(muat)
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-wallet" title="Riwayat Pengeluaran" subtitle="Semua kas keluar, belanja bahan, dan refund">
      <button v-for="p in PRESET" :key="p.label" class="pill" @click="rentang = p.range()">{{ p.label }}</button>
      <DatePicker v-model="rentang" selection-mode="range" date-format="dd/mm/yy" :manual-input="false" show-icon placeholder="Pilih tanggal" />
    </PageHeader>

    <section class="panel box">
      <div class="tools">
        <CariInput v-model="cari" placeholder="Cari keterangan…" class="cari" />
        <Select v-model="kategori" :options="opsiKategori" option-label="label" option-value="value" placeholder="Semua kategori" show-clear class="kat" />
        <span class="total">Total <strong>{{ rp(data.totalJumlah) }}</strong> · {{ data.total }} catatan</span>
        <Button label="Unduh Excel" icon="pi pi-file-excel" variant="outlined" :disabled="!data.total" @click="run(unduh)" />
      </div>

      <DataTable
        :value="data.items"
        :loading="loading"
        size="small"
        scrollable
        lazy
        paginator
        :rows="50"
        :total-records="data.total"
        :first="page * 50"
        empty-message="Tidak ada pengeluaran pada rentang ini"
        @page="(e) => ((page = e.page), muat())"
      >
        <Column header="Waktu"><template #body="{ data: p }">{{ tgl(p.at) }}</template></Column>
        <Column header="Kategori"><template #body="{ data: p }">{{ labelKategori(p.kategori) }}</template></Column>
        <Column field="keterangan" header="Keterangan" />
        <Column header="Jumlah" class="kanan"><template #body="{ data: p }">{{ rp(p.jumlah) }}</template></Column>
        <Column field="oleh" header="Dicatat oleh" />
        <Column class="aksi">
          <template #body="{ data: p }">
            <template v-if="p.kategori !== 'refund'">
              <Button icon="pi pi-pencil" text rounded aria-label="Ubah" @click="ubah = { ...p }" />
              <Button icon="pi pi-trash" text rounded severity="danger" aria-label="Hapus" @click="hapus(p)" />
            </template>
          </template>
        </Column>
      </DataTable>
    </section>

    <Dialog :visible="!!ubah" modal header="Ubah pengeluaran" :style="{ width: '28rem' }" :breakpoints="{ '600px': '96vw' }" @update:visible="ubah = null">
      <div v-if="ubah" class="stack">
        <label>Kategori</label>
        <Select v-model="ubah.kategori" :options="KATEGORI_PENGELUARAN" option-label="label" option-value="value" fluid />
        <label for="ket">Keterangan</label>
        <InputText id="ket" v-model="ubah.keterangan" maxlength="255" fluid />
        <label>Jumlah</label>
        <MoneyInput v-model="ubah.jumlah" />
        <Button label="Simpan" icon="pi pi-check" :disabled="!ubah.keterangan.trim() || !ubah.jumlah" @click="simpan" />
      </div>
    </Dialog>
  </div>
</template>

<style scoped>
.gap {
  gap: 14px;
}
.box {
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.tools {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}
.cari {
  margin-bottom: 0;
  flex: 1 1 220px;
}
.kat {
  width: 220px;
}
.total {
  margin-left: auto;
  color: var(--p-text-muted-color);
}
.total strong {
  color: var(--p-text-color);
}
:deep(.kanan) {
  text-align: right;
  white-space: nowrap;
}
:deep(.aksi) {
  white-space: nowrap;
  width: 1%;
}
</style>
