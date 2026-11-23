<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useCari } from '@/composables/useCari'
import { useConfirm } from 'primevue/useconfirm'
import { del, get, post, put } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { konfirmasi } from '@/utils/confirm'
import { hari, rp, tgl, ymd } from '@/utils/format'

const MODE = [
  { label: 'Karyawan', value: 'karyawan' },
  { label: 'Slip gaji', value: 'slip' },
]
const JENIS = [
  { label: 'Bulanan', value: 'bulanan' },
  { label: 'Harian', value: 'harian' },
]
const STATUS = { draft: { label: 'Belum dibayar', severity: 'warn' }, dibayar: { label: 'Dibayar', severity: 'success' } }

const confirm = useConfirm()
const run = useRun()

const mode = ref('karyawan')
const list = ref([])
const slips = ref([])
const form = ref(null)
const bon = ref(null)
const slip = ref(null)

const awalBulan = () => new Date(new Date().getFullYear(), new Date().getMonth(), 1)
const akhirBulan = () => new Date(new Date().getFullYear(), new Date().getMonth() + 1, 0)
const rentang = ref([awalBulan(), akhirBulan()])

const totalBon = computed(() => list.value.reduce((s, k) => s + k.saldoBon, 0))
const belumDibayar = computed(() => slips.value.filter((s) => s.status === 'draft').reduce((n, s) => n + s.netto, 0))

const muat = async () => (list.value = await get('/karyawan'))
async function muatSlip() {
  const [dari, sampai] = rentang.value
  if (!dari) return
  await run(async () => (slips.value = await get(`/gaji?dari=${ymd(dari)}&sampai=${ymd(sampai || dari)}`)))
}

async function simpan() {
  const { id, nama, jabatan, jenis, gaji } = form.value
  const body = { nama: nama.trim(), jabatan: jabatan.trim(), jenis, gaji: gaji || 0 }
  if (await run(() => (id ? put(`/karyawan/${id}`, body) : post('/karyawan', body)), 'Karyawan disimpan')) {
    form.value = null
    await run(muat)
  }
}

function hapus(k) {
  konfirmasi(confirm, {
    header: 'Hapus karyawan?',
    message: `"${k.nama}" dinonaktifkan. Slip gaji dan bon lamanya tetap tersimpan. Karyawan yang masih punya bon tidak bisa dihapus.`,
    accept: async () => (await run(() => del(`/karyawan/${k.id}`), 'Karyawan dihapus')) && run(muat),
  })
}

async function bukaBon(k) {
  bon.value = { k, jumlah: null, keterangan: '', log: [] }
  await run(async () => (bon.value.log = await get(`/karyawan/${k.id}/bon`)))
}
async function simpanBon() {
  const { k, jumlah, keterangan } = bon.value
  if (await run(() => post(`/karyawan/${k.id}/bon`, { jumlah, keterangan: keterangan.trim() }), 'Bon dicatat')) {
    bon.value.jumlah = null
    bon.value.keterangan = ''
    await run(async () => (bon.value.log = await get(`/karyawan/${k.id}/bon`)))
    await run(muat)
  }
}

const bukaSlip = (k) => {
  const hari = k.jenis === 'harian' ? 26 : 0
  // bon langsung diusulkan dipotong penuh, dibatasi gaji pokok agar netto tidak minus
  const potonganBon = Math.min(k.saldoBon, k.jenis === 'harian' ? k.gaji * hari : k.gaji)
  slip.value = { k, rentang: [awalBulan(), akhirBulan()], hari, tunjangan: 0, bonus: 0, potongan: 0, potonganBon, catatan: '' }
}

const hitung = computed(() => {
  const s = slip.value
  if (!s) return { pokok: 0, netto: 0 }
  const pokok = s.k.jenis === 'harian' ? s.k.gaji * (s.hari || 0) : s.k.gaji
  return { pokok, netto: pokok + (s.tunjangan || 0) + (s.bonus || 0) - (s.potongan || 0) - (s.potonganBon || 0) }
})
const slipValid = () => slip.value && slip.value.rentang[0] && slip.value.rentang[1] && hitung.value.netto >= 0 && (slip.value.k.jenis !== 'harian' || slip.value.hari > 0)

async function simpanSlip() {
  const s = slip.value
  const body = {
    karyawanId: s.k.id,
    periodeAwal: ymd(s.rentang[0]),
    periodeAkhir: ymd(s.rentang[1]),
    hari: s.k.jenis === 'harian' ? s.hari : 0,
    tunjangan: s.tunjangan || 0,
    bonus: s.bonus || 0,
    potongan: s.potongan || 0,
    potonganBon: s.potonganBon || 0,
    catatan: s.catatan.trim(),
  }
  if (await run(() => post('/gaji', body), 'Slip gaji dibuat')) {
    slip.value = null
    mode.value = 'slip'
    await run(muat)
    await muatSlip()
  }
}

function bayar(s) {
  konfirmasi(confirm, {
    header: 'Tandai dibayar?',
    message: `Gaji ${s.karyawan} sebesar ${rp(s.netto)} ditandai sudah dibayar${s.potonganBon ? ` dan bon ${rp(s.potonganBon)} dipotong` : ''}. Slip yang dibayar tidak bisa diubah.`,
    accept: async () => {
      if (await run(() => post(`/gaji/${s.id}/bayar`), 'Gaji dibayar')) {
        await muatSlip()
        await run(muat)
      }
    },
  })
}

function hapusSlip(s) {
  konfirmasi(confirm, {
    header: 'Hapus slip gaji?',
    message: `Slip ${s.karyawan} periode ${s.periodeAwal} s/d ${s.periodeAkhir} dihapus.`,
    accept: async () => (await run(() => del(`/gaji/${s.id}`), 'Slip dihapus')) && muatSlip(),
  })
}

watch(rentang, (r) => r[1] !== null && muatSlip())
watch(mode, (m) => m === 'slip' && muatSlip())
onMounted(() => run(muat))
const { cari: cariKaryawan, hasil: karyawanTampil } = useCari(list, 'nama', 'jabatan')
const { cari: cariSlip, hasil: slipTampil } = useCari(slips, 'karyawan', 'jabatan')
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-id-card" title="Karyawan dan Gaji" subtitle="Data karyawan, bon, dan slip gaji">
      <button class="pill" @click="form = { nama: '', jabatan: '', jenis: 'bulanan', gaji: 0 }"><i class="pi pi-plus" /> Tambah karyawan</button>
    </PageHeader>

    <div class="stats">
      <StatCard icon="pi pi-users" label="Karyawan aktif" :value="String(list.length)" />
      <StatCard icon="pi pi-wallet" label="Total bon karyawan" :value="rp(totalBon)" />
      <StatCard v-if="mode === 'slip'" icon="pi pi-clock" label="Gaji belum dibayar (daftar ini)" :value="rp(belumDibayar)" />
    </div>

    <SelectButton v-model="mode" :options="MODE" option-label="label" option-value="value" :allow-empty="false" />

    <section v-if="mode === 'karyawan'" class="panel box">
      <CariInput v-model="cariKaryawan" placeholder="Cari karyawan atau jabatan…" />
      <DataTable paginator :rows="10" :rows-per-page-options="[10, 25, 50]" :always-show-paginator="false" :value="karyawanTampil" size="small" empty-message="Belum ada karyawan">
        <Column field="nama" header="Nama" sortable />
        <Column field="jabatan" header="Jabatan" />
        <Column header="Gaji">
          <template #body="{ data }">{{ rp(data.gaji) }} <small class="hint">/ {{ data.jenis === 'harian' ? 'hari' : 'bulan' }}</small></template>
        </Column>
        <Column header="Sisa bon"><template #body="{ data }">{{ data.saldoBon ? rp(data.saldoBon) : '—' }}</template></Column>
        <Column>
          <template #body="{ data }">
            <div class="act">
              <Button label="Buat slip" icon="pi pi-file" size="small" text @click="bukaSlip(data)" />
              <Button label="Bon" icon="pi pi-money-bill" size="small" text severity="secondary" @click="bukaBon(data)" />
              <Button icon="pi pi-pencil" size="small" text severity="secondary" aria-label="Ubah" @click="form = { ...data }" />
              <Button icon="pi pi-trash" size="small" text severity="danger" aria-label="Hapus" @click="hapus(data)" />
            </div>
          </template>
        </Column>
      </DataTable>
    </section>

    <template v-else>
      <section class="panel box filter">
        <DatePicker v-model="rentang" selection-mode="range" date-format="dd/mm/yy" :manual-input="false" show-icon placeholder="Periode berakhir pada" />
        <small class="hint">Menampilkan slip yang periodenya berakhir dalam rentang ini.</small>
      </section>
      <section class="panel box">
        <CariInput v-model="cariSlip" placeholder="Cari karyawan…" />
        <DataTable paginator :rows="10" :rows-per-page-options="[10, 25, 50]" :always-show-paginator="false" :value="slipTampil" size="small" empty-message="Belum ada slip gaji pada rentang ini">
          <Column header="Periode"><template #body="{ data }">{{ data.periodeAwal }} s/d {{ data.periodeAkhir }}</template></Column>
          <Column header="Karyawan">
            <template #body="{ data }">
              <strong>{{ data.karyawan }}</strong>
              <div class="hint">{{ data.jabatan }}</div>
            </template>
          </Column>
          <Column header="Pokok"><template #body="{ data }">{{ rp(data.pokok) }}<small v-if="data.jenis === 'harian'" class="hint"> ({{ data.hari }} hari)</small></template></Column>
          <Column header="Tunjangan + bonus"><template #body="{ data }">{{ rp(data.tunjangan + data.bonus) }}</template></Column>
          <Column header="Potongan"><template #body="{ data }">{{ rp(data.potongan + data.potonganBon) }}</template></Column>
          <Column header="Netto"><template #body="{ data }"><strong>{{ rp(data.netto) }}</strong></template></Column>
          <Column header="Status">
            <template #body="{ data }">
              <Tag :value="STATUS[data.status].label" :severity="STATUS[data.status].severity" />
              <div v-if="data.dibayarAt" class="hint">{{ tgl(data.dibayarAt) }}</div>
            </template>
          </Column>
          <Column>
            <template #body="{ data }">
              <div v-if="data.status === 'draft'" class="act">
                <Button label="Bayar" icon="pi pi-check" size="small" text @click="bayar(data)" />
                <Button icon="pi pi-trash" size="small" text severity="danger" aria-label="Hapus slip" @click="hapusSlip(data)" />
              </div>
            </template>
          </Column>
        </DataTable>
      </section>
    </template>

    <Dialog :visible="!!form" modal :header="form?.id ? 'Ubah karyawan' : 'Tambah karyawan'" :style="{ width: '26rem' }" @update:visible="form = null">
      <div v-if="form" class="stack">
        <label for="nama">Nama</label>
        <InputText id="nama" v-model="form.nama" maxlength="100" fluid />
        <label for="jab">Jabatan</label>
        <InputText id="jab" v-model="form.jabatan" maxlength="50" placeholder="mis. Barista, Juru masak" fluid />
        <label>Jenis gaji</label>
        <SelectButton v-model="form.jenis" :options="JENIS" option-label="label" option-value="value" :allow-empty="false" />
        <label>{{ form.jenis === 'harian' ? 'Gaji per hari kerja' : 'Gaji pokok per bulan' }}</label>
        <MoneyInput v-model="form.gaji" />
      </div>
      <template #footer><Button label="Simpan" icon="pi pi-check" :disabled="!form?.nama.trim()" @click="simpan" /></template>
    </Dialog>

    <Dialog :visible="!!bon" modal :header="`Bon ${bon?.k.nama}`" :style="{ width: '34rem' }" :breakpoints="{ '600px': '96vw' }" @update:visible="bon = null">
      <div v-if="bon" class="stack">
        <div class="saldo"><span>Sisa bon</span><strong>{{ rp(list.find((x) => x.id === bon.k.id)?.saldoBon ?? 0) }}</strong></div>
        <label>Bon baru (uang yang diambil karyawan)</label>
        <MoneyInput v-model="bon.jumlah" />
        <InputText v-model="bon.keterangan" maxlength="255" placeholder="Keterangan, mis. keperluan keluarga" fluid />
        <Button label="Catat bon" icon="pi pi-plus" :disabled="!bon.jumlah" @click="simpanBon" />
        <DataTable paginator :rows="10" :rows-per-page-options="[10, 25, 50]" :always-show-paginator="false" :value="bon.log" size="small" empty-message="Belum ada catatan bon">
          <Column field="tanggal" header="Tanggal" />
          <Column header="Jumlah"><template #body="{ data }"><span :class="{ minus: data.jumlah < 0 }">{{ data.jumlah < 0 ? '-' : '' }}{{ rp(Math.abs(data.jumlah)) }}</span></template></Column>
          <Column field="keterangan" header="Keterangan" />
        </DataTable>
      </div>
    </Dialog>

    <Dialog :visible="!!slip" modal :header="`Slip gaji ${slip?.k.nama}`" :style="{ width: '30rem' }" :breakpoints="{ '600px': '96vw' }" @update:visible="slip = null">
      <div v-if="slip" class="stack">
        <label>Periode</label>
        <DatePicker v-model="slip.rentang" selection-mode="range" date-format="dd/mm/yy" :manual-input="false" show-icon fluid />
        <template v-if="slip.k.jenis === 'harian'">
          <label for="hari">Hari kerja (gaji {{ rp(slip.k.gaji) }} per hari)</label>
          <InputNumber id="hari" v-model="slip.hari" :min="0" :max="31" fluid />
        </template>
        <div class="saldo"><span>Gaji pokok</span><strong>{{ rp(hitung.pokok) }}</strong></div>
        <label>Tunjangan</label>
        <MoneyInput v-model="slip.tunjangan" />
        <label>Bonus atau lembur</label>
        <MoneyInput v-model="slip.bonus" />
        <label>Potongan lain (mis. terlambat)</label>
        <MoneyInput v-model="slip.potongan" />
        <label>Potong bon (sisa bon {{ rp(slip.k.saldoBon) }})</label>
        <MoneyInput v-model="slip.potonganBon" />
        <label for="cat">Catatan</label>
        <InputText id="cat" v-model="slip.catatan" maxlength="255" fluid />
        <div class="saldo netto"><span>Gaji diterima (netto)</span><strong>{{ rp(hitung.netto) }}</strong></div>
        <small v-if="slip.potonganBon > slip.k.saldoBon" class="minus">Potongan bon melebihi sisa bon.</small>
      </div>
      <template #footer>
        <Button label="Simpan slip" icon="pi pi-check" :disabled="!slipValid() || slip?.potonganBon > slip?.k.saldoBon" @click="simpanSlip" />
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
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
}
.box {
  padding: 14px;
}
.filter {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}
.act {
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 2px;
}
.hint {
  color: var(--p-text-muted-color);
}
.minus {
  color: var(--p-red-500);
}
.saldo {
  display: flex;
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: 10px;
  background: color-mix(in srgb, var(--p-primary-color) 8%, var(--p-content-background));
}
.netto {
  font-size: 1.05em;
}
</style>
