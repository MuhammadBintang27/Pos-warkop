import { computed, ref, unref, watch } from 'vue'
import { load, save } from '@/utils/storage'

export const SEMUA = -1
export const UTAMA = 0 // meja tanpa area (dibuat sebelum ada fitur area)

/**
 * Tab area meja dari respons GET /meja. Tanpa area sama sekali, `tabs` kosong dan semua meja cocok.
 * `semua`: tambahkan tab "Semua" (daftar meja); denah tidak memakainya karena posisi meja berlaku per area.
 * Pilihan diingat per perangkat lewat `kunci`.
 */
export function useArea(denah, { semua = false, kunci }) {
  const pilih = ref(load(kunci, null))
  watch(pilih, (v) => save(kunci, v))

  const tabs = computed(() => {
    const d = unref(denah)
    const area = d?.area ?? []
    if (!area.length) return []
    const t = area.map((a) => ({ id: a.id, nama: a.nama }))
    if (d.meja.some((m) => !m.areaId)) t.unshift({ id: UTAMA, nama: 'Utama' })
    if (semua) t.unshift({ id: SEMUA, nama: 'Semua' })
    return t
  })
  // pilihan lama bisa sudah tidak ada (area dihapus): pakai tab pertama
  const aktif = computed(() => (tabs.value.some((t) => t.id === pilih.value) ? pilih.value : (tabs.value[0]?.id ?? SEMUA)))
  const cocok = (m) => aktif.value === SEMUA || (m.areaId ?? UTAMA) === aktif.value
  const namaArea = (id) => unref(denah)?.area?.find((a) => a.id === id)?.nama ?? 'Utama'

  return { tabs, aktif, pilih, cocok, namaArea }
}
