<script setup>
import { onMounted, ref, watch } from 'vue'
import { get } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { hari, tgl, ymd } from '@/utils/format'

const PRESET = [
  { label: 'Hari ini', range: () => [hari(), hari()] },
  { label: '7 hari', range: () => [hari(-6), hari()] },
  { label: '30 hari', range: () => [hari(-29), hari()] },
]

const run = useRun()
const rentang = ref([hari(-6), hari()])
const oleh = ref(null)
const users = ref([])
const page = ref(0)
const data = ref({ items: [], total: 0 })
const loading = ref(false)

async function muat() {
  const [dari, sampai] = rentang.value
  if (!dari) return
  loading.value = true
  const params = { dari: ymd(dari), sampai: ymd(sampai || dari), page: page.value }
  if (oleh.value) params.oleh = oleh.value
  await run(async () => (data.value = await get('/audit', { params })))
  loading.value = false
}

watch([rentang, oleh], () => {
  if (rentang.value[1] === null) return
  page.value = 0
  muat()
})

onMounted(() => {
  muat()
  run(async () => (users.value = await get('/users')))
})
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-shield" title="Log Aktivitas" subtitle="Perubahan harga, diskon, pembatalan, refund, stok, dan pengguna">
      <button v-for="p in PRESET" :key="p.label" class="pill" @click="rentang = p.range()">{{ p.label }}</button>
      <DatePicker v-model="rentang" selection-mode="range" date-format="dd/mm/yy" :manual-input="false" show-icon placeholder="Pilih tanggal" />
    </PageHeader>

    <section class="panel box">
      <Select v-model="oleh" :options="users" option-label="nama" option-value="id" placeholder="Semua pengguna" show-clear class="oleh" />
      <DataTable
        :value="data.items"
        :loading="loading"
        size="small"
        lazy
        paginator
        :rows="50"
        :total-records="data.total"
        :first="page * 50"
        empty-message="Belum ada aktivitas pada rentang ini"
        @page="(e) => ((page = e.page), muat())"
      >
        <Column header="Waktu"><template #body="{ data: a }">{{ tgl(a.at) }}</template></Column>
        <Column field="nama" header="Pengguna" />
        <Column field="aksi" header="Aksi" />
        <Column header="Detail"><template #body="{ data: a }"><code class="detail">{{ a.detail }}</code></template></Column>
      </DataTable>
    </section>
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
.oleh {
  max-width: 240px;
}
.detail {
  font-size: 0.8rem;
  word-break: break-all;
  color: var(--p-text-muted-color);
}
</style>
