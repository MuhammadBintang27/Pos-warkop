<script setup>
import { useConfirm } from 'primevue/useconfirm'
import { post } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { useKasirStore } from '@/stores/kasir'
import { konfirmasi } from '@/utils/confirm'
import { printBatalDapur, printDapur } from '@/utils/print'

const visible = defineModel('visible')
const store = useKasirStore()
const run = useRun()
const confirm = useConfirm()

const JENIS = { dapur: 'Tiket dapur', batal: 'Tiket batal', struk: 'Struk', closing: 'Laporan', tes: 'Tes cetak' }
const menit = (t) => Math.max(1, Math.round((Date.now() - new Date(t)) / 60000))

async function abaikan(j) {
  await post(`/cetak/${j.id}/abaikan`)
  await store.loadCetak()
}

// payload tiket dapur berisi satu tenant; bentuknya disesuaikan dengan printDapur/printBatalDapur
function cetakBrowser(j) {
  const p = j.payload
  const t = {
    mejaNama: p.mejaNama,
    atasNama: p.atasNama,
    oleh: p.oleh,
    tambahan: { 0: p.judul === 'PESANAN TAMBAHAN' },
    items: p.items.map((i) => ({ ...i, tenantId: 0, tenantNama: p.tenant })),
  }
  ;(j.jenis === 'batal' ? printBatalDapur : printDapur)(t)
  // sudah keluar di kertas browser: jangan sampai tercetak dobel saat printer menyala lagi
  run(() => abaikan(j))
}

function tanyaAbaikan(j) {
  konfirmasi(confirm, {
    header: 'Abaikan tiket ini?',
    message: 'Tiket tidak akan dicetak printer walau printernya menyala lagi. Pastikan isinya sudah disampaikan.',
    label: 'Abaikan',
    accept: () => run(() => abaikan(j), 'Tiket diabaikan'),
  })
}
</script>

<template>
  <Dialog v-model:visible="visible" modal header="Tiket belum tercetak" :style="{ width: '34rem' }" :breakpoints="{ '600px': '96vw' }">
    <p v-if="!store.cetak.stasiunAktif" class="info warn">
      <i class="pi pi-exclamation-triangle" /> Stasiun cetak tidak terhubung. Pastikan HP/PC stasiun menyala dan ada internet.
    </p>
    <p v-else class="info">Stasiun mencoba lagi tiap 15 detik. Tiket keluar sendiri begitu masalah printernya beres.</p>

    <div v-if="store.cetak.daftar?.length" class="daftar">
      <div v-for="j in store.cetak.daftar" :key="j.id" class="tiket">
        <div class="atas">
          <strong>{{ JENIS[j.jenis] ?? j.jenis }} · {{ j.mejaNama || 'Bawa pulang' }}</strong>
          <small>{{ menit(j.dibuatAt) }} mnt lalu</small>
        </div>
        <div class="redup">{{ [j.tenant, `Printer: ${j.printer}`].filter(Boolean).join(' · ') }}</div>
        <div v-if="j.galat" class="galat"><i class="pi pi-times-circle" /> {{ j.galat }}</div>
        <div v-else class="redup">Belum ada percobaan cetak</div>
        <div class="aksi">
          <Button v-if="j.payload" label="Cetak lewat browser" icon="pi pi-print" size="small" @click="cetakBrowser(j)" />
          <Button label="Abaikan" size="small" severity="secondary" outlined @click="tanyaAbaikan(j)" />
        </div>
      </div>
    </div>
    <p v-else class="redup">Semua tiket sudah tercetak.</p>
    <small v-if="store.cetak.macet > (store.cetak.daftar?.length ?? 0)" class="redup">
      Ditampilkan {{ store.cetak.daftar.length }} dari {{ store.cetak.macet }} tiket.
    </small>
  </Dialog>
</template>

<style scoped>
.info {
  margin: 0 0 12px;
  color: var(--p-text-muted-color);
}
.info.warn {
  color: var(--p-red-600, #dc2626);
  font-weight: 600;
}
.daftar {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.tiket {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px;
  border: 1px solid var(--p-content-border-color);
  border-radius: 12px;
}
.atas {
  display: flex;
  justify-content: space-between;
  gap: 8px;
}
.atas small,
.redup {
  color: var(--p-text-muted-color);
}
.galat {
  color: var(--p-red-600, #dc2626);
  font-weight: 600;
}
.aksi {
  display: flex;
  gap: 8px;
  margin-top: 6px;
  flex-wrap: wrap;
}
</style>
