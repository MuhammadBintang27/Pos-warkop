<script setup>
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { get } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { applyTheme, isHex, PRIMARY_COLORS, THEMES } from '@/composables/useAppTheme'
import { usePengaturanStore } from '@/stores/pengaturan'

const KEYS = ['namaUsaha', 'alamat', 'logo', 'footerStruk', 'printerKasirId', 'theme', 'primary', 'mode']
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

const MAKS_LOGO = 200 * 1024
function pilihLogo(e) {
  const file = e.target.files[0]
  e.target.value = ''
  if (!file) return
  if (!file.type.startsWith('image/') || file.size > MAKS_LOGO) return run(() => Promise.reject(new Error('Logo harus gambar dan maksimal 200 KB')))
  const r = new FileReader()
  r.onload = () => (form.logo = r.result)
  r.readAsDataURL(file)
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
          <div class="logorow">
            <img v-if="form.logo" :src="form.logo" alt="Logo" class="logoimg" />
            <label class="pill upload">
              <i class="pi pi-upload" /> {{ form.logo ? 'Ganti' : 'Unggah' }}
              <input type="file" accept="image/*" hidden @change="pilihLogo" />
            </label>
            <Button v-if="form.logo" label="Hapus" text severity="secondary" @click="form.logo = ''" />
          </div>
          <label for="alamat">Alamat (di struk)</label>
          <InputText id="alamat" v-model="form.alamat" maxlength="255" fluid />
          <label for="pk">Printer kasir (struk dan bill closing)</label>
          <Select v-model="form.printerKasirId" input-id="pk" :options="printers" option-label="nama" option-value="id" placeholder="Tanpa printer: cetak lewat browser" show-clear fluid />
          <label for="footer">Catatan kaki struk</label>
          <InputText id="footer" v-model="form.footerStruk" maxlength="255" placeholder="Terima kasih sudah mampir" fluid />
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
.logoimg {
  width: 44px;
  height: 44px;
  object-fit: contain;
  border-radius: 8px;
  border: 1px solid var(--p-content-border-color);
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
