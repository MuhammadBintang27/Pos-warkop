package id.warkop.pos;

import android.content.Intent;
import android.content.SharedPreferences;
import android.net.Uri;
import android.os.Bundle;
import android.os.PowerManager;
import android.provider.Settings;
import android.webkit.JavascriptInterface;

import com.getcapacitor.BridgeActivity;

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
    }

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
