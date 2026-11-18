export const KATEGORI_PENGELUARAN = [
  { value: 'bahan', label: 'Belanja bahan' },
  { value: 'operasional', label: 'Operasional (listrik, air, gas)' },
  { value: 'sewa', label: 'Sewa' },
  { value: 'transport', label: 'Transport' },
  { value: 'perawatan', label: 'Perawatan dan perbaikan' },
  { value: 'lainnya', label: 'Lainnya' },
]

export const labelKategori = (v) => KATEGORI_PENGELUARAN.find((k) => k.value === v)?.label ?? v
