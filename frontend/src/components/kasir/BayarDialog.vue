<script setup>
import { computed, ref, watch } from 'vue'
import { get, post } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { useKasirStore } from '@/stores/kasir'
import { rp } from '@/utils/format'

const METODE = [
  { label: 'Tunai', value: 'tunai' },
  { label: 'QRIS', value: 'qris' },
  { label: 'Transfer', value: 'transfer' },
  { label: 'Kasbon', value: 'kasbon' },
]
// pecahan uang yang biasa disodorkan pelanggan: jumlah dibulatkan ke atas per pecahan, mis. 167.000 -> 170rb, 180rb, 200rb
const PECAHAN = [5000, 10000, 20000, 50000, 100000]
const cepat = (jumlah) =>
  [...new Set(PECAHAN.map((p) => Math.ceil((jumlah + 1) / p) * p))].sort((a, b) => a - b).slice(0, 4)

const visible = defineModel('visible')
const emit = defineEmits(['paid'])
const store = useKasirStore()
const run = useRun()

const baris = ref([])
const pelanggan = ref([])
const namaBaru = ref('')

const total = computed(() => store.pesanan?.total ?? 0)
const sisa = computed(() => total.value - baris.value.reduce((s, b) => s + (b.jumlah || 0), 0))
const adaKasbon = computed(() => baris.value.some((b) => b.metode === 'kasbon'))
const kembalian = (b) => (b.metode === 'tunai' && b.diterima > b.jumlah ? b.diterima - b.jumlah : 0)
const valid = computed(
  () =>
    sisa.value === 0 &&
    baris.value.every(
      (b) =>
        b.jumlah > 0 &&
        (b.metode !== 'kasbon' || b.pelangganId) &&
        (b.metode !== 'tunai' || !b.diterima || b.diterima >= b.jumlah),
    ),
)

const loadPelanggan = async () => (pelanggan.value = await get('/pelanggan'))

watch(visible, (v) => {
  if (!v) return
  baris.value = [{ metode: 'tunai', jumlah: total.value, diterima: 0, pelangganId: null }]
  run(loadPelanggan)
})

const tambah = () => baris.value.push({ metode: 'qris', jumlah: 0, diterima: 0, pelangganId: null })

async function pelangganBaru() {
  const nama = namaBaru.value.trim()
  if (!nama) return
  const ok = await run(async () => {
    await post('/pelanggan', { nama, telp: '' })
    await loadPelanggan()
  })
  if (!ok) return
  const dibuat = [...pelanggan.value].reverse().find((p) => p.nama === nama)
  const target = baris.value.find((b) => b.metode === 'kasbon' && !b.pelangganId)
  if (target && dibuat) target.pelangganId = dibuat.id
  namaBaru.value = ''
}

async function bayar() {
  let lunas
  const payload = baris.value.map((b) => ({
    metode: b.metode,
    jumlah: b.jumlah,
    diterima: b.metode === 'tunai' ? b.diterima || 0 : 0,
    pelangganId: b.metode === 'kasbon' ? b.pelangganId : null,
  }))
  if (await run(async () => (lunas = await store.bayar(payload)))) {
    visible.value = false
    emit('paid', lunas)
  }
}
</script>

<template>
  <Dialog v-model:visible="visible" modal header="Pembayaran" :style="{ width: '34rem' }" :breakpoints="{ '600px': '96vw' }">
    <div class="stack">
      <div class="tagihan">
        <span>Total tagihan</span>
        <strong>{{ rp(total) }}</strong>
      </div>

      <div v-for="(b, idx) in baris" :key="idx" class="stack baris">
        <div class="top">
          <SelectButton v-model="b.metode" :options="METODE" option-label="label" option-value="value" :allow-empty="false" />
          <Button v-if="baris.length > 1" icon="pi pi-times" text severity="danger" aria-label="Hapus baris" @click="baris.splice(idx, 1)" />
        </div>

        <label>Jumlah</label>
        <div class="inline">
          <MoneyInput v-model="b.jumlah" />
          <Button v-if="sisa > 0" label="Isi sisa" severity="secondary" outlined @click="b.jumlah += sisa" />
        </div>

        <template v-if="b.metode === 'tunai'">
          <label>Uang diterima</label>
          <MoneyInput v-model="b.diterima" />
          <div class="cepat">
            <Button label="Uang pas" severity="secondary" size="small" @click="b.diterima = 0" />
            <Button v-for="n in cepat(b.jumlah || 0)" :key="n" :label="rp(n)" severity="secondary" size="small" @click="b.diterima = n" />
          </div>
          <div v-if="kembalian(b)" class="kembali">Kembalian <strong>{{ rp(kembalian(b)) }}</strong></div>
        </template>

        <template v-if="b.metode === 'kasbon'">
          <label>Pelanggan</label>
          <Select
            v-model="b.pelangganId"
            :options="pelanggan"
            option-label="nama"
            option-value="id"
            placeholder="Pilih pelanggan"
            filter
            fluid
          />
        </template>
      </div>

      <div v-if="adaKasbon" class="inline">
        <InputText v-model="namaBaru" placeholder="Nama pelanggan baru" fluid @keyup.enter="pelangganBaru" />
        <Button label="Tambah" severity="secondary" outlined @click="pelangganBaru" />
      </div>

      <Button label="Tambah metode" icon="pi pi-plus" severity="secondary" text @click="tambah" />
      <Message v-if="sisa !== 0" :severity="sisa > 0 ? 'warn' : 'error'" size="small" variant="simple">
        {{ sisa > 0 ? `Kurang ${rp(sisa)}` : `Lebih ${rp(-sisa)}` }}
      </Message>
    </div>

    <template #footer>
      <Button label="Batal" severity="secondary" text @click="visible = false" />
      <Button label="Bayar" icon="pi pi-check" :disabled="!valid" @click="bayar" />
    </template>
  </Dialog>
</template>

<style scoped>
.tagihan {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  border-radius: 12px;
  background: color-mix(in srgb, var(--p-primary-color) 10%, var(--p-content-background));
  font-size: 1.3rem;
}
.tagihan strong {
  color: var(--accent-text);
}
.baris {
  padding: 12px;
  border: 1px solid var(--p-content-border-color);
  border-radius: 12px;
}
.top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.inline {
  display: flex;
  gap: 8px;
}
.cepat {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.kembali {
  padding: 8px 12px;
  border-radius: 8px;
  background: color-mix(in srgb, var(--p-green-500, #22c55e) 16%, transparent);
}
</style>
