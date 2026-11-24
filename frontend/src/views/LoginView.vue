<script setup>
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { errMsg } from '@/api/http'
import { useAuthStore } from '@/stores/auth'
import { usePengaturanStore } from '@/stores/pengaturan'
import { load, save } from '@/utils/storage'

const router = useRouter()
const auth = useAuthStore()
const pengaturan = usePengaturanStore()

const FITUR = [
  { icon: 'pi pi-receipt', judul: 'Pesanan per meja', isi: 'Buka bill, pindah, gabung, dan pisah dengan beberapa perangkat sekaligus.' },
  { icon: 'pi pi-print', judul: 'Tiket ke dapur', isi: 'Setiap tenant menerima tiketnya sendiri di printernya sendiri.' },
  { icon: 'pi pi-chart-line', judul: 'Laporan dan closing', isi: 'Penjualan harian, shift, dan setoran tenant tercatat rapi.' },
]

const username = ref(load('lastUser', ''))
const ingat = ref(!!username.value)
const password = ref('')
const error = ref('')
const loading = ref(false)
const capsLock = ref(false)

const initial = computed(() => (pengaturan.settings.namaUsaha || 'W').trim()[0].toUpperCase())
const cekCaps = (e) => (capsLock.value = !!e.getModifierState?.('CapsLock'))

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(username.value.trim(), password.value)
    save('lastUser', ingat.value ? username.value.trim() : '')
    await pengaturan.fetch()
    router.push('/')
  } catch (e) {
    error.value = errMsg(e)
    password.value = ''
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login">
    <aside class="brand">
      <div class="logo">
        <img v-if="pengaturan.settings.logo" :src="pengaturan.settings.logo" alt="" class="mark img" />
        <div v-else class="mark">{{ initial }}</div>
        <strong>{{ pengaturan.settings.namaUsaha }}</strong>
      </div>

      <div class="pitch">
        <h1>Kasir warkop yang tidak bikin repot.</h1>
        <p>Dari pesanan di meja sampai laporan akhir hari, semuanya ada di satu tempat.</p>
        <ul>
          <li v-for="f in FITUR" :key="f.judul">
            <i :class="f.icon" />
            <div>
              <b>{{ f.judul }}</b>
              <span>{{ f.isi }}</span>
            </div>
          </li>
        </ul>
      </div>

      <small class="foot">&copy; {{ new Date().getFullYear() }} POS Warkop</small>
    </aside>

    <main class="side">
      <form class="form" @submit.prevent="submit">
        <div class="logo compact">
          <img v-if="pengaturan.settings.logo" :src="pengaturan.settings.logo" alt="" class="mark img" />
          <div v-else class="mark">{{ initial }}</div>
          <strong>{{ pengaturan.settings.namaUsaha }}</strong>
        </div>

        <h2>Selamat datang</h2>
        <p class="sub">Masuk dengan akun yang diberikan owner.</p>

        <label for="username">Username</label>
        <IconField>
          <InputIcon class="pi pi-user" />
          <InputText id="username" v-model="username" autocomplete="username" placeholder="contoh: 1-budi" autofocus fluid />
        </IconField>

        <label for="password">Password</label>
        <Password
          v-model="password"
          input-id="password"
          :input-props="{ autocomplete: 'current-password', onKeyup: cekCaps, onKeydown: cekCaps }"
          :feedback="false"
          toggle-mask
          fluid
        />
        <small v-if="capsLock" class="caps"><i class="pi pi-exclamation-triangle" /> Caps Lock sedang aktif</small>

        <div class="ingat">
          <Checkbox v-model="ingat" input-id="ingat" binary />
          <label for="ingat">Ingat username</label>
        </div>

        <Message v-if="error" severity="error" size="small" variant="simple">{{ error }}</Message>
        <Button type="submit" label="Masuk" icon="pi pi-arrow-right" icon-pos="right" :loading="loading" :disabled="!username || !password" fluid />

        <p class="bantuan">Username berawalan nomor usaha, misalnya <b>1-budi</b>. Lupa password? Minta owner meresetnya.</p>
        <RouterLink to="/unduh" class="bantuan"><i class="pi pi-download" /> Unduh aplikasi Android / Windows</RouterLink>
      </form>
    </main>
  </div>
</template>

<style scoped>
.login {
  min-height: 100dvh;
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(0, 1fr);
  background: var(--p-content-background);
}

.brand {
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 32px;
  padding: 40px 48px;
  background: var(--p-primary-color);
  color: var(--p-primary-contrast-color);
}
.brand::before,
.brand::after {
  content: '';
  position: absolute;
  border-radius: 50%;
  background: rgb(255 255 255 / 0.1);
  pointer-events: none;
}
.brand::before {
  width: 420px;
  height: 420px;
  right: -140px;
  top: -120px;
}
.brand::after {
  width: 280px;
  height: 280px;
  left: -90px;
  bottom: -90px;
}
.brand > * {
  position: relative;
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 18px;
}
.mark {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  font-weight: 800;
  background: var(--p-primary-contrast-color);
  color: var(--p-primary-color);
}
.mark.img {
  object-fit: contain;
  background: transparent;
}

.pitch {
  max-width: 460px;
}
.pitch h1 {
  margin: 0 0 12px;
  font-size: 34px;
  line-height: 1.2;
  letter-spacing: -0.02em;
}
.pitch p {
  margin: 0 0 28px;
  font-size: 16px;
  opacity: 0.9;
}
.pitch ul {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.pitch li {
  display: flex;
  gap: 14px;
  align-items: flex-start;
}
.pitch li i {
  flex: none;
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  background: rgb(255 255 255 / 0.18);
}
.pitch li div {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.pitch li span {
  font-size: 14px;
  opacity: 0.85;
}
.foot {
  opacity: 0.7;
}

.side {
  display: grid;
  place-items: center;
  padding: 32px 20px;
}
.form {
  width: 100%;
  max-width: 380px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.form h2 {
  margin: 0;
  font-size: 26px;
  letter-spacing: -0.01em;
}
.sub {
  margin: 0 0 16px;
  color: var(--p-text-muted-color);
}
.form label {
  font-weight: 600;
  margin-top: 6px;
}
.ingat {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 6px 0 8px;
}
.ingat label {
  margin: 0;
  font-weight: 400;
}
.caps {
  color: var(--p-orange-500);
  display: flex;
  align-items: center;
  gap: 6px;
}
.bantuan {
  margin: 14px 0 0;
  font-size: 13px;
  color: var(--p-text-muted-color);
}
.compact {
  display: none;
  margin-bottom: 20px;
}
.compact .mark {
  background: var(--p-primary-color);
  color: var(--p-primary-contrast-color);
}

@media (max-width: 900px) {
  .login {
    grid-template-columns: minmax(0, 1fr);
  }
  .brand {
    display: none;
  }
  .compact {
    display: flex;
  }
}
</style>
