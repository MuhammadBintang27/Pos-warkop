<script setup>
import { onMounted, ref } from 'vue'
import { useCari } from '@/composables/useCari'
import { useConfirm } from 'primevue/useconfirm'
import { del, get, post, put } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { konfirmasi } from '@/utils/confirm'
import { rp } from '@/utils/format'

const confirm = useConfirm()
const run = useRun()

const tenant = ref([])
const form = ref(null)

const printer = ref([])
const menu = ref([])
const isiSemua = ref(null)

async function muat() {
  ;[tenant.value, printer.value, menu.value] = await Promise.all([
    get('/tenant'),
    get('/printer'),
    get('/menu').then((r) => r.kategori.flatMap((k) => k.menu)),
  ])
}

// salinan tarif per menu tenant ini; baru dikirim saat Simpan
function ubah(t) {
  isiSemua.value = null
  form.value = { ...t, menu: menu.value.filter((m) => m.tenantId === t.id).map((m) => ({ id: m.id, nama: m.nama, harga: m.harga, komisi: m.komisi })) }
}

const namaPrinter = (id) => printer.value.find((p) => p.id === id)?.nama

async function simpan() {
  const { id, nama, komisiPersen, komisiPerMenu, milikSendiri, printerId } = form.value
  const body = {
    nama: nama.trim(),
    komisiPersen: komisiPersen || 0,
    komisiPerMenu,
    milikSendiri,
    printerId: printerId ?? null,
    komisiMenu: (form.value.menu ?? []).map((m) => ({ menuId: m.id, komisi: m.komisi || 0 })),
  }
  if (await run(() => (id ? put(`/tenant/${id}`, body) : post('/tenant', body)), 'Tenant disimpan')) {
    form.value = null
    await run(muat)
  }
}

function hapus(t) {
  konfirmasi(confirm, {
    header: 'Hapus tenant?',
    message: `"${t.nama}" akan dihapus. Tenant yang masih punya menu aktif tidak bisa dihapus.`,
    accept: async () => (await run(() => del(`/tenant/${t.id}`), 'Tenant dihapus')) && run(muat),
  })
}

onMounted(() => run(muat))
const { cari: cariTenant, hasil: tenantTampil } = useCari(tenant, 'nama')
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-shop" title="Tenant" subtitle="Penjual di tempatmu dan bagi hasilnya">
      <button class="pill" @click="form = { nama: '', komisiPersen: 0, komisiPerMenu: false, milikSendiri: false, printerId: null }"><i class="pi pi-plus" /> Tambah tenant</button>
    </PageHeader>

    <section class="panel box">
      <CariInput v-model="cariTenant" placeholder="Cari tenant…" />
      <DataTable paginator :rows="10" :rows-per-page-options="[10, 25, 50]" :always-show-paginator="false" :value="tenantTampil" size="small" empty-message="Belum ada tenant">
        <Column field="nama" header="Nama" />
        <Column header="Printer">
          <template #body="{ data }">{{ namaPrinter(data.printerId) ?? 'Kasir' }}</template>
        </Column>
        <Column header="Komisi untuk pemilik tempat">
          <template #body="{ data }">
            <Tag v-if="data.milikSendiri" value="Milik sendiri" severity="info" />
            <template v-else-if="data.komisiPerMenu">Per menu</template>
            <template v-else>{{ data.komisiPersen }}%</template>
          </template>
        </Column>
        <Column>
          <template #body="{ data }">
            <Button icon="pi pi-pencil" text severity="secondary" aria-label="Ubah" @click="ubah(data)" />
            <Button icon="pi pi-trash" text severity="danger" aria-label="Hapus" @click="hapus(data)" />
          </template>
        </Column>
      </DataTable>
      <small class="hint">
        Setiap menu dimiliki satu tenant. Tiket dapur dicetak terpisah per tenant, dan laporan menghitung setoran
        tenant setelah dipotong komisi.
      </small>
    </section>

    <Dialog :visible="!!form" modal :header="form?.id ? 'Ubah tenant' : 'Tambah tenant'" :style="{ width: '24rem' }" @update:visible="form = null">
      <div v-if="form" class="stack">
        <label for="nama">Nama tenant</label>
        <InputText id="nama" v-model="form.nama" maxlength="100" autofocus fluid />
        <div class="sw">
          <ToggleSwitch v-model="form.milikSendiri" input-id="sendiri" />
          <label for="sendiri">Milik sendiri (tanpa bagi hasil dan setoran)</label>
        </div>
        <template v-if="!form.milikSendiri">
          <label>Jenis komisi untuk pemilik tempat</label>
          <SelectButton
            v-model="form.komisiPerMenu"
            :options="[{ label: 'Persen', value: false }, { label: 'Rupiah per menu', value: true }]"
            option-label="label"
            option-value="value"
            :allow-empty="false"
          />
        </template>
        <template v-if="form.milikSendiri || !form.komisiPerMenu">
          <label for="komisi">Komisi untuk pemilik tempat (%)</label>
          <InputNumber
            id="komisi"
            :model-value="form.milikSendiri ? 100 : form.komisiPersen"
            :disabled="form.milikSendiri"
            suffix="%"
            :min="0"
            :max="100"
            :max-fraction-digits="2"
            fluid
            @update:model-value="form.komisiPersen = $event"
          />
        </template>
        <template v-else>
          <small v-if="!form.menu?.length" class="hint">
            {{ form.id ? 'Tenant ini belum punya menu.' : 'Simpan tenant dulu, lalu tambahkan menunya.' }} Komisi tiap menu bisa diisi di sini atau di form menu.
          </small>
          <template v-else>
            <small class="hint">Rupiah per porsi terjual untuk pemilik tempat. Diskon bill ditanggung tenant.</small>
            <div class="sw">
              <MoneyInput v-model="isiSemua" placeholder="Isi semua" />
              <Button label="Terapkan" severity="secondary" outlined :disabled="isiSemua == null" @click="form.menu.forEach((m) => (m.komisi = isiSemua))" />
            </div>
            <div class="daftar">
              <div v-for="m in form.menu" :key="m.id" class="baris">
                <span class="nama">{{ m.nama }} <small class="hint">{{ rp(m.harga) }}</small></span>
                <MoneyInput v-model="m.komisi" />
              </div>
            </div>
          </template>
        </template>
        <label>Printer tiket dapur</label>
        <Select
          v-model="form.printerId"
          :options="printer"
          option-label="nama"
          option-value="id"
          show-clear
          placeholder="Tanpa printer (dicetak di kasir)"
          fluid
        />
      </div>
      <template #footer>
        <Button label="Simpan" icon="pi pi-check" :disabled="!form?.nama.trim()" @click="simpan" />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.gap {
  gap: 14px;
}
.box {
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.hint {
  color: var(--p-text-muted-color);
}
.sw {
  display: flex;
  align-items: center;
  gap: 10px;
}
.sw label {
  font-weight: 500;
}
.daftar {
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 18rem;
  overflow-y: auto;
}
.baris {
  display: grid;
  grid-template-columns: 1fr 8rem;
  align-items: center;
  gap: 8px;
}
.nama {
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
