# Cangkang Android (APK)

Hanya membuka alamat server POS. Tampilannya tetap dari server, jadi update frontend tidak perlu membuild ulang APK.

## Build

Butuh Android Studio (JDK dan Android SDK ikut terpasang). Gradle-nya Capacitor 8 butuh **Java 21**: pakai JDK bawaan Android Studio.

```
cd mobile
npm install
npm run sync
cd android
set JAVA_HOME=C:\Program Files\Android\Android Studio\jbr
gradlew.bat assembleDebug
```

Hasilnya: `mobile/android/app/build/outputs/apk/debug/app-debug.apk`.

## Merilis APK baru (pembaruan dari dalam aplikasi)

1. Naikkan `versionCode` (dan `versionName`) di `frontend/public/dl/versi.json`, lalu isi `catatan`. Gradle membaca versi dari file ini, jadi tidak ada angka versi lain yang perlu diubah.
2. Build APK seperti di atas, lalu salin ke `frontend/public/dl/pos-warkop.apk`.
3. Commit dan deploy. Pembaruan bersifat wajib. Saat aplikasi dibuka, HP dengan versi lebih lama langsung mengunduh APK di balik dialog yang tidak bisa ditutup. Setelah itu layar pasang Android muncul, dan pengguna cukup menekan **Install**.

APK harus selalu ditandatangani dengan kunci yang sama, karena Android menolak update dengan kunci berbeda. Saat ini dipakai kunci debug laptop pembuat (`%USERPROFILE%\.android\debug.keystore`). **Cadangkan file itu.** Kalau hilang, semua HP harus uninstall lalu pasang ulang secara manual.

## Catatan

- Alamat server tertanam di `www/index.html` (`SERVER`). Pindah domain berarti build ulang APK.
- `android:usesCleartextTraffic="true"` di `AndroidManifest.xml` mengizinkan alamat `http://` untuk uji jaringan lokal. **Cabut sebelum produksi** dan pakai HTTPS.
- `www/capacitor-core.js` dan `www/capacitor-push.js` adalah salinan `dist` dari `@capacitor/core` dan `@capacitor/push-notifications`, karena halaman awal tidak memakai bundler. Salin ulang setelah menaikkan versi Capacitor.
- Suara notifikasi: `suara/pesanan.wav` (chime buatan sendiri), disalin ke `android/app/src/main/res/raw/` oleh `npm run sync` dan dipakai channel `pesanan_v2`. Mengganti suara berarti harus memakai id channel baru di `www/index.html` dan `backend/internal/services/push.go`, karena Android mengunci suara channel.
- Push notifikasi butuh `android/app/google-services.json` dari Firebase (project `pos-warkop-5b964`, package `id.warkop.pos`).
- Mode offline kasir (`frontend/src/api/offline.js`): HP kasir mencetak tiket dan struk langsung ke IP printer lewat `StasiunAndroid.cetak(ip, port, base64)` (`MainActivity.java`), tanpa server. Fitur ini hanya aktif di APK yang punya fungsi itu (1.4 ke atas).
- Cetak dari HP lewat browser terbatas. Struk dan tiket dicetak oleh stasiun cetak: di PC kasir (lihat `tauri/`) atau di APK ini sendiri (`StasiunService.java`, foreground service, porting dari `tauri/src-tauri/src/main.rs`). Owner menyalakannya dari halaman Printer lewat `window.StasiunAndroid`.
