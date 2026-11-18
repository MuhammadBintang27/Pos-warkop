<script setup>
import { computed } from 'vue'
import { usePengaturanStore } from '@/stores/pengaturan'
import { rp } from '@/utils/format'
import { useRun } from '@/composables/useRun'
import { cetakStruk } from '@/utils/cetak'

const props = defineProps({ pesanan: Object })
const visible = defineModel('visible')
const peng = usePengaturanStore()
const run = useRun()

const kembalian = computed(() => props.pesanan?.pembayaran.reduce((s, b) => s + b.kembalian, 0) ?? 0)
</script>

<template>
  <Dialog v-model:visible="visible" modal header="Pembayaran berhasil" :style="{ width: '26rem' }" :breakpoints="{ '600px': '96vw' }">
    <div v-if="pesanan" class="stack">
      <div class="ok"><i class="pi pi-check-circle" /> {{ pesanan.mejaNama || 'Bawa pulang' }} · #{{ pesanan.id }} lunas</div>
      <div class="row total"><span>Total</span><strong>{{ rp(pesanan.total) }}</strong></div>
      <div v-for="(b, i) in pesanan.pembayaran" :key="i" class="row">
        <span class="cap">{{ b.metode }}{{ b.pelanggan ? ' · ' + b.pelanggan : '' }}</span><span>{{ rp(b.jumlah) }}</span>
      </div>
      <div v-if="kembalian" class="kembali">
        <span>Kembalian</span>
        <strong>{{ rp(kembalian) }}</strong>
      </div>
    </div>
    <template #footer>
      <Button label="Cetak struk" icon="pi pi-print" severity="secondary" outlined @click="run(() => cetakStruk(pesanan, peng.settings))" />
      <Button label="Selesai" icon="pi pi-check" @click="visible = false" />
    </template>
  </Dialog>
</template>

<style scoped>
.ok {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: var(--p-green-600, #16a34a);
}
.row {
  display: flex;
  justify-content: space-between;
}
.total {
  font-size: 1.3rem;
}
.cap {
  text-transform: capitalize;
}
.kembali {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  border-radius: 12px;
  font-size: 1.4rem;
  background: color-mix(in srgb, var(--p-primary-color) 12%, var(--p-content-background));
}
.kembali strong {
  color: var(--accent-text);
}
</style>
