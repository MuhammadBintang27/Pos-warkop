<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { get, post } from '@/api/http'
import { onEvent } from '@/composables/useEvents'
import { useRun } from '@/composables/useRun'

const run = useRun()
const data = ref({ tenant: '', items: [] })

const jam = (d) => new Date(d).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })

// satu kartu per bill, supaya item satu meja dikerjakan bersama
function kelompok(items) {
  const map = new Map()
  for (const it of items) {
    if (!map.has(it.pesananId)) map.set(it.pesananId, { id: it.pesananId, meja: it.mejaNama, atasNama: it.atasNama, at: it.dibuatAt, items: [] })
    map.get(it.pesananId).items.push(it)
  }
  return [...map.values()]
}
const antrean = computed(() => kelompok(data.value.items.filter((i) => !i.siapAt)))
const siap = computed(() => kelompok(data.value.items.filter((i) => i.siapAt)))

const muat = async () => (data.value = await get('/tenant-saya/pesanan'))
const tandai = (it, s) => run(async () => {
  await post(`/tenant-saya/item/${it.id}/siap`, { siap: s })
  await muat()
})

const offs = []
onMounted(() => {
  run(muat)
  offs.push(onEvent('pesanan', muat), onEvent('dapur', muat))
})
onBeforeUnmount(() => offs.forEach((off) => off()))
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-inbox" title="Pesanan Masuk" :subtitle="data.tenant" />

    <h3 class="judul">Antrean <Tag :value="String(antrean.reduce((n, b) => n + b.items.length, 0))" /></h3>
    <p v-if="!antrean.length" class="kosong">Belum ada pesanan yang perlu dibuat.</p>
    <div class="grid">
      <section v-for="b in antrean" :key="b.id" class="panel kartu">
        <header>
          <strong>{{ b.meja || 'Tanpa meja' }}</strong>
          <span v-if="b.atasNama" class="redup">{{ b.atasNama }}</span>
          <small class="redup kanan">#{{ b.id }} · {{ jam(b.at) }}</small>
        </header>
        <div v-for="it in b.items" :key="it.id" class="baris">
          <div>
            <b>{{ it.qty }}×</b> {{ it.nama }}
            <small v-if="it.opsi" class="redup"> · {{ it.opsi }}</small>
            <div v-if="it.catatan" class="catatan"><i class="pi pi-comment" /> {{ it.catatan }}</div>
          </div>
          <Button label="Siap" icon="pi pi-check" size="small" @click="tandai(it, true)" />
        </div>
      </section>
    </div>

    <template v-if="siap.length">
      <h3 class="judul">Sudah siap <small class="redup">(3 jam terakhir)</small></h3>
      <div class="grid">
        <section v-for="b in siap" :key="b.id" class="panel kartu selesai">
          <header>
            <strong>{{ b.meja || 'Tanpa meja' }}</strong>
            <span v-if="b.atasNama" class="redup">{{ b.atasNama }}</span>
            <small class="redup kanan">#{{ b.id }}</small>
          </header>
          <div v-for="it in b.items" :key="it.id" class="baris">
            <div><b>{{ it.qty }}×</b> {{ it.nama }} <small class="redup">· siap {{ jam(it.siapAt) }}</small></div>
            <Button label="Batal" icon="pi pi-undo" size="small" text severity="secondary" @click="tandai(it, false)" />
          </div>
        </section>
      </div>
    </template>
  </div>
</template>

<style scoped>
.gap {
  gap: 14px;
}
.judul {
  margin: 4px 0 0;
  display: flex;
  align-items: center;
  gap: 8px;
}
.kosong,
.redup {
  color: var(--p-text-muted-color);
}
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
}
.kartu {
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.kartu header {
  display: flex;
  align-items: baseline;
  gap: 8px;
  flex-wrap: wrap;
}
.kanan {
  margin-left: auto;
}
.baris {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding-top: 8px;
  border-top: 1px solid var(--p-content-border-color);
}
.catatan {
  margin-top: 2px;
  font-size: 0.9rem;
  color: var(--accent-text);
}
.selesai {
  opacity: 0.75;
}
</style>
