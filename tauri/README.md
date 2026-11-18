# Aplikasi desktop POS Warkop (Tauri)

Satu program Windows kecil yang berisi:

- **Aplikasi kasir**: jendela yang membuka alamat server POS (tampilan tetap dari server, jadi update frontend tidak perlu membuild ulang program ini).
- **Stasiun cetak**: berjalan di latar belakang (system tray). Mengambil tiket dari server tiap 2 detik dan mengirimnya ke printer thermal lewat jaringan (port 9100). Menutup jendela hanya menyembunyikannya, dan program menyala otomatis saat Windows dihidupkan setelah stasiun diaktifkan.

## Syarat build (sekali saja)

- Node.js
- Rust (`rustup`, toolchain `stable-x86_64-pc-windows-msvc`)
- Visual Studio Build Tools dengan komponen **MSVC C++** dan **Windows SDK**
- WebView2 (sudah ada di Windows 10/11 terbaru)

## Build

```
cd tauri
npm install
npx tauri icon ikon.png      # sekali, membuat folder src-tauri/icons
npx tauri build              # hasil: src-tauri/target/release/bundle/nsis/*.exe
```

Percobaan jalan tanpa installer: `npx tauri dev`.

## Pemakaian

1. Pasang installer dan buka **POS Warkop**. Isi alamat server (misalnya `https://pos.warkopmu.com`).
2. Di web: Master Data › Printer, tambah printer lalu **Buat token stasiun**.
3. Klik kanan ikon POS Warkop di system tray › **Stasiun cetak…**, tempel token, centang aktifkan, Simpan.
4. Di web: Master Data › Tenant, pilih printer untuk tiap tenant, lalu **Tes cetak** di Master Data › Printer.

Pengaturan (alamat server dan token) disimpan di `%APPDATA%\id.warkop.pos\config.json`.
