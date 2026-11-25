#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use std::{
    collections::{HashMap, HashSet},
    fs,
    io::{Read, Write},
    net::{TcpStream, ToSocketAddrs},
    path::PathBuf,
    sync::{Arc, Mutex},
    thread,
    time::{Duration, Instant},
};

use base64::{engine::general_purpose::STANDARD, Engine};
use serde::{Deserialize, Serialize};
use tauri::{
    menu::{Menu, MenuItem},
    tray::{MouseButton, MouseButtonState, TrayIconBuilder, TrayIconEvent},
    AppHandle, Manager, State, WebviewUrl, WebviewWindowBuilder, WindowEvent,
};
use tauri_plugin_autostart::{MacosLauncher, ManagerExt};

const JEDA_POLL: Duration = Duration::from_secs(2);
const JEDA_ULANG: Duration = Duration::from_secs(15);
const BATAS_LOG: usize = 30;
// Alamat server tertanam di aplikasi; pindah domain berarti build ulang installer.
const SERVER: &str = "https://poswarkop.142-91-103-131.sslip.io";

#[derive(Serialize, Deserialize, Default, Clone)]
struct Config {
    #[serde(default)]
    server: String,
    #[serde(default)]
    token: String,
    #[serde(default)]
    aktif: bool,
}

#[derive(Serialize, Default, Clone)]
struct Status {
    aktif: bool,
    terhubung: bool,
    pesan: String,
    log: Vec<String>,
}

struct Bersama {
    config: Mutex<Config>,
    status: Mutex<Status>,
    berkas: PathBuf,
}

#[derive(Deserialize)]
struct Amplop<T> {
    data: T,
}

#[derive(Deserialize)]
struct Pekerjaan {
    id: i64,
    ip: String,
    port: u16,
    #[serde(default)]
    data: Option<String>,
    #[serde(default)]
    payload: serde_json::Value,
}

type Bagi = Arc<Bersama>;

impl Bersama {
    fn simpan(&self) {
        if let Ok(c) = self.config.lock() {
            let _ = fs::write(&self.berkas, serde_json::to_string_pretty(&*c).unwrap_or_default());
        }
    }

    fn catat(&self, teks: String) {
        let mut s = self.status.lock().unwrap();
        s.log.insert(0, format!("{}  {}", chrono::Local::now().format("%H:%M:%S"), teks));
        s.log.truncate(BATAS_LOG);
    }
}

// ---- stasiun cetak: ambil tiket dari server, kirim ke printer, lapor ----

fn galat_http(e: ureq::Error) -> String {
    match e {
        ureq::Error::Status(kode, res) => {
            let isi: String = res.into_string().unwrap_or_default().chars().take(120).collect();
            format!("Server menjawab {kode}: {isi}")
        }
        ureq::Error::Transport(t) => t.to_string(),
    }
}

fn alamat_api(cfg: &Config, path: &str) -> String {
    format!("{}/api/stasiun{}", cfg.server.trim_end_matches('/'), path)
}

fn api_get<T: serde::de::DeserializeOwned>(cfg: &Config, path: &str) -> Result<T, String> {
    let res = ureq::get(&alamat_api(cfg, path))
        .set("X-Stasiun-Token", &cfg.token)
        .timeout(Duration::from_secs(10))
        .call()
        .map_err(galat_http)?;
    let amplop: Amplop<T> = res.into_json().map_err(|e| e.to_string())?;
    Ok(amplop.data)
}

fn api_post(cfg: &Config, path: &str, isi: Option<serde_json::Value>) -> Result<(), String> {
    let req = ureq::post(&alamat_api(cfg, path))
        .set("X-Stasiun-Token", &cfg.token)
        .timeout(Duration::from_secs(10));
    let hasil = match isi {
        Some(b) => req.send_json(b),
        None => req.call(),
    };
    hasil.map(|_| ()).map_err(galat_http)
}

fn kirim_ke_printer(ip: &str, port: u16, data: &[u8]) -> Result<(), String> {
    let alamat = (ip, port)
        .to_socket_addrs()
        .map_err(|e| format!("Alamat printer tidak valid: {e}"))?
        .next()
        .ok_or_else(|| "Alamat printer tidak ditemukan".to_string())?;
    let mut soket = TcpStream::connect_timeout(&alamat, Duration::from_secs(5))
        .map_err(|e| format!("Tidak bisa terhubung ke {ip}:{port} ({e})"))?;
    soket.set_write_timeout(Some(Duration::from_secs(5))).ok();
    cek_status(&mut soket)?;
    soket.write_all(data).map_err(|e| format!("Gagal mengirim ke printer: {e}"))?;
    soket.flush().map_err(|e| e.to_string())
}

// Printer tetap menerima data walau kertas habis/tutup terbuka, jadi tanya dulu statusnya (ESC/POS DLE EOT).
// Printer yang tidak menjawab dianggap siap, supaya printer murah tanpa fitur status tetap bisa mencetak.
fn cek_status(soket: &mut TcpStream) -> Result<(), String> {
    soket.set_read_timeout(Some(Duration::from_millis(800))).ok();
    let mut tanya = |n: u8| -> Option<u8> {
        soket.write_all(&[0x10, 0x04, n]).ok()?;
        let mut b = [0u8; 1];
        soket.read_exact(&mut b).ok()?;
        // byte status ESC/POS selalu berpola 0xx1 0x10 (bit 1 & 4 = 1, bit 0 & 7 = 0); selain itu bukan jawaban status
        (b[0] & 0x93 == 0x12).then_some(b[0])
    };
    if let Some(b) = tanya(2) {
        let galat = galat_offline(b);
        if !galat.is_empty() {
            return Err(galat.into());
        }
    }
    if tanya(4).is_some_and(|b| b & 0x60 == 0x60) {
        return Err("Kertas printer habis".into());
    }
    Ok(())
}

fn galat_offline(b: u8) -> &'static str {
    if b & 0x04 != 0 {
        "Tutup printer terbuka"
    } else if b & 0x20 != 0 {
        "Kertas printer habis"
    } else if b & 0x40 != 0 {
        "Printer error (cek kertas macet/pemotong)"
    } else {
        ""
    }
}

#[cfg(test)]
mod tes {
    use super::galat_offline;

    #[test]
    fn baca_status_offline() {
        assert_eq!(galat_offline(0x12), "");
        assert_eq!(galat_offline(0x16), "Tutup printer terbuka");
        assert_eq!(galat_offline(0x32), "Kertas printer habis");
        assert_eq!(galat_offline(0x52), "Printer error (cek kertas macet/pemotong)");
    }
}

fn ringkas(p: &serde_json::Value) -> String {
    let teks = |k: &str| p.get(k).and_then(|v| v.as_str()).unwrap_or("").to_string();
    format!("{} · {} · {}", teks("judul"), teks("tenant"), teks("mejaNama"))
}

fn proses(cfg: &Config, sh: &Bersama, job: &Pekerjaan, sudah: &mut HashSet<i64>) -> Result<(), String> {
    // sudah tercetak tetapi laporan ke server belum berhasil: jangan dicetak dua kali
    if !sudah.contains(&job.id) {
        let data = job
            .data
            .as_deref()
            .filter(|d| !d.is_empty())
            .ok_or_else(|| "Tugas tidak punya data cetak".to_string())?;
        let bait = STANDARD.decode(data).map_err(|e| format!("Data cetak rusak: {e}"))?;
        kirim_ke_printer(&job.ip, job.port, &bait)?;
        sudah.insert(job.id);
        sh.catat(format!("Tercetak: {}", ringkas(&job.payload)));
    }
    api_post(cfg, &format!("/jobs/{}/selesai", job.id), None)
}

fn putaran(cfg: &Config, sh: &Bersama, sudah: &mut HashSet<i64>, tunda: &mut HashMap<i64, Instant>) -> Result<(), String> {
    let jobs: Vec<Pekerjaan> = api_get(cfg, "/jobs")?;
    {
        let mut s = sh.status.lock().unwrap();
        s.aktif = true;
        s.terhubung = true;
        s.pesan = if jobs.is_empty() { "Siap".into() } else { format!("{} tiket menunggu", jobs.len()) };
    }
    for job in jobs {
        if tunda.get(&job.id).is_some_and(|t| *t > Instant::now()) {
            continue;
        }
        match proses(cfg, sh, &job, sudah) {
            Ok(()) => {
                sudah.remove(&job.id);
                tunda.remove(&job.id);
            }
            Err(e) => {
                tunda.insert(job.id, Instant::now() + JEDA_ULANG);
                sh.catat(format!("Gagal: {e}"));
                let galat: String = e.chars().take(250).collect();
                let _ = api_post(cfg, &format!("/jobs/{}/gagal", job.id), Some(serde_json::json!({ "galat": galat })));
            }
        }
    }
    Ok(())
}

fn jalankan_stasiun(sh: Bagi) {
    let mut sudah: HashSet<i64> = HashSet::new();
    let mut tunda: HashMap<i64, Instant> = HashMap::new();
    loop {
        let cfg = sh.config.lock().unwrap().clone();
        if cfg.aktif && !cfg.token.is_empty() && !cfg.server.is_empty() {
            if let Err(e) = putaran(&cfg, &sh, &mut sudah, &mut tunda) {
                let mut s = sh.status.lock().unwrap();
                s.aktif = true;
                s.terhubung = false;
                s.pesan = format!("Tidak tersambung ke server: {e}");
            }
        } else {
            let mut s = sh.status.lock().unwrap();
            s.aktif = false;
            s.terhubung = false;
            s.pesan = "Mati".into();
        }
        thread::sleep(JEDA_POLL);
    }
}

// ---- perintah untuk halaman lokal (launcher dan pengaturan stasiun) ----

#[tauri::command]
async fn ambil_config(sh: State<'_, Bagi>) -> Result<serde_json::Value, String> {
    let c = sh.config.lock().unwrap();
    Ok(serde_json::json!({ "server": c.server, "aktif": c.aktif, "adaToken": !c.token.is_empty() }))
}

#[tauri::command]
async fn simpan_stasiun(app: AppHandle, sh: State<'_, Bagi>, aktif: bool, token: String) -> Result<(), String> {
    {
        let mut c = sh.config.lock().unwrap();
        c.aktif = aktif;
        if !token.is_empty() {
            c.token = token;
        }
    }
    sh.simpan();
    // stasiun harus menyala sendiri saat PC dihidupkan
    let _ = if aktif { app.autolaunch().enable() } else { app.autolaunch().disable() };
    Ok(())
}

#[tauri::command]
async fn status_stasiun(sh: State<'_, Bagi>) -> Result<Status, String> {
    Ok(sh.status.lock().unwrap().clone())
}

// ---- jendela dan tray ----

fn tampilkan_utama(app: &AppHandle) {
    if let Some(w) = app.get_webview_window("main") {
        let _ = w.show();
        let _ = w.unminimize();
        let _ = w.set_focus();
    }
}

fn buka_stasiun(app: &AppHandle) {
    if let Some(w) = app.get_webview_window("stasiun") {
        let _ = w.show();
        let _ = w.set_focus();
        return;
    }
    let _ = WebviewWindowBuilder::new(app, "stasiun", WebviewUrl::App("stasiun.html".into()))
        .title("Stasiun cetak")
        .inner_size(560.0, 640.0)
        .build();
}

fn pasang_tray(app: &AppHandle) -> tauri::Result<()> {
    let buka = MenuItem::with_id(app, "buka", "Buka POS Warkop", true, None::<&str>)?;
    let stasiun = MenuItem::with_id(app, "stasiun", "Stasiun cetak…", true, None::<&str>)?;
    let keluar = MenuItem::with_id(app, "keluar", "Keluar", true, None::<&str>)?;
    let menu = Menu::with_items(app, &[&buka, &stasiun, &keluar])?;
    TrayIconBuilder::new()
        .icon(app.default_window_icon().unwrap().clone())
        .tooltip("POS Warkop")
        .menu(&menu)
        .on_menu_event(|app, event| match event.id.as_ref() {
            "buka" => tampilkan_utama(app),
            "stasiun" => buka_stasiun(app),
            "keluar" => app.exit(0),
            _ => {}
        })
        .on_tray_icon_event(|tray, event| {
            if let TrayIconEvent::Click { button: MouseButton::Left, button_state: MouseButtonState::Up, .. } = event {
                tampilkan_utama(tray.app_handle());
            }
        })
        .build(app)?;
    Ok(())
}

fn main() {
    tauri::Builder::default()
        .plugin(tauri_plugin_single_instance::init(|app, _argv, _cwd| tampilkan_utama(app)))
        .plugin(tauri_plugin_autostart::init(MacosLauncher::LaunchAgent, Some(vec!["--sembunyi"])))
        .invoke_handler(tauri::generate_handler![ambil_config, simpan_stasiun, status_stasiun])
        .on_window_event(|window, event| {
            // menutup jendela utama hanya menyembunyikannya, supaya stasiun cetak tetap hidup di tray
            if let WindowEvent::CloseRequested { api, .. } = event {
                if window.label() == "main" {
                    api.prevent_close();
                    let _ = window.hide();
                }
            }
        })
        .setup(|app| {
            let dir = app.path().app_config_dir()?;
            fs::create_dir_all(&dir).ok();
            let berkas = dir.join("config.json");
            let mut config: Config = fs::read_to_string(&berkas)
                .ok()
                .and_then(|t| serde_json::from_str(&t).ok())
                .unwrap_or_default();
            config.server = SERVER.into(); // menimpa alamat lama yang mungkin tersimpan di config.json
            let sh: Bagi = Arc::new(Bersama {
                config: Mutex::new(config),
                status: Mutex::new(Status { pesan: "Mati".into(), ..Default::default() }),
                berkas,
            });
            app.manage(sh.clone());

            let sembunyi = std::env::args().any(|a| a == "--sembunyi");
            WebviewWindowBuilder::new(app, "main", WebviewUrl::App("index.html".into()))
                .title("POS Warkop")
                .inner_size(1366.0, 820.0)
                .visible(!sembunyi)
                .build()?;

            pasang_tray(app.handle())?;
            thread::spawn(move || jalankan_stasiun(sh));
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("gagal menjalankan aplikasi POS Warkop");
}
