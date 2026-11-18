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
