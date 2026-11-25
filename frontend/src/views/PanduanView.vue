<script setup>
import { computed, ref } from 'vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

// Isi panduan statis. **teks** dicetak tebal; roles menentukan siapa yang melihat topik/bagian itu.
const KASIR = ['pelayan', 'kasir', 'owner']
const TOPIK = [
  {
    id: 'awal',
    judul: 'Persiapan awal',
    icon: 'pi pi-flag',
    roles: ['owner'],
    bagian: [
      {
        sub: '1. Identitas usaha',
        langkah: [
          'Buka **Pengaturan › Tampilan**, isi nama usaha, alamat, dan catatan kaki struk.',
          'Unggah **logo**. Pakai PNG atau JPG dengan garis tegas: di struk printer logo dicetak hitam putih.',
          'Pilih **Printer kasir** untuk struk dan bill closing (printernya didaftarkan dulu, lihat topik Printer).',
        ],
      },
      {
        sub: '2. Tenant',
        langkah: [
          'Buka **Master Data › Tenant**, tambahkan setiap penjual di tempatmu.',
          'Atur bagi hasil: **persen** dari penjualan, atau **rupiah per menu**. Diskon bill ditanggung tenant.',
          'Pilih printer dapur tenant. Tenant tanpa printer mencetak tiket lewat dialog browser di perangkat kasir.',
        ],
      },
      {
        sub: '3. Menu dan bahan',
        langkah: [
          'Buka **Stok Bahan** bila ingin stok bahan berkurang otomatis. Pakai satuan terkecil (gram, ml).',
          'Buka **Master Data › Menu**: buat kategori, lalu menu dengan tenant dan harganya.',
          'Tambahkan **varian** bila perlu (mis. Panas/Es, level pedas), wajib atau opsional, boleh dengan harga tambahan.',
          'Isi **resep** dari stok bahan, atau **harga beli** untuk barang jadi (rokok, minuman botol) supaya laba rugi akurat.',
        ],
      },
      {
        sub: '4. Meja dan akun',
        langkah: [
          'Buka **Master Data › Meja**: buat area (mis. Lantai 1) lalu mejanya. Susun letaknya di **Denah Meja**.',
          'Buka **Master Data › Pengguna**: buat akun **pelayan**, **kasir**, dan **tenant**. Username diberi awalan kode usaha secara otomatis.',
          'Akun tenant hanya melihat pesanan masuk dan rekapan tenantnya sendiri.',
        ],
      },
    ],
  },
  {
    id: 'kasir',
    judul: 'Kasir harian',
    icon: 'pi pi-shopping-cart',
    roles: KASIR,
    bagian: [
      {
        sub: 'Mulai hari',
        roles: ['kasir', 'owner'],
        langkah: [
          'Tekan **Buka Shift** di halaman Kasir dan isi uang tunai yang ada di laci (kas awal).',
          'Satu shift dipakai bersama semua kasir. Pembayaran tunai, kas keluar, dan kasbon butuh shift yang terbuka.',
          'Menu atau tenant yang hari ini tidak jualan dimatikan di **Ketersediaan**.',
        ],
      },
      {
        sub: 'Mencatat pesanan',
        langkah: [
          'Tekan **Pilih meja** (atau **Bawa pulang**), isi atas nama bila perlu, lalu **Pesanan baru**.',
          'Ketuk menu untuk menambah. Menu bervarian membuka pilihan varian dulu.',
          'Catatan per item (mis. "tidak pedas") lewat ikon pensil, sebelum dikirim.',
          'Tekan **Kirim ke dapur**. Tiket keluar di printer tenant dan pesanan muncul di layar tenant.',
          'Pesanan tambahan: tambah menu lalu **Kirim ke dapur** lagi. Tiketnya berjudul PESANAN TAMBAHAN.',
        ],
      },
      {
        sub: 'Arti ikon di keranjang',
        langkah: [
          '<i class="pi pi-send"></i> sudah dikirim ke dapur, belum dibuat.',
          '<i class="pi pi-hourglass ikon-proses"></i> sedang dibuat tenant.',
          '<i class="pi pi-check-circle ikon-antar"></i> sudah diantar tenant.',
        ],
      },
      {
        sub: 'Pembayaran',
        roles: ['kasir', 'owner'],
        langkah: [
          'Tekan **Lanjut ke pembayaran**, pilih metode: tunai, QRIS, transfer, atau kasbon.',
          'Satu bill boleh dibayar dengan beberapa metode lewat **Tambah metode**.',
          'Kasbon dicatat atas nama pelanggan. Pelunasannya lewat tombol **Kasbon** di halaman Kasir.',
          'Setelah bayar, tekan **Cetak struk** bila pelanggan minta.',
          'Diskon bill diisi di keranjang sebelum bayar.',
        ],
      },
      {
        sub: 'Mengatur bill',
        roles: ['kasir', 'owner'],
        langkah: [
          'Lewat tombol titik tiga di keranjang: **Pindah meja**, **Gabung bill** (dua meja jadi satu), **Pisah bill** (bayar sendiri-sendiri).',
          'Uang keluar dari laci (beli gas, es batu) dicatat di **Kas Keluar**.',
        ],
      },
      {
        sub: 'Pelayan',
        roles: ['pelayan'],
        langkah: [
          'Pelayan bisa mencatat pesanan dan mengirimnya ke dapur.',
          'Pembayaran, diskon, dan pembatalan item yang sudah dikirim dilakukan kasir.',
        ],
      },
      {
        sub: 'Akhir hari',
        roles: ['kasir', 'owner'],
        langkah: [
          'Tekan **Tutup Shift**, hitung semua uang tunai di laci dan isi jumlahnya.',
          'Sistem menghitung kas yang seharusnya ada dan menampilkan selisihnya. Riwayatnya ada di **Laporan › Riwayat Shift**.',
        ],
      },
    ],
  },
  {
    id: 'batal',
    judul: 'Pembatalan dan refund',
    icon: 'pi pi-times-circle',
    roles: KASIR,
    bagian: [
      {
        sub: 'Belum dikirim ke dapur',
        langkah: ['Hapus saja itemnya (tanda ×). Tidak perlu alasan dan dapur tidak menerima apa-apa.'],
      },
      {
        sub: 'Sudah dikirim, bill belum dibayar',
        langkah: [
          'Kasir menekan × atau − pada item, atau **Batalkan pesanan** di menu titik tiga untuk seluruh bill.',
          'Alasan wajib diisi. Centang **Sudah dibuat dapur** bila makanannya sudah jadi, supaya stok tidak dikembalikan.',
          'Tenant menerima tiket **BATAL** di printernya, kartu merah di bagian Dibatalkan, dan notifikasi di HP.',
          'Bila tenant sudah menekan **Proses**, muncul peringatan merah dan "Sudah dibuat" langsung tercentang. Beri tahu pelanggan bahwa pesanannya sudah dimasak.',
        ],
      },
      {
        sub: 'Bill sudah lunas',
        roles: ['owner'],
        langkah: [
          'Buka **Riwayat**, pilih bill-nya.',
          '**Buka ulang**: untuk salah input. Pembayaran dihapus, bill dikoreksi di kasir lalu dibayar ulang. Hanya untuk bill yang dibayar di shift yang masih buka.',
          '**Refund**: membatalkan seluruh bill. Uang tunai dicatat keluar dari kas shift aktif, utang kasbon dihapus, QRIS/transfer dikembalikan manual.',
          'Saat refund, hapus centang **Sudah dibuat/disajikan** bila dapur masih memasaknya: tiket BATAL akan dicetak.',
          'Bill yang sudah masuk closing tenant harus dibatalkan closing-nya dulu.',
        ],
      },
      {
        sub: 'Bill sudah lunas',
        roles: ['pelayan', 'kasir'],
        langkah: ['Pembatalan bill yang sudah lunas (buka ulang atau refund) hanya bisa dilakukan owner.'],
      },
      {
        sub: 'Jejak',
        roles: ['owner'],
        langkah: ['Semua pembatalan, diskon, dan refund tercatat di **Pengaturan › Log Aktivitas** dan laporan tenant.'],
      },
    ],
  },
  {
    id: 'tenant',
    judul: 'Pesanan Masuk (tenant)',
    icon: 'pi pi-inbox',
    roles: ['tenant', 'owner'],
    bagian: [
      {
        sub: 'Supaya notifikasi masuk',
        langkah: [
          'Pasang aplikasi Android PSD POS dari halaman **/unduh**, lalu login dengan akun tenant.',
          'Izinkan notifikasi. Pesanan baru dan pembatalan masuk ke HP walau layar mati.',
        ],
      },
      {
        sub: 'Alur di layar Pesanan Masuk',
        langkah: [
          '**Antrean**: pesanan baru dari kasir. Kartu yang baru masuk berkedip dan berbunyi.',
          'Tekan **Proses** saat mulai membuat. Item pindah ke **Diproses** dan kasir melihat bahwa pesanannya sedang dibuat.',
          'Tekan **Selesai** saat sudah diantar ke meja. Item pindah ke **Selesai**.',
          'Salah tekan? **urungkan** di Diproses, atau **Batal** di Selesai.',
        ],
      },
      {
        sub: 'Bagian Dibatalkan',
        langkah: [
          'Kartu merah berarti kasir atau owner membatalkan pesanan itu, lengkap dengan alasan dan nama pembatalnya.',
          '"Jangan dibuat": hentikan pembuatannya. "Sudah dibuat": hanya pemberitahuan, penjualan itu tidak dihitung.',
        ],
      },
      {
        sub: 'Rekapan',
        roles: ['tenant'],
        langkah: ['Menu **Rekapan** menampilkan penjualan, komisi pemilik tempat, dan bagian Anda per hari dan per menu.'],
      },
    ],
  },
  {
    id: 'laporan',
    judul: 'Laporan dan closing tenant',
    icon: 'pi pi-chart-line',
    roles: ['owner'],
    bagian: [
      {
        sub: 'Closing tenant',
        langkah: [
          'Buka **Laporan › Tenant**, pilih tanggal, lalu **Closing** per tenant (atau **Closing semua**). Angkanya dibekukan supaya setoran tidak berubah diam-diam.',
          'Cetak bill closing, serahkan uangnya ke tenant, lalu **Tandai disetor**.',
          'Setor mingguan: pilih rentang beberapa hari, lalu **Cetak rekap periode** dan **Tandai disetor** sekaligus.',
          'Closing yang belum disetor masih bisa dibatalkan.',
        ],
      },
      {
        sub: 'Laporan lain',
        langkah: [
          '**Ringkasan & Tren**, **Kategori, Tenant & Jam**, **Kinerja Karyawan**, **Kasbon & Stok**.',
          '**Riwayat Shift**: penjualan dan selisih kas tiap shift.',
          '**Laba Rugi**: penjualan dikurangi HPP (resep/harga beli), pengeluaran, dan gaji yang sudah dibayar.',
        ],
      },
    ],
  },
  {
    id: 'qr',
    judul: 'Menu untuk pelanggan (QR)',
    icon: 'pi pi-qrcode',
    roles: ['owner'],
    bagian: [
      {
        langkah: [
          'Buka **Pengaturan › Tampilan**, bagian **Menu untuk pelanggan**.',
          '**Unduh QR**, cetak, lalu tempel di meja. Pelanggan bisa melihat menu dan harga tanpa login.',
          'Menu yang dimatikan di Ketersediaan atau stoknya habis tampil sebagai habis di menu pelanggan, begitu juga tenant yang tutup.',
        ],
      },
    ],
  },
  {
    id: 'printer',
    judul: 'Printer dan kalau ada masalah',
    icon: 'pi pi-print',
    roles: ['owner', 'kasir'],
    bagian: [
      {
        sub: 'Cara kerja',
        roles: ['owner'],
        langkah: [
          'Printer thermal (LAN/WiFi, ESC/POS) didaftarkan di **Master Data › Printer** dengan IP-nya. Kunci IP printer di router supaya tidak berubah.',
          'Perlu satu **stasiun cetak** di jaringan warung: HP/tablet Android (**Jadikan perangkat ini stasiun**) atau PC Windows (token stasiun lewat ikon di system tray).',
          'Tekan **Tes cetak** pada setiap printer. Panduan lengkapnya ada di dokumen pemasangan printer.',
        ],
      },
      {
        sub: 'Kalau tiket atau struk tidak keluar',
        langkah: [
          'Di halaman Kasir muncul peringatan printer bila stasiun cetak tidak terhubung.',
          'Pastikan HP/PC stasiun menyala, ada internet, dan berada di WiFi yang sama dengan printer (bukan WiFi tamu).',
          'Tiket yang gagal dicoba ulang otomatis tiap 15 detik. Tiket lebih dari 6 jam tidak dicetak lagi: pakai **Cetak ulang** di Riwayat cetak.',
          'Huruf acak: printer belum di mode ESC/POS. Teks terpotong: lebar kertas (58/80 mm) salah.',
        ],
      },
      {
        sub: 'Kalau notifikasi tenant tidak masuk',
        roles: ['owner'],
        langkah: [
          'Tenant harus login di aplikasi Android (bukan browser) dan mengizinkan notifikasi.',
          'Di HP Xiaomi, Oppo, Vivo, Realme, Huawei: nyalakan Autostart dan setel baterai aplikasi ke Tanpa batasan.',
        ],
      },
    ],
  },
]

const boleh = (x) => !x.roles || x.roles.includes(auth.role)
const topik = computed(() =>
  TOPIK.filter(boleh)
    .map((t) => ({ ...t, bagian: t.bagian.filter(boleh) }))
    .filter((t) => t.bagian.length),
)
const terbuka = ref(new Set([topik.value[0]?.id]))
function toggle(id) {
  const s = new Set(terbuka.value)
  s.has(id) ? s.delete(id) : s.add(id)
  terbuka.value = s
}

// isi panduan ditulis sendiri (bukan input pengguna), jadi aman dirender sebagai HTML
const tebal = (s) => s.replace(/\*\*(.+?)\*\*/g, '<b>$1</b>')
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-question-circle" title="Panduan" subtitle="Cara memakai aplikasi, sesuai peran akunmu" />

    <section v-for="t in topik" :key="t.id" class="panel topik">
      <button class="kepala" :aria-expanded="terbuka.has(t.id)" @click="toggle(t.id)">
        <i :class="t.icon" class="ikon" />
        <strong>{{ t.judul }}</strong>
        <i class="pi kanan" :class="terbuka.has(t.id) ? 'pi-chevron-up' : 'pi-chevron-down'" />
      </button>
      <div v-if="terbuka.has(t.id)" class="isi">
        <div v-for="(b, n) in t.bagian" :key="n" class="bagian">
          <h4 v-if="b.sub">{{ b.sub }}</h4>
          <ul>
            <li v-for="(l, k) in b.langkah" :key="k" v-html="tebal(l)" />
          </ul>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.gap {
  gap: 12px;
}
.topik {
  padding: 0;
  overflow: hidden;
}
.kepala {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 14px 16px;
  border: 0;
  background: none;
  font: inherit;
  font-size: 1.05rem;
  color: inherit;
  text-align: left;
  cursor: pointer;
}
.ikon {
  color: var(--accent-text);
}
.kanan {
  margin-left: auto;
  color: var(--p-text-muted-color);
}
.isi {
  padding: 0 16px 14px;
  border-top: 1px solid var(--p-content-border-color);
}
.bagian h4 {
  margin: 14px 0 6px;
}
.bagian ul {
  margin: 0;
  padding-left: 20px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  line-height: 1.5;
}
:deep(.ikon-proses) {
  color: var(--p-orange-500);
}
:deep(.ikon-antar) {
  color: var(--p-green-500);
}
</style>
