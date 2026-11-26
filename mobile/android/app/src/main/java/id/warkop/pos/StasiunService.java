package id.warkop.pos;

import android.app.Notification;
import android.app.NotificationManager;
import android.app.PendingIntent;
import android.app.Service;
import android.content.Context;
import android.content.Intent;
import android.content.SharedPreferences;
import android.content.pm.ServiceInfo;
import android.net.wifi.WifiManager;
import android.os.Build;
import android.os.IBinder;
import android.os.PowerManager;
import android.util.Base64;

import androidx.core.app.NotificationChannelCompat;
import androidx.core.app.NotificationCompat;
import androidx.core.app.NotificationManagerCompat;
import androidx.core.app.ServiceCompat;
import androidx.core.content.ContextCompat;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.ByteArrayOutputStream;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.InetSocketAddress;
import java.net.Socket;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
import java.util.Set;

/**
 * Stasiun cetak di HP/tablet kasir: sama dengan stasiun di aplikasi desktop (tauri/src-tauri/src/main.rs).
 * Ambil tiket dari server, kirim byte ESC/POS ke printer LAN, lalu lapor selesai/gagal.
 */
public class StasiunService extends Service {
    private static final String CHANNEL = "stasiun";
    private static final int NOTIF_ID = 9100;
    private static final long JEDA_POLL = 2000;
    private static final long JEDA_ULANG = 15000;

    static volatile String pesan = "Mati";

    private volatile boolean jalan;
    private Thread thread;
    private PowerManager.WakeLock wake;
    private WifiManager.WifiLock wifi;

    static SharedPreferences prefs(Context c) {
        return c.getSharedPreferences("stasiun", MODE_PRIVATE);
    }

    /** Menyalakan atau mematikan service sesuai pengaturan tersimpan. */
    static void terapkan(Context c) {
        Intent i = new Intent(c, StasiunService.class);
        SharedPreferences p = prefs(c);
        if (p.getBoolean("aktif", false) && !p.getString("token", "").isEmpty()) {
            if ("Mati".equals(pesan)) pesan = "Menyambung…";
            ContextCompat.startForegroundService(c, i);
        } else {
            c.stopService(i);
            pesan = "Mati";
        }
    }

    @Override
    public void onCreate() {
        super.onCreate();
        NotificationManagerCompat.from(this).createNotificationChannel(
                new NotificationChannelCompat.Builder(CHANNEL, NotificationManagerCompat.IMPORTANCE_LOW).setName("Stasiun cetak").build());
        int tipe = Build.VERSION.SDK_INT >= 29 ? ServiceInfo.FOREGROUND_SERVICE_TYPE_CONNECTED_DEVICE : 0;
        ServiceCompat.startForeground(this, NOTIF_ID, notif("Menyambung…"), tipe);

        // layar mati tidak boleh menghentikan cetak
        wake = ((PowerManager) getSystemService(POWER_SERVICE)).newWakeLock(PowerManager.PARTIAL_WAKE_LOCK, "warkop:stasiun");
        wake.acquire();
        wifi = ((WifiManager) getApplicationContext().getSystemService(WIFI_SERVICE)).createWifiLock(WifiManager.WIFI_MODE_FULL_HIGH_PERF, "warkop:stasiun");
        wifi.acquire();

        jalan = true;
        thread = new Thread(this::loop, "stasiun-cetak");
        thread.start();
    }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        return START_STICKY;
    }

    @Override
    public void onDestroy() {
        jalan = false;
        if (thread != null) thread.interrupt();
        if (wake != null && wake.isHeld()) wake.release();
        if (wifi != null && wifi.isHeld()) wifi.release();
        pesan = "Mati";
        super.onDestroy();
    }

    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }

    private Notification notif(String teks) {
        PendingIntent buka = PendingIntent.getActivity(this, 0, new Intent(this, MainActivity.class), PendingIntent.FLAG_IMMUTABLE);
        return new NotificationCompat.Builder(this, CHANNEL)
                .setSmallIcon(android.R.drawable.stat_sys_upload_done)
                .setContentTitle("Stasiun cetak aktif")
                .setContentText(teks)
                .setContentIntent(buka)
                .setOngoing(true)
                .build();
    }

    private void setPesan(String teks) {
        if (teks.equals(pesan)) return;
        pesan = teks;
        ((NotificationManager) getSystemService(NOTIFICATION_SERVICE)).notify(NOTIF_ID, notif(teks));
    }

    // ---- loop stasiun ----

    private void loop() {
        Set<Long> sudah = new HashSet<>();
        Map<Long, Long> tunda = new HashMap<>();
        while (jalan) {
            SharedPreferences p = prefs(this);
            String server = p.getString("server", ""), token = p.getString("token", "");
            try {
                putaran(server, token, sudah, tunda);
            } catch (Exception e) {
                setPesan("Tidak tersambung ke server: " + e.getMessage());
            }
            try {
                Thread.sleep(JEDA_POLL);
            } catch (InterruptedException e) {
                return;
            }
        }
    }

    private void putaran(String server, String token, Set<Long> sudah, Map<Long, Long> tunda) throws Exception {
        JSONArray jobs = new JSONObject(http("GET", server, token, "/jobs", null)).getJSONArray("data");
        setPesan(jobs.length() == 0 ? "Siap" : jobs.length() + " tiket menunggu");
        for (int i = 0; i < jobs.length(); i++) {
            JSONObject job = jobs.getJSONObject(i);
            long id = job.getLong("id");
            Long sampai = tunda.get(id);
            if (sampai != null && sampai > System.currentTimeMillis()) continue;
            try {
                // sudah tercetak tetapi laporan ke server belum berhasil: jangan dicetak dua kali
                if (!sudah.contains(id)) {
                    String data = job.optString("data", "");
                    if (data.isEmpty() || job.isNull("data")) throw new Exception("Tugas tidak punya data cetak");
                    kirimKePrinter(job.getString("ip"), job.getInt("port"), Base64.decode(data, Base64.DEFAULT));
                    sudah.add(id);
                }
                http("POST", server, token, "/jobs/" + id + "/selesai", null);
                sudah.remove(id);
                tunda.remove(id);
            } catch (Exception e) {
                tunda.put(id, System.currentTimeMillis() + JEDA_ULANG);
                String galat = String.valueOf(e.getMessage());
                if (galat.length() > 250) galat = galat.substring(0, 250);
                try {
                    http("POST", server, token, "/jobs/" + id + "/gagal", new JSONObject().put("galat", galat).toString());
                } catch (Exception ignored) {
                }
            }
        }
    }

    static void kirimKePrinter(String ip, int port, byte[] data) throws Exception {
        try (Socket s = new Socket()) {
            try {
                s.connect(new InetSocketAddress(ip, port), 5000);
            } catch (Exception e) {
                throw new Exception("Tidak bisa terhubung ke " + ip + ":" + port + " (" + e.getMessage() + ")");
            }
            OutputStream out = s.getOutputStream();
            cekStatus(s, out);
            s.setSoTimeout(5000);
            out.write(data);
            out.flush();
        }
    }

    /**
     * Printer tetap menerima data walau kertas habis/tutup terbuka, jadi tanya dulu statusnya (ESC/POS DLE EOT).
     * Printer yang tidak menjawab dianggap siap, supaya printer tanpa fitur status tetap bisa mencetak.
     * Sama dengan cek_status di tauri/src-tauri/src/main.rs.
     */
    private static void cekStatus(Socket s, OutputStream out) throws Exception {
        s.setSoTimeout(800);
        InputStream in = s.getInputStream();
        int b = tanya(out, in, 2);
        if (b >= 0) {
            if ((b & 0x04) != 0) throw new Exception("Tutup printer terbuka");
            if ((b & 0x20) != 0) throw new Exception("Kertas printer habis");
            if ((b & 0x40) != 0) throw new Exception("Printer error (cek kertas macet/pemotong)");
        }
        b = tanya(out, in, 4);
        if (b >= 0 && (b & 0x60) == 0x60) throw new Exception("Kertas printer habis");
    }

    /** Mengembalikan byte status, atau -1 bila printer tidak menjawab/jawabannya bukan byte status. */
    private static int tanya(OutputStream out, InputStream in, int n) {
        try {
            out.write(new byte[]{0x10, 0x04, (byte) n});
            out.flush();
            int b = in.read();
            // byte status ESC/POS selalu berpola 0xx1 0x10 (bit 1 & 4 = 1, bit 0 & 7 = 0)
            return b >= 0 && (b & 0x93) == 0x12 ? b : -1;
        } catch (Exception e) {
            return -1;
        }
    }

    private static String http(String metode, String server, String token, String path, String json) throws Exception {
        HttpURLConnection c = (HttpURLConnection) new URL(server.replaceAll("/+$", "") + "/api/stasiun" + path).openConnection();
        try {
            c.setRequestMethod(metode);
            c.setConnectTimeout(10000);
            c.setReadTimeout(10000);
            c.setRequestProperty("X-Stasiun-Token", token);
            if (json != null) {
                c.setDoOutput(true);
                c.setRequestProperty("Content-Type", "application/json");
                c.getOutputStream().write(json.getBytes(StandardCharsets.UTF_8));
            }
            int kode = c.getResponseCode();
            InputStream in = kode < 400 ? c.getInputStream() : c.getErrorStream();
            String isi = baca(in);
            if (kode >= 400) throw new Exception("Server menjawab " + kode + ": " + (isi.length() > 120 ? isi.substring(0, 120) : isi));
            return isi;
        } finally {
            c.disconnect();
        }
    }

    private static String baca(InputStream in) throws Exception {
        if (in == null) return "";
        try (InputStream i = in) {
            ByteArrayOutputStream b = new ByteArrayOutputStream();
            byte[] buf = new byte[4096];
            for (int n; (n = i.read(buf)) > 0; ) b.write(buf, 0, n);
            return b.toString("UTF-8");
        }
    }
}
