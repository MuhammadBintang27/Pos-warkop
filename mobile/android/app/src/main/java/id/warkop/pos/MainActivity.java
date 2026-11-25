package id.warkop.pos;

import android.app.DownloadManager;
import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.content.IntentFilter;
import android.content.SharedPreferences;
import android.database.Cursor;
import android.net.Uri;
import android.os.Build;
import android.os.Bundle;
import android.os.Environment;
import android.os.PowerManager;
import android.provider.Settings;
import android.webkit.JavascriptInterface;

import androidx.core.content.ContextCompat;
import androidx.core.content.FileProvider;
import androidx.core.content.pm.PackageInfoCompat;

import com.getcapacitor.BridgeActivity;

import java.io.File;

import org.json.JSONObject;

public class MainActivity extends BridgeActivity {
    @Override
    public void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        // window.StasiunAndroid untuk halaman Printer di server; dipasang sebelum halaman server dimuat
        bridge.getWebView().addJavascriptInterface(new StasiunJs(), "StasiunAndroid");
    }

    class StasiunJs {
        @JavascriptInterface
        public String status() {
            SharedPreferences p = StasiunService.prefs(MainActivity.this);
            try {
                return new JSONObject()
                        .put("aktif", p.getBoolean("aktif", false))
                        .put("adaToken", !p.getString("token", "").isEmpty())
                        .put("pesan", StasiunService.pesan)
                        .toString();
            } catch (Exception e) {
                return "{}";
            }
        }

        /** token kosong = pakai token tersimpan. */
        @JavascriptInterface
        public void simpan(String server, String token, boolean aktif) {
            SharedPreferences.Editor e = StasiunService.prefs(MainActivity.this).edit().putBoolean("aktif", aktif).putString("server", server);
            if (token != null && !token.isEmpty()) e.putString("token", token);
            e.apply();
            if (aktif) mintaBebasHematBaterai();
            StasiunService.terapkan(MainActivity.this);
        }

        /** versionCode APK yang terpasang; dibandingkan halaman web dengan /dl/versi.json di server. */
        @JavascriptInterface
        public long versi() {
            try {
                return PackageInfoCompat.getLongVersionCode(getPackageManager().getPackageInfo(getPackageName(), 0));
            } catch (Exception e) {
                return 0;
            }
        }

        /** versionName APK yang terpasang, untuk ditampilkan di sidebar. */
        @JavascriptInterface
        public String versiNama() {
            try {
                return getPackageManager().getPackageInfo(getPackageName(), 0).versionName;
            } catch (Exception e) {
                return "";
            }
        }

        /**
         * Unduh APK lalu buka layar pasang Android (pengguna tetap menekan Install: aturan Android di luar Play Store).
         * Hasil: "izin" bila pengguna harus mengizinkan pemasangan dari aplikasi ini dulu, "unduh" bila mulai mengunduh.
         */
        @JavascriptInterface
        public String perbarui(String url) {
            if (Build.VERSION.SDK_INT >= 26 && !getPackageManager().canRequestPackageInstalls()) {
                startActivity(new Intent(Settings.ACTION_MANAGE_UNKNOWN_APP_SOURCES, Uri.parse("package:" + getPackageName())));
                return "izin";
            }
            unduhDanPasang(url);
            return "unduh";
        }
    }

    private void unduhDanPasang(String url) {
        File tujuan = new File(getExternalFilesDir(Environment.DIRECTORY_DOWNLOADS), "pos-warkop.apk");
        tujuan.delete(); // DownloadManager tidak menimpa file lama
        DownloadManager dm = (DownloadManager) getSystemService(DOWNLOAD_SERVICE);
        long id = dm.enqueue(new DownloadManager.Request(Uri.parse(url))
                .setTitle("Pembaruan POS Warkop")
                .setMimeType(APK)
                .setNotificationVisibility(DownloadManager.Request.VISIBILITY_VISIBLE_NOTIFY_COMPLETED)
                .setDestinationUri(Uri.fromFile(tujuan)));
        ContextCompat.registerReceiver(this, new BroadcastReceiver() {
            @Override
            public void onReceive(Context c, Intent i) {
                if (i.getLongExtra(DownloadManager.EXTRA_DOWNLOAD_ID, -1) != id) return;
                unregisterReceiver(this);
                if (!berhasil(dm, id)) return; // unduhan gagal; pengguna bisa menekan Perbarui lagi
                Uri apk = FileProvider.getUriForFile(MainActivity.this, getPackageName() + ".fileprovider", tujuan);
                startActivity(new Intent(Intent.ACTION_VIEW)
                        .setDataAndType(apk, APK)
                        .addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION | Intent.FLAG_ACTIVITY_NEW_TASK));
            }
        }, new IntentFilter(DownloadManager.ACTION_DOWNLOAD_COMPLETE), ContextCompat.RECEIVER_EXPORTED);
    }

    private static boolean berhasil(DownloadManager dm, long id) {
        try (Cursor c = dm.query(new DownloadManager.Query().setFilterById(id))) {
            return c.moveToFirst() && c.getInt(c.getColumnIndexOrThrow(DownloadManager.COLUMN_STATUS)) == DownloadManager.STATUS_SUCCESSFUL;
        }
    }

    private static final String APK = "application/vnd.android.package-archive";

    // tanpa ini Android bisa membekukan stasiun saat layar mati lama
    private void mintaBebasHematBaterai() {
        PowerManager pm = (PowerManager) getSystemService(POWER_SERVICE);
        if (pm.isIgnoringBatteryOptimizations(getPackageName())) return;
        try {
            startActivity(new Intent(Settings.ACTION_REQUEST_IGNORE_BATTERY_OPTIMIZATIONS, Uri.parse("package:" + getPackageName())));
        } catch (Exception ignored) {
        }
    }
}
