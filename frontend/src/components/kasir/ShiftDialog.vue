<script setup>
import { ref, watch } from 'vue'
import { post } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { useKasirStore } from '@/stores/kasir'
import { rp, tgl } from '@/utils/format'

const visible = defineModel('visible')
const store = useKasirStore()
const run = useRun()

const nominal = ref(0)
const hasil = ref(null)

watch(visible, (v) => {
  if (v) {
    nominal.value = 0
    hasil.value = null
  }
})

async function buka() {
  if (await run(async () => (store.shift = await post('/shift/buka', { kasAwal: nominal.value || 0 })), 'Shift dibuka')) {
    visible.value = false
  } else {
    // kemungkinan orang lain sudah membuka shift lebih dulu: tampilkan shift yang sedang berjalan
    await store.loadShift().catch(() => {})
  }
}

async function tutup() {
  await run(async () => {
    hasil.value = await post('/shift/tutup', { kasHitung: nominal.value || 0 })
    store.shift = null
  })
}
</script>

<template>
  <Dialog
    v-model:visible="visible"
    modal
    :header="hasil ? 'Shift ditutup' : store.shift ? 'Tutup shift' : 'Buka shift'"
    :style="{ width: '26rem' }"
    :breakpoints="{ '600px': '96vw' }"
  >
    <div v-if="hasil" class="stack">
      <div class="row"><span>Kas awal</span><strong>{{ rp(hasil.kasAwal) }}</strong></div>
      <div class="row"><span>Kas menurut sistem</span><strong>{{ rp(hasil.kasSistem) }}</strong></div>
      <div class="row"><span>Kas hasil hitung</span><strong>{{ rp(hasil.kasHitung) }}</strong></div>
      <div class="row selisih" :class="{ ok: hasil.selisih === 0 }">
        <span>Selisih</span><strong>{{ hasil.selisih > 0 ? '+' : '' }}{{ rp(hasil.selisih) }}</strong>
      </div>
    </div>

    <div v-else-if="store.shift" class="stack">
      <small>Dibuka oleh {{ store.shift.kasir }}, {{ tgl(store.shift.bukaAt) }} · kas awal {{ rp(store.shift.kasAwal) }}</small>
      <small>Shift ini dipakai bersama semua kasir. Hitung seluruh uang tunai di laci.</small>
      <label>Uang tunai di laci saat ini</label>
      <MoneyInput v-model="nominal" />
      <small>Sistem menghitung kas yang seharusnya ada, lalu menampilkan selisihnya.</small>
    </div>

    <div v-else class="stack">
      <label>Kas awal (uang di laci)</label>
      <MoneyInput v-model="nominal" />
    </div>

    <template #footer>
      <Button v-if="hasil" label="Selesai" @click="visible = false" />
      <template v-else>
        <Button label="Batal" severity="secondary" text @click="visible = false" />
        <Button v-if="store.shift" label="Tutup shift" icon="pi pi-lock" severity="warn" @click="tutup" />
        <Button v-else label="Buka shift" icon="pi pi-lock-open" @click="buka" />
      </template>
    </template>
  </Dialog>
</template>

<style scoped>
.row {
  display: flex;
  justify-content: space-between;
}
.selisih {
  padding: 10px 12px;
  border-radius: 10px;
  background: color-mix(in srgb, var(--p-red-500, #ef4444) 14%, transparent);
}
.selisih.ok {
  background: color-mix(in srgb, var(--p-green-500, #22c55e) 16%, transparent);
}
</style>
