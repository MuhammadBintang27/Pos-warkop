import { createApp } from 'vue'
import { createPinia } from 'pinia'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import ConfirmationService from 'primevue/confirmationservice'
import 'primeicons/primeicons.css'
import '@/assets/main.css'
import App from './App.vue'
import router from './router'
import { bindConfig, initTheme } from '@/composables/useAppTheme'
import { save } from '@/utils/storage'

// Aplikasi Android membawa token FCM lewat ?fcm=...; disimpan lalu dibuang dari alamat.
const url = new URL(location.href)
if (url.searchParams.has('fcm')) {
  save('fcmToken', url.searchParams.get('fcm'))
  url.searchParams.delete('fcm')
  history.replaceState(null, '', url)
}

// Tampilan aplikasi disimpan di perangkat (src/sw.js) supaya tetap bisa dibuka saat internet putus.
if (import.meta.env.PROD && 'serviceWorker' in navigator) navigator.serviceWorker.register('/sw.js').catch(() => {})

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(PrimeVue, {
  ripple: true,
  theme: { preset: initTheme(), options: { darkModeSelector: '.app-dark', cssLayer: false } },
})
app.use(ToastService)
app.use(ConfirmationService)
bindConfig(app.config.globalProperties.$primevue.config)

app.mount('#app')
