import { ref } from 'vue'
import { defineStore } from 'pinia'
import { get, put } from '@/api/http'
import { applyTheme, DEFAULT_SETTINGS, loadCachedSettings } from '@/composables/useAppTheme'

export const usePengaturanStore = defineStore('pengaturan', () => {
  const settings = ref(loadCachedSettings())

  async function fetch() {
    try {
      settings.value = { ...DEFAULT_SETTINGS, ...(await get('/pengaturan')) }
      applyTheme(settings.value)
    } catch {
      /* server tidak terjangkau: tetap pakai tema dari cache */
    }
  }

  async function save(form) {
    settings.value = { ...DEFAULT_SETTINGS, ...(await put('/pengaturan', form)) }
    applyTheme(settings.value)
  }

  return { settings, fetch, save }
})
