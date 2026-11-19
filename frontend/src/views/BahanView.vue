<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useCari } from '@/composables/useCari'
import { useConfirm } from 'primevue/useconfirm'
import { del, get, post, put } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { useAuthStore } from '@/stores/auth'
import { konfirmasi } from '@/utils/confirm'
import { hari, rp, tgl, ymd } from '@/utils/format'

const STATUS = {
  ok: { label: 'Cukup', severity: 'success' },
  menipis: { label: 'Menipis', severity: 'warn' },
  habis: { label: 'Habis', severity: 'danger' },
}
const JENIS = {
  restock: { label: 'Restock', severity: 'success' },
  opname: { label: 'Opname', severity: 'info' },
  pakai: { label: 'Terpakai', severity: 'secondary' },
  kembali: { label: 'Dikembalikan', severity: 'secondary' },
}
const MODE = [
  { label: 'Stok', value: 'stok' },
  { label: 'Pemakaian', value: 'pemakaian' },
]
const PRESET = [
  { label: 'Hari ini', range: () => [hari(), hari()] },
  { label: '7 hari', range: () => [hari(-6), hari()] },
  { label: 'Bulan ini', range: () => [new Date(new Date().getFullYear(), new Date().getMonth(), 1), hari()] },
]

const confirm = useConfirm()
const run = useRun()
const auth = useAuthStore()

const mode = ref('stok')
const list = ref([])
const form = ref(null)
const restock = ref(null)
const opname = ref(null)
const riwayat = ref(null)
const rentang = ref([hari(-6), hari()])
const pemakaian = ref([])

const n = (v) => Number(v).toLocaleString('id-ID')
const menipis = computed(() => list.value.filter((b) => b.status === 'menipis').length)
const habis = computed(() => list.value.filter((b) => b.status === 'habis').length)

const muat = async () => (list.value = await get('/bahan'))

async function muatPemakaian() {
  const [dari, sampai] = rentang.value
  if (!dari) return
  await run(async () => (pemakaian.value = await get(`/bahan/pemakaian?dari=${ymd(dari)}&sampai=${ymd(sampai || dari)}`)))
}

async function simpan() {
  const { id, nama, satuan, stok, minimum, hargaRata } = form.value
  const body = { nama: nama.trim(), satuan: satuan.trim(), stok: stok || 0, minimum: minimum || 0, hargaRata: hargaRata || 0 }
  if (await run(() => (id ? put(`/bahan/${id}`, body) : post('/bahan', body)), 'Bahan disimpan')) {
    form.value = null
    await run(muat)
  }
}

async function simpanRestock() {
  const { id, jumlah, hargaBeli, catatan, dariKas } = restock.value
  if (await run(() => post(`/bahan/${id}/restock`, { jumlah, hargaBeli: hargaBeli || 0, catatan, dariKas }), 'Stok bertambah')) {
    restock.value = null
    await run(muat)
  }
}

async function simpanOpname() {
  const { id, stokAktual, catatan } = opname.value
  if (await run(() => post(`/bahan/${id}/opname`, { stokAktual: stokAktual ?? 0, catatan }), 'Stok disesuaikan')) {
    opname.value = null
    await run(muat)
  }
}

async function bukaRiwayat(b) {
  riwayat.value = { bahan: b, log: [] }
  await run(async () => (riwayat.value.log = await get(`/bahan/${b.id}/riwayat`)))
}

function hapus(b) {
  konfirmasi(confirm, {
    header: 'Hapus bahan?',
    message: `"${b.nama}" dihapus dari daftar. Bahan yang masih dipakai di resep tidak bisa dihapus.`,
    accept: async () => (await run(() => del(`/bahan/${b.id}`), 'Bahan dihapus')) && run(muat),
  })
}

watch(rentang, (r) => r[1] !== null && muatPemakaian())
watch(mode, (m) => m === 'pemakaian' && muatPemakaian())
onMounted(() => run(muat))
const { cari: cariBahan, hasil: bahanTampil } = useCari(list, 'nama')
const { cari: cariPakai, hasil: pemakaianTampil } = useCari(pemakaian, 'nama')
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-box" title="Stok Bahan" subtitle="Bahan baku untuk resep menu">
      <button v-if="auth.isOwner" class="pill" @click="form = { nama: '', satuan: '', stok: 0, minimum: 0, hargaRata: 0 }"><i class="pi pi-plus" /> Tambah bahan</button>
    </PageHeader>

    <div class="stats">
      <StatCard icon="pi pi-box" label="Jenis bahan" :value="String(list.length)" />
      <StatCard icon="pi pi-exclamation-triangle" label="Menipis" :value="String(menipis)" />
      <StatCard icon="pi pi-times-circle" label="Habis" :value="String(habis)" />
    </div>

    <SelectButton v-model="mode" :options="MODE" option-label="label" option-value="value" :allow-empty="false" />

    <section v-if="mode === 'stok'" class="panel box">
      <CariInput v-model="cariBahan" placeholder="Cari bahan…" />
      <DataTable paginator :rows="10" :rows-per-page-options="[10, 25, 50]" :always-show-paginator="false" :value="bahanTampil" size="small" empty-message="Belum ada bahan">
        <Column field="nama" header="Bahan" sortable />
        <Column header="Stok" sortable field="stok">
          <template #body="{ data }">{{ n(data.stok) }} {{ data.satuan }}</template>
        </Column>
        <Column header="Harga / satuan">
          <template #body="{ data }">{{ data.hargaRata ? rp(Math.round(data.hargaRata * 100) / 100) : '—' }}</template>
        </Column>
        <Column header="Minimum">
          <template #body="{ data }">{{ data.minimum ? `${n(data.minimum)} ${data.satuan}` : '—' }}</template>
        </Column>
        <Column header="Status">
          <template #body="{ data }"><Tag :value="STATUS[data.status].label" :severity="STATUS[data.status].severity" /></template>
        </Column>
        <Column>
          <template #body="{ data }">
            <div class="act">
              <Button label="Restock" icon="pi pi-plus" size="small" text @click="restock = { id: data.id, nama: data.nama, satuan: data.satuan, jumlah: null, hargaBeli: 0, catatan: '', dariKas: true }" />
              <Button label="Opname" icon="pi pi-check-square" size="small" text severity="secondary" @click="opname = { ...data, stokAktual: data.stok, catatan: '' }" />
              <Button icon="pi pi-history" size="small" text severity="secondary" aria-label="Riwayat" @click="bukaRiwayat(data)" />
              <template v-if="auth.isOwner">
                <Button icon="pi pi-pencil" size="small" text severity="secondary" aria-label="Ubah" @click="form = { ...data }" />
                <Button icon="pi pi-trash" size="small" text severity="danger" aria-label="Hapus" @click="hapus(data)" />
              </template>
            </div>
          </template>
        </Column>
      </DataTable>
    </section>

    <template v-else>
      <section class="panel box filter">
        <button v-for="p in PRESET" :key="p.label" class="pill" @click="rentang = p.range()">{{ p.label }}</button>
        <DatePicker v-model="rentang" selection-mode="range" date-format="dd/mm/yy" :manual-input="false" show-icon placeholder="Pilih tanggal" />
      </section>
      <section class="panel box">
        <CariInput v-model="cariPakai" placeholder="Cari bahan…" />
        <DataTable paginator :rows="10" :rows-per-page-options="[10, 25, 50]" :always-show-paginator="false" :value="pemakaianTampil" size="small" empty-message="Belum ada bahan">
          <Column field="nama" header="Bahan" sortable />
          <Column header="Terpakai" sortable field="dipakai"><template #body="{ data }">{{ n(data.dipakai) }} {{ data.satuan }}</template></Column>
          <Column header="Restock"><template #body="{ data }">{{ n(data.restock) }} {{ data.satuan }}</template></Column>
          <Column header="Selisih opname">
            <template #body="{ data }">
              <span :class="{ minus: data.selisihOpname < 0 }">{{ data.selisihOpname > 0 ? '+' : '' }}{{ n(data.selisihOpname) }} {{ data.satuan }}</span>
            </template>
          </Column>
          <Column header="Stok sekarang"><template #body="{ data }">{{ n(data.stokSekarang) }} {{ data.satuan }}</template></Column>
        </DataTable>
        <small class="hint">Selisih opname negatif berarti stok fisik lebih sedikit dari catatan sistem, misalnya karena tumpah atau takaran yang tidak pas.</small>
      </section>
    </template>

    <Dialog :visible="!!form" modal :header="form?.id ? 'Ubah bahan' : 'Tambah bahan'" :style="{ width: '24rem' }" @update:visible="form = null">
      <div v-if="form" class="stack">
        <label for="nama">Nama bahan</label>
        <InputText id="nama" v-model="form.nama" maxlength="100" placeholder="Biji kopi" fluid />
        <label for="satuan">Satuan</label>
        <InputText id="satuan" v-model="form.satuan" maxlength="10" placeholder="gram, ml, pcs" fluid />
        <template v-if="!form.id">
          <label for="stok">Stok awal</label>
          <InputNumber id="stok" v-model="form.stok" :min="0" fluid />
        </template>
        <label for="hr">Harga per satuan (untuk HPP)</label>
        <InputNumber id="hr" v-model="form.hargaRata" :min="0" :max-fraction-digits="2" mode="currency" currency="IDR" locale="id-ID" fluid />
        <small class="hint">Misalnya biji kopi Rp120 per gram. Setelah ini, restock dengan harga beli menghitung ulang rata-ratanya otomatis.</small>
        <label for="min">Batas minimum (peringatan menipis)</label>
        <InputNumber id="min" v-model="form.minimum" :min="0" fluid />
        <small class="hint">Gunakan satuan terkecil dan bilangan bulat, misalnya gram dan ml. Stok berikutnya diubah lewat Restock atau Opname.</small>
      </div>
      <template #footer><Button label="Simpan" icon="pi pi-check" :disabled="!form?.nama.trim() || !form?.satuan.trim()" @click="simpan" /></template>
    </Dialog>

    <Dialog :visible="!!restock" modal :header="`Restock ${restock?.nama}`" :style="{ width: '24rem' }" @update:visible="restock = null">
      <div v-if="restock" class="stack">
        <label for="rj">Jumlah masuk ({{ restock.satuan }})</label>
        <InputNumber id="rj" v-model="restock.jumlah" :min="1" fluid autofocus />
        <label>Total harga beli{{ restock.dariKas ? '' : ' (opsional)' }}</label>
        <MoneyInput v-model="restock.hargaBeli" />
        <label for="rc">Catatan</label>
        <InputText id="rc" v-model="restock.catatan" maxlength="255" placeholder="mis. beli di pasar" fluid />
        <div class="cek">
          <Checkbox v-model="restock.dariKas" input-id="rk" binary />
          <label for="rk">Dibayar dari kas laci (dicatat sebagai pengeluaran shift)</label>
        </div>
      </div>
      <template #footer><Button label="Tambah stok" icon="pi pi-check" :disabled="!restock?.jumlah || (restock.dariKas && !restock.hargaBeli)" @click="simpanRestock" /></template>
    </Dialog>

    <Dialog :visible="!!opname" modal :header="`Opname ${opname?.nama}`" :style="{ width: '24rem' }" @update:visible="opname = null">
      <div v-if="opname" class="stack">
        <small class="hint">Hitung stok fisik, lalu isi jumlah sebenarnya. Selisih dengan catatan sistem tersimpan di riwayat.</small>
        <label for="os">Stok sebenarnya ({{ opname.satuan }})</label>
        <InputNumber id="os" v-model="opname.stokAktual" :min="0" fluid autofocus />
        <small class="hint">
          Sistem: {{ n(opname.stok) }} · selisih:
          <b>{{ (opname.stokAktual ?? 0) - opname.stok > 0 ? '+' : '' }}{{ n((opname.stokAktual ?? 0) - opname.stok) }}</b>
        </small>
        <label for="oc">Catatan</label>
        <InputText id="oc" v-model="opname.catatan" maxlength="255" fluid />
      </div>
      <template #footer><Button label="Simpan" icon="pi pi-check" @click="simpanOpname" /></template>
    </Dialog>

    <Dialog :visible="!!riwayat" modal :header="`Riwayat ${riwayat?.bahan.nama}`" :style="{ width: '44rem' }" :breakpoints="{ '700px': '96vw' }" @update:visible="riwayat = null">
      <DataTable paginator :rows="10" :rows-per-page-options="[10, 25, 50]" :always-show-paginator="false" v-if="riwayat" :value="riwayat.log" size="small" empty-message="Belum ada catatan">
        <Column header="Waktu"><template #body="{ data }">{{ tgl(data.at) }}</template></Column>
        <Column header="Jenis"><template #body="{ data }"><Tag :value="JENIS[data.jenis].label" :severity="JENIS[data.jenis].severity" /></template></Column>
        <Column header="Perubahan">
          <template #body="{ data }"><span :class="{ minus: data.jumlah < 0 }">{{ data.jumlah > 0 ? '+' : '' }}{{ n(data.jumlah) }}</span></template>
        </Column>
        <Column header="Stok akhir"><template #body="{ data }">{{ n(data.stokAkhir) }}</template></Column>
        <Column header="Harga beli"><template #body="{ data }">{{ data.hargaBeli ? rp(data.hargaBeli) : '—' }}</template></Column>
        <Column field="oleh" header="Oleh" />
        <Column field="catatan" header="Catatan" />
      </DataTable>
    </Dialog>
  </div>
</template>

<style scoped>
.cek {
  display: flex;
  align-items: center;
  gap: 10px;
}
.gap {
  gap: 14px;
}
.stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
}
.box {
  padding: 14px;
}
.filter {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}
.act {
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 2px;
}
.hint {
  color: var(--p-text-muted-color);
}
.minus {
  color: var(--p-red-500);
}
</style>
