import { useToast } from 'primevue/usetoast'
import { errMsg } from '@/api/http'

/** Jalankan aksi async; error tampil sebagai toast. `success` boleh fungsi bila pesannya bergantung hasil. Mengembalikan true bila berhasil. */
export function useRun() {
  const toast = useToast()
  return async (fn, success) => {
    try {
      await fn()
      const pesan = typeof success === 'function' ? success() : success
      if (pesan) toast.add({ severity: 'success', summary: pesan, life: 2000 })
      return true
    } catch (e) {
      toast.add({ severity: 'error', summary: 'Gagal', detail: errMsg(e), life: 4000 })
      return false
    }
  }
}
