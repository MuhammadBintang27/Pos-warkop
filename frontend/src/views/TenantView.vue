<script setup>
import { onMounted, ref } from 'vue'
import { useCari } from '@/composables/useCari'
import { useConfirm } from 'primevue/useconfirm'
import { del, get, post, put } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { konfirmasi } from '@/utils/confirm'

const confirm = useConfirm()
const run = useRun()

const tenant = ref([])
const form = ref(null)

const printer = ref([])

async function muat() {
  ;[tenant.value, printer.value] = await Promise.all([get('/tenant'), get('/printer')])
}

const namaPrinter = (id) => printer.value.find((p) => p.id === id)?.nama

async function simpan() {
  const { id, nama, komisiPersen, milikSendiri, printerId } = form.value
  const body = { nama: nama.trim(), komisiPersen: komisiPersen || 0, milikSendiri, printerId: printerId ?? null }
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
      <button class="pill" @click="form = { nama: '', komisiPersen: 0, milikSendiri: false, printerId: null }"><i class="pi pi-plus" /> Tambah tenant</button>
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
            <template v-else>{{ data.komisiPersen }}%</template>
          </template>
        </Column>
        <Column>
          <template #body="{ data }">
            <Button icon="pi pi-pencil" text severity="secondary" aria-label="Ubah" @click="form = { ...data }" />
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
</style>
