<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useCari } from '@/composables/useCari'
import { useRoute } from 'vue-router'
import { get } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { hari, rp, tgl, ymd } from '@/utils/format'
import { unduhExcel } from '@/utils/ekspor'

const run = useRun()

const PRESET = [
  { label: 'Hari ini', range: () => [hari(), hari()] },
  { label: '7 hari', range: () => [hari(-6), hari()] },
  { label: 'Bulan ini', range: () => [new Date(new Date().getFullYear(), new Date().getMonth(), 1), hari()] },
]
const MODE = [
  { label: 'Ringkasan & Tren', value: 'ringkasan', icon: 'pi pi-chart-line' },
  { label: 'Kategori, Tenant & Jam', value: 'waktu', icon: 'pi pi-th-large' },
  { label: 'Kinerja Karyawan', value: 'karyawan', icon: 'pi pi-users' },
  { label: 'Kasbon & Stok', value: 'kasbon', icon: 'pi pi-book' },
]
const HARI = ['Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu', 'Minggu']
const ROLE = { owner: 'Owner', kasir: 'Kasir', pelayan: 'Pelayan' }

const route = useRoute()
const rentang = ref([hari(-6), hari()])
const mode = computed(() => route.params.tab || 'ringkasan') // tab dipilih dari menu navigasi
const judul = computed(() => MODE.find((m) => m.value === mode.value)?.label)
const data = ref(null) // /laporan, hanya untuk tab ringkasan
const setoran = ref(null)
const an = ref({}) // gabungan bagian analitik yang sudah dimuat
const memuat = ref(false)
const totalSetor = computed(() => setoran.value?.tenant.reduce((s, t) => s + t.setor, 0) ?? 0)
const subtitle = computed(() => {
  const [d, s] = rentang.value
  return d ? `${ymd(d)} s/d ${ymd(s || d)}` : ''
})

// tiap tab hanya mengambil datanya sendiri, sekali per rentang tanggal; pindah tab berikutnya tidak memanggil server lagi
const KUNCI = {
  ringkasan: ['tren', 'ringkasan', 'sebelumnya', 'sebelumnyaDari', 'sebelumnyaSampai'],
  waktu: ['jamRamai', 'kategori', 'tenant'],
  karyawan: ['karyawan'],
  kasbon: ['kasbon', 'stok'],
}
const termuat = new Set()
let versi = 0

async function muatTab(tab) {
  const [dari, sampai] = rentang.value
  if (!dari || termuat.has(tab)) return
  const v = versi
  const q = `dari=${ymd(dari)}&sampai=${ymd(sampai || dari)}`
  termuat.add(tab)
  memuat.value = true
  const ok = await run(async () => {
    const [a, l, s] = await Promise.all([
      get(`/laporan/analitik?${q}&bagian=${tab}`),
      tab === 'ringkasan' ? get(`/laporan?${q}`) : null,
      tab === 'ringkasan' ? get(`/laporan/tenant?${q}`) : null,
    ])
    if (v !== versi) return // rentang sudah diganti saat menunggu
    an.value = { ...an.value, ...Object.fromEntries(KUNCI[tab].map((k) => [k, a[k]])) }
    if (tab === 'ringkasan') [data.value, setoran.value] = [l, s]
  })
  if (!ok) termuat.delete(tab)
  if (v === versi) memuat.value = false
}

watch(rentang, (r) => {
  if (r[1] === null) return
  versi++
  termuat.clear()
  an.value = {}
  data.value = setoran.value = null
  muatTab(mode.value)
})
watch(mode, muatTab)
onMounted(() => muatTab(mode.value))

// ---------- ringkasan & perbandingan ----------
function delta(k) {
  const a = an.value?.ringkasan[k]
  const b = an.value?.sebelumnya[k]
  if (a == null || !b) return null
  return Math.round(((a - b) / Math.abs(b)) * 100)
}
const r = computed(() => an.value?.ringkasan ?? {})

// ---------- grafik ----------
const singkat = (n) => {
  const a = Math.abs(n)
  if (a >= 1e6) return (n / 1e6).toFixed(1).replace('.', ',') + 'jt'
  if (a >= 1e3) return Math.round(n / 1e3) + 'rb'
  return String(n)
}
function warna() {
  const s = getComputedStyle(document.documentElement)
  const v = (n, d) => s.getPropertyValue(n).trim() || d
  return { primary: v('--p-primary-color', '#6366f1'), teks: v('--p-text-muted-color', '#888'), garis: v('--p-content-border-color', '#e5e7eb') }
}
const alpha = (c, a) => (/^#[0-9a-f]{6}$/i.test(c) ? c + Math.round(a * 255).toString(16).padStart(2, '0') : c)
const PALET = ['#6366f1', '#22c55e', '#f59e0b', '#ef4444', '#06b6d4', '#a855f7', '#ec4899', '#84cc16', '#64748b']

function opsiSumbu(w, extra = {}) {
  return {
    maintainAspectRatio: false,
    interaction: { mode: 'index', intersect: false },
    plugins: {
      legend: { labels: { color: w.teks, usePointStyle: true } },
      tooltip: { callbacks: { label: (c) => `${c.dataset.label}: ${rp(c.parsed.y ?? c.parsed.x)}` } },
    },
    scales: {
      x: { ticks: { color: w.teks }, grid: { display: false } },
      y: { ticks: { color: w.teks, callback: singkat }, grid: { color: w.garis } },
    },
    ...extra,
  }
}

const trenChart = computed(() => {
  const t = an.value?.tren ?? []
  const w = warna()
  return {
    data: {
      labels: t.map((x) => x.tanggal.slice(8) + '/' + x.tanggal.slice(5, 7)),
      datasets: [
        { type: 'bar', label: 'Omzet', data: t.map((x) => x.omzet), backgroundColor: alpha(w.primary, 0.35), borderRadius: 6, order: 3 },
        // garis lurus (bukan kurva): data per hari, kurva memberi kesan ada nilai di antara hari
        // laba kotor putus-putus biru supaya tetap terlihat saat sama dengan laba bersih (hari tanpa biaya)
        { type: 'line', label: 'Laba kotor', data: t.map((x) => x.labaKotor), borderColor: '#3b82f6', backgroundColor: '#3b82f6', borderDash: [6, 4], pointRadius: 3, order: 1 },
        { type: 'line', label: 'Laba bersih', data: t.map((x) => x.labaBersih), borderColor: '#16a34a', backgroundColor: '#16a34a', borderWidth: 3, pointRadius: 3, order: 2 },
      ],
    },
    options: opsiSumbu(w),
  }
})

const metodeChart = computed(() => {
  const m = data.value?.perMetode ?? []
  const w = warna()
  return {
    data: { labels: m.map((x) => x.metode.toUpperCase()), datasets: [{ data: m.map((x) => x.total), backgroundColor: PALET, borderWidth: 0 }] },
    options: {
      maintainAspectRatio: false,
      cutout: '62%',
      plugins: { legend: { position: 'right', labels: { color: w.teks, usePointStyle: true } }, tooltip: { callbacks: { label: (c) => `${c.label}: ${rp(c.parsed)}` } } },
    },
  }
})

const kategoriChart = computed(() => {
  const k = an.value?.kategori ?? []
  const w = warna()
  return {
    data: { labels: k.map((x) => x.nama), datasets: [{ label: 'Penjualan', data: k.map((x) => x.total), backgroundColor: k.map((_, i) => PALET[i % PALET.length]), borderRadius: 6 }] },
    options: opsiSumbu(w, {
      indexAxis: 'y',
      plugins: { legend: { display: false }, tooltip: { callbacks: { label: (c) => rp(c.parsed.x) } } },
      scales: { x: { ticks: { color: w.teks, callback: singkat }, grid: { color: w.garis } }, y: { ticks: { color: w.teks }, grid: { display: false } } },
    }),
  }
})

const tenantChart = computed(() => {
  const t = an.value?.tenant ?? []
  const w = warna()
  return {
    data: { labels: t.map((x) => x.nama), datasets: [{ data: t.map((x) => x.total), backgroundColor: PALET, borderWidth: 0 }] },
    options: {
      maintainAspectRatio: false,
      cutout: '62%',
      plugins: { legend: { position: 'right', labels: { color: w.teks, usePointStyle: true } }, tooltip: { callbacks: { label: (c) => `${c.label}: ${rp(c.parsed)}` } } },
    },
  }
})
const totalTenant = computed(() => Math.max(1, (an.value?.tenant ?? []).reduce((s, x) => s + x.total, 0)))

// ---------- jam ramai (heatmap hari x jam) ----------
const heat = computed(() => {
  const j = an.value?.jamRamai ?? []
  const sel = Object.fromEntries(j.map((x) => [`${x.hari}-${x.jam}`, x]))
  const jamAda = j.map((x) => x.jam)
  const dari = jamAda.length ? Math.min(...jamAda) : 8
  const sampai = jamAda.length ? Math.max(...jamAda) : 22
  const jam = Array.from({ length: sampai - dari + 1 }, (_, i) => dari + i)
  const maks = Math.max(1, ...j.map((x) => x.bill))
  const puncak = [...j].sort((a, b) => b.bill - a.bill)[0]
  const perJam = jam.map((h) => j.filter((x) => x.jam === h).reduce((s, x) => s + x.bill, 0))
  const jamTersibuk = perJam.length ? jam[perJam.indexOf(Math.max(...perJam))] : null
  const perHari = HARI.map((_, d) => j.filter((x) => x.hari === d).reduce((s, x) => s + x.omzet, 0))
  const hariTerbaik = perHari.some((x) => x) ? HARI[perHari.indexOf(Math.max(...perHari))] : null
  return { sel, jam, maks, puncak, jamTersibuk, hariTerbaik }
})
const isiSel = (d, h) => heat.value.sel[`${d}-${h}`]
const gayaSel = (d, h) => {
  const c = isiSel(d, h)
  return c ? { background: `color-mix(in srgb, var(--p-primary-color) ${Math.round(12 + (c.bill / heat.value.maks) * 88)}%, transparent)` } : {}
}

// ---------- kasbon & stok ----------
const totalKasbon = computed(() => (an.value?.kasbon ?? []).reduce((s, k) => s + k.sisa, 0))
const kasbonLama = computed(() => (an.value?.kasbon ?? []).filter((k) => k.umurHari > 30).reduce((s, k) => s + k.sisa, 0))
const nilaiStok = computed(() => (an.value?.stok ?? []).reduce((s, b) => s + b.nilai, 0))
const totalSusut = computed(() => (an.value?.stok ?? []).reduce((s, b) => s + b.susutNilai, 0))
const umurTag = (h) => (h > 30 ? 'danger' : h > 7 ? 'warn' : 'success')

// ---------- ekspor ----------
const uang = (header, key) => ({ header, key, uang: true, lebar: 16 })
async function excel() {
  // Excel berisi semua tab, jadi tab yang belum dibuka dimuat dulu
  await Promise.all(Object.keys(KUNCI).map(muatTab))
  const d = data.value
  const a = an.value
  if (!d || !a.tren || !a.jamRamai || !a.tenant || !a.karyawan || !a.kasbon) return
  run(
    () =>
      unduhExcel(`laporan-${d.dari}_${d.sampai}.xlsx`, [
        {
          nama: 'Ringkasan',
          kolom: [{ header: 'Keterangan', key: 'k', lebar: 28 }, uang('Periode ini', 'a'), uang('Periode lalu', 'b')],
          baris: [
            { k: `Periode (${d.dari} s/d ${d.sampai} vs ${a.sebelumnyaDari} s/d ${a.sebelumnyaSampai})` },
            { k: 'Omzet', a: a.ringkasan.omzet, b: a.sebelumnya.omzet },
            { k: 'Transaksi', a: a.ringkasan.transaksi, b: a.sebelumnya.transaksi },
            { k: 'Rata-rata per transaksi', a: a.ringkasan.rataRata, b: a.sebelumnya.rataRata },
            { k: 'Pendapatan', a: a.ringkasan.pendapatan, b: a.sebelumnya.pendapatan },
            { k: 'HPP + susut', a: a.ringkasan.hpp, b: a.sebelumnya.hpp },
            { k: 'Laba kotor', a: a.ringkasan.labaKotor, b: a.sebelumnya.labaKotor },
            { k: 'Laba bersih', a: a.ringkasan.labaBersih, b: a.sebelumnya.labaBersih },
            { k: 'Diskon', a: d.diskon },
            { k: 'Pengeluaran (uang keluar)', a: d.pengeluaran },
          ],
        },
        {
          nama: 'Tren harian',
          kolom: [{ header: 'Tanggal', key: 'tanggal' }, uang('Omzet', 'omzet'), { header: 'Transaksi', key: 'transaksi' }, uang('Pendapatan', 'pendapatan'), uang('HPP + susut', 'hpp'), uang('Biaya', 'biaya'), uang('Laba kotor', 'labaKotor'), uang('Laba bersih', 'labaBersih')],
          baris: a.tren,
        },
        { nama: 'Metode bayar', kolom: [{ header: 'Metode', key: 'metode' }, uang('Total', 'total')], baris: d.perMetode },
        { nama: 'Menu terlaris', kolom: [{ header: 'Menu', key: 'nama', lebar: 30 }, { header: 'Terjual', key: 'qty' }, uang('Nilai', 'total')], baris: d.terlaris },
        { nama: 'Tenant', kolom: [{ header: 'Tenant', key: 'nama', lebar: 24 }, { header: 'Bill', key: 'bill' }, { header: 'Item terjual', key: 'qty' }, uang('Penjualan', 'total')], baris: a.tenant },
        { nama: 'Kategori', kolom: [{ header: 'Kategori', key: 'nama', lebar: 24 }, { header: 'Terjual', key: 'qty' }, uang('Penjualan', 'total')], baris: a.kategori },
        {
          nama: 'Jam ramai',
          kolom: [{ header: 'Hari', key: 'hari' }, { header: 'Jam', key: 'jam' }, { header: 'Bill', key: 'bill' }, uang('Omzet', 'omzet')],
          baris: a.jamRamai.map((x) => ({ ...x, hari: HARI[x.hari], jam: `${String(x.jam).padStart(2, '0')}:00` })),
        },
        {
          nama: 'Karyawan',
          kolom: [
            { header: 'Nama', key: 'nama', lebar: 22 }, { header: 'Role', key: 'role' }, { header: 'Bill dibuat', key: 'billDibuat' }, uang('Nilai dibuat', 'nilaiDibuat'),
            { header: 'Bill dibayar', key: 'billDibayar' }, uang('Uang diterima', 'nilaiDibayar'), uang('Diskon', 'diskon'), { header: 'Item batal', key: 'batalQty' },
            uang('Nilai batal', 'batalNilai'), { header: 'Shift', key: 'shift' }, uang('Selisih kas', 'selisihKas'),
          ],
          baris: a.karyawan,
        },
        {
          nama: 'Shift',
          kolom: [{ header: 'Dibuka oleh', key: 'kasir', lebar: 20 }, { header: 'Buka', key: 'buka', lebar: 20 }, { header: 'Ditutup oleh', key: 'tutupOleh', lebar: 20 }, uang('Kas sistem', 'kasSistem'), uang('Kas dihitung', 'kasHitung'), uang('Selisih', 'selisih')],
          baris: d.shift.map((s) => ({ ...s, buka: tgl(s.bukaAt) })),
        },
        {
          nama: 'Kasbon',
          kolom: [{ header: 'Pelanggan', key: 'nama', lebar: 22 }, { header: 'Telp', key: 'telp', lebar: 16 }, uang('Sisa utang', 'sisa'), { header: 'Umur (hari)', key: 'umurHari' }],
          baris: a.kasbon,
        },
        {
          nama: 'Stok',
          kolom: [{ header: 'Bahan', key: 'nama', lebar: 24 }, { header: 'Satuan', key: 'satuan' }, { header: 'Stok', key: 'stok' }, uang('Nilai stok', 'nilai'), { header: 'Susut', key: 'susutQty' }, uang('Nilai susut', 'susutNilai')],
          baris: a.stok,
        },
        {
          nama: 'Setoran tenant',
          kolom: [{ header: 'Tenant', key: 'nama', lebar: 22 }, uang('Penjualan', 'kotor'), uang('Diskon', 'diskon'), { header: 'Komisi %', key: 'komisiPersen' }, uang('Komisi', 'komisi'), uang('Setor', 'setor')],
          baris: setoran.value?.tenant ?? [],
        },
      ]),
    'Excel diunduh',
  )
}
const cetak = () => window.print()
const { cari: cariKaryawan, hasil: karyawanTampil } = useCari(computed(() => an.value.karyawan), 'nama')
const { cari: cariKasbon, hasil: kasbonTampil } = useCari(computed(() => an.value.kasbon), 'nama', 'telp')
const { cari: cariStok, hasil: stokTampil } = useCari(computed(() => an.value.stok), 'nama')
</script>

<template>
  <div class="stack gap">
    <PageHeader :icon="MODE.find((m) => m.value === mode)?.icon" :title="judul" :subtitle="subtitle">
      <button v-for="p in PRESET" :key="p.label" class="pill no-print" @click="rentang = p.range()">{{ p.label }}</button>
      <DatePicker v-model="rentang" class="no-print" selection-mode="range" date-format="dd/mm/yy" :manual-input="false" show-icon placeholder="Pilih tanggal" />
      <Button class="no-print" label="Excel" icon="pi pi-file-excel" severity="secondary" outlined :loading="memuat" @click="excel" />
      <Button class="no-print" label="PDF" icon="pi pi-file-pdf" severity="secondary" outlined @click="cetak" />
    </PageHeader>

    <p v-if="memuat" class="none"><i class="pi pi-spin pi-spinner" /> Memuat…</p>

      <!-- ================= RINGKASAN ================= -->
      <template v-if="mode === 'ringkasan' && data && an.tren">
        <div class="stats">
          <StatCard icon="pi pi-wallet" label="Omzet" :value="rp(r.omzet)" :delta="delta('omzet')" />
          <StatCard icon="pi pi-receipt" label="Transaksi" :value="String(r.transaksi)" :delta="delta('transaksi')" />
          <StatCard icon="pi pi-calculator" label="Rata-rata / transaksi" :value="rp(r.rataRata)" :delta="delta('rataRata')" />
          <StatCard icon="pi pi-chart-line" label="Laba kotor" :value="rp(r.labaKotor)" :delta="delta('labaKotor')" />
          <StatCard icon="pi pi-check-circle" label="Laba bersih" :value="rp(r.labaBersih)" :delta="delta('labaBersih')" />
          <StatCard icon="pi pi-arrow-up-right" label="Pengeluaran" :value="rp(data.pengeluaran)" />
        </div>
        <small class="none">Dibanding {{ an.sebelumnyaDari }} s/d {{ an.sebelumnyaSampai }}. Laba memakai rumus yang sama dengan halaman Laba Rugi.</small>

        <section class="panel box">
          <h3>Tren omzet dan laba</h3>
          <Chart type="bar" :data="trenChart.data" :options="trenChart.options" class="chart" />
          <small class="none">
            Batang = omzet. Garis biru putus-putus = laba kotor (setelah HPP). Garis hijau = laba bersih (setelah pengeluaran dan gaji).
            Bila kedua garis berimpit, hari itu tidak ada pengeluaran. Arahkan kursor ke tanggal untuk melihat angkanya.
          </small>
        </section>

        <div class="two">
          <section class="panel box">
            <h3>Per metode bayar</h3>
            <p v-if="!data.perMetode.length" class="none">Belum ada transaksi.</p>
            <Chart v-else type="doughnut" :data="metodeChart.data" :options="metodeChart.options" class="chart pendek" />
          </section>
          <section class="panel box">
            <h3>Menu terlaris</h3>
            <DataTable paginator :rows="10" :rows-per-page-options="[10, 25, 50]" :always-show-paginator="false" :value="data.terlaris" size="small" empty-message="Belum ada penjualan">
              <Column header="#"><template #body="{ index }">{{ index + 1 }}</template></Column>
              <Column field="nama" header="Menu" />
              <Column field="qty" header="Terjual" />
              <Column header="Nilai"><template #body="{ data: x }">{{ rp(x.total) }}</template></Column>
            </DataTable>
          </section>
        </div>

        <section class="panel box">
          <div class="judul-link">
            <h3>Shift terakhir</h3>
            <RouterLink to="/laporan/shift" class="no-print">Lihat semua shift →</RouterLink>
          </div>
          <DataTable paginator :rows="10" :rows-per-page-options="[10, 25, 50]" :always-show-paginator="false" :value="data.shift" size="small" empty-message="Tidak ada shift">
            <Column field="kasir" header="Dibuka oleh" />
            <Column header="Buka"><template #body="{ data: s }">{{ tgl(s.bukaAt) }}</template></Column>
            <Column header="Ditutup oleh"><template #body="{ data: s }">{{ s.tutupOleh || '—' }}</template></Column>
            <Column header="Kas sistem"><template #body="{ data: s }">{{ s.kasSistem == null ? '—' : rp(s.kasSistem) }}</template></Column>
            <Column header="Selisih">
              <template #body="{ data: s }">
                <Tag v-if="s.tutupAt == null" value="Berjalan" severity="info" />
                <Tag v-else :value="rp(s.selisih)" :severity="s.selisih === 0 ? 'success' : 'danger'" />
              </template>
            </Column>
          </DataTable>
        </section>

        <section v-if="setoran && setoran.tenant.length > 1" class="panel box">
          <h3>Setoran tenant</h3>
          <DataTable paginator :rows="10" :rows-per-page-options="[10, 25, 50]" :always-show-paginator="false" :value="setoran.tenant" size="small">
            <Column field="nama" header="Tenant" />
            <Column header="Penjualan"><template #body="{ data: t }">{{ rp(t.kotor) }}</template></Column>
            <Column header="Diskon"><template #body="{ data: t }">{{ rp(t.diskon) }}</template></Column>
            <Column header="Komisi">
              <template #body="{ data: t }">{{ rp(t.komisi) }} <small class="none">({{ t.komisiPersen }}%)</small></template>
            </Column>
            <Column header="Setor ke tenant"><template #body="{ data: t }"><strong>{{ rp(t.setor) }}</strong></template></Column>
          </DataTable>
          <p class="total-setor">Total setoran: <strong>{{ rp(totalSetor) }}</strong></p>
          <small class="none">Komisi dihitung dari penjualan setelah diskon. Diskon bill dibagi menurut nilai item tiap tenant.</small>
        </section>
      </template>

      <!-- ================= WAKTU & KATEGORI ================= -->
      <template v-else-if="mode === 'waktu' && an.jamRamai">
        <div class="stats">
          <StatCard icon="pi pi-clock" label="Jam tersibuk" :value="heat.jamTersibuk == null ? '—' : `${String(heat.jamTersibuk).padStart(2, '0')}:00`" />
          <StatCard icon="pi pi-calendar" label="Hari omzet tertinggi" :value="heat.hariTerbaik ?? '—'" />
          <StatCard icon="pi pi-bolt" label="Puncak" :value="heat.puncak ? `${HARI[heat.puncak.hari]} ${String(heat.puncak.jam).padStart(2, '0')}:00 · ${heat.puncak.bill} bill` : '—'" />
        </div>

        <section class="panel box">
          <h3>Jam ramai</h3>
          <p v-if="!an.jamRamai.length" class="none">Belum ada transaksi.</p>
          <div v-else class="heat-wrap">
            <div class="heat" :style="{ gridTemplateColumns: `64px repeat(${heat.jam.length}, minmax(26px, 1fr))` }">
              <span />
              <span v-for="h in heat.jam" :key="'h' + h" class="jam">{{ String(h).padStart(2, '0') }}</span>
              <template v-for="(nama, d) in HARI" :key="nama">
                <span class="hari">{{ nama.slice(0, 3) }}</span>
                <span
                  v-for="h in heat.jam"
                  :key="d + '-' + h"
                  class="sel"
                  :style="gayaSel(d, h)"
                  :title="isiSel(d, h) ? `${nama} ${h}:00 · ${isiSel(d, h).bill} bill · ${rp(isiSel(d, h).omzet)}` : `${nama} ${h}:00 · sepi`"
                >{{ isiSel(d, h)?.bill || '' }}</span>
              </template>
            </div>
          </div>
          <small class="none">Angka = jumlah bill lunas pada jam itu, menurut jam pembayaran. Makin pekat makin ramai.</small>
        </section>

        <div class="two">
          <section class="panel box">
            <h3>Penjualan per kategori</h3>
            <p v-if="!an.kategori.length" class="none">Belum ada penjualan.</p>
            <Chart v-else type="bar" :data="kategoriChart.data" :options="kategoriChart.options" class="chart" :style="{ height: Math.max(200, an.kategori.length * 42) + 'px' }" />
          </section>
          <section class="panel box">
            <h3>Rincian kategori</h3>
            <DataTable paginator :rows="10" :rows-per-page-options="[10, 25, 50]" :always-show-paginator="false" :value="an.kategori" size="small" empty-message="Belum ada penjualan">
              <Column field="nama" header="Kategori" />
              <Column field="qty" header="Terjual" />
              <Column header="Penjualan"><template #body="{ data: k }">{{ rp(k.total) }}</template></Column>
              <Column header="Porsi">
                <template #body="{ data: k }">{{ Math.round((k.total / Math.max(1, an.kategori.reduce((s, x) => s + x.total, 0))) * 100) }}%</template>
              </Column>
            </DataTable>
            <small class="none">Nilai sebelum diskon bill.</small>
          </section>
        </div>

        <div v-if="an.tenant?.length > 1" class="two">
          <section class="panel box">
            <h3>Penjualan per tenant</h3>
            <Chart type="doughnut" :data="tenantChart.data" :options="tenantChart.options" class="chart pendek" />
          </section>
          <section class="panel box">
            <h3>Rincian tenant</h3>
            <DataTable :value="an.tenant" size="small">
              <Column field="nama" header="Tenant" />
              <Column field="bill" header="Bill" sortable />
              <Column field="qty" header="Item terjual" sortable />
              <Column field="total" header="Penjualan" sortable><template #body="{ data: t }">{{ rp(t.total) }}</template></Column>
              <Column header="Porsi"><template #body="{ data: t }">{{ Math.round((t.total / totalTenant) * 100) }}%</template></Column>
            </DataTable>
            <small class="none">Nilai sebelum diskon bill. Satu bill bisa berisi menu beberapa tenant, jadi jumlah bill per tenant tidak dijumlahkan. Komisi dan setoran ada di Laporan Tenant.</small>
          </section>
        </div>
      </template>

      <!-- ================= KARYAWAN ================= -->
      <template v-else-if="mode === 'karyawan' && an.karyawan">
        <section class="panel box">
          <h3>Kinerja karyawan</h3>
          <CariInput v-model="cariKaryawan" placeholder="Cari karyawan…" />
          <DataTable paginator :rows="10" :rows-per-page-options="[10, 25, 50]" :always-show-paginator="false" :value="karyawanTampil" size="small" scrollable empty-message="Tidak ada aktivitas pada rentang ini" sort-field="nilaiDibayar" :sort-order="-1">
            <Column header="Nama" frozen>
              <template #body="{ data: k }">{{ k.nama }} <Tag :value="ROLE[k.role] ?? k.role" severity="secondary" /></template>
            </Column>
            <Column field="billDibuat" header="Bill diinput" sortable />
            <Column field="nilaiDibuat" header="Nilai diinput" sortable><template #body="{ data: k }">{{ rp(k.nilaiDibuat) }}</template></Column>
            <Column field="billDibayar" header="Bill dibayar" sortable />
            <Column field="nilaiDibayar" header="Uang diterima" sortable><template #body="{ data: k }">{{ rp(k.nilaiDibayar) }}</template></Column>
            <Column field="diskon" header="Diskon" sortable><template #body="{ data: k }">{{ k.diskon ? rp(k.diskon) : '—' }}</template></Column>
            <Column field="batalNilai" header="Pembatalan" sortable>
              <template #body="{ data: k }">
                <template v-if="k.batalQty">{{ k.batalQty }} item · {{ rp(k.batalNilai) }}</template><template v-else>—</template>
              </template>
            </Column>
            <Column field="shift" header="Shift" sortable />
            <Column field="selisihKas" header="Selisih kas" sortable>
              <template #body="{ data: k }">
                <span v-if="!k.shift">—</span>
                <Tag v-else :value="rp(k.selisihKas)" :severity="k.selisihKas === 0 ? 'success' : k.selisihKas > 0 ? 'warn' : 'danger'" />
              </template>
            </Column>
          </DataTable>
          <small class="none">
            Bill diinput = bill lunas yang dia buat. Uang diterima = pembayaran yang dia terima sebagai kasir. Diskon dihitung pada bill yang pembayarannya dia terima.
            Pembatalan termasuk refund. Selisih kas negatif = uang kurang saat tutup shift.
          </small>
        </section>
      </template>

      <!-- ================= KASBON & STOK ================= -->
      <template v-else-if="mode === 'kasbon' && an.kasbon">
        <div class="stats">
          <StatCard icon="pi pi-book" label="Total piutang kasbon" :value="rp(totalKasbon)" />
          <StatCard icon="pi pi-exclamation-triangle" label="Piutang lebih dari 30 hari" :value="rp(kasbonLama)" />
          <StatCard icon="pi pi-box" label="Nilai persediaan bahan" :value="rp(nilaiStok)" />
          <StatCard icon="pi pi-trash" label="Bahan susut (rentang ini)" :value="rp(totalSusut)" />
        </div>
        <div class="two">
          <section class="panel box">
            <h3>Umur piutang kasbon</h3>
            <CariInput v-model="cariKasbon" placeholder="Cari pelanggan…" />
            <DataTable :value="kasbonTampil" size="small" paginator :rows="15" empty-message="Tidak ada kasbon yang belum lunas">
              <Column field="nama" header="Pelanggan">
                <template #body="{ data: k }">{{ k.nama }}<small v-if="k.telp" class="none"><br />{{ k.telp }}</small></template>
              </Column>
              <Column field="sisa" header="Sisa utang" sortable><template #body="{ data: k }"><strong>{{ rp(k.sisa) }}</strong></template></Column>
              <Column field="umurHari" header="Umur" sortable>
                <template #body="{ data: k }"><Tag :value="`${k.umurHari} hari`" :severity="umurTag(k.umurHari)" /></template>
              </Column>
            </DataTable>
            <small class="none">Posisi hari ini. Umur dihitung dari kasbon tertua yang belum tertutup cicilan.</small>
          </section>
          <section class="panel box">
            <h3>Persediaan dan susut bahan</h3>
            <CariInput v-model="cariStok" placeholder="Cari bahan…" />
            <DataTable :value="stokTampil" size="small" paginator :rows="15" empty-message="Belum ada bahan">
              <Column field="nama" header="Bahan" sortable />
              <Column header="Stok"><template #body="{ data: b }">{{ b.stok.toLocaleString('id-ID') }} {{ b.satuan }}</template></Column>
              <Column field="nilai" header="Nilai" sortable><template #body="{ data: b }">{{ rp(b.nilai) }}</template></Column>
              <Column field="susutNilai" header="Susut" sortable>
                <template #body="{ data: b }">
                  <span v-if="!b.susutQty" class="none">—</span>
                  <span v-else :class="b.susutNilai > 0 ? 'merah' : ''">{{ b.susutQty.toLocaleString('id-ID') }} {{ b.satuan }} · {{ rp(b.susutNilai) }}</span>
                </template>
              </Column>
            </DataTable>
            <small class="none">Nilai = stok × harga rata-rata. Susut dari selisih opname pada rentang tanggal; negatif berarti stok fisik lebih.</small>
          </section>
        </div>
      </template>
  </div>
</template>

<style scoped>
.total-setor {
  text-align: right;
  margin: 8px 0;
}
.gap {
  gap: 14px;
}
.stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 14px;
}
.two {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(340px, 1fr));
  gap: 14px;
}
.box {
  padding: 14px;
  min-width: 0;
}
h3 {
  margin: 0 0 12px;
}
.none {
  color: var(--p-text-muted-color);
}
.merah {
  color: var(--p-red-500, #ef4444);
}
.chart {
  height: 320px;
}
.chart.pendek {
  height: 240px;
}
.heat-wrap {
  overflow-x: auto;
}
.heat {
  display: grid;
  gap: 3px;
  min-width: max-content;
}
.jam,
.hari {
  font-size: 0.72rem;
  color: var(--p-text-muted-color);
  text-align: center;
  align-self: center;
}
.hari {
  text-align: left;
}
.sel {
  height: 30px;
  border-radius: 6px;
  display: grid;
  place-items: center;
  font-size: 0.72rem;
  font-weight: 600;
  background: color-mix(in srgb, var(--p-text-muted-color) 7%, transparent);
}
@media (max-width: 600px) {
  .two {
    grid-template-columns: 1fr;
  }
  .chart {
    height: 260px;
  }
}
.judul-link {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 10px;
}
.judul-link a {
  color: var(--accent-text);
  font-weight: 600;
}
</style>
