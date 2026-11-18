<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { get, post } from '@/api/http'
import { onEvent } from '@/composables/useEvents'
import { useRun } from '@/composables/useRun'

const run = useRun()

const tenant = ref([])
const kategori = ref([])
const q = ref('')

const kelompok = computed(() => {
  const term = q.value.trim().toLowerCase()
  const semua = kategori.value.flatMap((k) => k.menu.map((m) => ({ ...m, kategori: k.nama })))
  return tenant.value.map((t) => ({
    ...t,
    menu: semua.filter((m) => m.tenantId === t.id && (!term || m.nama.toLowerCase().includes(term))),
  }))
})

async function muat() {
  ;[tenant.value, kategori.value] = await Promise.all([get('/tenant'), get('/menu').then((r) => r.kategori)])
}

const ubahTenant = (t, buka) =>
  run(async () => {
    await post(`/tenant/${t.id}/buka`, { buka })
    await muat()
  })

const ubahMenu = (m, tersedia) =>
  run(async () => {
    await post(`/menu/${m.id}/stok`, { tersedia })
    await muat()
  })

const semua = (t, tersedia) =>
  run(async () => {
    await Promise.all(t.menu.filter((m) => m.tersedia !== tersedia).map((m) => post(`/menu/${m.id}/stok`, { tersedia })))
    await muat()
  }, tersedia ? 'Semua menu dinyalakan' : 'Semua menu dimatikan')

const offs = []
onMounted(() => {
  run(muat)
  offs.push(onEvent('menu', () => muat().catch(() => {})))
})
onBeforeUnmount(() => offs.forEach((off) => off()))
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-power-off" title="Ketersediaan" subtitle="Buka atau tutup tenant, nyalakan atau matikan menu" />

    <IconField>
      <InputIcon class="pi pi-search" />
      <InputText v-model="q" placeholder="Cari menu…" fluid />
    </IconField>

    <section v-for="t in kelompok" :key="t.id" class="panel tenant" :class="{ tutup: !t.buka }">
      <header class="head">
        <strong>{{ t.nama }}</strong>
        <Tag :value="t.buka ? 'Buka' : 'Tutup'" :severity="t.buka ? 'success' : 'danger'" />
        <span class="kanan">
          <Button label="Nyalakan semua" size="small" severity="secondary" text @click="semua(t, true)" />
          <Button label="Matikan semua" size="small" severity="secondary" text @click="semua(t, false)" />
          <label class="sw">
            <ToggleSwitch :model-value="t.buka" @update:model-value="ubahTenant(t, $event)" />
            <span>{{ t.buka ? 'Buka' : 'Tutup' }}</span>
          </label>
        </span>
      </header>

      <p v-if="!t.buka" class="info">
        Tenant tutup, jadi semua menunya tidak bisa dipesan. Pengaturan tiap menu di bawah tetap tersimpan dan berlaku lagi saat tenant dibuka.
      </p>

      <ul class="daftar">
        <li v-for="m in t.menu" :key="m.id" :class="{ mati: !m.tersedia }">
          <span class="nama">
            {{ m.nama }}
            <small>{{ m.kategori }}</small>
          </span>
          <Tag v-if="m.trackStok" :value="`Stok ${m.stok}`" :severity="m.stok > 0 ? 'secondary' : 'danger'" />
          <ToggleSwitch :model-value="m.tersedia" @update:model-value="ubahMenu(m, $event)" />
        </li>
      </ul>
      <p v-if="!t.menu.length" class="info">Tidak ada menu.</p>
    </section>
  </div>
</template>

<style scoped>
.gap {
  gap: 14px;
}
.tenant {
  padding: 14px;
}
.head {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
  font-size: 1.1rem;
}
.kanan {
  margin-left: auto;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
}
.sw {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  font-size: 0.95rem;
}
.info {
  color: var(--p-text-muted-color);
  margin: 10px 0 0;
}
.daftar {
  list-style: none;
  margin: 10px 0 0;
  padding: 0;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 8px;
}
.daftar li {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border: 1px solid var(--p-content-border-color);
  border-radius: 10px;
}
.daftar li.mati .nama {
  color: var(--p-text-muted-color);
  text-decoration: line-through;
}
.nama {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  font-weight: 600;
}
.nama small {
  font-weight: 400;
  color: var(--p-text-muted-color);
}
.tutup .daftar {
  opacity: 0.55;
}
</style>
