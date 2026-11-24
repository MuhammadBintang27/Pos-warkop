# Pemasangan Printer POS Warkop

Panduan memasang printer thermal di warung, dari printernya sampai tersambung ke aplikasi, baik lewat HP/tablet Android maupun PC Windows.

## Cara kerjanya

```
 Kasir / pelanggan pesan
          │
          ▼
 ┌─────────────────┐      internet      ┌──────────────────────────┐
 │   Server POS    │ ◄────────────────► │  Stasiun cetak           │
 │ (antrean tiket) │   ambil tiket      │  HP/tablet Android       │
 └─────────────────┘   tiap 2 detik     │  ATAU PC Windows         │
                                        └────────────┬─────────────┘
                                                     │ WiFi / LAN warung
                                    ┌────────────────┼────────────────┐
                                    ▼                ▼                ▼
                              Printer kasir   Printer dapur A   Printer dapur B
```

Server ada di internet, sedangkan printer ada di jaringan lokal warung. Server tidak bisa menjangkau printer secara langsung. Karena itu harus ada satu **stasiun cetak** di warung, yaitu HP/tablet Android atau PC yang menjalankan aplikasi POS Warkop. Stasiun mengambil tiket dari server lalu meneruskannya ke printer yang benar.

**Satu stasiun cukup untuk semua printer.** Setiap tiket sudah membawa alamat printer tujuannya.

## 1. Yang dibutuhkan

- **Printer thermal ESC/POS yang punya port LAN (RJ45) atau WiFi**, kertas 58 mm atau 80 mm. Printer yang hanya punya USB atau Bluetooth tidak bisa dipakai.
- **Router WiFi** warung, dan kabel LAN untuk printer LAN.
- **Satu stasiun cetak**, pilih salah satu:
  - HP/tablet Android (Android 7 ke atas), atau
  - PC/laptop Windows.
- Koneksi internet di router, karena stasiun harus bisa menjangkau server.

## 2. Memasang printer (hardware)

1. Pasang kertas, lalu nyalakan printer.
2. Sambungkan printer ke jaringan warung:
   - **Printer LAN:** colok kabel LAN dari printer ke port LAN di router.
   - **Printer WiFi:** sambungkan ke WiFi warung memakai aplikasi atau tool bawaan merek printer (lihat buku manualnya).
3. **Cetak lembar self-test untuk melihat alamat IP printer.** Biasanya caranya: matikan printer, tahan tombol **FEED**, nyalakan printer, lalu lepas tombolnya setelah kertas keluar. Cari tulisan seperti `IP Address: 192.168.1.87`.
4. **Kunci alamat IP-nya.** Ini langkah yang paling penting. Kalau IP printer berubah, misalnya setelah router mati listrik, tiket akan gagal tercetak. Pilih salah satu cara:
   - Di halaman admin router, buka menu *DHCP Reservation / Address Reservation / Static Lease*. Kunci IP tersebut untuk printer (dikenali dari MAC address yang juga tercetak di self-test).
   - Atau atur IP statis di printer lewat halaman setelannya. Ketik `http://<IP printer>` di browser, atau pakai tool dari merek printer.
5. Catat setiap printer: **nama** (mis. "Dapur Tenant A"), **IP**, dan **lebar kertas**. Port-nya hampir selalu **9100**.

> **Syarat jaringan:** stasiun cetak dan semua printer harus berada di **jaringan yang sama**. Biasanya angka depan IP-nya sama, misalnya semua `192.168.1.x`. Jangan sambungkan stasiun ke WiFi *tamu/guest*. Banyak router memblokir perangkat di WiFi tamu agar tidak bisa saling terhubung (fitur *AP isolation / client isolation*).

## 3. Mendaftarkan printer di aplikasi

Login sebagai **owner**.

1. Buka **Master Data › Printer**, lalu **Tambah printer**.
2. Isi nama, alamat IP, port `9100`, dan lebar kertas (58 atau 80 mm). Simpan.
3. Ulangi untuk setiap printer.

Lalu tentukan printer mana mencetak apa:

| Yang dicetak | Diatur di |
|---|---|
| Tiket dapur per tenant | **Master Data › Tenant**: ubah tenant, lalu pilih printernya. Dua tenant boleh memakai satu printer. |
| Struk pembayaran dan bill closing | **Pengaturan › Tampilan › Printer kasir** |

Tenant yang tidak diberi printer tetap mencetak lewat dialog print browser di perangkat kasir.

## 4. Menyalakan stasiun cetak

Pilih **satu** perangkat saja. Menyalakan stasiun baru otomatis mematikan stasiun lama, karena tokennya diganti.

Aplikasi Android (APK) dan installer Windows bisa diunduh di halaman `/unduh` pada alamat server POS.

### Pilihan A: HP/tablet Android (tanpa PC)

1. Pasang APK POS Warkop di HP/tablet kasir.
2. Sambungkan HP/tablet ke **WiFi warung yang sama** dengan printer.
3. Buka aplikasinya, login sebagai **owner**, lalu buka **Master Data › Printer**.
4. Tekan **Jadikan perangkat ini stasiun**, lalu konfirmasi.
5. Kalau Android bertanya soal **penghemat baterai**, pilih **Izinkan**. Kalau ada permintaan izin notifikasi, izinkan juga.
6. Muncul notifikasi tetap **"Stasiun cetak aktif"** yang menunjukkan statusnya ("Siap", "2 tiket menunggu", dan seterusnya).
7. Sesudah itu boleh logout dan ganti ke akun kasir. Stasiun tetap jalan, dan menyala lagi sendiri saat HP dinyalakan ulang.

**Khusus HP Xiaomi, Oppo, Vivo, Realme, dan Huawei:** merek-merek ini punya pembunuh aplikasi sendiri. Agar stasiun tidak mati diam-diam:
- nyalakan **Autostart** untuk POS Warkop (di Setelan › Aplikasi › POS Warkop),
- setel penghemat baterai aplikasi ke **Tanpa batasan**,
- kunci aplikasi di daftar aplikasi terbaru (ikon gembok).

**Kalau stasiun mati di tengah jalan**, kasir tidak perlu memanggil owner. Di layar Kasir muncul tombol **Nyalakan stasiun**. Tombol itu menyalakan lagi stasiun di HP tersebut memakai token yang sudah tersimpan. Tombol ini hanya muncul di HP yang pernah diaktifkan owner.

Untuk mematikan stasiun: buka **Master Data › Printer** di perangkat itu, lalu tekan **Matikan**.

### Pilihan B: PC Windows

1. Pasang installer POS Warkop di PC kasir. PC harus tersambung ke jaringan yang sama dengan printer, boleh lewat kabel atau WiFi.
2. Di browser atau aplikasinya, login sebagai **owner**, buka **Master Data › Printer**, lalu tekan **Buat token stasiun**. Salin tokennya. Token hanya ditampilkan sekali.
3. Klik kanan ikon POS Warkop di *system tray* Windows (pojok kanan bawah), lalu pilih **Stasiun cetak…**.
4. Tempel token, centang **Aktifkan stasiun cetak di PC ini**, lalu tekan **Simpan**.
5. Statusnya harus menjadi **Terhubung**. Menutup jendela aplikasi tidak mematikan stasiun, karena ia tetap berjalan di tray dan menyala sendiri saat Windows dihidupkan.

## 5. Mengetes

1. Di **Master Data › Printer**, label di bagian *Stasiun cetak* harus **Terhubung**. Artinya stasiun melapor ke server dalam 15 detik terakhir.
2. Tekan **Tes cetak** di setiap printer. Tiket percobaan harus keluar dalam beberapa detik.
3. Buat satu pesanan sungguhan. Tiket dapur harus keluar di printer tenantnya, dan struk di printer kasir.
4. **Khusus Android:** matikan layar HP selama 5–10 menit, lalu tes cetak lagi. Kalau tidak keluar, periksa lagi pengaturan baterai di Pilihan A.

## 6. Kalau tidak tercetak

Lihat **Riwayat cetak** di halaman Printer. Tiket yang gagal menampilkan jumlah percobaan dan pesan galatnya.

| Gejala | Penyebab umum | Yang dilakukan |
|---|---|---|
| Stasiun **Tidak terhubung** | Stasiun mati, tidak ada internet, atau token sudah diganti | Pastikan HP/PC menyala dan ada internet. Di Android, cek notifikasinya. Kalau token pernah dibuat ulang, nyalakan lagi stasiunnya. |
| "Tidak bisa terhubung ke 192.168.x.x:9100" | IP printer berubah, printer mati, atau beda jaringan | Cetak self-test lagi dan cocokkan IP-nya. Pastikan stasiun tidak tersambung ke WiFi tamu. Kunci IP printer (langkah 2.4). |
| Tiket keluar tapi huruf acak | Printer bukan ESC/POS atau mode emulasinya salah | Setel printer ke mode ESC/POS (lihat manual printer). |
| Teks terpotong atau terlalu sempit | Lebar kertas salah | Ubah lebar kertas printer di aplikasi (58/80 mm). |
| Tiket tercetak dobel | Seharusnya tidak terjadi, karena hanya satu stasiun yang bisa aktif | Laporkan ke pengembang. |
| Tiket lama tidak tercetak | Tiket lebih dari 6 jam dianggap kedaluwarsa | Pakai **Cetak ulang** di Riwayat cetak. |

Tiket yang gagal dicoba lagi otomatis setiap 15 detik selama stasiun menyala, jadi setelah masalahnya beres tiket yang tertunda akan keluar sendiri.
