<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import { del, get, post, put } from '@/api/http'
import { onEvent } from '@/composables/useEvents'
import { useRun } from '@/composables/useRun'
import { konfirmasi } from '@/utils/confirm'
import { tgl } from '@/utils/format'

const LEBAR = [
  { label: '58 mm', value: 58 },
  { label: '80 mm', value: 80 },
]
const STATUS = {
  antri: { label: 'Menunggu', severity: 'warn' },
  selesai: { label: 'Tercetak', severity: 'success' },
  kedaluwarsa: { label: 'Kedaluwarsa', severity: 'secondary' },
}

const confirm = useConfirm()
const toast = useToast()
const run = useRun()

const printer = ref([])
const riwayat = ref([])
const status = ref({ stasiunAktif: false })
const form = ref(null)
const token = ref('')

async function muat() {
  ;[printer.value, riwayat.value, status.value] = await Promise.all([get('/printer'), get('/cetak/riwayat'), get('/cetak/status')])
}

async function simpan() {
  const { id, nama, ip, port, lebarMm } = form.value
  const body = { nama: nama.trim(), ip: ip.trim(), port, lebarMm }
  if (await run(() => (id ? put(`/printer/${id}`, body) : post('/printer', body)), 'Printer disimpan')) {
    form.value = null
    await run(muat)
  }
}

function hapus(p) {
  konfirmasi(confirm, {
    header: 'Hapus printer?',
    message: `"${p.nama}" dihapus dan tenant yang memakainya kembali dicetak di kasir.`,
    accept: async () => (await run(() => del(`/printer/${p.id}`), 'Printer dihapus')) && run(muat),
  })
}

const tes = (p) =>
  run(async () => {
    await post(`/printer/${p.id}/tes`)
    await muat()
  }, 'Tiket percobaan dikirim ke antrean')

const ulang = (j) =>
  run(async () => {
    await post(`/cetak/${j.id}/ulang`)
    await muat()
  }, 'Dikirim ulang ke antrean')

async function buatToken() {
  await run(async () => (token.value = (await post('/stasiun/token')).token))
}

async function salinToken() {
  try {
    await navigator.clipboard.writeText(token.value)
    toast.add({ severity: 'success', summary: 'Token disalin', life: 2000 })
  } catch {
    toast.add({ severity: 'warn', summary: 'Salin manual dari kotak token', life: 3000 })
  }
}

const offs = []
onMounted(() => {
  run(muat)
  offs.push(onEvent('cetak', () => muat().catch(() => {})))
})
onBeforeUnmount(() => offs.forEach((off) => off()))
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-print" title="Printer" subtitle="Printer tiket dapur untuk tiap tenant">
      <button class="pill" @click="form = { nama: '', ip: '', port: 9100, lebarMm: 58 }"><i class="pi pi-plus" /> Tambah printer</button>
    </PageHeader>

    <section class="panel box">
      <div class="head">
        <h3>Stasiun cetak</h3>
        <Tag :value="status.stasiunAktif ? 'Terhubung' : 'Tidak terhubung'" :severity="status.stasiunAktif ? 'success' : 'danger'" />
        <Button label="Buat token stasiun" icon="pi pi-key" size="small" severity="secondary" outlined class="kanan" @click="buatToken" />
      </div>
      <small class="redup">
        Stasiun cetak adalah aplikasi desktop POS Warkop di PC kasir. Ia mengambil tiket dari server lalu mencetaknya ke printer di jaringan.
        Klik kanan ikon POS Warkop di system tray Windows, pilih <strong>Stasiun cetak…</strong>, tempel token di bawah, lalu aktifkan.
      </small>
      <div v-if="token" class="token">
        <code>{{ token }}</code>
        <Button icon="pi pi-copy" text aria-label="Salin token" @click="salinToken" />
        <small>Token hanya ditampilkan sekali. Membuat token baru membuat token lama tidak berlaku.</small>
      </div>
    </section>

    <section class="panel box">
      <h3>Daftar printer</h3>
      <DataTable :value="printer" size="small" empty-message="Belum ada printer">
        <Column field="nama" header="Nama" />
        <Column header="Alamat"><template #body="{ data }">{{ data.ip }}:{{ data.port }}</template></Column>
        <Column header="Kertas"><template #body="{ data }">{{ data.lebarMm }} mm</template></Column>
        <Column header="Antrean">
          <template #body="{ data }">
            <Tag v-if="data.antri" :value="`${data.antri} menunggu`" severity="warn" />
            <span v-else class="redup">—</span>
          </template>
        </Column>
        <Column>
          <template #body="{ data }">
            <Button label="Tes cetak" size="small" outlined severity="secondary" @click="tes(data)" />
            <Button icon="pi pi-pencil" text severity="secondary" aria-label="Ubah" @click="form = { ...data }" />
            <Button icon="pi pi-trash" text severity="danger" aria-label="Hapus" @click="hapus(data)" />
          </template>
        </Column>
      </DataTable>
      <small class="redup">Hubungkan tiap tenant ke printernya di Master Data › Tenant. Dua tenant boleh memakai satu printer.</small>
    </section>

    <section class="panel box">
      <h3>Riwayat cetak</h3>
      <DataTable :value="riwayat" size="small" paginator :rows="10" empty-message="Belum ada tiket">
        <Column header="Waktu"><template #body="{ data }">{{ tgl(data.dibuatAt) }}</template></Column>
        <Column field="judul" header="Tiket" />
        <Column field="tenant" header="Tenant" />
        <Column field="mejaNama" header="Meja" />
        <Column field="printer" header="Printer" />
        <Column header="Status">
          <template #body="{ data }">
            <Tag :value="STATUS[data.status].label" :severity="STATUS[data.status].severity" />
            <small v-if="data.status === 'antri' && data.percobaan" class="redup"><br />coba {{ data.percobaan }}× · {{ data.galat }}</small>
          </template>
        </Column>
        <Column>
          <template #body="{ data }">
            <Button label="Cetak ulang" size="small" text @click="ulang(data)" />
          </template>
        </Column>
      </DataTable>
    </section>

    <Dialog :visible="!!form" modal :header="form?.id ? 'Ubah printer' : 'Tambah printer'" :style="{ width: '26rem' }" @update:visible="form = null">
      <div v-if="form" class="stack">
        <label for="nama">Nama (mis. Printer Tenant A dan B)</label>
        <InputText id="nama" v-model="form.nama" maxlength="100" autofocus fluid />
        <label for="ip">Alamat IP printer</label>
        <InputText id="ip" v-model="form.ip" maxlength="100" placeholder="192.168.1.50" fluid />
        <label for="port">Port</label>
        <InputNumber id="port" v-model="form.port" :min="1" :max="65535" :use-grouping="false" fluid />
        <label>Lebar kertas</label>
        <SelectButton v-model="form.lebarMm" :options="LEBAR" option-label="label" option-value="value" :allow-empty="false" />
      </div>
      <template #footer>
        <Button label="Simpan" icon="pi pi-check" :disabled="!form?.nama.trim() || !form?.ip.trim() || !form?.port" @click="simpan" />
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
  gap: 10px;
}
.head {
  display: flex;
  align-items: center;
  gap: 10px;
}
.head h3 {
  margin: 0;
}
.kanan {
  margin-left: auto;
}
h3 {
  margin: 0;
}
.redup {
  color: var(--p-text-muted-color);
}
.token {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border: 1px dashed var(--p-primary-color);
  border-radius: 10px;
}
.token code {
  font-size: 0.95rem;
  word-break: break-all;
}
</style>
