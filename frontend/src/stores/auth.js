import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { http, post } from '@/api/http'
import { load, remove, save } from '@/utils/storage'
import { lupakanData } from '@/api/offline'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(load('token', ''))
  const user = ref(load('user', null))

  const isLoggedIn = computed(() => !!token.value)
  const role = computed(() => user.value?.role || '')
  const isKasir = computed(() => role.value === 'kasir' || role.value === 'owner')
  const isOwner = computed(() => role.value === 'owner')

  async function login(username, password) {
    const res = await post('/auth/login', { username, password })
    lupakanData() // data offline milik akun sebelumnya jangan terbawa
    token.value = res.token
    user.value = res.user
    save('token', res.token)
    save('user', res.user)
  }

  // HP tenant yang logout berhenti menerima notifikasi; header dikirim langsung karena token segera dihapus.
  function lepasPerangkat() {
    const fcm = load('fcmToken', '')
    if (role.value !== 'tenant' || !fcm) return
    http.post('/tenant-saya/perangkat/hapus', { token: fcm }, { headers: { Authorization: `Bearer ${token.value}` } }).catch(() => {})
  }

  function logout() {
    lepasPerangkat()
    lupakanData()
    token.value = ''
    user.value = null
    remove('token')
    remove('user')
  }

  return { token, user, role, isLoggedIn, isKasir, isOwner, login, logout }
})
