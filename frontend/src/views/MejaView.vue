<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useCari } from '@/composables/useCari'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import { del, get, post, put } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { konfirmasi } from '@/utils/confirm'

const confirm = useConfirm()
const toast = useToast()
const run = useRun()

const meja = ref([])
const form = ref(null)
const gen = ref(null)

const muat = async () => (meja.value = (await get('/meja')).meja)

const escapeRe = (s) => s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
const nextNumber = (awalan) => {
  const re = new RegExp(`^${escapeRe(awalan.trim())} (\\d+)$`)
  return meja.value.reduce((max, m) => Math.max(max, Number(m.nama.match(re)?.[1] ?? 0)), 0) + 1
}
const bukaGen = () => (gen.value = { awalan: 'Meja', mulai: nextNumber('Meja'), jumlah: 10 })
const rentang = computed(() => {
  const { awalan, mulai, jumlah } = gen.value
  return `${awalan.trim()} ${mulai} sampai ${awalan.trim()} ${mulai + (jumlah || 1) - 1}`
})
const genValid = computed(() => gen.value?.awalan.trim() && gen.value.mulai >= 1 && gen.value.jumlah >= 1)

watch(() => gen.value?.awalan, (a) => a != null && (gen.value.mulai = nextNumber(a)))

async function buatBanyak() {
  const { awalan, mulai, jumlah } = gen.value
  let hasil
  if (await run(async () => (hasil = await post('/meja/generate', { awalan: awalan.trim(), mulai, jumlah })))) {
    gen.value = null
    const lewat = hasil.dilewati ? `, ${hasil.dilewati} dilewati karena sudah ada` : ''
    toast.add({ severity: 'success', summary: `${hasil.dibuat} meja dibuat${lewat}`, life: 3500 })
    await run(muat)
  }
}

async function simpan() {
  const { id, nama } = form.value
  if (await run(() => (id ? put(`/meja/${id}`, { nama: nama.trim() }) : post('/meja', { nama: nama.trim() })), 'Meja disimpan')) {
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
      <button class="pill" @click="form = { nama: '' }"><i class="pi pi-plus" /> Tambah meja</button>
    </PageHeader>

    <section class="panel box">
      <CariInput v-model="cariMeja" placeholder="Cari meja…" />
      <DataTable :value="mejaTampil" size="small" paginator :rows="20" empty-message="Belum ada meja">
        <Column field="nama" header="Nama" />
        <Column header="Status">
          <template #body="{ data }">
            <Tag v-if="data.pesanan.length" value="Terisi" severity="warn" />
            <Tag v-else value="Kosong" severity="success" />
          </template>
        </Column>
        <Column>
          <template #body="{ data }">
            <Button icon="pi pi-pencil" text severity="secondary" aria-label="Ubah" @click="form = { id: data.id, nama: data.nama }" />
            <Button icon="pi pi-trash" text severity="danger" aria-label="Hapus" @click="hapus(data)" />
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
      </div>
      <template #footer>
        <Button label="Simpan" icon="pi pi-check" :disabled="!form?.nama.trim()" @click="simpan" />
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
</style>
