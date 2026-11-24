<script setup>
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import QRCode from 'qrcode'
import { get, post } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { applyTheme, isHex, PRIMARY_COLORS, THEMES } from '@/composables/useAppTheme'
import { usePengaturanStore } from '@/stores/pengaturan'

const KEYS = ['namaUsaha', 'alamat', 'logo', 'footerStruk', 'printerKasirId', 'dialogDapur', 'theme', 'primary', 'mode']
const MODES = [
  { label: 'Terang', value: 'light' },
  { label: 'Gelap', value: 'dark' },
  { label: 'Ikuti perangkat', value: 'auto' },
]

const run = useRun()
const store = usePengaturanStore()
const form = reactive({ ...store.settings })
const printers = ref([])
run(async () => (printers.value = await get('/printer')))
const saving = ref(false)

// logo diunggah sebagai file ke server (folder uploads), yang disimpan di pengaturan hanya URL-nya
const MAKS_LOGO = 1024 * 1024 // sama dengan services.MaksLogo
const mengunggah = ref(false)
function pilihLogo(e) {
  const file = e.target.files[0]
  e.target.value = ''
  if (!file) return
  if (!['image/png', 'image/jpeg'].includes(file.type)) return run(() => Promise.reject(new Error('Logo harus PNG atau JPG')))
  if (file.size > MAKS_LOGO) return run(() => Promise.reject(new Error('Logo maksimal 1 MB')))
  const fd = new FormData()
  fd.append('logo', file)
  mengunggah.value = true
  run(async () => (form.logo = (await post('/pengaturan/logo', fd)).url)).finally(() => (mengunggah.value = false))
}

const dirty = computed(() => KEYS.some((k) => form[k] !== store.settings[k]))
const customHex = computed({
  get: () => (isHex(form.primary) ? form.primary : '#8b4513'),
  set: (v) => (form.primary = v),
})

watch(
  () => [form.theme, form.primary, form.mode],
  () => applyTheme({ ...form }, { persist: false }),
)

onBeforeUnmount(() => {
  if (dirty.value) applyTheme(store.settings, { persist: false })
})

async function save() {
  saving.value = true
  await run(() => store.save({ ...form }), 'Tersimpan, semua perangkat ikut berganti')
  saving.value = false
}

const reset = () => Object.assign(form, store.settings)

// link menu untuk pelanggan dan QR-nya
const linkMenu = computed(() => (store.settings.kode ? `${location.origin}/m/${store.settings.kode}` : ''))
const qr = ref('')
watch(linkMenu, async (url) => (qr.value = url ? await QRCode.toDataURL(url, { width: 600, margin: 2 }) : ''), { immediate: true })
const salinLink = () => run(() => navigator.clipboard.writeText(linkMenu.value), 'Link disalin')
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-palette" title="Tampilan" subtitle="Tema, warna, dan identitas usaha">
      <button class="pill" :disabled="!dirty" @click="reset">Batal</button>
      <button class="pill" :disabled="!dirty || !form.namaUsaha.trim() || saving" @click="save"><i class="pi pi-check" /> Simpan</button>
    </PageHeader>

    <div class="grid">
      <section class="panel box">
        <h3>Tema</h3>
        <div class="stack">
          <label>Bentuk komponen</label>
          <SelectButton v-model="form.theme" :options="THEMES" option-label="label" option-value="id" :allow-empty="false" />
          <label>Mode</label>
          <SelectButton v-model="form.mode" :options="MODES" option-label="label" option-value="value" :allow-empty="false" />
          <label>Warna utama</label>
          <div class="swatches">
            <button
              v-for="c in PRIMARY_COLORS"
              :key="c.name"
              type="button"
              class="swatch"
              :class="{ on: form.primary === c.name }"
              :style="{ background: c.hex }"
              :aria-label="c.name"
              :title="c.name"
              @click="form.primary = c.name"
            />
            <label class="swatch custom" :class="{ on: isHex(form.primary) }" title="Warna brand sendiri">
              <i class="pi pi-palette" />
              <input v-model="customHex" type="color" aria-label="Warna brand sendiri" />
            </label>
          </div>
          <small v-if="isHex(form.primary)">Warna brand: {{ form.primary }}</small>
        </div>
      </section>

      <section class="panel box">
        <h3>Identitas usaha</h3>
        <div class="stack">
          <label for="nama">Nama usaha</label>
          <InputText id="nama" v-model="form.namaUsaha" maxlength="100" fluid />
          <label>Logo</label>
          <div class="logobox">
            <div class="pratinjau">
              <i v-if="mengunggah" class="pi pi-spin pi-spinner" />
              <template v-else-if="form.logo">
                <figure>
                  <img :src="form.logo" alt="Logo" />
                  <figcaption>Layar</figcaption>
                </figure>
                <figure>
                  <img :src="form.logo" alt="" class="struk" />
                  <figcaption>Struk</figcaption>
                </figure>
              </template>
              <span v-else class="redup">Belum ada logo</span>
            </div>
            <div class="logorow">
              <label class="pill upload">
                <i :class="mengunggah ? 'pi pi-spin pi-spinner' : 'pi pi-upload'" />
                {{ mengunggah ? 'Mengunggah…' : form.logo ? 'Ganti' : 'Unggah' }}
                <input type="file" accept="image/png,image/jpeg" hidden :disabled="mengunggah" @change="pilihLogo" />
              </label>
              <Button v-if="form.logo" label="Hapus" text severity="secondary" @click="form.logo = ''" />
            </div>
            <small class="redup">PNG atau JPG, maksimal 1 MB. Logo baru dipakai setelah klik Simpan.</small>
          </div>
          <label for="alamat">Alamat (di struk)</label>
          <InputText id="alamat" v-model="form.alamat" maxlength="255" fluid />
          <label for="pk">Printer kasir (struk dan bill closing)</label>
          <Select v-model="form.printerKasirId" input-id="pk" :options="printers" option-label="nama" option-value="id" placeholder="Tanpa printer: cetak lewat browser" show-clear fluid />
          <div class="sw">
            <ToggleSwitch v-model="form.dialogDapur" input-id="dd" />
            <label for="dd">Munculkan dialog cetak tiket dapur</label>
          </div>
          <small class="hint">Untuk tenant tanpa printer: tiket pesanan dan BATAL dicetak lewat dialog browser di perangkat pengirim. Matikan bila tenant tersebut tidak butuh kertas tiket.</small>
          <label for="footer">Catatan kaki struk</label>
          <InputText id="footer" v-model="form.footerStruk" maxlength="255" placeholder="Terima kasih sudah mampir" fluid />
        </div>
      </section>

      <section v-if="linkMenu" class="panel box">
        <h3>Menu untuk pelanggan</h3>
        <div class="stack">
          <small class="hint">Pelanggan bisa melihat menu dan harga tanpa login. Cetak QR-nya lalu tempel di meja.</small>
          <InputGroup>
            <InputText :model-value="linkMenu" readonly />
            <Button icon="pi pi-copy" severity="secondary" aria-label="Salin link" @click="salinLink" />
            <Button icon="pi pi-external-link" severity="secondary" aria-label="Buka menu" as="a" :href="linkMenu" target="_blank" />
          </InputGroup>
          <img v-if="qr" :src="qr" alt="QR menu" class="qr" />
          <Button v-if="qr" label="Unduh QR" icon="pi pi-download" severity="secondary" outlined as="a" :href="qr" :download="`qr-menu-${store.settings.kode}.png`" />
        </div>
      </section>

      <section class="panel box wide">
        <h3>Pratinjau</h3>
        <div class="row">
          <Button label="Utama" />
          <Button label="Sekunder" severity="secondary" />
          <Button label="Garis" outlined />
          <Button label="Teks" text />
          <Tag value="Terisi" />
          <Tag value="Lunas" severity="success" />
        </div>
        <InputText placeholder="Contoh input" fluid class="mt" />
        <ProgressBar :value="60" class="mt" />
      </section>
    </div>
  </div>
</template>

<style scoped>
.sw {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 6px;
}
.hint {
  color: var(--p-text-muted-color);
}
.qr {
  width: 220px;
  max-width: 100%;
  align-self: center;
  border-radius: 8px;
}
.gap {
  gap: 14px;
}
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 14px;
}
.box {
  padding: 16px;
}
.wide {
  grid-column: 1 / -1;
}
h3 {
  margin: 0 0 12px;
}
.swatches {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.swatch {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
  padding: 0;
  position: relative;
}
.swatch.on {
  outline: 3px solid var(--p-text-color);
  outline-offset: 2px;
}
.swatch.custom {
  display: grid;
  place-items: center;
  background: var(--p-content-background);
  border: 2px dashed var(--p-content-border-color);
  font-weight: normal;
  margin: 0;
}
.swatch.custom input {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
  width: 100%;
  height: 100%;
}
.logorow {
  display: flex;
  align-items: center;
  gap: 10px;
}
.logobox {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.pratinjau {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 24px;
  min-height: 120px;
  padding: 12px;
  border: 1px dashed var(--p-content-border-color);
  border-radius: 8px;
}
.pratinjau figure {
  margin: 0;
  text-align: center;
}
.pratinjau img {
  display: block;
  width: 88px;
  height: 88px;
  object-fit: contain;
}
/* perkiraan hasil printer thermal: hitam putih di atas kertas putih, sama dengan print.js */
.pratinjau img.struk {
  background: #fff;
  filter: grayscale(1) contrast(100);
}
.redup {
  color: var(--p-text-muted-color);
}
.pratinjau figcaption {
  font-size: 12px;
  color: var(--p-text-muted-color);
}
.upload {
  cursor: pointer;
  margin: 0;
}
.row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}
.mt {
  margin-top: 12px;
}
</style>
