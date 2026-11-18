import { describe, expect, it } from 'vitest'
import { gabungItem } from './print'

const item = (o) => ({ menuId: 1, nama: 'Dimsum Ayam', harga: 20000, qty: 1, opsi: '', catatan: '', ...o })

describe('gabungItem', () => {
  it('menggabung baris yang sama dan menjumlahkan jumlahnya', () => {
    const hasil = gabungItem([item({ qty: 3 }), item({ qty: 5 })])
    expect(hasil).toHaveLength(1)
    expect(hasil[0].qty).toBe(8)
  })

  it('tidak menggabung varian, catatan, atau harga yang berbeda', () => {
    const hasil = gabungItem([item({}), item({ opsi: 'Es' }), item({ catatan: 'tidak pedas' }), item({ harga: 25000 }), item({ menuId: 2 })])
    expect(hasil).toHaveLength(5)
  })

  it('tidak mengubah data asli', () => {
    const asli = [item({ qty: 3 }), item({ qty: 5 })]
    gabungItem(asli)
    expect(asli[0].qty).toBe(3)
  })
})
