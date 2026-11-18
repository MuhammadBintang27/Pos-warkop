<script setup>
import { onBeforeUnmount, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { usePengaturanStore } from '@/stores/pengaturan'
import { connectEvents, disconnectEvents, onEvent } from '@/composables/useEvents'
import { post } from '@/api/http'
import { load } from '@/utils/storage'

const router = useRouter()
const auth = useAuthStore()
const pengaturan = usePengaturanStore()

const expired = () => {
  auth.logout()
  router.push('/login')
}

onMounted(() => {
  pengaturan.fetch()
  window.addEventListener('auth:expired', expired)
})
onBeforeUnmount(() => window.removeEventListener('auth:expired', expired))

watch(() => auth.token, (t) => (t ? connectEvents(t) : disconnectEvents()), { immediate: true })
// tema milik usaha baru bisa diambil setelah login (sebelumnya yang terambil tema bawaan)
watch(() => auth.token, (t) => t && pengaturan.fetch())

// HP tenant didaftarkan untuk push notifikasi setiap kali aplikasi dibuka atau login
watch(
  () => auth.token,
  (t) => {
    const fcm = load('fcmToken', '')
    if (t && auth.role === 'tenant' && fcm) post('/tenant-saya/perangkat', { token: fcm }).catch(() => {})
  },
  { immediate: true },
)
onEvent('pengaturan', pengaturan.fetch)
</script>

<template>
  <Toast position="top-center" />
  <ConfirmDialog />
  <RouterView />
</template>
