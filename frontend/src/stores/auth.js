import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { post } from '@/api/http'
import { load, remove, save } from '@/utils/storage'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(load('token', ''))
  const user = ref(load('user', null))

  const isLoggedIn = computed(() => !!token.value)
  const role = computed(() => user.value?.role || '')
  const isKasir = computed(() => role.value === 'kasir' || role.value === 'owner')
  const isOwner = computed(() => role.value === 'owner')

  async function login(username, password) {
    const res = await post('/auth/login', { username, password })
    token.value = res.token
    user.value = res.user
    save('token', res.token)
    save('user', res.user)
  }

  function logout() {
    token.value = ''
    user.value = null
    remove('token')
    remove('user')
  }

  return { token, user, role, isLoggedIn, isKasir, isOwner, login, logout }
})
