<script setup>
import { ref, watch } from 'vue'
import { get } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { usePengaturanStore } from '@/stores/pengaturan'
import { labelKomisi, rp, tgl } from '@/utils/format'
import { cetakClosing } from '@/utils/cetak'

const STATUS = {
  closing: { label: 'Closing', severity: 'info' },
  disetor: { label: 'Disetor', severity: 'success' },
  berubah: { label: 'Berubah setelah closing', severity: 'danger' },
}

const props = defineProps({ baris: Object, tertunda: Object })
const visible = defineModel('visible')
const peng = usePengaturanStore()
const run = useRun()
const detail = ref(null)

watch(visible, (v) => {
  if (!v || !props.baris) return
  detail.value = null
  run(async () => {
    const { tenantId, tanggal } = props.baris
    detail.value = await get('/laporan/tenant/detail', { params: { tenantId, dari: tanggal, sampai: tanggal } })
  })
})
</script>

<template>
  <Dialog v-model:visible="visible" modal :header="baris?.milikSendiri ? 'Rekap harian' : 'Bill closing tenant'" :style="{ width: '30rem' }" :breakpoints="{ '600px': '96vw' }">
    <div v-if="baris" class="stack">
      <div class="head">
        <strong>{{ baris.nama }}</strong>
        <Tag v-if="STATUS[baris.status]" :value="STATUS[baris.status].label" :severity="STATUS[baris.status].severity" />
      </div>
      <small class="redup">
        {{ baris.tanggal.split('-').reverse().join('/') }} · {{ baris.transaksi }} bill
      </small>

      <Message v-if="baris.status === 'berubah'" severity="warn" size="small">
        Ada transaksi yang masuk atau berubah setelah closing. Angka di bawah adalah angka saat closing, sedangkan rincian menu mengikuti data terbaru.
      </Message>

      <ul class="items">
        <li v-for="m in detail?.menu.filter((x) => x.qty) ?? []" :key="m.nama">
          <span>{{ m.qty }}× {{ m.nama }}</span><span>{{ rp(m.nilai) }}</span>
        </li>
        <li v-if="detail && !detail.menu.some((x) => x.qty)" class="redup">Tidak ada rincian</li>
      </ul>

      <div class="row"><span>Penjualan</span><span>{{ rp(baris.kotor) }}</span></div>
      <div class="row"><span>Diskon</span><span>-{{ rp(baris.diskon) }}</span></div>
      <div v-if="baris.milikSendiri" class="row total"><span>Pendapatan pemilik</span><strong>{{ rp(baris.komisi) }}</strong></div>
      <template v-else>
        <div class="row"><span>Komisi {{ labelKomisi(baris) }}</span><span>-{{ rp(baris.komisi) }}</span></div>
        <div class="row total"><span>Setor ke tenant</span><strong>{{ rp(baris.setor) }}</strong></div>
      </template>
      <div v-if="detail?.talangan?.length" class="talangan">
        <small class="redup">Termasuk {{ detail.talangan.length }} bill belum dibayar pelanggan (ditalangi pemilik):</small>
        <div v-for="t in detail.talangan" :key="t.pesananId" class="trow">
          <span>{{ t.mejaNama || 'Bawa pulang' }} · #{{ t.pesananId }}<template v-if="t.atasNama"> · {{ t.atasNama }}</template></span>
          <span>{{ rp(t.kotor) }}</span>
        </div>
      </div>
      <small v-if="detail?.batalQty" class="redup">Dibatalkan: {{ detail.batalQty }} porsi ({{ rp(detail.batalNilai) }})</small>
      <Message v-if="tertunda" severity="info" size="small">
        Belum dibayar pelanggan: {{ tertunda.bill }} bill ≈ {{ rp(tertunda.nilai) }}. Belum termasuk angka di atas, dan ikut closing saat dibayar.
      </Message>

      <div class="jejak">
        <div v-if="baris.dibuatAt"><i class="pi pi-lock" /> Closing oleh {{ baris.dibuatOleh }} · {{ tgl(baris.dibuatAt) }}</div>
        <div v-if="baris.disetorAt"><i class="pi pi-check-circle" /> Disetor oleh {{ baris.disetorOleh }} · {{ tgl(baris.disetorAt) }}</div>
      </div>
    </div>

    <template #footer>
      <Button :label="baris?.milikSendiri ? 'Cetak rekap' : 'Cetak bill'" icon="pi pi-print" severity="secondary" outlined :disabled="!detail" @click="run(() => cetakClosing(baris, detail, peng.settings, tertunda))" />
      <Button label="Selesai" icon="pi pi-check" @click="visible = false" />
    </template>
  </Dialog>
</template>

<style scoped>
.head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 1.15rem;
}
.talangan {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.trow {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  font-size: 0.9em;
}
.redup {
  color: var(--p-text-muted-color);
}
.items {
  list-style: none;
  margin: 4px 0;
  padding: 8px 0;
  border-block: 1px dashed var(--p-content-border-color);
  max-height: 14rem;
  overflow-y: auto;
}
.items li {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 3px 0;
}
.row {
  display: flex;
  justify-content: space-between;
}
.total {
  padding-top: 8px;
  border-top: 1px solid var(--p-content-border-color);
  font-size: 1.25rem;
  color: var(--accent-text);
}
.jejak {
  display: flex;
  flex-direction: column;
  gap: 4px;
  color: var(--p-text-muted-color);
  font-size: 0.9rem;
}
</style>
