<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useCari } from '@/composables/useCari'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import { del, get, post, put } from '@/api/http'
import { SEMUA, UTAMA, useArea } from '@/composables/useArea'
import { useRun } from '@/composables/useRun'
import { useAuthStore } from '@/stores/auth'
import { konfirmasi } from '@/utils/confirm'

const confirm = useConfirm()
const toast = useToast()
const run = useRun()
const auth = useAuthStore()

const denah = ref({ area: [], meja: [] })
const form = ref(null)
const gen = ref(null)
const areaForm = ref(null)

const muat = async () => (denah.value = await get('/meja'))
const { tabs, aktif, pilih, cocok, namaArea } = useArea(denah, { semua: true, kunci: 'meja.area' })
const meja = computed(() => denah.value.meja.filter(cocok))
const adaArea = computed(() => denah.value.area.length > 0)
const opsiArea = computed(() => [{ id: null, nama: 'Utama (tanpa area)' }, ...denah.value.area])
// meja baru masuk ke area yang sedang dibuka
const areaBawaan = () => (aktif.value > UTAMA ? aktif.value : null)
const jumlahMeja = (a) => denah.value.meja.filter((m) => m.areaId === a.id).length

async function simpanArea() {
  const { id, nama, urutan } = areaForm.value
  const body = { nama: nama.trim(), urutan: urutan || 0 }
  if (await run(() => (id ? put(`/area/${id}`, body) : post('/area', body)), 'Area disimpan')) {
    areaForm.value = null
    await run(muat)
  }
}

function hapusArea(a) {
  konfirmasi(confirm, {
    header: 'Hapus area?',
    message: `Area "${a.nama}" akan dihapus. Area yang masih berisi meja tidak bisa dihapus.`,
    accept: async () => (await run(() => del(`/area/${a.id}`), 'Area dihapus')) && run(muat),
  })
}

const escapeRe = (s) => s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
const nextNumber = (awalan) => {
  const re = new RegExp(`^${escapeRe(awalan.trim())} (\\d+)$`)
  return denah.value.meja.reduce((max, m) => Math.max(max, Number(m.nama.match(re)?.[1] ?? 0)), 0) + 1
}
const bukaGen = () => (gen.value = { awalan: 'Meja', mulai: nextNumber('Meja'), jumlah: 10, areaId: areaBawaan() })
const rentang = computed(() => {
  const { awalan, mulai, jumlah } = gen.value
  return `${awalan.trim()} ${mulai} sampai ${awalan.trim()} ${mulai + (jumlah || 1) - 1}`
})
const genValid = computed(() => gen.value?.awalan.trim() && gen.value.mulai >= 1 && gen.value.jumlah >= 1)

watch(() => gen.value?.awalan, (a) => a != null && (gen.value.mulai = nextNumber(a)))

async function buatBanyak() {
  const { awalan, mulai, jumlah, areaId } = gen.value
  let hasil
  if (await run(async () => (hasil = await post('/meja/generate', { awalan: awalan.trim(), mulai, jumlah, areaId })))) {
    gen.value = null
    const lewat = hasil.dilewati ? `, ${hasil.dilewati} dilewati karena sudah ada` : ''
    toast.add({ severity: 'success', summary: `${hasil.dibuat} meja dibuat${lewat}`, life: 3500 })
    await run(muat)
  }
}

async function simpan() {
  const { id, nama, areaId } = form.value
  const body = { nama: nama.trim(), areaId }
  if (await run(() => (id ? put(`/meja/${id}`, body) : post('/meja', body)), 'Meja disimpan')) {
    form.value = null
    await run(muat)
  }
}

function hapus(m) {
  konfirmasi(confirm, {
    header: 'Hapus meja?',
    message: `"${m.nama}" akan dihapus. Meja yang masih punya pesanan terbuka tidak bisa dihapus.`,
    accept: async () => (await run(() => del(`/meja/${m.id}`), 'Meja dihapus')) && run(muat),
  })
}

onMounted(() => run(muat))
const { cari: cariMeja, hasil: mejaTampil } = useCari(meja, 'nama')
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-table" title="Meja" subtitle="Daftar meja yang tampil di kasir">
      <button class="pill" @click="bukaGen"><i class="pi pi-clone" /> Buat banyak</button>
      <button class="pill" @click="form = { nama: '', areaId: areaBawaan() }"><i class="pi pi-plus" /> Tambah meja</button>
    </PageHeader>

    <section class="panel box">
      <div class="area-head">
        <strong>Area / lantai</strong>
        <Button label="Tambah area" icon="pi pi-plus" size="small" text @click="areaForm = { nama: '', urutan: denah.area.length + 1 }" />
      </div>
      <p v-if="!adaArea" class="hint">
        Belum ada area, semua meja ada di satu denah. Tambahkan area (mis. Lantai 1, Lantai 2, Teras) bila tempatnya punya beberapa bagian.
      </p>
      <div v-else class="area-list">
        <span v-for="a in denah.area" :key="a.id" class="area-chip">
          {{ a.nama }} <small>{{ jumlahMeja(a) }} meja</small>
          <Button icon="pi pi-pencil" text rounded size="small" severity="secondary" aria-label="Ubah area" @click="areaForm = { id: a.id, nama: a.nama, urutan: a.urutan }" />
          <Button v-if="auth.isOwner" icon="pi pi-trash" text rounded size="small" severity="danger" aria-label="Hapus area" @click="hapusArea(a)" />
        </span>
      </div>
    </section>

    <section class="panel box">
      <SelectButton
        v-if="tabs.length"
        :model-value="aktif"
        :options="tabs"
        option-label="nama"
        option-value="id"
        :allow-empty="false"
        class="tabs"
        @update:model-value="pilih = $event"
      />
      <CariInput v-model="cariMeja" placeholder="Cari meja…" />
      <DataTable :value="mejaTampil" size="small" paginator :rows="20" empty-message="Belum ada meja">
        <Column field="nama" header="Nama" />
        <Column v-if="adaArea && aktif === SEMUA" header="Area"><template #body="{ data }">{{ namaArea(data.areaId) }}</template></Column>
        <Column header="Status">
          <template #body="{ data }">
            <Tag v-if="data.pesanan.length" value="Terisi" severity="warn" />
            <Tag v-else value="Kosong" severity="success" />
          </template>
        </Column>
        <Column>
          <template #body="{ data }">
            <Button icon="pi pi-pencil" text severity="secondary" aria-label="Ubah" @click="form = { id: data.id, nama: data.nama, areaId: data.areaId ?? null }" />
            <Button v-if="auth.isOwner" icon="pi pi-trash" text severity="danger" aria-label="Hapus" @click="hapus(data)" />
          </template>
        </Column>
      </DataTable>
    </section>

    <Dialog :visible="!!gen" modal header="Buat banyak meja" :style="{ width: '26rem' }" @update:visible="gen = null">
      <div v-if="gen" class="stack">
        <label for="awalan">Awalan nama</label>
        <InputText id="awalan" v-model="gen.awalan" maxlength="40" fluid />
        <label for="mulai">Mulai dari nomor</label>
        <InputNumber id="mulai" v-model="gen.mulai" :min="1" :max="9999" fluid />
        <label for="jumlah">Jumlah meja (maks. 200)</label>
        <InputNumber id="jumlah" v-model="gen.jumlah" show-buttons :min="1" :max="200" fluid />
        <template v-if="adaArea">
          <label>Area</label>
          <Select v-model="gen.areaId" :options="opsiArea" option-label="nama" option-value="id" fluid />
          <small class="hint">Nama meja harus unik di semua area, jadi pakai awalan berbeda per lantai (mis. "L2 Meja").</small>
        </template>
        <Message v-if="genValid" severity="secondary" size="small" variant="simple">
          Akan dibuat: {{ rentang }}. Nama yang sudah ada dilewati.
        </Message>
      </div>
      <template #footer>
        <Button label="Buat meja" icon="pi pi-check" :disabled="!genValid" @click="buatBanyak" />
      </template>
    </Dialog>

    <Dialog :visible="!!form" modal :header="form?.id ? 'Ubah meja' : 'Tambah meja'" :style="{ width: '24rem' }" @update:visible="form = null">
      <div v-if="form" class="stack">
        <label for="nama">Nama meja</label>
        <InputText id="nama" v-model="form.nama" maxlength="50" placeholder="mis. Meja 1, Lesehan A" autofocus fluid @keyup.enter="simpan" />
        <template v-if="adaArea">
          <label>Area</label>
          <Select v-model="form.areaId" :options="opsiArea" option-label="nama" option-value="id" fluid />
          <small v-if="form.id" class="hint">Pindah area membuat letak meja di denah disusun ulang di area barunya.</small>
        </template>
      </div>
      <template #footer>
        <Button label="Simpan" icon="pi pi-check" :disabled="!form?.nama.trim()" @click="simpan" />
      </template>
    </Dialog>

    <Dialog :visible="!!areaForm" modal :header="areaForm?.id ? 'Ubah area' : 'Tambah area'" :style="{ width: '24rem' }" @update:visible="areaForm = null">
      <div v-if="areaForm" class="stack">
        <label for="anama">Nama area</label>
        <InputText id="anama" v-model="areaForm.nama" maxlength="50" placeholder="mis. Lantai 2, Teras, VIP" autofocus fluid @keyup.enter="simpanArea" />
        <label for="aurut">Urutan tampil</label>
        <InputNumber id="aurut" v-model="areaForm.urutan" :min="0" :max="999" fluid />
      </div>
      <template #footer>
        <Button label="Simpan" icon="pi pi-check" :disabled="!areaForm?.nama.trim()" @click="simpanArea" />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.gap {
  gap: 14px;
}
.box {
  padding: 14px;
}
.area-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.area-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}
.area-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 4px 2px 12px;
  border: 1px solid var(--p-content-border-color);
  border-radius: 999px;
  font-weight: 600;
}
.area-chip small {
  color: var(--p-text-muted-color);
  font-weight: 400;
}
.hint {
  color: var(--p-text-muted-color);
  margin: 6px 0 0;
}
.tabs {
  flex-wrap: wrap;
  margin-bottom: 10px;
}
</style>
