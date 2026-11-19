import { computed, ref, unref } from 'vue'

/** Pencarian di sisi browser untuk tabel: `hasil` berisi baris yang salah satu kolomnya memuat kata yang dicari. */
export function useCari(sumber, ...kolom) {
  const cari = ref('')
  const hasil = computed(() => {
    const list = unref(sumber) ?? []
    const q = cari.value.trim().toLowerCase()
    if (!q) return list
    return list.filter((r) => kolom.some((k) => String(r[k] ?? '').toLowerCase().includes(q)))
  })
  return { cari, hasil }
}
