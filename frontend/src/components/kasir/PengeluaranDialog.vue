<script setup>
import { ref, watch } from 'vue'
import { get, post } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { useKasirStore } from '@/stores/kasir'
import { rp, tgl } from '@/utils/format'
import { KATEGORI_PENGELUARAN, labelKategori } from '@/utils/kategori'

const visible = defineModel('visible')
const store = useKasirStore()
const run = useRun()

const daftar = ref([])
const keterangan = ref('')
const jumlah = ref(0)
const kategori = ref('lainnya')

const muat = async () => (daftar.value = await get('/pengeluaran'))

watch(visible, (v) => {
  if (v) {
    keterangan.value = ''
    jumlah.value = 0
    kategori.value = 'lainnya'
    run(muat)
  }
})

async function simpan() {
  if (
    await run(async () => {
      await post('/pengeluaran', { keterangan: keterangan.value.trim(), jumlah: jumlah.value, kategori: kategori.value })
      await muat()
    }, 'Pengeluaran dicatat')
  ) {
    keterangan.value = ''
    jumlah.value = 0
  }
}
</script>

<template>
  <Dialog v-model:visible="visible" modal header="Kas keluar" :style="{ width: '32rem' }" :breakpoints="{ '600px': '96vw' }">
    <Message v-if="!store.shift" severity="warn" size="small">Buka shift dulu untuk mencatat pengeluaran.</Message>
    <div v-else class="stack">
      <label>Kategori</label>
      <Select v-model="kategori" :options="KATEGORI_PENGELUARAN" option-label="label" option-value="value" fluid />
      <label for="ket">Keterangan</label>
      <InputText id="ket" v-model="keterangan" maxlength="255" placeholder="mis. beli gas, es batu" fluid />
      <label>Jumlah</label>
      <MoneyInput v-model="jumlah" />
      <Button label="Catat" icon="pi pi-check" :disabled="!keterangan.trim() || !jumlah" @click="simpan" />

      <h4>Shift ini</h4>
      <p v-if="!daftar.length" class="none">Belum ada pengeluaran.</p>
      <ul class="list">
        <li v-for="p in daftar" :key="p.id">
          <span>{{ p.keterangan }}<br /><small>{{ labelKategori(p.kategori) }} · {{ tgl(p.at) }}<template v-if="p.oleh"> · {{ p.oleh }}</template></small></span>
          <strong>{{ rp(p.jumlah) }}</strong>
        </li>
      </ul>
    </div>
  </Dialog>
</template>

<style scoped>
h4 {
  margin: 12px 0 0;
}
.none,
small {
  color: var(--p-text-muted-color);
}
.list {
  list-style: none;
  margin: 0;
  padding: 0;
}
.list li {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 0;
  border-bottom: 1px solid var(--p-content-border-color);
}
</style>
