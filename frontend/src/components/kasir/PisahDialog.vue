<script setup>
import { computed, reactive, watch } from 'vue'
import { useRun } from '@/composables/useRun'
import { useKasirStore } from '@/stores/kasir'
import { rp } from '@/utils/format'

const visible = defineModel('visible')
const store = useKasirStore()
const run = useRun()
const jumlah = reactive({})

const items = computed(() => store.pesanan?.items ?? [])
const pilihan = computed(() => items.value.filter((i) => jumlah[i.id] > 0))
const sisaQty = computed(() => items.value.reduce((s, i) => s + i.qty - (jumlah[i.id] || 0), 0))
const subtotalBaru = computed(() => pilihan.value.reduce((s, i) => s + i.harga * jumlah[i.id], 0))

watch(visible, (v) => v && items.value.forEach((i) => (jumlah[i.id] = 0)))

async function pisah() {
  if (await run(() => store.pisah(pilihan.value.map((i) => ({ itemId: i.id, qty: jumlah[i.id] }))), 'Bill dipisah')) {
    visible.value = false
  }
}
</script>

<template>
  <Dialog v-model:visible="visible" modal header="Pisah bill" :style="{ width: '30rem' }" :breakpoints="{ '600px': '96vw' }">
    <p class="hint">Pilih jumlah item yang dipindah ke bill baru di meja yang sama.</p>
    <ul class="list">
      <li v-for="i in items" :key="i.id">
        <span>
          <strong>{{ i.nama }}</strong>
          <small v-if="i.opsi"> · {{ i.opsi }}</small>
          <br /><small>{{ rp(i.harga) }} · ada {{ i.qty }}</small>
        </span>
        <InputNumber v-model="jumlah[i.id]" show-buttons button-layout="horizontal" :min="0" :max="i.qty" input-class="qty" />
      </li>
    </ul>
    <div v-if="pilihan.length" class="ringkas">Bill baru: <strong>{{ rp(subtotalBaru) }}</strong></div>
    <template #footer>
      <Button label="Batal" severity="secondary" text @click="visible = false" />
      <Button label="Pisahkan" icon="pi pi-share-alt" :disabled="!pilihan.length || sisaQty === 0" @click="pisah" />
    </template>
  </Dialog>
</template>

<style scoped>
.hint {
  margin-top: 0;
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
  align-items: center;
  gap: 12px;
  padding: 10px 0;
  border-bottom: 1px solid var(--p-content-border-color);
}
.list small {
  color: var(--p-text-muted-color);
}
.ringkas {
  margin-top: 12px;
  text-align: right;
}
:deep(.qty) {
  width: 3rem;
  text-align: center;
}
</style>
