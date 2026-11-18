import { useToast } from 'primevue/usetoast'
import { errMsg } from '@/api/http'

/** Jalankan aksi async; error tampil sebagai toast. Mengembalikan true bila berhasil. */
export function useRun() {
  const toast = useToast()
  return async (fn, success) => {
    try {
      await fn()
      if (success) toast.add({ severity: 'success', summary: success, life: 2000 })
      return true
    } catch (e) {
      toast.add({ severity: 'error', summary: 'Gagal', detail: errMsg(e), life: 4000 })
      return false
    }
  }
}
