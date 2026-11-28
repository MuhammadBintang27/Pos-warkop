<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { rp } from '@/utils/format'

const props = defineProps({ menu: Object })
const visible = defineModel('visible')
const emit = defineEmits(['add'])

const pilih = reactive({})
const qty = ref(1)
const catatan = ref('')

watch(() => props.menu, (m) => {
  Object.keys(pilih).forEach((k) => delete pilih[k])
  ;(m?.varian || []).forEach((g) => (pilih[g.id] = g.multi ? [] : 0))
  qty.value = 1
  catatan.value = ''
})

const varian = computed(() => props.menu?.varian || [])
const opsiIds = computed(() => Object.values(pilih).flat().filter(Boolean))
const hargaSatuan = computed(() => {
  let dasar = props.menu?.harga || 0
  let tambah = 0
  for (const g of varian.value) {
    for (const o of g.opsi.filter((x) => opsiIds.value.includes(x.id))) {
      if (g.gantiHarga) dasar = o.tambahan
      else tambah += o.tambahan
    }
  }
  return dasar + tambah
})
const total = computed(() => hargaSatuan.value * (qty.value || 1))
const lengkap = computed(() =>
  varian.value.every((g) => !g.wajib || (g.multi ? pilih[g.id]?.length > 0 : pilih[g.id] > 0)),
)

function submit() {
  emit('add', { menuId: props.menu.id, qty: qty.value || 1, opsiIds: opsiIds.value, catatan: catatan.value })
  visible.value = false
}
</script>

<template>
  <Dialog v-model:visible="visible" modal :header="menu?.nama" :style="{ width: '28rem' }" :breakpoints="{ '600px': '95vw' }">
    <div v-if="menu" class="stack">
      <div v-for="g in varian" :key="g.id" class="stack grp">
        <label>
          {{ g.nama }}
          <small class="hint">{{ g.wajib ? 'wajib' : 'opsional' }}{{ g.multi ? ', boleh lebih dari satu' : '' }}</small>
        </label>
        <template v-if="!g.multi">
          <label v-if="!g.wajib" class="opt"><RadioButton v-model="pilih[g.id]" :value="0" /> Tanpa</label>
          <label v-for="o in g.opsi" :key="o.id" class="opt">
            <RadioButton v-model="pilih[g.id]" :value="o.id" /> {{ o.nama }}
            <small v-if="g.gantiHarga" class="plus">{{ rp(o.tambahan) }}</small>
            <small v-else-if="o.tambahan" class="plus">+{{ rp(o.tambahan) }}</small>
          </label>
        </template>
        <template v-else>
          <label v-for="o in g.opsi" :key="o.id" class="opt">
            <Checkbox v-model="pilih[g.id]" :value="o.id" /> {{ o.nama }}
            <small v-if="o.tambahan" class="plus">+{{ rp(o.tambahan) }}</small>
          </label>
        </template>
      </div>

      <label for="catatan">Catatan</label>
      <InputText id="catatan" v-model="catatan" maxlength="255" placeholder="mis. tidak pedas" fluid />
    </div>

    <template #footer>
      <InputNumber v-model="qty" show-buttons button-layout="horizontal" :min="1" :max="999" input-class="qty-input" />
      <Button :label="`Tambah · ${rp(total)}`" icon="pi pi-plus" :disabled="!lengkap" @click="submit" />
    </template>
  </Dialog>
</template>

<style scoped>
.grp {
  padding-bottom: 6px;
}
.opt {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 500;
  cursor: pointer;
}
.hint {
  font-weight: 400;
  color: var(--p-text-muted-color);
}
.plus {
  margin-left: auto;
  color: var(--accent-text);
  font-weight: 600;
}
:deep(.qty-input) {
  width: 3.5rem;
  text-align: center;
}
</style>
