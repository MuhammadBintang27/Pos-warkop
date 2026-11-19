<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { get, post } from '@/api/http'
import { load, save } from '@/utils/storage'
import { onEvent } from '@/composables/useEvents'
import { useRun } from '@/composables/useRun'

const run = useRun()

const tenant = ref([])
const kategori = ref([])
const q = ref('')
// pilihan tenant diingat per perangkat; 0 = semua
const pilihTenant = ref(load('ketersediaan.tenant', 0))
watch(pilihTenant, (v) => save('ketersediaan.tenant', v))
// tenant yang diingat bisa sudah dihapus: kembali ke semua supaya daftar tidak kosong
watch(tenant, (t) => {
  if (pilihTenant.value && !t.some((x) => x.id === pilihTenant.value)) pilihTenant.value = 0
})
const opsiTenant = computed(() => [{ id: 0, nama: 'Semua' }, ...tenant.value])

const kelompok = computed(() => {
  const term = q.value.trim().toLowerCase()
  const semua = kategori.value.flatMap((k) => k.menu.map((m) => ({ ...m, kategori: k.nama })))
  return tenant.value.filter((t) => !pilihTenant.value || t.id === pilihTenant.value).map((t) => ({
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

// ubah stok menu yang dilacak stoknya; stok 0 tidak otomatis mematikan menu, toggle tetap terpisah
const stokForm = ref(null)
const simpanStok = () =>
  run(async () => {
    const { id, tersedia, stok, trackStok } = stokForm.value
    await post(`/menu/${id}/stok`, { tersedia, trackStok, stok: stok ?? 0 })
    stokForm.value = null
    await muat()
  }, 'Stok disimpan')

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

    <SelectButton
      v-if="tenant.length > 1"
      v-model="pilihTenant"
      :options="opsiTenant"
      option-label="nama"
      option-value="id"
      :allow-empty="false"
      class="pilih-tenant"
    />
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
          <Button
            :label="m.trackStok ? `Stok ${m.stok}` : 'Stok'"
            :icon="m.trackStok ? 'pi pi-pencil' : 'pi pi-plus'"
            icon-pos="right"
            size="small"
            :severity="!m.trackStok ? 'secondary' : m.stok > 0 ? 'secondary' : 'danger'"
            :outlined="m.trackStok"
            :text="!m.trackStok"
            :aria-label="m.trackStok ? 'Ubah stok' : 'Mulai lacak stok'"
            @click="stokForm = { id: m.id, nama: m.nama, tersedia: m.tersedia, stok: m.stok, trackStok: true, awal: m.trackStok }"
          />
          <ToggleSwitch :model-value="m.tersedia" @update:model-value="ubahMenu(m, $event)" />
        </li>
      </ul>
      <p v-if="!t.menu.length" class="info">Tidak ada menu.</p>
    </section>

    <Dialog :visible="!!stokForm" modal :header="`Stok ${stokForm?.nama}`" :style="{ width: '22rem' }" @update:visible="stokForm = null">
      <div v-if="stokForm" class="stack">
        <div class="lacak">
          <ToggleSwitch v-model="stokForm.trackStok" input-id="klacak" />
          <label for="klacak">Lacak stok menu ini</label>
        </div>
        <small class="info">
          {{ stokForm.trackStok ? 'Stok berkurang tiap kali dipesan; pesanan ditolak bila stok tidak cukup.' : 'Menu bisa dipesan tanpa batas jumlah (hanya mengikuti tombol tersedia).' }}
        </small>
      </div>
      <div v-if="stokForm?.trackStok" class="stack">
        <label for="kstok">Stok sekarang</label>
        <InputNumber id="kstok" v-model="stokForm.stok" :min="0" show-buttons fluid autofocus @keydown.enter="simpanStok" />
        <div class="cepat">
          <Button v-for="n in [5, 10, 20]" :key="n" :label="`+${n}`" size="small" severity="secondary" outlined @click="stokForm.stok = (stokForm.stok || 0) + n" />
          <Button label="Habis (0)" size="small" severity="danger" outlined @click="stokForm.stok = 0" />
        </div>
      </div>
      <template #footer><Button label="Simpan" icon="pi pi-check" @click="simpanStok" /></template>
    </Dialog>
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
.pilih-tenant {
  align-self: flex-start;
  flex-wrap: wrap;
}
.lacak {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 600;
}
.cepat {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.tutup .daftar {
  opacity: 0.55;
}
</style>
