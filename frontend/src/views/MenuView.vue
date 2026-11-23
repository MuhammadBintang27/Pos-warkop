<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useConfirm } from 'primevue/useconfirm'
import { del, get, post, put } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { konfirmasi } from '@/utils/confirm'
import { hargaTampil, rp } from '@/utils/format'
import { load, save } from '@/utils/storage'

const confirm = useConfirm()
const run = useRun()

const kategori = ref([])
const aktif = ref(0)
const q = ref('')
const kat = ref(null)
const form = ref(null)

const tenant = ref([])
const bahan = ref([])
// filter tenant diingat per perangkat; 0 = semua tenant
const pilihTenant = ref(load('menu.tenant', 0))
watch(pilihTenant, (v) => save('menu.tenant', v))
const opsiTenant = computed(() => [{ id: 0, nama: 'Semua tenant' }, ...tenant.value])
// tenant yang diingat bisa sudah dihapus: kembali ke semua supaya daftar tidak kosong
watch(tenant, (t) => {
  if (pilihTenant.value && !t.some((x) => x.id === pilihTenant.value)) pilihTenant.value = 0
})
const milikPilihan = (m) => !pilihTenant.value || m.tenantId === pilihTenant.value
const jumlahDi = (k) => k.menu.filter(milikPilihan).length

async function muat() {
  ;[kategori.value, tenant.value, bahan.value] = await Promise.all([get('/menu').then((r) => r.kategori), get('/tenant'), get('/bahan')])
}

const satuan = (id) => bahan.value.find((b) => b.id === id)?.satuan ?? ''
const namaTenant = (id) => tenant.value.find((t) => t.id === id)?.nama ?? '—'
const total = computed(() => kategori.value.reduce((s, k) => s + jumlahDi(k), 0))
const rows = computed(() => {
  const term = q.value.trim().toLowerCase()
  return kategori.value
    .filter((k) => !aktif.value || k.id === aktif.value)
    .flatMap((k) => k.menu.filter(milikPilihan).map((m) => ({ ...m, kategori: k.nama })))
    .filter((m) => !term || m.nama.toLowerCase().includes(term))
})

async function simpanKat() {
  const { id, nama, urutan } = kat.value
  const body = { nama: nama.trim(), urutan: urutan || 0 }
  if (await run(() => (id ? put(`/kategori/${id}`, body) : post('/kategori', body)), 'Kategori disimpan')) {
    kat.value = null
    await run(muat)
  }
}

function hapusKat(k) {
  konfirmasi(confirm, {
    header: 'Hapus kategori?',
    message: `Kategori "${k.nama}" akan dihapus. Kategori yang masih punya menu tidak bisa dihapus.`,
    accept: async () => {
      if (await run(() => del(`/kategori/${k.id}`), 'Kategori dihapus')) {
        if (aktif.value === k.id) aktif.value = 0
        await run(muat)
      }
    },
  })
}

function bukaMenu(m) {
  form.value = m
    ? {
        id: m.id, tenantId: m.tenantId, kategoriId: m.kategoriId, nama: m.nama, harga: m.harga, hargaModal: m.hargaModal, komisi: m.komisi, tersedia: m.tersedia, trackStok: m.trackStok, stok: m.stok,
        resep: m.resep.map((r) => ({ bahanId: r.bahanId, jumlah: r.jumlah })),
        varian: m.varian.map((g) => ({
          nama: g.nama, wajib: g.wajib, multi: g.multi, gantiHarga: g.gantiHarga,
          opsi: g.opsi.map((o) => ({ nama: o.nama, tambahan: o.tambahan, bahan: o.bahan.map((b) => ({ bahanId: b.bahanId, jumlah: b.jumlah })), buka: o.bahan.length > 0 })),
        })),
      }
    : {
        tenantId: pilihTenant.value || tenant.value[0]?.id || null,
        kategoriId: aktif.value || kategori.value[0]?.id || null,
        nama: '', harga: 0, hargaModal: 0, komisi: 0, tersedia: true, trackStok: false, stok: 0, varian: [], resep: [],
      }
}

const hargaDariOpsi = computed(() => form.value?.varian.some((g) => g.gantiHarga && g.wajib))

const tambahGrup = () =>
  form.value.varian.push({ nama: '', wajib: false, multi: false, gantiHarga: false, opsi: [{ nama: '', tambahan: 0, bahan: [], buka: false }] })

// hanya satu grup yang boleh menentukan harga
const setGanti = (gi, nilai) => form.value.varian.forEach((g, i) => (g.gantiHarga = nilai && i === gi))

async function simpanMenu() {
  const { id, ...f } = form.value
  const body = {
    ...f,
    nama: f.nama.trim(),
    harga: f.harga || 0,
    hargaModal: f.hargaModal || 0,
    komisi: f.komisi || 0,
    stok: f.trackStok ? f.stok || 0 : 0,
    resep: f.resep.filter((r) => r.bahanId && r.jumlah > 0),
    varian: f.varian
      .map((g) => ({
        ...g,
        nama: g.nama.trim(),
        opsi: g.opsi
          .filter((o) => o.nama.trim())
          .map((o) => ({ nama: o.nama.trim(), tambahan: o.tambahan || 0, bahan: o.bahan.filter((b) => b.bahanId && b.jumlah) })),
      }))
      .filter((g) => g.nama && g.opsi.length),
  }
  if (body.varian.some((g) => g.gantiHarga && g.wajib)) body.harga = 0
  if (await run(() => (id ? put(`/menu/${id}`, body) : post('/menu', body)), 'Menu disimpan')) {
    form.value = null
    await run(muat)
  }
}

function hapusMenu(m) {
  konfirmasi(confirm, {
    header: 'Hapus menu?',
    message: `"${m.nama}" tidak akan muncul lagi di kasir. Riwayat transaksi tetap tersimpan.`,
    accept: async () => (await run(() => del(`/menu/${m.id}`), 'Menu dihapus')) && run(muat),
  })
}

const ubahTersedia = async (m, tersedia) => {
  await run(() => post(`/menu/${m.id}/stok`, { tersedia }))
  await run(muat)
}

onMounted(() => run(muat))
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-book" title="Menu" subtitle="Kelola menu, varian, dan stok">
      <button class="pill" @click="kat = { nama: '', urutan: 0 }"><i class="pi pi-folder-plus" /> Kategori baru</button>
      <button class="pill" :disabled="!kategori.length" @click="bukaMenu()"><i class="pi pi-plus" /> Tambah menu</button>
    </PageHeader>

    <div class="layout">
      <aside class="panel kat">
        <button class="k" :class="{ on: !aktif }" @click="aktif = 0">Semua <small>{{ total }}</small></button>
        <div v-for="k in kategori" :key="k.id" class="krow">
          <button class="k" :class="{ on: aktif === k.id }" @click="aktif = k.id">{{ k.nama }} <small>{{ jumlahDi(k) }}</small></button>
          <Button icon="pi pi-pencil" text size="small" severity="secondary" aria-label="Ubah" @click="kat = { id: k.id, nama: k.nama, urutan: k.urutan }" />
          <Button icon="pi pi-trash" text size="small" severity="danger" aria-label="Hapus" @click="hapusKat(k)" />
        </div>
        <p v-if="!kategori.length" class="none">Buat kategori dulu, misalnya Kopi, Makanan, Gorengan.</p>
      </aside>

      <section class="panel tabel">
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
        <DataTable :value="rows" size="small" paginator :rows="15" empty-message="Belum ada menu">
          <Column field="nama" header="Nama" sortable />
          <Column field="kategori" header="Kategori" />
          <Column v-if="tenant.length > 1" header="Tenant"><template #body="{ data }">{{ namaTenant(data.tenantId) }}</template></Column>
          <Column header="Harga" sortable field="harga"><template #body="{ data }">{{ hargaTampil(data) }}</template></Column>
          <Column header="Stok">
            <template #body="{ data }">
              <Tag v-if="!data.trackStok" value="—" severity="secondary" />
              <Tag v-else :value="String(data.stok)" :severity="data.stok > 0 ? 'success' : 'danger'" />
              <small v-if="data.porsiBahan !== null" class="redup"> · {{ data.porsiBahan }} porsi dari bahan</small>
            </template>
          </Column>
          <Column header="Varian"><template #body="{ data }">{{ data.varian.length || '—' }}</template></Column>
          <Column header="Tersedia">
            <template #body="{ data }">
              <ToggleSwitch :model-value="data.tersedia" @update:model-value="ubahTersedia(data, $event)" />
            </template>
          </Column>
          <Column>
            <template #body="{ data }">
              <Button icon="pi pi-pencil" text severity="secondary" aria-label="Ubah" @click="bukaMenu(data)" />
              <Button icon="pi pi-trash" text severity="danger" aria-label="Hapus" @click="hapusMenu(data)" />
            </template>
          </Column>
        </DataTable>
      </section>
    </div>

    <Dialog :visible="!!kat" modal :header="kat?.id ? 'Ubah kategori' : 'Kategori baru'" :style="{ width: '24rem' }" @update:visible="kat = null">
      <div v-if="kat" class="stack">
        <label for="knama">Nama</label>
        <InputText id="knama" v-model="kat.nama" maxlength="100" autofocus fluid @keyup.enter="simpanKat" />
        <label for="kurut">Urutan tampil</label>
        <InputNumber id="kurut" v-model="kat.urutan" :min="0" fluid />
      </div>
      <template #footer>
        <Button label="Simpan" icon="pi pi-check" :disabled="!kat?.nama.trim()" @click="simpanKat" />
      </template>
    </Dialog>

    <Dialog :visible="!!form" modal :header="form?.id ? 'Ubah menu' : 'Tambah menu'" :style="{ width: '40rem' }" :breakpoints="{ '700px': '96vw' }" @update:visible="form = null">
      <div v-if="form" class="stack">
        <label>Kategori</label>
        <Select v-model="form.kategoriId" :options="kategori" option-label="nama" option-value="id" fluid />
        <template v-if="tenant.length > 1">
          <label>Tenant</label>
          <Select v-model="form.tenantId" :options="tenant" option-label="nama" option-value="id" fluid />
        </template>
        <label for="mnama">Nama menu</label>
        <InputText id="mnama" v-model="form.nama" maxlength="100" fluid />
        <template v-if="!hargaDariOpsi">
          <label>Harga dasar</label>
          <MoneyInput v-model="form.harga" />
        </template>
        <small v-else class="hasil">Harga ditentukan oleh opsi di bagian Varian, jadi harga dasar tidak perlu diisi.</small>
        <div class="sw"><ToggleSwitch v-model="form.tersedia" input-id="tersedia" /><label for="tersedia">Tersedia</label></div>
        <div class="sw"><ToggleSwitch v-model="form.trackStok" input-id="track" /><label for="track">Lacak stok</label></div>
        <template v-if="form.trackStok">
          <label for="stok">Stok saat ini</label>
          <InputNumber id="stok" v-model="form.stok" :min="0" fluid />
        </template>

        <template v-if="tenant.find((t) => t.id === form.tenantId)?.komisiPerMenu">
          <label>Komisi pemilik tempat per porsi</label>
          <MoneyInput v-model="form.komisi" />
          <small class="hasil">Tenant ini memakai komisi rupiah per menu. Bisa juga diatur sekaligus di halaman Tenant.</small>
        </template>

        <h4>Resep</h4>
        <small class="hasil">
          Bahan yang terpakai per porsi. Stok bahan berkurang saat item masuk ke bill, dan menu otomatis habis bila bahannya kurang.
          {{ bahan.length ? '' : 'Tambahkan bahan dulu di menu Stok Bahan.' }}
        </small>
        <div v-for="(r, ri) in form.resep" :key="ri" class="inline">
          <Select v-model="r.bahanId" :options="bahan" option-label="nama" option-value="id" placeholder="Bahan" filter fluid />
          <InputNumber v-model="r.jumlah" :min="1" :max="99999" fluid />
          <small class="hasil">{{ satuan(r.bahanId) }}</small>
          <Button icon="pi pi-times" text severity="secondary" aria-label="Hapus bahan" @click="form.resep.splice(ri, 1)" />
        </div>
        <Button label="Tambah bahan ke resep" icon="pi pi-plus" severity="secondary" outlined :disabled="!bahan.length" @click="form.resep.push({ bahanId: null, jumlah: null })" />
        <!-- tenant lain: warkop hanya dapat komisi, HPP-nya tidak dihitung di laba rugi -->
        <template v-if="tenant.find((t) => t.id === form.tenantId)?.milikSendiri !== false && !form.resep.some((r) => r.bahanId && r.jumlah > 0)">
          <label>Harga modal per porsi</label>
          <MoneyInput v-model="form.hargaModal" />
          <small class="hasil">Untuk barang jadi tanpa resep (rokok, kerupuk, minuman botol): harga beli per item, dipakai sebagai HPP di laba rugi.</small>
        </template>

        <h4>Varian</h4>
        <small class="hasil">
          Secara bawaan tiap opsi menambah harga dasar. Aktifkan "Harga ditentukan opsi" bila tiap opsi punya harga sendiri,
          misalnya panas Rp5.000 dan es Rp10.000.
        </small>
        <div v-for="(g, gi) in form.varian" :key="gi" class="grup">
          <div class="inline">
            <InputText v-model="g.nama" maxlength="50" placeholder="Nama grup, mis. Suhu" fluid />
            <Button icon="pi pi-trash" text severity="danger" aria-label="Hapus grup" @click="form.varian.splice(gi, 1)" />
          </div>
          <div class="inline wrap">
            <div class="sw"><ToggleSwitch v-model="g.wajib" :input-id="`w${gi}`" /><label :for="`w${gi}`">Wajib dipilih</label></div>
            <div class="sw">
              <ToggleSwitch v-model="g.multi" :disabled="g.gantiHarga" :input-id="`m${gi}`" />
              <label :for="`m${gi}`">Boleh lebih dari satu</label>
            </div>
            <div class="sw">
              <ToggleSwitch :model-value="g.gantiHarga" :disabled="g.multi" :input-id="`h${gi}`" @update:model-value="setGanti(gi, $event)" />
              <label :for="`h${gi}`">Harga ditentukan opsi</label>
            </div>
          </div>
          <template v-for="(o, oi) in g.opsi" :key="oi">
          <div class="inline">
            <InputText v-model="o.nama" maxlength="50" placeholder="Opsi, mis. Es" fluid />
            <MoneyInput v-model="o.tambahan" />
            <small class="hasil">{{ g.gantiHarga ? 'harga jadi' : '= ' + rp((form.harga || 0) + (o.tambahan || 0)) }}</small>
            <Button icon="pi pi-box" text :severity="o.bahan.length ? 'primary' : 'secondary'" aria-label="Bahan tambahan opsi" @click="o.buka = !o.buka" />
            <Button icon="pi pi-times" text severity="secondary" aria-label="Hapus opsi" @click="g.opsi.splice(oi, 1)" />
          </div>
          <div v-if="o.buka" class="resep dalam">
            <small class="hasil">Bahan yang bertambah (atau berkurang bila negatif) saat opsi ini dipilih, di luar resep dasar.</small>
            <div v-for="(b, bi) in o.bahan" :key="bi" class="inline">
              <Select v-model="b.bahanId" :options="bahan" option-label="nama" option-value="id" placeholder="Bahan" filter fluid />
              <InputNumber v-model="b.jumlah" :min="-99999" :max="99999" fluid />
              <small class="hasil">{{ satuan(b.bahanId) }}</small>
              <Button icon="pi pi-times" text severity="secondary" aria-label="Hapus bahan" @click="o.bahan.splice(bi, 1)" />
            </div>
            <Button label="Tambah bahan" icon="pi pi-plus" text size="small" :disabled="!bahan.length" @click="o.bahan.push({ bahanId: null, jumlah: null })" />
          </div>
          </template>
          <Button label="Tambah opsi" icon="pi pi-plus" text size="small" @click="g.opsi.push({ nama: '', tambahan: 0, bahan: [], buka: false })" />
        </div>
        <Button label="Tambah grup varian" icon="pi pi-plus" severity="secondary" outlined @click="tambahGrup" />
      </div>
      <template #footer>
        <Button label="Batal" severity="secondary" text @click="form = null" />
        <Button label="Simpan" icon="pi pi-check" :disabled="!form?.nama.trim() || !form?.kategoriId || !form?.tenantId" @click="simpanMenu" />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.gap {
  gap: 14px;
}
.layout {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  gap: 14px;
  align-items: start;
}
.kat {
  padding: 10px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.krow {
  display: flex;
  align-items: center;
}
.k {
  flex: 1;
  display: flex;
  justify-content: space-between;
  gap: 8px;
  padding: 10px 12px;
  border: 0;
  border-radius: 10px;
  background: none;
  font: inherit;
  font-weight: 600;
  text-align: left;
  color: inherit;
  cursor: pointer;
}
.k:hover {
  background: color-mix(in srgb, var(--p-primary-color) 10%, transparent);
}
.k.on {
  background: var(--p-primary-color);
  color: var(--p-primary-contrast-color);
}
.none {
  margin: 8px;
  color: var(--p-text-muted-color);
  font-size: 0.9rem;
}
.tabel {
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.sw {
  display: flex;
  align-items: center;
  gap: 10px;
}
.sw label {
  font-weight: 500;
}
h4 {
  margin: 10px 0 0;
}
.grup {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  border: 1px solid var(--p-content-border-color);
  border-radius: 12px;
}
.inline {
  display: flex;
  align-items: center;
  gap: 8px;
}
.redup {
  color: var(--p-text-muted-color);
}
.dalam {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding-left: 12px;
}
.hasil {
  color: var(--p-text-muted-color);
}
.inline .hasil {
  white-space: nowrap;
}
.wrap {
  flex-wrap: wrap;
}
@media (max-width: 800px) {
  .layout {
    grid-template-columns: 1fr;
  }
}
.pilih-tenant {
  flex-wrap: wrap;
  margin-bottom: 10px;
}
</style>
