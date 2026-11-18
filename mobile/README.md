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
- Cetak dari HP lewat browser terbatas. Struk dan tiket dicetak di PC kasir oleh stasiun cetak (lihat `tauri/`).
