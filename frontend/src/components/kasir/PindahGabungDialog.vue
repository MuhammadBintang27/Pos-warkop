<script setup>
import { computed } from 'vue'
import { useArea } from '@/composables/useArea'
import { useRun } from '@/composables/useRun'
import { useKasirStore } from '@/stores/kasir'
import { rp } from '@/utils/format'

const props = defineProps({ mode: { type: String, default: 'pindah' } })
const visible = defineModel('visible')
const store = useKasirStore()
const run = useRun()

const { tabs, aktif, pilih, cocok, namaArea } = useArea(computed(() => store.denah), { semua: true, kunci: 'kasir.area' })
const kosong = computed(() => store.denah.meja.filter((m) => !m.pesanan.length && cocok(m)))
// nama meja bisa sama di area berbeda, jadi area ikut ditulis bila sudah memakai area
const lokasi = (m) => (tabs.value.length ? `${m.nama} · ${namaArea(m.areaId)}` : m.nama)
const bills = computed(() => {
  const dariMeja = store.denah.meja.flatMap((m) => m.pesanan.map((p) => ({ ...p, lokasi: lokasi(m) })))
  const take = store.denah.takeaway.map((p) => ({ ...p, lokasi: 'Bawa pulang' }))
  return [...dariMeja, ...take].filter((p) => p.id !== store.pesanan?.id)
})

async function pindah(mejaId) {
  if (await run(() => store.pindah(mejaId), 'Pesanan dipindahkan')) visible.value = false
}

async function gabung(id) {
  if (await run(() => store.gabung(id), 'Bill digabung')) visible.value = false
}
</script>

<template>
  <Dialog
    v-model:visible="visible"
    modal
    :header="mode === 'pindah' ? 'Pindah meja' : 'Gabung bill'"
    :style="{ width: '30rem' }"
    :breakpoints="{ '600px': '96vw' }"
  >
    <template v-if="mode === 'pindah'">
      <p class="hint">Pilih meja kosong tujuan.</p>
      <SelectButton
        v-if="tabs.length"
        :model-value="aktif"
        :options="tabs"
        option-label="nama"
        option-value="id"
        :allow-empty="false"
        class="area"
        @update:model-value="pilih = $event"
      />
      <div class="opts">
        <Button v-for="m in kosong" :key="m.id" :label="m.nama" severity="secondary" outlined @click="pindah(m.id)" />
      </div>
      <p v-if="!kosong.length" class="hint">Tidak ada meja kosong. Gunakan gabung bill.</p>
    </template>
    <template v-else>
      <p class="hint">Semua item bill ini dipindah ke bill yang dipilih.</p>
      <ul class="list">
        <li v-for="p in bills" :key="p.id">
          <span>
            <strong>{{ p.lokasi }}</strong> · #{{ p.id }} <small v-if="p.atasNama">({{ p.atasNama }})</small>
          </span>
          <Button :label="rp(p.total)" severity="secondary" outlined size="small" @click="gabung(p.id)" />
        </li>
      </ul>
      <p v-if="!bills.length" class="hint">Tidak ada bill lain yang terbuka.</p>
    </template>
  </Dialog>
</template>

<style scoped>
.area {
  flex-wrap: wrap;
  margin-bottom: 12px;
}
.hint {
  margin-top: 0;
  color: var(--p-text-muted-color);
}
.opts {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.list {
  list-style: none;
  margin: 0;
  padding: 0;
}
.list li {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 10px 0;
  border-bottom: 1px solid var(--p-content-border-color);
}
</style>
