<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { get, post } from '@/api/http'
import { onEvent } from '@/composables/useEvents'
import { useRun } from '@/composables/useRun'
import { useAuthStore } from '@/stores/auth'
import { useKasirStore } from '@/stores/kasir'
import { usePengaturanStore } from '@/stores/pengaturan'
import { hari, rp, tgl, ymd } from '@/utils/format'
import { cetakStruk } from '@/utils/cetak'

const STATUS = [
  { label: 'Semua', value: '' },
  { label: 'Berjalan', value: 'open' },
  { label: 'Lunas', value: 'lunas' },
  { label: 'Batal', value: 'batal' },
]
const STATUS_TAG = {
  open: { label: 'Berjalan', severity: 'warn' },
  lunas: { label: 'Lunas', severity: 'success' },
  batal: { label: 'Batal', severity: 'danger' },
}
const PRESET = [
  { label: 'Hari ini', range: () => [hari(), hari()] },
  { label: 'Kemarin', range: () => [hari(-1), hari(-1)] },
  { label: '7 hari', range: () => [hari(-6), hari()] },
]

const router = useRouter()
const auth = useAuthStore()
const kasir = useKasirStore()
const peng = usePengaturanStore()
const run = useRun()

const route = useRoute()
// dibuka dari Laporan Tenant: ?tenant=<id>&tanggal=YYYY-MM-DD
const tglAwal = route.query.tanggal ? new Date(`${route.query.tanggal}T00:00:00`) : null
const rentang = ref(tglAwal && !isNaN(tglAwal) ? [tglAwal, tglAwal] : [hari(), hari()])
const tenant = ref(Number(route.query.tenant) || null)
const tenants = ref([])
const namaTenant = computed(() => tenants.value.find((t) => t.id === tenant.value)?.nama)
const status = ref('')
const q = ref('')
const oleh = ref(null)
const users = ref([])
const page = ref(1)
const limit = ref(20)
const loading = ref(false)
const data = ref({ items: [], total: 0, lunas: 0, omzet: 0 })
const detail = ref(null)
const detailBuka = ref(false)

const subtitle = computed(
  () => ({ owner: 'Semua transaksi', kasir: 'Transaksi yang kamu buat atau terima pembayarannya' })[auth.role] ?? 'Pesanan yang kamu input',
)

async function muat() {
  const [dari, sampai] = rentang.value
  if (!dari) return
  loading.value = true
  const params = { dari: ymd(dari), sampai: ymd(sampai || dari), status: status.value, q: q.value.trim(), page: page.value, limit: limit.value }
  if (oleh.value) params.oleh = oleh.value
  if (tenant.value) params.tenant = tenant.value
  await run(async () => (data.value = await get('/riwayat', { params })))
  loading.value = false
}

function ulang() {
  if (rentang.value[1] === null) return
  page.value = 1
  muat()
}

let timer
watch(q, () => {
  clearTimeout(timer)
  timer = setTimeout(ulang, 350)
})
watch([rentang, status, oleh, tenant], ulang)

function ganti(e) {
  page.value = e.page + 1
  limit.value = e.rows
  muat()
}

async function lihat(row) {
  detail.value = null
  detailBuka.value = true
  await run(async () => (detail.value = await get(`/pesanan/${row.id}`)))
}

function lanjutkan() {
  run(async () => {
    await kasir.bukaPesanan(detail.value.id)
    router.push('/')
  })
}

// koreksi bill lunas oleh owner: 'buka-ulang' (salah input, bayar ulang) atau 'refund' (bill dibatalkan, uang kembali)
const koreksi = ref(null)
const tunaiDetail = computed(() => (detail.value?.pembayaran ?? []).filter((b) => b.metode === 'tunai').reduce((s, b) => s + b.jumlah, 0))

async function simpanKoreksi() {
  const { jenis, alasan, sudahDibuat } = koreksi.value
  const ok = await run(async () => {
    detail.value = await post(`/pesanan/${detail.value.id}/${jenis}`, { alasan: alasan.trim(), sudahDibuat })
  }, jenis === 'refund' ? 'Bill direfund' : 'Bill dibuka ulang')
  if (!ok) return
  koreksi.value = null
  if (jenis === 'buka-ulang') lanjutkan()
}

const offs = []
onMounted(() => {
  muat()
  if (auth.isOwner) run(async () => (users.value = await get('/users')))
  run(async () => (tenants.value = await get('/tenant')))
  offs.push(onEvent('pesanan', () => muat()))
})
onBeforeUnmount(() => {
  offs.forEach((off) => off())
  clearTimeout(timer)
})
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-history" title="Riwayat" :subtitle="subtitle">
      <button v-for="p in PRESET" :key="p.label" class="pill" @click="rentang = p.range()">{{ p.label }}</button>
      <DatePicker v-model="rentang" selection-mode="range" date-format="dd/mm/yy" :manual-input="false" show-icon placeholder="Pilih tanggal" />
    </PageHeader>

    <div class="stats">
      <StatCard icon="pi pi-receipt" label="Transaksi lunas" :value="String(data.lunas)" />
      <StatCard icon="pi pi-wallet" label="Nilai transaksi lunas" :value="rp(data.omzet)" />
    </div>

    <section class="panel box">
      <div class="filters">
        <SelectButton v-model="status" :options="STATUS" option-label="label" option-value="value" :allow-empty="false" />
        <IconField class="cari">
          <InputIcon class="pi pi-search" />
          <InputText v-model="q" placeholder="Cari meja, nama, atau nomor…" fluid />
        </IconField>
        <Select
          v-if="auth.isOwner"
          v-model="oleh"
          :options="users"
          option-label="nama"
          option-value="id"
          placeholder="Semua pembuat"
          show-clear
          class="oleh"
        />
        <Select
          v-if="tenants.length > 1"
          v-model="tenant"
          :options="tenants"
          option-label="nama"
          option-value="id"
          placeholder="Semua tenant"
          show-clear
          class="oleh"
        />
      </div>

      <DataTable
        :value="data.items"
        lazy
        paginator
        :rows="limit"
        :first="(page - 1) * limit"
        :total-records="data.total"
        :loading="loading"
        :rows-per-page-options="[20, 50, 100]"
        size="small"
        empty-message="Tidak ada transaksi pada rentang ini"
        @page="ganti"
      >
        <Column header="Waktu"><template #body="{ data: r }">{{ tgl(r.lunasAt || r.dibuatAt) }}</template></Column>
        <Column header="No."><template #body="{ data: r }">#{{ r.id }}</template></Column>
        <Column header="Meja">
          <template #body="{ data: r }">
            {{ r.mejaNama || 'Bawa pulang' }}<small v-if="r.atasNama" class="sub"> · {{ r.atasNama }}</small>
          </template>
        </Column>
        <Column field="jumlahItem" header="Item" />
        <Column header="Total"><template #body="{ data: r }"><strong>{{ rp(r.total) }}</strong></template></Column>
        <Column v-if="tenant" :header="`Item ${namaTenant ?? 'tenant'}`">
          <template #body="{ data: r }">{{ rp(r.nilaiTenant) }}</template>
        </Column>
        <Column header="Bayar">
          <template #body="{ data: r }">
            <Tag v-for="m in r.metode ? r.metode.split(',') : []" :key="m" :value="m" severity="secondary" class="metode" />
          </template>
        </Column>
        <Column field="pembuat" header="Diinput oleh" />
        <Column v-if="auth.isOwner || auth.isKasir" field="kasir" header="Kasir" />
        <Column header="Status">
          <template #body="{ data: r }"><Tag :value="STATUS_TAG[r.status].label" :severity="STATUS_TAG[r.status].severity" /></template>
        </Column>
        <Column>
          <template #body="{ data: r }">
            <Button icon="pi pi-eye" text severity="secondary" aria-label="Detail" @click="lihat(r)" />
          </template>
        </Column>
      </DataTable>
    </section>

    <Dialog v-model:visible="detailBuka" modal header="Detail transaksi" :style="{ width: '30rem' }" :breakpoints="{ '600px': '96vw' }">
      <div v-if="detail" class="stack">
        <div class="head">
          <strong>{{ detail.mejaNama || 'Bawa pulang' }} · #{{ detail.id }}</strong>
          <Tag :value="STATUS_TAG[detail.status]?.label ?? detail.status" :severity="STATUS_TAG[detail.status]?.severity" />
        </div>
        <small class="sub">{{ tgl(detail.lunasAt || detail.dibuatAt) }}<template v-if="detail.atasNama"> · {{ detail.atasNama }}</template></small>

        <ul class="items">
          <li v-for="i in detail.items" :key="i.id">
            <span>
              {{ i.qty }}× {{ i.nama }}
              <small v-if="i.opsi" class="sub"><br />{{ i.opsi }}</small>
              <small v-if="i.catatan" class="sub"><br />“{{ i.catatan }}”</small>
            </span>
            <span>{{ rp(i.harga * i.qty) }}</span>
          </li>
        </ul>

        <div class="row"><span>Subtotal</span><span>{{ rp(detail.subtotal) }}</span></div>
        <div v-if="detail.diskon" class="row"><span>Diskon</span><span>-{{ rp(detail.diskon) }}</span></div>
        <div class="row total"><span>Total</span><strong>{{ rp(detail.total) }}</strong></div>
        <div v-for="(b, i) in detail.pembayaran" :key="i" class="row">
          <span class="cap">{{ b.metode }}{{ b.pelanggan ? ' · ' + b.pelanggan : '' }}</span><span>{{ rp(b.jumlah) }}</span>
        </div>
        <small v-if="detail.kasir" class="sub">Kasir: {{ detail.kasir }}</small>

        <template v-if="detail.pembatalan.length">
          <strong class="batal-judul">Dibatalkan</strong>
          <ul class="items">
            <li v-for="(b, i) in detail.pembatalan" :key="i">
              <span>
                {{ b.qty }}× {{ b.nama }}
                <small class="sub">
                  <br />{{ b.alasan }} · {{ b.oleh }} · {{ tgl(b.at) }}
                  <template v-if="b.sudahDikirim"> · sudah dikirim ke dapur</template>
                  <template v-if="b.sudahDibuat"> · sudah dibuat</template>
                </small>
              </span>
              <span>{{ rp(b.harga * b.qty) }}</span>
            </li>
          </ul>
        </template>
      </div>
      <template #footer>
        <template v-if="detail?.status === 'lunas' && auth.isOwner">
          <Button label="Refund" icon="pi pi-replay" severity="danger" text @click="koreksi = { jenis: 'refund', alasan: '', sudahDibuat: true }" />
          <Button label="Buka ulang" icon="pi pi-pencil" severity="secondary" text @click="koreksi = { jenis: 'buka-ulang', alasan: '', sudahDibuat: false }" />
        </template>
        <Button v-if="detail?.status === 'lunas'" label="Cetak ulang struk" icon="pi pi-print" severity="secondary" outlined @click="run(() => cetakStruk(detail, peng.settings))" />
        <Button v-if="detail?.status === 'open'" label="Lanjutkan di kasir" icon="pi pi-arrow-right" icon-pos="right" @click="lanjutkan" />
      </template>
    </Dialog>

    <Dialog :visible="!!koreksi" modal :header="koreksi?.jenis === 'refund' ? 'Refund bill' : 'Buka ulang bill'" :style="{ width: '26rem' }" :breakpoints="{ '600px': '96vw' }" @update:visible="koreksi = null">
      <div v-if="koreksi" class="stack">
        <small v-if="koreksi.jenis === 'buka-ulang'" class="sub">
          Pembayaran bill ini dihapus dan bill kembali berjalan, lalu dibuka di kasir untuk dikoreksi dan dibayar ulang.
          Hanya untuk bill yang dibayar di shift yang sedang buka.
        </small>
        <small v-else class="sub">
          Bill dibatalkan dan tidak dihitung sebagai penjualan.
          <template v-if="tunaiDetail"> Uang tunai {{ rp(tunaiDetail) }} dicatat keluar dari kas laci shift ini.</template>
          Utang kasbon dari bill ini dihapus; QRIS/transfer dikembalikan manual.
        </small>
        <label for="kalasan">Alasan</label>
        <InputText id="kalasan" v-model="koreksi.alasan" maxlength="100" placeholder="mis. salah metode bayar" fluid autofocus />
        <div v-if="koreksi.jenis === 'refund'" class="cek">
          <Checkbox v-model="koreksi.sudahDibuat" input-id="kdibuat" binary />
          <label for="kdibuat">Sudah dibuat/disajikan (stok tidak dikembalikan)</label>
        </div>
      </div>
      <template #footer>
        <Button :label="koreksi?.jenis === 'refund' ? 'Refund' : 'Buka ulang'" :severity="koreksi?.jenis === 'refund' ? 'danger' : undefined" icon="pi pi-check" :disabled="!koreksi?.alasan.trim()" @click="simpanKoreksi" />
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
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.filters {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
}
.cari {
  flex: 1;
  min-width: 200px;
}
.oleh {
  min-width: 180px;
}
.sub {
  color: var(--p-text-muted-color);
}
.metode {
  margin-right: 4px;
  text-transform: capitalize;
}
.head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 1.1rem;
}
.items {
  list-style: none;
  margin: 4px 0;
  padding: 8px 0;
  border-block: 1px dashed var(--p-content-border-color);
}
.items li {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 4px 0;
}
.row {
  display: flex;
  justify-content: space-between;
}
.batal-judul {
  margin-top: 6px;
  color: var(--p-red-500, #ef4444);
}
.total {
  font-size: 1.2rem;
  color: var(--accent-text);
}
.cap {
  text-transform: capitalize;
}
.cek {
  display: flex;
  align-items: center;
  gap: 10px;
}
</style>
