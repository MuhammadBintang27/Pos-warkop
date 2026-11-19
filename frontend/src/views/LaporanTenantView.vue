<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useCari } from '@/composables/useCari'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import { del, errMsg, get, post } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { konfirmasi } from '@/utils/confirm'
import { hari, rp, tgl, ymd } from '@/utils/format'

const STATUS = {
  belum: { label: 'Belum closing', severity: 'secondary' },
  closing: { label: 'Closing', severity: 'info' },
  disetor: { label: 'Disetor', severity: 'success' },
  berubah: { label: 'Berubah setelah closing', severity: 'danger' },
  sendiri: { label: 'Tanpa closing', severity: 'secondary' },
}
const PRESET = [
  { label: 'Hari ini', range: () => [hari(), hari()] },
  { label: 'Kemarin', range: () => [hari(-1), hari(-1)] },
  { label: '7 hari', range: () => [hari(-6), hari()] },
  { label: 'Bulan ini', range: () => [new Date(new Date().getFullYear(), new Date().getMonth(), 1), hari()] },
]

const confirm = useConfirm()
const toast = useToast()
const run = useRun()

const tab = ref('harian')
const rentang = ref([hari(-6), hari()])
const tenant = ref([])
const filterTenant = ref(null)
const harian = ref({ baris: [] })
const tenantId = ref(null)
const detail = ref(null)

const tanggalIndo = (s) => s.split('-').reverse().join('/')
const query = () => {
  const [dari, sampai] = rentang.value
  return dari ? { dari: ymd(dari), sampai: ymd(sampai || dari) } : null
}

const baris = computed(() => harian.value.baris.filter((b) => !filterTenant.value || b.tenantId === filterTenant.value))
const belum = computed(() => baris.value.filter((b) => b.status === 'belum'))
const tertunda = computed(() =>
  (harian.value.tertunda ?? [])
    .filter((t) => !filterTenant.value || t.tenantId === filterTenant.value)
    .map((t) => ({ ...t, nama: tenant.value.find((x) => x.id === t.tenantId)?.nama ?? '—' })),
)
const tertundaBill = computed(() => harian.value.tertunda?.find((t) => t.tenantId === bill.value?.tenantId) ?? null)
const total = computed(() => ({
  kotor: baris.value.reduce((s, b) => s + b.kotor, 0),
  komisi: baris.value.reduce((s, b) => s + b.komisi, 0),
  setor: baris.value.reduce((s, b) => s + b.setor, 0),
}))

async function muatHarian() {
  const params = query()
  if (params) harian.value = await get('/laporan/tenant/harian', { params })
}

async function muatDetail() {
  const params = query()
  detail.value = params && tenantId.value ? await get('/laporan/tenant/detail', { params: { ...params, tenantId: tenantId.value } }) : null
}

const muat = () => run(() => Promise.all([muatHarian(), muatDetail()]))

watch(rentang, (r) => r[1] !== null && muat())
watch(tenantId, () => run(muatDetail))

const bill = ref(null)
const billBuka = ref(false)

// ambil barisnya lagi dari data terbaru supaya bill menampilkan status, siapa, dan waktu yang sudah diperbarui
function bukaBill(b) {
  bill.value = harian.value.baris.find((x) => x.tenantId === b.tenantId && x.tanggal === b.tanggal) ?? b
  billBuka.value = true
}

const pilihBill = ref({ baris: null, bills: [] })
const pilihBuka = ref(false)

async function kirimClosing(b, pesananIds) {
  await post('/closing', { tenantId: b.tenantId, tanggal: b.tanggal, pesananIds })
  await muatHarian()
  bukaBill(b)
}

// bila ada bill belum dibayar, owner memilih mana yang ikut ditalangi; bila tidak ada, langsung closing
const closing = (b) =>
  run(async () => {
    const bills = await get('/laporan/tenant/belum-bayar', { params: { tenantId: b.tenantId } })
    if (!bills.length) return kirimClosing(b, [])
    pilihBill.value = { baris: b, bills }
    pilihBuka.value = true
  })

const submitPilihan = (ids) =>
  run(async () => {
    await kirimClosing(pilihBill.value.baris, ids)
    pilihBuka.value = false
  }, 'Closing berhasil')

const setor = (b) =>
  run(async () => {
    await post(`/closing/${b.closingId}/setor`)
    await muatHarian()
    bukaBill(b)
  }, 'Ditandai sudah disetor')

function batalClosing(b) {
  konfirmasi(confirm, {
    header: 'Batalkan closing?',
    message: `Closing ${b.nama} tanggal ${tanggalIndo(b.tanggal)} dibatalkan dan angkanya dihitung ulang. Closing yang sudah disetor tidak bisa dibatalkan.`,
    label: 'Batalkan closing',
    accept: () =>
      run(async () => {
        await del(`/closing/${b.closingId}`)
        await muatHarian()
      }, 'Closing dibatalkan'),
  })
}

function closingSemua() {
  konfirmasi(confirm, {
    header: 'Closing semua?',
    message: `${belum.value.length} penjualan tenant yang belum di-closing akan ditutup. Bill yang belum dibayar pelanggan tidak ikut (tidak ditalangi) dan masuk closing saat dibayar. Untuk menalangi, pakai tombol Closing di barisnya.`,
    label: 'Closing semua',
    accept: async () => {
      let berhasil = 0
      const gagal = []
      for (const b of belum.value) {
        try {
          await post('/closing', { tenantId: b.tenantId, tanggal: b.tanggal })
          berhasil++
        } catch (e) {
          gagal.push(`${b.nama} ${tanggalIndo(b.tanggal)}: ${errMsg(e)}`)
        }
      }
      await run(muatHarian)
      toast.add({
        severity: gagal.length ? 'warn' : 'success',
        summary: `${berhasil} closing berhasil${gagal.length ? `, ${gagal.length} dilewati` : ''}`,
        detail: gagal.join('\n'),
        life: 8000,
      })
    },
  })
}

onMounted(async () => {
  await run(async () => {
    tenant.value = await get('/tenant')
    await muatHarian()
  })
  tenantId.value = tenant.value[0]?.id ?? null
})
const { cari: cariBaris, hasil: barisTampil } = useCari(baris, 'nama', 'tanggal')
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-shop" title="Laporan Tenant" subtitle="Penjualan harian, closing, dan detail per menu">
      <button v-for="p in PRESET" :key="p.label" class="pill" @click="rentang = p.range()">{{ p.label }}</button>
      <DatePicker v-model="rentang" selection-mode="range" date-format="dd/mm/yy" :manual-input="false" show-icon placeholder="Pilih tanggal" />
    </PageHeader>

    <Tabs v-model:value="tab">
      <TabList>
        <Tab value="harian">Harian dan closing</Tab>
        <Tab value="detail">Detail penjualan</Tab>
      </TabList>
      <TabPanels>
        <TabPanel value="harian">
          <div class="stack gap">
            <Message v-if="tertunda.length" severity="warn" size="small">
              Belum dibayar pelanggan, jadi belum masuk penjualan. Nilainya ikut closing pada hari bill itu dibayar:
              <strong v-for="t in tertunda" :key="t.tenantId" class="tt">{{ t.nama }} {{ t.bill }} bill ≈ {{ rp(t.nilai) }}</strong>
            </Message>

            <div class="stats">
              <StatCard icon="pi pi-wallet" label="Penjualan tenant" :value="rp(total.kotor)" />
              <StatCard icon="pi pi-percentage" label="Pendapatan pemilik tempat" :value="rp(total.komisi)" />
              <StatCard icon="pi pi-send" label="Setor ke tenant" :value="rp(total.setor)" />
            </div>

            <section class="panel box">
              <div class="bar">
                <Select
                  v-model="filterTenant"
                  :options="tenant"
                  option-label="nama"
                  option-value="id"
                  placeholder="Semua tenant"
                  show-clear
                  class="pilih"
                />
                <Button
                  :label="`Closing semua (${belum.length})`"
                  icon="pi pi-lock"
                  :disabled="!belum.length"
                  class="kanan"
                  @click="closingSemua"
                />
              </div>

              <CariInput v-model="cariBaris" placeholder="Cari tenant atau tanggal (YYYY-MM-DD)…" />
              <DataTable :value="barisTampil" size="small" paginator :rows="20" empty-message="Belum ada penjualan tenant pada rentang ini">
                <Column header="Tanggal"><template #body="{ data: b }">{{ tanggalIndo(b.tanggal) }}</template></Column>
                <Column field="nama" header="Tenant" />
                <Column field="transaksi" header="Bill" />
                <Column header="Penjualan"><template #body="{ data: b }">{{ rp(b.kotor) }}</template></Column>
                <Column header="Diskon"><template #body="{ data: b }">{{ rp(b.diskon) }}</template></Column>
                <Column header="Komisi">
                  <template #body="{ data: b }">
                    <small v-if="b.milikSendiri" class="redup">milik sendiri</small>
                    <template v-else>{{ rp(b.komisi) }} <small class="redup">({{ b.komisiPersen }}%)</small></template>
                  </template>
                </Column>
                <Column header="Setor ke tenant">
                  <template #body="{ data: b }">
                    <span v-if="b.milikSendiri" class="redup">—</span>
                    <strong v-else>{{ rp(b.setor) }}</strong>
                  </template>
                </Column>
                <Column header="Status">
                  <template #body="{ data: b }">
                    <Tag :value="STATUS[b.status].label" :severity="STATUS[b.status].severity" />
                    <small v-if="b.disetorAt" class="redup"><br />{{ tgl(b.disetorAt) }}</small>
                  </template>
                </Column>
                <Column>
                  <template #body="{ data: b }">
                    <Button
                      icon="pi pi-list"
                      size="small"
                      text
                      severity="secondary"
                      aria-label="Lihat transaksi"
                      title="Lihat transaksi hari ini"
                      @click="$router.push({ path: '/riwayat', query: { tenant: b.tenantId, tanggal: b.tanggal } })"
                    />
                    <!-- dapur sendiri: tidak ada setoran, cukup rekap harian yang bisa dicetak -->
                    <Button v-if="b.status === 'sendiri'" label="Rekap" size="small" icon="pi pi-print" severity="secondary" outlined @click="bukaBill(b)" />
                    <Button v-else-if="b.status === 'belum'" label="Closing" size="small" icon="pi pi-lock" @click="closing(b)" />
                    <template v-else>
                      <Button icon="pi pi-file" size="small" text severity="secondary" aria-label="Lihat bill closing" @click="bukaBill(b)" />
                      <Button v-if="b.status === 'closing' && !b.milikSendiri" label="Tandai disetor" size="small" outlined @click="setor(b)" />
                      <Button
                        v-if="b.status !== 'disetor'"
                        icon="pi pi-undo"
                        size="small"
                        text
                        severity="secondary"
                        aria-label="Batalkan closing"
                        @click="batalClosing(b)"
                      />
                    </template>
                  </template>
                </Column>
              </DataTable>

              <small class="redup">
                Tanggal mengikuti hari kerja, yaitu tanggal shift dibuka, jadi shift yang melewati tengah malam tetap satu hari.
                Closing bisa dilakukan kapan saja, tidak menunggu shift kasir ditutup. Closing membekukan angka tenant pada satu tanggal dan menghitung bill yang sudah lunas, ditambah bill belum dibayar yang kamu pilih untuk ditalangi. Statusnya berubah menjadi "Berubah setelah closing" bila ada transaksi yang masuk
                sesudahnya. Komisi dihitung dari penjualan setelah diskon, dengan diskon bill dibagi menurut nilai item tiap tenant.
              </small>
            </section>
          </div>
        </TabPanel>

        <TabPanel value="detail">
          <div class="stack gap">
            <Select
              v-model="tenantId"
              :options="tenant"
              option-label="nama"
              option-value="id"
              placeholder="Pilih tenant"
              class="pilih"
            />

            <template v-if="detail">
              <div class="stats">
                <StatCard icon="pi pi-shopping-bag" label="Porsi terjual" :value="String(detail.qty)" />
                <StatCard icon="pi pi-wallet" label="Nilai penjualan" :value="rp(detail.nilai)" />
                <StatCard icon="pi pi-times-circle" label="Porsi dibatalkan" :value="String(detail.batalQty)" />
                <StatCard icon="pi pi-ban" label="Nilai dibatalkan" :value="rp(detail.batalNilai)" />
              </div>

              <section class="panel box">
                <h3>Penjualan per menu</h3>
                <DataTable paginator :rows="10" :rows-per-page-options="[10, 25, 50]" :always-show-paginator="false" :value="detail.menu" size="small" empty-message="Belum ada penjualan">
                  <Column field="nama" header="Menu" />
                  <Column field="qty" header="Terjual" />
                  <Column header="Nilai"><template #body="{ data: m }">{{ rp(m.nilai) }}</template></Column>
                  <Column header="Dibatalkan">
                    <template #body="{ data: m }">
                      <Tag v-if="m.batalQty" :value="String(m.batalQty)" severity="danger" />
                      <span v-else class="redup">—</span>
                    </template>
                  </Column>
                  <Column header="Nilai batal"><template #body="{ data: m }">{{ m.batalNilai ? rp(m.batalNilai) : '—' }}</template></Column>
                </DataTable>
              </section>

              <section class="panel box">
                <h3>Pembatalan</h3>
                <DataTable paginator :rows="10" :rows-per-page-options="[10, 25, 50]" :always-show-paginator="false" :value="detail.batal" size="small" empty-message="Tidak ada pembatalan pada rentang ini">
                  <Column header="Waktu"><template #body="{ data: b }">{{ tgl(b.at) }}</template></Column>
                  <Column header="Bill"><template #body="{ data: b }">#{{ b.pesananId }}</template></Column>
                  <Column header="Menu">
                    <template #body="{ data: b }">{{ b.nama }}<small v-if="b.opsi" class="redup"> · {{ b.opsi }}</small></template>
                  </Column>
                  <Column field="qty" header="Jumlah" />
                  <Column header="Nilai"><template #body="{ data: b }">{{ rp(b.nilai) }}</template></Column>
                  <Column field="alasan" header="Alasan" />
                  <Column field="oleh" header="Oleh" />
                  <Column header="Keterangan">
                    <template #body="{ data: b }">
                      <Tag v-if="b.sudahDibuat" value="Sudah dibuat" severity="warn" class="tag" />
                      <Tag v-else-if="b.sudahDikirim" value="Sudah dikirim" severity="secondary" class="tag" />
                    </template>
                  </Column>
                </DataTable>
              </section>
            </template>
          </div>
        </TabPanel>
      </TabPanels>
    </Tabs>

    <ClosingDialog v-model:visible="pilihBuka" :baris="pilihBill.baris" :bills="pilihBill.bills" @submit="submitPilihan" />
    <BillClosingDialog v-model:visible="billBuka" :baris="bill" :tertunda="bill?.milikSendiri ? null : tertundaBill" />
  </div>
</template>

<style scoped>
.gap {
  gap: 14px;
}
.stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 14px;
}
.box {
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.bar {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
}
.kanan {
  margin-left: auto;
}
.pilih {
  min-width: 220px;
}
h3 {
  margin: 0;
}
.redup {
  color: var(--p-text-muted-color);
}
.tag {
  margin-right: 4px;
}
.tt {
  margin-left: 10px;
}
</style>
