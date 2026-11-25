<script setup>
import { onMounted, ref } from 'vue'

// Hanya di aplikasi Android versi yang sudah punya fitur pembaruan (window.StasiunAndroid.versi).
// Tampilan web sudah selalu terbaru dari server; ini untuk bagian native (stasiun cetak, izin, notifikasi).
const hp = window.StasiunAndroid
const baru = ref(null)
const tampil = ref(false)
const pesan = ref('')

// Pembaruan wajib: dialog tidak bisa ditutup dan unduhan langsung dimulai; tombol Perbarui untuk mencoba lagi.
onMounted(async () => {
  if (!hp?.versi) return
  try {
    const v = await (await fetch('/dl/versi.json', { cache: 'no-store' })).json()
    if (v.versionCode > Number(hp.versi())) {
      baru.value = v
      tampil.value = true
      perbarui()
    }
  } catch {
    // tanpa versi.json (offline/server lama) cukup tidak menawarkan pembaruan
  }
})

function perbarui() {
  const hasil = hp.perbarui(`${location.origin}/dl/pos-warkop.apk`)
  if (hasil === 'izin') {
    pesan.value = 'Izinkan "Instal aplikasi tidak dikenal" untuk POS Warkop, lalu kembali dan tekan Perbarui.'
    return
  }
  pesan.value = 'Mengunduh pembaruan… Layar pemasangan muncul otomatis setelah selesai, lalu tekan Install.'
}
</script>

<template>
  <Dialog v-model:visible="tampil" modal header="Memperbarui aplikasi" :closable="false" :close-on-escape="false" :style="{ width: '26rem' }" :breakpoints="{ '600px': '94vw' }">
    <p class="judul">POS Warkop {{ baru?.versionName }}</p>
    <p v-if="baru?.catatan" class="redup">{{ baru.catatan }}</p>
    <p v-if="pesan" class="pesan">{{ pesan }}</p>
    <template #footer>
      <Button label="Perbarui" icon="pi pi-download" @click="perbarui" />
    </template>
  </Dialog>
</template>

<style scoped>
.judul {
  margin: 0 0 6px;
  font-weight: 700;
}
.redup {
  margin: 0;
  color: var(--p-text-muted-color);
}
.pesan {
  margin: 12px 0 0;
  font-weight: 600;
  color: var(--accent-text);
}
</style>
