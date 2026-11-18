<script setup>
import { ref } from 'vue'
import { useKasirStore } from '@/stores/kasir'
import { rp } from '@/utils/format'

const visible = defineModel('visible')
const emit = defineEmits(['meja', 'pesanan', 'baru'])
const store = useKasirStore()
const atasNama = ref('')

const sum = (m) => m.pesanan.reduce((s, p) => s + p.total, 0)

function baru() {
  emit('baru', atasNama.value.trim())
  atasNama.value = ''
}
</script>

<template>
  <Dialog v-model:visible="visible" modal header="Pilih meja" :style="{ width: '46rem' }" :breakpoints="{ '760px': '96vw' }">
    <div class="meja-grid">
      <div
        v-for="m in store.denah.meja"
        :key="m.id"
        class="meja"
        :class="{ terisi: m.pesanan.length, current: m.pesanan.some((p) => p.id === store.pesanan?.id) || store.draft?.mejaId === m.id }"
      >
        <button class="main" @click="emit('meja', m)">
          <strong>{{ m.nama }}</strong>
          <small v-if="m.pesanan.length">{{ rp(sum(m)) }}</small>
          <small v-else>Kosong</small>
        </button>
        <div v-if="m.pesanan.length > 1" class="bills">
          <button v-for="p in m.pesanan" :key="p.id" @click="emit('pesanan', p.id)">#{{ p.id }} · {{ rp(p.total) }}</button>
        </div>
      </div>
    </div>
    <p v-if="!store.denah.meja.length" class="none">Belum ada meja. Tambahkan di Master Data › Meja.</p>

    <h4>Bawa pulang</h4>
    <div class="take">
      <InputText v-model="atasNama" maxlength="100" placeholder="Atas nama (opsional)" class="grow" @keyup.enter="baru" />
      <Button label="Pesanan baru" icon="pi pi-plus" @click="baru" />
    </div>
    <div v-if="store.denah.takeaway.length" class="bills open">
      <button v-for="p in store.denah.takeaway" :key="p.id" @click="emit('pesanan', p.id)">
        #{{ p.id }} {{ p.atasNama }} · {{ rp(p.total) }}
      </button>
    </div>
  </Dialog>
</template>

<style scoped>
.meja-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
  gap: 10px;
}
.meja {
  border: 1px solid var(--p-content-border-color);
  border-radius: 12px;
  overflow: hidden;
  background: var(--p-content-background);
}
.meja.terisi {
  background: color-mix(in srgb, var(--p-primary-color) 12%, var(--p-content-background));
  border-color: var(--p-primary-color);
}
.meja.current {
  outline: 3px solid var(--p-primary-color);
  outline-offset: 1px;
}
.main {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 14px 10px;
  border: 0;
  background: none;
  font: inherit;
  color: inherit;
  cursor: pointer;
}
.main small {
  color: var(--p-text-muted-color);
}
.meja.terisi .main small {
  color: var(--accent-text);
  font-weight: 700;
}
.bills {
  display: flex;
  flex-direction: column;
  border-top: 1px solid var(--p-content-border-color);
}
.bills.open {
  flex-direction: row;
  flex-wrap: wrap;
  gap: 8px;
  border: 0;
  margin-top: 10px;
}
.bills button {
  padding: 7px 10px;
  border: 1px solid var(--p-content-border-color);
  border-radius: 8px;
  background: var(--p-content-background);
  color: var(--accent-text);
  font: inherit;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
}
.none {
  color: var(--p-text-muted-color);
}
h4 {
  margin: 20px 0 8px;
}
.take {
  display: flex;
  gap: 8px;
}
.grow {
  flex: 1;
}
</style>
