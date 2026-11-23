import { describe, expect, it } from 'vitest'
import { gabungItem, gabungTerkirim } from './print'

const item = (o) => ({ menuId: 1, nama: 'Dimsum Ayam', harga: 20000, qty: 1, opsi: '', catatan: '', ...o })

describe('gabungItem', () => {
  it('menggabung baris yang sama dan menjumlahkan jumlahnya', () => {
    const hasil = gabungItem([item({ qty: 3 }), item({ qty: 5 })])
    expect(hasil).toHaveLength(1)
    expect(hasil[0].qty).toBe(8)
  })

  it('tidak menggabung varian, nama, atau harga yang berbeda', () => {
    const hasil = gabungItem([item({}), item({ opsi: 'Es' }), item({ harga: 25000 }), item({ nama: 'Dimsum Udang' })])
    expect(hasil).toHaveLength(4)
  })

  it('catatan tidak membedakan baris struk karena tidak dicetak', () => {
    expect(gabungItem([item({}), item({ catatan: 'tidak pedas' })])).toHaveLength(1)
  })

  it('tidak mengubah data asli', () => {
    const asli = [item({ qty: 3 }), item({ qty: 5 })]
    gabungItem(asli)
    expect(asli[0].qty).toBe(3)
  })
})

describe('gabungTerkirim', () => {
  it('menggabung baris terkirim yang sama, baris belum terkirim tetap terpisah', () => {
    const hasil = gabungTerkirim([
      item({ id: 1, dikirim: true }),
      item({ id: 2, dikirim: true, qty: 2 }),
      item({ id: 3, dikirim: false }),
      item({ id: 4, dikirim: true, opsi: 'Es' }),
    ])
    expect(hasil.map((i) => i.qty)).toEqual([3, 1, 1])
    expect(hasil[0].rows.map((r) => r.id)).toEqual([1, 2])
    expect(hasil[1].id).toBe(3)
  })

  it('baris yang sudah diantar tenant tidak digabung dengan yang belum', () => {
    const hasil = gabungTerkirim([
      item({ id: 1, dikirim: true, siapAt: '2026-11-23T10:00:00' }),
      item({ id: 2, dikirim: true }),
      item({ id: 3, dikirim: true, siapAt: '2026-11-23T10:05:00' }),
    ])
    expect(hasil.map((i) => i.rows.map((r) => r.id))).toEqual([[1, 3], [2]])
  })
})
