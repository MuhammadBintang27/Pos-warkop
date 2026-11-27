<script setup>
import { computed, ref, watch } from 'vue'
import { useToast } from 'primevue/usetoast'
import { useRun } from '@/composables/useRun'
import { useAuthStore } from '@/stores/auth'
import { useKasirStore } from '@/stores/kasir'
import { usePengaturanStore } from '@/stores/pengaturan'
import { cetakStruk } from '@/utils/cetak'
import { rp } from '@/utils/format'
import { gabungTerkirim, printDapur, sisaLokal } from '@/utils/print'

defineProps({ bisaTutup: Boolean, lebar: Boolean })
const emit = defineEmits(['bayar', 'pisah', 'pindah', 'gabung', 'denah', 'tutup', 'lebar'])
const store = useKasirStore()
const auth = useAuthStore()
const toast = useToast()
const run = useRun()
const jam = (d) => new Date(d).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })

const p = computed(() => store.pesanan)
const baris = computed(() => gabungTerkirim(p.value?.items ?? []))
const belumKirim = computed(() => p.value?.items.filter((i) => !i.dikirim).length ?? 0)
const jumlahItem = computed(() => p.value?.items.reduce((s, i) => s + i.qty, 0) ?? 0)
const aktif = computed(() => p.value ?? store.draft)
const judul = computed(() => (aktif.value ? aktif.value.mejaNama || 'Bawa pulang' : 'Keranjang'))

// bill lain di meja yang sama (hasil pisah bill), supaya bisa berpindah tanpa membuka denah
const saudara = computed(() => {
  const m = store.denah.meja.find((x) => x.id === p.value?.mejaId)
  return m && m.pesanan.length > 1 ? m.pesanan : []
})
const gantiBill = (id) => run(() => store.bukaPesanan(id))

const diskon = ref(0)
watch(() => p.value?.diskon, (v) => (diskon.value = v ?? 0), { immediate: true })

const menu = ref()
const edit = ref(null)

const aksi = [
  // bill tagihan sebelum bayar, untuk pelanggan yang minta lihat dulu
  { label: 'Cetak bill', icon: 'pi pi-print', command: () => p.value.items.length && run(() => cetakStruk(p.value, usePengaturanStore().settings)) },
  { separator: true },
  { label: 'Pindah meja', icon: 'pi pi-arrow-right-arrow-left', command: () => emit('pindah') },
  { label: 'Gabung bill', icon: 'pi pi-clone', command: () => emit('gabung') },
  { label: 'Pisah bill', icon: 'pi pi-share-alt', command: () => emit('pisah') },
  { separator: true },
  { label: 'Batalkan pesanan', icon: 'pi pi-trash', command: () => bukaBatal(null) },
]

async function simpanDiskon() {
  const nilai = diskon.value || 0
  if (nilai === p.value.diskon) return
  if (!(await run(() => store.setDiskon(nilai)))) diskon.value = p.value.diskon
}

const batal = ref({ buka: false, item: null, awal: 1 })

function bukaBatal(item, awal = 1) {
  if (item?.dikirim && !auth.isKasir) {
    return toast.add({ severity: 'warn', summary: 'Minta kasir untuk membatalkan item yang sudah dikirim', life: 4000 })
  }
  batal.value = { buka: true, item, awal }
}

function hapus(item) {
  return item.dikirim ? bukaBatal(item, item.qty) : run(() => store.hapusItem(item))
}

function ubah(item, delta) {
  if (item.qty + delta <= 0) return hapus(item)
  if (delta < 0 && item.dikirim) return bukaBatal(item, 1)
  // baris gabungan: tambah lewat baris aslinya yang terakhir, server membuatnya baris baru yang belum dikirim
  const asli = item.rows?.at(-1) ?? item
  return run(() => store.updateItem(asli, asli.qty + delta, asli.catatan))
}

async function simpanCatatan() {
  const { item, catatan } = edit.value
  if (await run(() => store.updateItem(item, item.qty, catatan))) edit.value = null
}

async function kirim() {
  let tiket
  const pesan = () => (tiket?.ragu ? 'Disimpan, tiket menunggu kepastian server' : 'Terkirim ke dapur')
  if (!(await run(async () => (tiket = await store.kirim()), pesan))) return
  const lokal = sisaLokal(tiket)
  if (lokal.items.length && usePengaturanStore().settings.dialogDapur) printDapur(lokal)
}

</script>

<template>
  <section class="panel cart">
    <header class="head">
      <i class="pi pi-shopping-cart" />
      <strong>{{ judul }}</strong>
      <Tag :value="`${jumlahItem} item`" severity="secondary" />
      <span v-if="aktif?.atasNama" class="nama">{{ aktif.atasNama }}</span>
      <span class="kanan">
        <Button
          v-if="p && auth.isKasir"
          icon="pi pi-ellipsis-v"
          text
          rounded
          severity="secondary"
          aria-label="Aksi pesanan"
          @click="menu.toggle($event)"
        />
        <Menu ref="menu" :model="aksi" popup />
        <Button
          v-if="bisaTutup"
          :icon="lebar ? 'pi pi-window-minimize' : 'pi pi-window-maximize'"
          text
          rounded
          severity="secondary"
          :aria-label="lebar ? 'Kecilkan keranjang' : 'Perbesar keranjang'"
          @click="emit('lebar')"
        />
        <Button
          v-if="bisaTutup && !lebar"
          icon="pi pi-angle-double-right"
          text
          rounded
          severity="secondary"
          aria-label="Tutup keranjang"
          @click="emit('tutup')"
        />
      </span>
    </header>

    <nav v-if="saudara.length" class="bills" aria-label="Bill di meja ini">
      <button
        v-for="(b, n) in saudara"
        :key="b.id"
        class="bill"
        :class="{ on: b.id === p.id }"
        @click="b.id !== p.id && gantiBill(b.id)"
      >
        Bill {{ n + 1 }} <small>{{ rp(b.total) }}</small>
      </button>
    </nav>

    <div v-if="!aktif" class="empty">
      <i class="pi pi-table" />
      <strong>Belum ada meja dipilih</strong>
      <small>Pilih meja atau bawa pulang untuk mulai memesan</small>
      <Button label="Pilih meja" icon="pi pi-th-large" @click="emit('denah')" />
    </div>
    <div v-else-if="!p?.items.length" class="empty">
      <i class="pi pi-shopping-cart" />
      <strong>Keranjang masih kosong</strong>
      <small>Ketuk menu di sebelah untuk menambah pesanan</small>
    </div>

    <ul v-else class="items">
      <li v-for="i in baris" :key="i.id" class="item" :class="{ lebar }">
        <span class="barisnama" :title="[i.nama, i.opsi, i.catatan && `“${i.catatan}”`, `${rp(i.harga)} per porsi`].filter(Boolean).join(' · ')">
          <i v-if="i.siapAt" class="pi pi-check-circle kirim antar" :title="`Sudah diantar tenant ${jam(i.siapAt)}`" />
          <i v-else-if="i.diprosesAt" class="pi pi-hourglass kirim proses" :title="`Sedang dibuat tenant sejak ${jam(i.diprosesAt)}`" />
          <i v-else-if="i.dikirim" class="pi pi-send kirim" title="Sudah dikirim ke dapur, belum dibuat" />
          <strong>{{ i.nama }}</strong>
          <small v-if="i.opsi">{{ i.opsi }}</small>
          <small v-if="i.catatan" class="note">“{{ i.catatan }}”</small>
        </span>
        <span v-if="lebar" class="satuan">{{ rp(i.harga) }}</span>
        <span class="qty">
          <button class="step" aria-label="Kurangi" @click="ubah(i, -1)"><i class="pi pi-minus" /></button>
          <b>{{ i.qty }}</b>
          <button class="step" aria-label="Tambah" @click="ubah(i, 1)"><i class="pi pi-plus" /></button>
        </span>
        <strong class="sub">{{ rp(i.harga * i.qty) }}</strong>
        <span class="aksi">
          <button v-if="!i.dikirim" class="ikon" aria-label="Catatan" @click="edit = { item: i, catatan: i.catatan }"><i class="pi pi-pencil" /></button>
          <button class="ikon bahaya" aria-label="Hapus" @click="hapus(i)"><i class="pi pi-times" /></button>
        </span>
      </li>
    </ul>

    <footer class="foot">
      <div class="row"><span>Subtotal</span><strong>{{ rp(p?.subtotal) }}</strong></div>
      <div v-if="auth.isKasir" class="row">
        <span>Potongan</span>
        <InputNumber
          v-model="diskon"
          class="diskon"
          mode="currency"
          currency="IDR"
          locale="id-ID"
          :max-fraction-digits="0"
          :min="0"
          :disabled="!p"
          size="small"
          @blur="simpanDiskon"
        />
      </div>
      <div class="row total"><span>Grand Total</span><strong>{{ rp(p?.total) }}</strong></div>
      <Button
        :label="belumKirim ? `Kirim ke dapur (${belumKirim})` : 'Kirim ke dapur'"
        icon="pi pi-send"
        :severity="auth.isKasir ? 'secondary' : undefined"
        :variant="auth.isKasir ? 'outlined' : undefined"
        :disabled="!belumKirim"
        size="small"
        fluid
        @click="kirim"
      />
      <Button
        v-if="auth.isKasir"
        label="Lanjut ke pembayaran"
        icon="pi pi-arrow-right"
        icon-pos="right"
        :disabled="!p || !p.items.length"
        fluid
        @click="emit('bayar')"
      />
    </footer>

    <BatalDialog v-model:visible="batal.buka" :item="batal.item" :awal="batal.awal" />

    <Dialog :visible="!!edit" modal header="Catatan item" :style="{ width: '24rem' }" @update:visible="edit = null">
      <div v-if="edit" class="stack">
        <strong>{{ edit.item.nama }}</strong>
        <InputText v-model="edit.catatan" maxlength="255" placeholder="mis. tidak pedas" autofocus fluid @keyup.enter="simpanCatatan" />
      </div>
      <template #footer>
        <Button label="Simpan" icon="pi pi-check" @click="simpanCatatan" />
      </template>
    </Dialog>
  </section>
</template>

<style scoped>
.cart {
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}
.head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 14px;
  border-bottom: 1px solid var(--p-content-border-color);
}
.head .pi-shopping-cart {
  color: var(--accent-text);
}
.nama {
  color: var(--p-text-muted-color);
  font-size: 0.9rem;
}
.kanan {
  margin-left: auto;
  display: flex;
  align-items: center;
}
.empty {
  flex: 1;
  display: grid;
  place-content: center;
  justify-items: center;
  gap: 6px;
  padding: 24px;
  text-align: center;
  color: var(--p-text-muted-color);
}
.empty .pi {
  font-size: 2.4rem;
}
.empty .p-button {
  margin-top: 10px;
}
.items {
  flex: 1;
  overflow-y: auto;
  margin: 0;
  padding: 0;
  list-style: none;
}
.item {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto auto auto;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-bottom: 1px solid var(--p-content-border-color);
}
.bills {
  display: flex;
  gap: 6px;
  padding: 8px 12px;
  overflow-x: auto;
  border-bottom: 1px solid var(--p-content-border-color);
}
.bill {
  flex: none;
  padding: 5px 12px;
  border: 1px solid var(--p-content-border-color);
  border-radius: 999px;
  background: var(--p-content-background);
  color: var(--p-text-color);
  font: inherit;
  font-weight: 600;
  cursor: pointer;
}
.bill small {
  margin-left: 4px;
  font-weight: 500;
  opacity: 0.8;
}
.bill.on {
  background: var(--p-primary-color);
  border-color: var(--p-primary-color);
  color: var(--p-primary-contrast-color);
}
.item.lebar {
  grid-template-columns: minmax(0, 1fr) 5.5rem auto auto auto;
  padding: 10px 18px;
}
.satuan {
  text-align: right;
  color: var(--p-text-muted-color);
}
.barisnama {
  display: flex;
  align-items: baseline;
  gap: 6px;
  min-width: 0;
  white-space: nowrap;
}
.barisnama strong {
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 0 1 auto;
}
.barisnama small {
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 0 5 auto;
  color: var(--p-text-muted-color);
}
.note {
  font-style: italic;
}
.kirim {
  flex: none;
  font-size: 0.7rem;
  color: var(--p-text-muted-color);
}
.kirim.antar {
  color: var(--p-green-500);
}
.kirim.proses {
  color: var(--p-orange-500);
}
.qty {
  display: flex;
  align-items: center;
  gap: 5px;
}
.qty b {
  min-width: 1.4ch;
  text-align: center;
}
.step,
.ikon {
  display: grid;
  place-items: center;
  width: 24px;
  height: 24px;
  border: 0;
  border-radius: 50%;
  font-size: 0.7rem;
  color: var(--p-text-color);
  background: color-mix(in srgb, var(--p-text-color) 8%, transparent);
  cursor: pointer;
}
.step:hover {
  background: color-mix(in srgb, var(--p-primary-color) 22%, transparent);
}
.ikon {
  background: none;
  color: var(--p-text-muted-color);
}
.ikon:hover {
  background: color-mix(in srgb, var(--p-text-color) 8%, transparent);
}
.ikon.bahaya:hover {
  color: var(--p-red-500, #ef4444);
}
.sub {
  min-width: 4.6rem;
  text-align: right;
  font-variant-numeric: tabular-nums;
}
.aksi {
  display: flex;
  justify-content: flex-end;
  gap: 2px;
  min-width: 52px;
}
.foot {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 10px 12px;
  border-top: 1px solid var(--p-content-border-color);
  background: color-mix(in srgb, var(--p-primary-color) 4%, var(--p-content-background));
}
.row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.diskon {
  width: 11rem;
}
.diskon :deep(input) {
  width: 100%;
  text-align: right;
}
.total {
  padding-top: 6px;
  border-top: 1px solid var(--p-content-border-color);
  font-size: 1.1rem;
  color: var(--accent-text);
  font-weight: 700;
}
</style>
