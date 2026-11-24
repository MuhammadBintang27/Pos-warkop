<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { onEvent } from '@/composables/useEvents'
import { useMobile } from '@/composables/useMobile'
import { useRun } from '@/composables/useRun'
import { useAuthStore } from '@/stores/auth'
import { useKasirStore } from '@/stores/kasir'
import { usePengaturanStore } from '@/stores/pengaturan'
import { cetakStruk } from '@/utils/cetak'
import { rp } from '@/utils/format'
import { load, save } from '@/utils/storage'

const auth = useAuthStore()
const store = useKasirStore()
const run = useRun()
const peng = usePengaturanStore()
const mobile = useMobile()

const dlg = reactive({
  denah: false, opsi: false, bayar: false, pisah: false, pindah: false,
  gabung: false, shift: false, kas: false, kasbon: false, struk: false, cart: false,
})
const menuDipilih = ref(null)
const lunas = ref(null)
const cartTutup = ref(load('cartTutup', false))
const cartLebar = ref(false)
watch(cartTutup, (v) => {
  save('cartTutup', v)
  if (v) cartLebar.value = false
})

const peringatanCetak = computed(() => {
  const { macet, stasiunAktif } = store.cetak
  if (!macet) return ''
  if (stasiunAktif) return `${macet} tiket belum tercetak, cek printer`
  return `Stasiun cetak tidak terhubung, ${macet} tiket menunggu` + (hp && !hpPunyaToken() ? '. Minta owner mengaktifkan stasiun dari halaman Printer' : '')
})

// stasiun cetak di HP ini (aplikasi Android); kasir boleh menyalakan ulang dengan token yang sudah disimpan owner
const hp = window.StasiunAndroid
const hpPunyaToken = () => JSON.parse(hp.status()).adaToken
const bisaNyalakanHp = computed(() => hp && !store.cetak.stasiunAktif && hpPunyaToken())
function nyalakanHp() {
  hp.simpan(location.origin, '', true)
  setTimeout(() => store.loadCetak().catch(() => {}), 4000)
}
const aktif = computed(() => store.pesanan ?? store.draft)
const subtitle = computed(() => {
  const p = aktif.value
  if (!p) return 'Pilih meja untuk mulai memesan'
  return [p.mejaNama || 'Bawa pulang', p.atasNama, store.pesanan && `#${p.id}`].filter(Boolean).join(' · ')
})
const jumlahItem = computed(() => store.pesanan?.items.reduce((s, i) => s + i.qty, 0) ?? 0)

function pilihMenu(m) {
  if (!aktif.value) return (dlg.denah = true)
  if (m.varian.length) {
    menuDipilih.value = m
    dlg.opsi = true
  } else {
    run(() => store.addItem({ menuId: m.id, qty: 1, opsiIds: [], catatan: '' }))
  }
}

const tambah = (body) => run(() => store.addItem(body))

async function pilihDenah(buka) {
  if (await run(buka)) dlg.denah = false
}

function selesaiBayar(p) {
  lunas.value = p
  if (peng.settings.printerKasirId) run(() => cetakStruk(p, peng.settings))
  dlg.struk = true
  dlg.cart = false
  cartLebar.value = false
}

const offs = []
let timerCetak
onMounted(async () => {
  offs.push(
    onEvent('pesanan', () => {
      store.refreshPesanan()
      store.loadDenah()
    }),
    // tenant menekan Selesai: perbarui tanda "sudah diantar" di keranjang
    onEvent('dapur', store.refreshPesanan),
    onEvent('meja', store.loadDenah),
    onEvent('menu', () => Promise.all([store.loadMenu(), store.loadTenant()])),
    onEvent('shift', () => auth.isKasir && store.loadShift()),
    onEvent('cetak', () => auth.isKasir && store.loadCetak()),
  )
  await run(() => Promise.all([store.loadMenu(), store.loadTenant(), store.loadDenah(), auth.isKasir && store.loadShift()]))
  if (!aktif.value) dlg.denah = true
  if (auth.isKasir) {
    store.loadCetak().catch(() => {})
    timerCetak = setInterval(() => store.loadCetak().catch(() => {}), 10000) // status stasiun berubah tanpa event
  }
})
onBeforeUnmount(() => {
  offs.forEach((off) => off())
  clearInterval(timerCetak)
})
</script>

<template>
  <div class="kasir">
    <PageHeader class="kasir-head" icon="pi pi-shopping-cart" title="Kasir" :subtitle="subtitle">
      <button class="pill" :title="aktif ? 'Ganti meja' : 'Pilih meja'" @click="dlg.denah = true"><i class="pi pi-th-large" /> <span class="lbl">{{ aktif ? 'Ganti meja' : 'Pilih meja' }}</span></button>
      <button class="pill" title="Bawa pulang" @click="store.baru('')"><i class="pi pi-shopping-bag" /> <span class="lbl">Bawa pulang</span></button>
      <template v-if="auth.isKasir">
        <button class="pill" title="Kas Keluar" @click="dlg.kas = true"><i class="pi pi-wallet" /> <span class="lbl">Kas Keluar</span></button>
        <button class="pill" title="Kasbon" @click="dlg.kasbon = true"><i class="pi pi-clock" /> <span class="lbl">Kasbon</span></button>
        <span v-if="peringatanCetak" class="chip warn" :title="peringatanCetak"><i class="pi pi-print" /> <span class="lbl">{{ peringatanCetak }}</span></span>
        <span v-else-if="store.cetak.stasiunAktif" class="chip ok" title="Stasiun cetak terhubung"><i class="pi pi-print" /> <span class="lbl">Printer siap</span></span>
        <button v-if="bisaNyalakanHp" class="pill warn" title="Nyalakan stasiun cetak di perangkat ini" @click="nyalakanHp"><i class="pi pi-power-off" /> <span class="lbl">Nyalakan stasiun</span></button>
        <span v-if="store.shift" class="chip ok" :title="`Shift aktif — ${store.shift.kasir}`"><i class="pi pi-check-circle" /> <span class="lbl">Shift aktif — {{ store.shift.kasir }}</span></span>
        <button class="pill warn" @click="dlg.shift = true">
          <i :class="store.shift ? 'pi pi-lock' : 'pi pi-lock-open'" /> {{ store.shift ? 'Tutup Shift' : 'Buka Shift' }}
        </button>
      </template>
    </PageHeader>

    <div class="body" :class="{ ringkas: cartTutup && !mobile }">
      <MenuGrid @pick="pilihMenu" />
      <template v-if="!mobile">
        <div v-if="cartLebar" class="scrim" @click="cartLebar = false" />
        <CartPanel
          v-show="!cartTutup"
          bisa-tutup
          :class="{ 'cart-lebar': cartLebar }"
          :lebar="cartLebar"
          @lebar="cartLebar = !cartLebar"
          @tutup="cartTutup = true"
          @denah="dlg.denah = true"
          @bayar="dlg.bayar = true"
          @pisah="dlg.pisah = true"
          @pindah="dlg.pindah = true"
          @gabung="dlg.gabung = true"
        />
        <button v-if="cartTutup" class="rail panel" aria-label="Buka keranjang" @click="cartTutup = false">
          <i class="pi pi-angle-double-left" />
          <span class="ikon">
            <i class="pi pi-shopping-cart" />
            <b v-if="jumlahItem" class="badge">{{ jumlahItem }}</b>
          </span>
          <strong v-if="store.pesanan" class="total">{{ rp(store.pesanan.total) }}</strong>
        </button>
      </template>
    </div>

    <button v-if="mobile" class="cart-bar" @click="dlg.cart = true">
      <span><i class="pi pi-shopping-cart" /> {{ jumlahItem }} item</span>
      <strong>{{ rp(store.pesanan?.total) }}</strong>
    </button>
    <Drawer v-model:visible="dlg.cart" position="bottom" header="Pesanan" class="cart-drawer">
      <CartPanel
        @denah="(dlg.cart = false), (dlg.denah = true)"
        @bayar="dlg.bayar = true"
        @pisah="dlg.pisah = true"
        @pindah="dlg.pindah = true"
        @gabung="dlg.gabung = true"
      />
    </Drawer>

    <DenahDialog
      v-model:visible="dlg.denah"
      @meja="(m) => pilihDenah(() => store.bukaMeja(m))"
      @pesanan="(id) => pilihDenah(() => store.bukaPesanan(id))"
      @baru="(nama) => pilihDenah(() => store.baru(nama))"
    />
    <OpsiDialog v-model:visible="dlg.opsi" :menu="menuDipilih" @add="tambah" />
    <BayarDialog v-model:visible="dlg.bayar" @paid="selesaiBayar" />
    <PisahDialog v-model:visible="dlg.pisah" />
    <PindahGabungDialog v-model:visible="dlg.pindah" mode="pindah" />
    <PindahGabungDialog v-model:visible="dlg.gabung" mode="gabung" />
    <ShiftDialog v-model:visible="dlg.shift" />
    <PengeluaranDialog v-model:visible="dlg.kas" />
    <KasbonDialog v-model:visible="dlg.kasbon" />
    <StrukDialog v-model:visible="dlg.struk" :pesanan="lunas" />
  </div>
</template>

<style scoped>
.kasir {
  height: 100%;
  min-height: 560px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-height: 0;
}
/* header tetap satu baris: di ruang sempit (iPad + sidebar) tombol jadi ikon saja, tombol shift tetap berlabel */
section.kasir-head {
  container-type: inline-size;
  flex-wrap: nowrap;
}
section.kasir-head :deep(.actions) {
  flex-wrap: nowrap;
  flex: none;
}
section.kasir-head :deep(.title) {
  min-width: 0;
}
section.kasir-head :deep(.title small) {
  display: block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
@media (max-width: 600px) {
  section.kasir-head,
  section.kasir-head :deep(.actions) {
    flex-wrap: wrap;
  }
}
@container (max-width: 1150px) {
  .kasir-head .lbl {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip-path: inset(50%);
  }
  .kasir-head .pill,
  .kasir-head .chip {
    gap: 0;
  }
}
.body {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 380px;
  grid-template-rows: minmax(0, 1fr);
  position: relative;
  gap: 14px;
}
.scrim {
  position: absolute;
  inset: 0;
  z-index: 14;
  border-radius: 14px;
  background: rgb(0 0 0 / 0.28);
  animation: muncul 0.18s ease;
}
.cart-lebar {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  z-index: 15;
  width: min(760px, 74%);
  box-shadow: -14px 0 36px rgb(0 0 0 / 0.28);
  animation: masuk 0.2s ease;
}
@keyframes masuk {
  from {
    transform: translateX(32px);
    opacity: 0;
  }
}
@keyframes muncul {
  from {
    opacity: 0;
  }
}
.body.ringkas {
  grid-template-columns: minmax(0, 1fr) 64px;
}
.rail {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 14px 6px;
  color: var(--p-text-color);
  font: inherit;
  cursor: pointer;
}
.rail:hover {
  border-color: var(--p-primary-color);
}
.rail .ikon {
  position: relative;
  font-size: 1.5rem;
  color: var(--accent-text);
}
.rail .badge {
  position: absolute;
  top: -10px;
  right: -12px;
  min-width: 20px;
  padding: 1px 5px;
  border-radius: 999px;
  font-size: 0.7rem;
  text-align: center;
  background: var(--p-primary-color);
  color: var(--p-primary-contrast-color);
}
.rail .total {
  writing-mode: vertical-rl;
  color: var(--accent-text);
}
.cart-bar {
  position: fixed;
  left: 12px;
  right: 12px;
  bottom: calc(12px + env(safe-area-inset-bottom));
  z-index: 20;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 18px;
  border: 0;
  border-radius: 14px;
  font: inherit;
  font-size: 1.05rem;
  background: var(--p-primary-color);
  color: var(--p-primary-contrast-color);
  box-shadow: 0 6px 20px rgb(0 0 0 / 0.25);
}
@media (max-width: 900px) {
  .body {
    grid-template-columns: 1fr;
    padding-bottom: 72px;
  }
}
</style>

<style>
.cart-drawer {
  height: 88dvh !important;
}
</style>
