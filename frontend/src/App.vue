<script setup>
import { onBeforeUnmount, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { usePengaturanStore } from '@/stores/pengaturan'
import { connectEvents, disconnectEvents, onEvent } from '@/composables/useEvents'

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
onEvent('pengaturan', pengaturan.fetch)
</script>

<template>
  <Toast position="top-center" />
  <ConfirmDialog />
  <RouterView />
</template>
