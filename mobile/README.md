# Cangkang Android (APK)

Hanya membuka alamat server POS. Tampilannya tetap dari server, jadi update frontend tidak perlu membuild ulang APK.

## Build

Butuh Android Studio (JDK dan Android SDK ikut terpasang). Gradle-nya Capacitor 8 butuh **Java 21**: pakai JDK bawaan Android Studio.

```
cd mobile
npm install
npx cap sync
cd android
set JAVA_HOME=C:\Program Files\Android\Android Studio\jbr
gradlew.bat assembleDebug
```

Hasilnya: `mobile/android/app/build/outputs/apk/debug/app-debug.apk`.

## Catatan

- Layar pertama meminta alamat server dan mengingatnya. Untuk menggantinya: hapus data aplikasi (Setelan > Aplikasi > POS Warkop > Hapus data) atau pasang ulang.
- `android:usesCleartextTraffic="true"` di `AndroidManifest.xml` mengizinkan alamat `http://` untuk uji jaringan lokal. **Cabut sebelum produksi** dan pakai HTTPS.
- `www/capacitor-core.js` dan `www/capacitor-push.js` adalah salinan `dist` dari `@capacitor/core` dan `@capacitor/push-notifications`, karena halaman awal tidak memakai bundler. Salin ulang setelah menaikkan versi Capacitor.
- Push notifikasi butuh `android/app/google-services.json` dari Firebase (project `pos-warkop-5b964`, package `id.warkop.pos`).
- Cetak dari HP lewat browser terbatas. Struk dan tiket dicetak di PC kasir oleh stasiun cetak (lihat `tauri/`).
