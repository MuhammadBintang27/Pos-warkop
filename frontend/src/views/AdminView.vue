<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useCari } from '@/composables/useCari'
import { useRouter } from 'vue-router'
import { get, post, put } from '@/api/http'
import { useRun } from '@/composables/useRun'

const LANGGANAN = {
  belum: { label: 'Belum Ada', severity: 'secondary' },
  aktif: { label: 'Aktif', severity: 'success' },
  segera: { label: 'Segera Berakhir', severity: 'warn' },
  berakhir: { label: 'Berakhir', severity: 'danger' },
}
const DURASI = [
  { label: '1 bulan', value: 1 },
  { label: '3 bulan', value: 3 },
  { label: '6 bulan', value: 6 },
  { label: '12 bulan', value: 12 },
]
const AWAL = [{ label: 'Belum ada', value: 0 }, ...DURASI]

const run = useRun()
const router = useRouter()

const list = ref([])
const baru = ref(null)
const edit = ref(null)
const perpanjang = ref(null)
const kredensial = ref(null)
const menu = ref()
const menuItems = ref([])
let kodeManual = false

const muat = async () => (list.value = await get('/admin/klien'))
const tgl = (v) => (v ? new Date(v).toLocaleDateString('id-ID', { dateStyle: 'medium' }) : '-')
const nomorBerikut = computed(() => list.value.reduce((n, k) => Math.max(n, k.id), 0) + 1)

const aktif = computed(() => list.value.filter((k) => k.aktif).length)
const totalPengguna = computed(() => list.value.reduce((n, k) => n + k.pengguna, 0))
const totalTransaksi = computed(() => list.value.reduce((n, k) => n + k.transaksi, 0))

const bukaBaru = () => {
  kodeManual = false
  baru.value = { nama: '', kode: '', alamat: '', bulan: 1, ownerNama: '', ownerUsername: 'owner', ownerPassword: '' }
}
watch(
  () => baru.value?.nama,
  (n) => {
    if (baru.value && !kodeManual) baru.value.kode = (n || '').toLowerCase().replace(/[^a-z0-9]/g, '').slice(0, 30)
  },
)
const validBaru = () => {
  const f = baru.value
  return f && f.nama.trim() && /^[a-z0-9]{3,30}$/.test(f.kode) && /^[a-z0-9._-]{3,30}$/.test((f.ownerUsername || 'owner').trim().toLowerCase()) && (!f.ownerPassword || f.ownerPassword.length >= 6)
}

async function simpanBaru() {
  const f = baru.value
  let hasil
  const ok = await run(
    async () =>
      (hasil = await post('/admin/klien', {
        kode: f.kode,
        nama: f.nama.trim(),
        alamat: f.alamat.trim(),
        bulan: f.bulan,
        ownerNama: f.ownerNama.trim(),
        ownerUsername: f.ownerUsername.trim().toLowerCase(),
        ownerPassword: f.ownerPassword,
      })),
  )
  if (!ok) return
  baru.value = null
  kredensial.value = { judul: 'Klien dibuat', ...hasil }
  await run(muat)
}

async function simpanEdit() {
  const { id, nama, alamat } = edit.value
  if (await run(() => put(`/admin/klien/${id}`, { nama: nama.trim(), alamat: alamat.trim() }), 'Data klien disimpan')) {
    edit.value = null
    await run(muat)
  }
}

// berakhir baru = dari tanggal berakhir saat ini, atau hari ini bila sudah lewat / belum ada
const berakhirBaru = computed(() => {
  const p = perpanjang.value
  if (!p) return ''
  const hariIni = new Date()
  hariIni.setHours(0, 0, 0, 0)
  const dasar = p.langgananSampai && new Date(p.langgananSampai) > hariIni ? new Date(p.langgananSampai) : hariIni
  const d = new Date(dasar)
  d.setMonth(d.getMonth() + p.bulan)
  return tgl(d)
})

async function simpanPerpanjang() {
  const { id, bulan } = perpanjang.value
  if (await run(() => post(`/admin/klien/${id}/langganan`, { bulan }), 'Langganan diperpanjang')) {
    perpanjang.value = null
    await run(muat)
  }
}

async function ubahAkses(k) {
  if (await run(() => post(`/admin/klien/${k.id}/akses`, { aktif: !k.aktif }), k.aktif ? 'Akses dimatikan' : 'Akses dibuka')) await run(muat)
}

async function reset(k) {
  let hasil
  if (await run(async () => (hasil = await post(`/admin/klien/${k.id}/reset-owner`)))) {
    kredensial.value = { judul: `Password owner ${k.nama} direset`, ...hasil }
  }
}

function bukaMenu(e, k) {
  menuItems.value = [
    k.aktif
      ? { label: 'Matikan Akses', icon: 'pi pi-lock', class: 'bahaya', command: () => ubahAkses(k) }
      : { label: 'Buka Akses', icon: 'pi pi-lock-open', command: () => ubahAkses(k) },
    { label: 'Perpanjang Langganan', icon: 'pi pi-calendar-plus', command: () => (perpanjang.value = { ...k, bulan: 1 }) },
    { label: 'Edit / Detail', icon: 'pi pi-pencil', command: () => (edit.value = { ...k }) },
    { separator: true },
    { label: 'Pengguna', icon: 'pi pi-users', command: () => router.push({ path: '/admin/pengguna', query: { klien: k.id } }) },
    { label: 'Reset password owner', icon: 'pi pi-key', command: () => reset(k) },
  ]
  menu.value.toggle(e)
}

onMounted(() => run(muat))
const { cari: cariKlien, hasil: klienTampil } = useCari(list, 'nama', 'kode', 'id')
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-building" title="Klien" subtitle="Usaha yang memakai sistem ini">
      <button class="pill" @click="bukaBaru"><i class="pi pi-plus" /> Tambah klien</button>
    </PageHeader>

    <div class="stats">
      <StatCard icon="pi pi-building" label="Klien dengan akses aktif" :value="`${aktif} dari ${list.length}`" />
      <StatCard icon="pi pi-users" label="Pengguna aktif" :value="String(totalPengguna)" />
      <StatCard icon="pi pi-receipt" label="Transaksi lunas" :value="String(totalTransaksi)" />
    </div>

    <section class="panel box">
      <CariInput v-model="cariKlien" placeholder="Cari nama, kode, atau ID klien…" />
      <DataTable paginator :rows="10" :rows-per-page-options="[10, 25, 50]" :always-show-paginator="false" :value="klienTampil" size="small" empty-message="Belum ada klien">
        <Column header="#"><template #body="{ index }">{{ index + 1 }}</template></Column>
        <Column field="id" header="ID Klien" />
        <Column header="Nama usaha">
          <template #body="{ data }">
            <strong>{{ data.nama }}</strong>
            <div class="kode">{{ data.kode }}</div>
          </template>
        </Column>
        <Column field="alamat" header="Alamat" />
        <Column header="Langganan">
          <template #body="{ data }"><Tag :value="LANGGANAN[data.langganan].label" :severity="LANGGANAN[data.langganan].severity" /></template>
        </Column>
        <Column header="Berakhir"><template #body="{ data }">{{ tgl(data.langgananSampai) }}</template></Column>
        <Column header="Akses">
          <template #body="{ data }"><Tag :value="data.aktif ? 'Aktif' : 'Nonaktif'" :severity="data.aktif ? 'success' : 'danger'" /></template>
        </Column>
        <Column header="Aksi">
          <template #body="{ data }">
            <Button label="Aksi" icon="pi pi-chevron-down" icon-pos="right" size="small" severity="secondary" variant="outlined" @click="bukaMenu($event, data)" />
          </template>
        </Column>
      </DataTable>
      <Menu ref="menu" :model="menuItems" popup />
    </section>

    <Dialog :visible="!!baru" modal header="Tambah klien" :style="{ width: '30rem' }" @update:visible="baru = null">
      <div v-if="baru" class="stack">
        <label for="nama">Nama usaha</label>
        <InputText id="nama" v-model="baru.nama" maxlength="100" fluid />
        <label for="kode">Kode</label>
        <InputText id="kode" v-model="baru.kode" maxlength="30" fluid @input="kodeManual = true" />
        <label for="alamat">Alamat</label>
        <InputText id="alamat" v-model="baru.alamat" maxlength="255" fluid />
        <label>Langganan awal</label>
        <Select v-model="baru.bulan" :options="AWAL" option-label="label" option-value="value" fluid />

        <div class="pisah">Akun owner</div>
        <label for="onama">Nama owner</label>
        <InputText id="onama" v-model="baru.ownerNama" maxlength="100" placeholder="Owner" fluid />
        <label for="ouser">Username</label>
        <InputGroup>
          <InputGroupAddon>{{ nomorBerikut }}-</InputGroupAddon>
          <InputText id="ouser" v-model="baru.ownerUsername" maxlength="30" />
        </InputGroup>
        <label for="opw">Password</label>
        <InputText id="opw" v-model="baru.ownerPassword" autocomplete="off" placeholder="Kosongkan untuk membuat acak" fluid />
        <small class="hint">Nomor klien dibuat otomatis (nomor terbesar yang ada + 1).</small>
      </div>
      <template #footer><Button label="Buat klien" icon="pi pi-check" :disabled="!validBaru()" @click="simpanBaru" /></template>
    </Dialog>

    <Dialog :visible="!!edit" modal header="Edit / Detail klien" :style="{ width: '26rem' }" @update:visible="edit = null">
      <div v-if="edit" class="stack">
        <label>ID Klien / kode</label>
        <InputText :model-value="`${edit.id} / ${edit.kode}`" disabled fluid />
        <label for="enama">Nama usaha</label>
        <InputText id="enama" v-model="edit.nama" maxlength="100" fluid />
        <label for="ealamat">Alamat</label>
        <InputText id="ealamat" v-model="edit.alamat" maxlength="255" fluid />
        <small class="hint">Langganan berakhir: {{ tgl(edit.langgananSampai) }}. Pengguna dan password dikelola di menu Pengguna.</small>
      </div>
      <template #footer><Button label="Simpan" icon="pi pi-check" :disabled="!edit?.nama.trim()" @click="simpanEdit" /></template>
    </Dialog>

    <Dialog :visible="!!perpanjang" modal header="Perpanjang langganan" :style="{ width: '24rem' }" @update:visible="perpanjang = null">
      <div v-if="perpanjang" class="stack">
        <strong>{{ perpanjang.nama }}</strong>
        <small class="hint">Berakhir saat ini: {{ tgl(perpanjang.langgananSampai) }}</small>
        <label>Tambah</label>
        <Select v-model="perpanjang.bulan" :options="DURASI" option-label="label" option-value="value" fluid />
        <small class="hint">Menjadi berakhir pada <b>{{ berakhirBaru }}</b>. Bila sudah lewat, dihitung dari hari ini.</small>
      </div>
      <template #footer><Button label="Perpanjang" icon="pi pi-calendar-plus" @click="simpanPerpanjang" /></template>
    </Dialog>

    <Dialog :visible="!!kredensial" modal :header="kredensial?.judul" :style="{ width: '26rem' }" @update:visible="kredensial = null">
      <div v-if="kredensial" class="stack">
        <label>Username owner</label>
        <InputText :model-value="kredensial.username" readonly fluid />
        <label>Password</label>
        <InputText v-if="kredensial.password" :model-value="kredensial.password" readonly fluid />
        <small class="hint">
          {{ kredensial.password ? 'Salin sekarang. Password tidak bisa ditampilkan lagi, tetapi bisa direset.' : 'Password sesuai yang Anda isi.' }}
        </small>
      </div>
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
.kode {
  color: var(--p-text-muted-color);
  font-size: 12px;
}
.pisah {
  margin-top: 8px;
  padding-top: 12px;
  border-top: 1px solid var(--p-content-border-color);
  font-weight: 700;
}
.hint {
  color: var(--p-text-muted-color);
}
:deep(.bahaya .p-menu-item-label),
:deep(.bahaya .p-menu-item-icon) {
  color: var(--p-red-500);
}
</style>
