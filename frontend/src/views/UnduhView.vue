<script setup>
const APP = [
  {
    icon: 'pi pi-android',
    judul: 'Android',
    file: '/dl/pos-warkop.apk',
    ukuran: '8 MB',
    langkah: [
      'Unduh file APK di HP.',
      'Buka filenya, izinkan "Instal aplikasi tidak dikenal" bila diminta.',
      'Saat pertama dibuka, isi alamat server dari owner.',
    ],
  },
  {
    icon: 'pi pi-microsoft',
    judul: 'Windows',
    file: '/dl/pos-warkop-setup.exe',
    ukuran: '2 MB',
    langkah: [
      'Unduh installer di komputer kasir.',
      'Jalankan, bila muncul SmartScreen pilih "More info" lalu "Run anyway".',
      'Ikuti langkah instalasi sampai selesai.',
    ],
  },
]
</script>

<template>
  <div class="unduh">
    <header>
      <h1>Unduh POS Warkop</h1>
      <p>Pasang aplikasi kasir di HP atau komputer. Akun tetap dari owner.</p>
    </header>

    <div class="grid">
      <section v-for="a in APP" :key="a.judul" class="kartu">
        <div class="kepala">
          <i :class="a.icon" />
          <div>
            <h2>{{ a.judul }}</h2>
            <small>{{ a.ukuran }}</small>
          </div>
        </div>
        <ol>
          <li v-for="l in a.langkah" :key="l">{{ l }}</li>
        </ol>
        <a :href="a.file" download class="p-button tombol"><i class="pi pi-download" /> Unduh untuk {{ a.judul }}</a>
      </section>
    </div>

    <RouterLink to="/login" class="masuk">Atau buka lewat browser <i class="pi pi-arrow-right" /></RouterLink>
  </div>
</template>

<style scoped>
.unduh {
  min-height: 100dvh;
  max-width: 880px;
  margin: 0 auto;
  padding: 48px 16px;
  display: flex;
  flex-direction: column;
  gap: 28px;
}
header h1 {
  margin: 0 0 6px;
  font-size: 30px;
  letter-spacing: -0.02em;
}
header p {
  margin: 0;
  color: var(--p-text-muted-color);
}
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 16px;
}
.kartu {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 20px;
  border: 1px solid var(--p-content-border-color);
  border-radius: 14px;
  background: var(--p-content-background);
}
.kepala {
  display: flex;
  align-items: center;
  gap: 12px;
}
.kepala > i {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  font-size: 1.4rem;
  background: var(--p-primary-color);
  color: var(--p-primary-contrast-color);
}
.kepala h2 {
  margin: 0;
  font-size: 18px;
}
.kepala small,
ol {
  color: var(--p-text-muted-color);
}
ol {
  margin: 0;
  padding-left: 20px;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.tombol {
  justify-content: center;
  gap: 8px;
  text-decoration: none;
}
.masuk {
  align-self: center;
  color: var(--p-primary-color);
  text-decoration: none;
}
</style>
