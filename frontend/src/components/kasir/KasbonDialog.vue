<script setup>
import { ref, watch } from 'vue'
import { get, post } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { useKasirStore } from '@/stores/kasir'
import { rp } from '@/utils/format'

const METODE = [
  { label: 'Tunai', value: 'tunai' },
  { label: 'QRIS', value: 'qris' },
  { label: 'Transfer', value: 'transfer' },
]

const visible = defineModel('visible')
const store = useKasirStore()
const run = useRun()

const daftar = ref([])
const nama = ref('')
const telp = ref('')
const lunasi = ref(null)

const muat = async () => (daftar.value = await get('/pelanggan'))

watch(visible, (v) => {
  if (v) {
    lunasi.value = null
    run(muat)
  }
})

async function tambah() {
  if (
    await run(async () => {
      await post('/pelanggan', { nama: nama.value.trim(), telp: telp.value.trim() })
      await muat()
    }, 'Pelanggan ditambahkan')
  ) {
    nama.value = ''
    telp.value = ''
  }
}

async function bayar() {
  const { p, jumlah, metode } = lunasi.value
  if (await run(async () => {
    await post(`/pelanggan/${p.id}/lunasi`, { jumlah, metode })
    await muat()
  }, 'Pelunasan dicatat')) lunasi.value = null
}
</script>

<template>
  <Dialog v-model:visible="visible" modal header="Kasbon pelanggan" :style="{ width: '38rem' }" :breakpoints="{ '700px': '96vw' }">
    <div class="inline">
      <InputText v-model="nama" placeholder="Nama pelanggan" fluid @keyup.enter="tambah" />
      <InputText v-model="telp" placeholder="No. HP (opsional)" fluid />
      <Button label="Tambah" icon="pi pi-plus" :disabled="!nama.trim()" @click="tambah" />
    </div>

    <DataTable :value="daftar" size="small" class="tabel" empty-message="Belum ada pelanggan">
      <Column field="nama" header="Nama" />
      <Column field="telp" header="HP" />
      <Column header="Sisa utang">
        <template #body="{ data }">
          <strong :class="{ utang: data.sisaUtang > 0 }">{{ rp(data.sisaUtang) }}</strong>
        </template>
      </Column>
      <Column>
        <template #body="{ data }">
          <Button
            v-if="data.sisaUtang > 0"
            label="Bayar"
            size="small"
            outlined
            @click="lunasi = { p: data, jumlah: data.sisaUtang, metode: 'tunai' }"
          />
        </template>
      </Column>
    </DataTable>

    <div v-if="lunasi" class="stack bayar">
      <strong>Pelunasan {{ lunasi.p.nama }}</strong>
      <SelectButton v-model="lunasi.metode" :options="METODE" option-label="label" option-value="value" :allow-empty="false" />
      <MoneyInput v-model="lunasi.jumlah" />
      <small v-if="!store.shift" class="warn">Buka shift dulu untuk mencatat pelunasan.</small>
      <div class="inline">
        <Button label="Batal" severity="secondary" text @click="lunasi = null" />
        <Button label="Catat pelunasan" icon="pi pi-check" :disabled="!store.shift || !lunasi.jumlah || lunasi.jumlah > lunasi.p.sisaUtang" @click="bayar" />
      </div>
    </div>
  </Dialog>
</template>

<style scoped>
.inline {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.tabel {
  margin-top: 14px;
}
.utang {
  color: var(--p-red-500, #ef4444);
}
.bayar {
  margin-top: 14px;
  padding: 12px;
  border: 1px solid var(--p-primary-color);
  border-radius: 12px;
}
.warn {
  color: var(--p-orange-600, #ea580c);
}
</style>
