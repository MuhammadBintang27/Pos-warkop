<script setup>
import { computed, ref, watch } from 'vue'
import { useRun } from '@/composables/useRun'
import { useKasirStore } from '@/stores/kasir'
import { printBatalDapur, sisaLokal } from '@/utils/print'

const ALASAN = ['Pelanggan batal', 'Salah input', 'Menu habis', 'Lainnya']

const props = defineProps({ item: Object, awal: { type: Number, default: 1 } })
const visible = defineModel('visible')
const store = useKasirStore()
const run = useRun()

const qty = ref(1)
const pilih = ref('')
const lain = ref('')
const sudahDibuat = ref(false)

const terkirim = computed(() => (props.item ? props.item.dikirim : store.pesanan?.items.some((i) => i.dikirim)))
const alasan = computed(() => (pilih.value === 'Lainnya' ? lain.value.trim() : pilih.value))
const valid = computed(() => !(props.item || terkirim.value) || !!alasan.value)

watch(visible, (v) => {
  if (!v) return
  qty.value = props.awal
  pilih.value = ''
  lain.value = ''
  sudahDibuat.value = false
})

async function batalkan() {
  let tiket
  const body = { alasan: alasan.value, sudahDibuat: sudahDibuat.value }
  const ok = await run(async () => {
    tiket = props.item ? await store.batalItem(props.item, { ...body, qty: qty.value }) : await store.batal(body)
  }, 'Dibatalkan')
  if (!ok) return
  visible.value = false
  const lokal = tiket && sisaLokal(tiket)
  if (lokal?.items.length) printBatalDapur(lokal)
}
</script>

<template>
  <Dialog
    v-model:visible="visible"
    modal
    :header="item ? `Batalkan ${item.nama}` : 'Batalkan pesanan'"
    :style="{ width: '28rem' }"
    :breakpoints="{ '600px': '96vw' }"
  >
    <div class="stack">
      <Message v-if="terkirim" severity="warn" size="small">
        Sudah dikirim ke dapur. Tiket BATAL akan dicetak supaya dapur berhenti membuatnya.
      </Message>

      <template v-if="item">
        <label>Jumlah yang dibatalkan (dari {{ item.qty }})</label>
        <InputNumber v-model="qty" show-buttons button-layout="horizontal" :min="1" :max="item.qty" input-class="qty" />
      </template>

      <label>Alasan{{ item || terkirim ? '' : ' (opsional)' }}</label>
      <SelectButton v-model="pilih" :options="ALASAN" :allow-empty="true" />
      <InputText v-if="pilih === 'Lainnya'" v-model="lain" maxlength="100" placeholder="Tulis alasan" fluid />

      <div v-if="terkirim" class="cek">
        <Checkbox v-model="sudahDibuat" input-id="dibuat" binary />
        <label for="dibuat">Sudah dibuat dapur (stok tidak dikembalikan)</label>
      </div>
    </div>

    <template #footer>
      <Button label="Kembali" severity="secondary" text @click="visible = false" />
      <Button label="Batalkan" icon="pi pi-trash" severity="danger" :disabled="!valid" @click="batalkan" />
    </template>
  </Dialog>
</template>

<style scoped>
.cek {
  display: flex;
  align-items: center;
  gap: 10px;
}
.cek label {
  font-weight: 500;
}
:deep(.qty) {
  width: 3.5rem;
  text-align: center;
}
</style>
