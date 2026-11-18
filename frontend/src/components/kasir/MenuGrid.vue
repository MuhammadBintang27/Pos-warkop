<script setup>
import { computed, ref } from 'vue'
import { useRun } from '@/composables/useRun'
import { useKasirStore } from '@/stores/kasir'
import { hargaTampil } from '@/utils/format'

const emit = defineEmits(['pick'])
const store = useKasirStore()
const run = useRun()

const q = ref('')
const aktif = ref(0)
const tenantAktif = ref(0)
const opsiTenant = computed(() => [{ id: 0, nama: 'Semua tenant' }, ...store.tenant])

const habis = (m) => !m.tersedia || (m.trackStok && m.stok <= 0) || m.porsiBahan === 0
const tutup = (m) => store.tenant.some((t) => t.id === m.tenantId && !t.buka)

const menus = computed(() => {
  const term = q.value.trim().toLowerCase()
  return store.kategori
    .filter((k) => !aktif.value || k.id === aktif.value)
    .flatMap((k) => k.menu)
    .filter((m) => !tenantAktif.value || m.tenantId === tenantAktif.value)
    .filter((m) => !term || m.nama.toLowerCase().includes(term))
})
</script>

<template>
  <section class="panel grid-panel">
    <div class="tools">
      <IconField class="grow">
        <InputIcon class="pi pi-search" />
        <InputText v-model="q" placeholder="Cari menu…" fluid />
      </IconField>
      <Select
        v-if="store.tenant.length > 1"
        v-model="tenantAktif"
        :options="opsiTenant"
        option-label="nama"
        option-value="id"
        class="tenant"
      />
      <Button icon="pi pi-refresh" variant="outlined" aria-label="Muat ulang" @click="run(store.loadMenu)" />
    </div>

    <div class="chips">
      <button class="chip-btn" :class="{ on: !aktif }" @click="aktif = 0">Semua</button>
      <button v-for="k in store.kategori" :key="k.id" class="chip-btn" :class="{ on: aktif === k.id }" @click="aktif = k.id">
        {{ k.nama }}
      </button>
    </div>

    <div v-if="menus.length" class="cards">
      <button v-for="m in menus" :key="m.id" class="card" :disabled="habis(m) || tutup(m)" @click="emit('pick', m)">
        <strong>{{ m.nama }}</strong>
        <span class="price">{{ hargaTampil(m) }}</span>
        <span class="meta">
          <i v-if="m.varian.length" class="pi pi-sliders-h" title="Ada pilihan" />
          <small v-if="m.trackStok && m.stok > 0">Sisa {{ m.stok }}</small>
        </span>
        <span v-if="habis(m) || tutup(m)" class="sold">{{ tutup(m) ? 'Tutup' : 'Habis' }}</span>
      </button>
    </div>
    <div v-else class="empty">
      <i class="pi pi-inbox" />
      <span>{{ q ? 'Menu tidak ditemukan' : 'Belum ada menu' }}</span>
    </div>
  </section>
</template>

<style scoped>
.grid-panel {
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 0;
  overflow: hidden;
}
.tools {
  flex: none;
  display: flex;
  gap: 8px;
}
.grow {
  flex: 1;
}
.tenant {
  min-width: 160px;
}
.chips {
  flex: none;
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 6px;
  scrollbar-width: thin;
}
.chip-btn {
  flex: none;
  white-space: nowrap;
  padding: 7px 14px;
  border-radius: 999px;
  border: 1px solid var(--p-content-border-color);
  background: var(--p-content-background);
  color: var(--p-text-color);
  font: inherit;
  font-weight: 600;
  cursor: pointer;
}
.chip-btn.on {
  background: var(--p-primary-color);
  border-color: var(--p-primary-color);
  color: var(--p-primary-contrast-color);
}
.cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 10px;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  align-content: start;
  padding: 2px;
}
.card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-height: 96px;
  padding: 12px;
  text-align: left;
  border: 1px solid var(--p-content-border-color);
  border-radius: 12px;
  background: var(--p-content-background);
  color: var(--p-text-color);
  font: inherit;
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s;
}
.card:hover:not(:disabled) {
  border-color: var(--p-primary-color);
  box-shadow: 0 2px 10px color-mix(in srgb, var(--p-primary-color) 22%, transparent);
}
.card:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}
.price {
  color: var(--accent-text);
  font-weight: 700;
}
.meta {
  margin-top: auto;
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--p-text-muted-color);
}
.sold {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 2px 8px;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 700;
  background: var(--p-red-500, #ef4444);
  color: #fff;
}
.empty {
  flex: 1;
  display: grid;
  place-content: center;
  justify-items: center;
  gap: 8px;
  color: var(--p-text-muted-color);
}
.empty .pi {
  font-size: 2.2rem;
}
</style>
