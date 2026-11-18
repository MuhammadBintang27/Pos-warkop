<script setup>
import { computed, ref, watch } from 'vue'
import { rp, tgl } from '@/utils/format'

const props = defineProps({ baris: Object, bills: Array })
const emit = defineEmits(['submit'])
const visible = defineModel('visible')

const dipilih = ref([])
watch(visible, (v) => v && (dipilih.value = []))

const terpilih = computed(() => props.bills?.filter((b) => dipilih.value.includes(b.pesananId)) ?? [])
const talangan = computed(() => terpilih.value.reduce((s, b) => s + b.kotor - b.diskon, 0))
const semua = computed(() => dipilih.value.length === props.bills?.length)

function toggleSemua() {
  dipilih.value = semua.value ? [] : props.bills.map((b) => b.pesananId)
}
</script>

<template>
  <Dialog v-model:visible="visible" modal header="Closing tenant" :style="{ width: '36rem' }" :breakpoints="{ '700px': '96vw' }">
    <div v-if="baris" class="stack">
      <div class="head">
        <strong>{{ baris.nama }}</strong>
        <small class="redup">{{ baris.tanggal.split('-').reverse().join('/') }}</small>
      </div>

      <Message severity="info" size="small">
        Bill di bawah sudah dikirim ke dapur tetapi belum dibayar pelanggan. Yang dicentang ikut dihitung dan disetor sekarang, artinya
        <b>pemilik menalangi</b> sampai pelanggan membayar. Yang tidak dicentang tetap tertunda dan masuk closing saat dibayar.
      </Message>

      <div class="bar">
        <Button :label="semua ? 'Lepas semua' : 'Pilih semua'" size="small" text @click="toggleSemua" />
        <small class="redup">{{ dipilih.length }} dari {{ bills?.length }} bill dipilih</small>
      </div>

      <div class="daftar">
        <label v-for="b in bills" :key="b.pesananId" class="bill" :class="{ on: dipilih.includes(b.pesananId) }">
          <Checkbox v-model="dipilih" :value="b.pesananId" />
          <div class="isi">
            <div class="atas">
              <strong>{{ b.mejaNama || 'Bawa pulang' }} · #{{ b.pesananId }}<template v-if="b.atasNama"> · {{ b.atasNama }}</template></strong>
              <strong>{{ rp(b.kotor - b.diskon) }}</strong>
            </div>
            <small class="redup">Dibuat {{ tgl(b.dibuatAt) }}</small>
            <div v-for="(i, n) in b.items" :key="n" class="item">
              <span>{{ i.qty }}x {{ i.nama }}<template v-if="i.opsi"> ({{ i.opsi }})</template></span>
              <span>{{ rp(i.harga * i.qty) }}</span>
            </div>
            <small v-if="b.diskon" class="redup">Sudah dikurangi bagian diskon bill {{ rp(b.diskon) }}</small>
          </div>
        </label>
      </div>

      <div class="ringkas">
        <span>Penjualan lunas</span><span>{{ rp(baris.kotor - baris.diskon) }}</span>
        <span>Talangan bill belum dibayar</span><span>{{ rp(talangan) }}</span>
      </div>
    </div>

    <template #footer>
      <Button label="Batal" severity="secondary" text @click="visible = false" />
      <Button
        :label="dipilih.length ? `Closing + ${dipilih.length} bill` : 'Closing tanpa talangan'"
        icon="pi pi-lock"
        @click="emit('submit', dipilih)"
      />
    </template>
  </Dialog>
</template>

<style scoped>
.head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}
.redup {
  color: var(--p-text-muted-color);
}
.bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.daftar {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 46vh;
  overflow-y: auto;
}
.bill {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  padding: 10px 12px;
  border: 1px solid var(--p-content-border-color);
  border-radius: 10px;
  cursor: pointer;
}
.bill.on {
  border-color: var(--p-primary-color);
  background: color-mix(in srgb, var(--p-primary-color) 7%, var(--p-content-background));
}
.isi {
  flex: 1;
  min-width: 0;
}
.atas,
.item {
  display: flex;
  justify-content: space-between;
  gap: 10px;
}
.item {
  font-size: 0.9em;
}
.ringkas {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 4px 12px;
  padding-top: 8px;
  border-top: 1px solid var(--p-content-border-color);
}
</style>
