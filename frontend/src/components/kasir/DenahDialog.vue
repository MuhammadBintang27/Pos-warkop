<script setup>
import { computed, ref } from 'vue'
import { useArea } from '@/composables/useArea'
import { cocokNama } from '@/composables/useCari'
import { useKasirStore } from '@/stores/kasir'
import { rp } from '@/utils/format'

const visible = defineModel('visible')
const emit = defineEmits(['meja', 'pesanan', 'baru'])
const store = useKasirStore()
const atasNama = ref('')
// pelayan di lantai 2 langsung melihat meja lantai 2; pilihan diingat per HP
const { tabs, aktif, pilih, cocok } = useArea(computed(() => store.denah), { semua: true, kunci: 'kasir.area' })
const cari = ref('')
const cocokMeja = (m) => cocokNama(cari.value, m.nama, ...m.pesanan.map((p) => p.atasNama))
// saat mencari, semua area ikut dicari supaya pelanggan di area lain tetap ketemu
const mejaTampil = computed(() => store.denah.meja.filter(cari.value.trim() ? cocokMeja : cocok))
const takeTampil = computed(() => store.denah.takeaway.filter((p) => cocokNama(cari.value, p.atasNama, `#${p.id}`)))

const sum = (m) => m.pesanan.reduce((s, p) => s + p.total, 0)

function baru() {
  emit('baru', atasNama.value.trim())
  atasNama.value = ''
}
</script>

<template>
  <Dialog v-model:visible="visible" modal header="Pilih meja" :style="{ width: '46rem' }" :breakpoints="{ '760px': '96vw' }">
    <CariInput v-model="cari" placeholder="Cari meja / atas nama…" />
    <SelectButton
      v-if="tabs.length && !cari.trim()"
      :model-value="aktif"
      :options="tabs"
      option-label="nama"
      option-value="id"
      :allow-empty="false"
      class="area"
      @update:model-value="pilih = $event"
    />
    <div class="meja-grid">
      <div
        v-for="m in mejaTampil"
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
    <p v-if="!mejaTampil.length && cari.trim()" class="none">Tidak ada meja yang cocok.</p>
    <p v-else-if="!mejaTampil.length" class="none">Belum ada meja{{ tabs.length ? ' di area ini' : '' }}. Tambahkan di Master Data › Meja.</p>

    <h4>Bawa pulang</h4>
    <div class="take">
      <InputText v-model="atasNama" maxlength="100" placeholder="Atas nama (opsional)" class="grow" @keyup.enter="baru" />
      <Button label="Pesanan baru" icon="pi pi-plus" @click="baru" />
    </div>
    <div v-if="takeTampil.length" class="bills open">
      <button v-for="p in takeTampil" :key="p.id" @click="emit('pesanan', p.id)">
        #{{ p.id }} {{ p.atasNama }} · {{ rp(p.total) }}
      </button>
    </div>
  </Dialog>
</template>

<style scoped>
/* pilihan area tetap terlihat saat daftar meja di-scroll */
.area {
  position: sticky;
  top: 0;
  display: flex; /* selebar dialog, supaya latarnya menutup meja yang lewat di bawahnya */
  z-index: 2;
  flex-wrap: wrap;
  margin-bottom: 12px;
  padding-bottom: 8px;
  background: var(--p-dialog-background, var(--p-content-background));
  box-shadow: 0 6px 6px -6px rgb(0 0 0 / 0.15);
}
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
