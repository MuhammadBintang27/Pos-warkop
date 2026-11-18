package services

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

// Pusher mengirim notifikasi FCM ke HP akun tenant. Tanpa service account, semua pengiriman dilewati.
type Pusher struct {
	db   *sql.DB
	akun *akunFCM
	http *http.Client

	mu      sync.Mutex
	akses   string
	kadaluw time.Time
}

type akunFCM struct {
	ProjectID   string `json:"project_id"`
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	TokenURI    string `json:"token_uri"`
}

type PesanPush struct {
	Judul, Isi string
}

func NewPusher(db *sql.DB, serviceAccountJSON string) *Pusher {
	p := &Pusher{db: db, http: &http.Client{Timeout: 15 * time.Second}}
	if strings.TrimSpace(serviceAccountJSON) == "" {
		log.Warn().Msg("FCM_SERVICE_ACCOUNT kosong: push notifikasi tenant dimatikan")
		return p
	}
	var a akunFCM
	if err := json.Unmarshal([]byte(serviceAccountJSON), &a); err != nil || a.ProjectID == "" || a.PrivateKey == "" {
		log.Error().Err(err).Msg("FCM_SERVICE_ACCOUNT tidak valid: push notifikasi tenant dimatikan")
		return p
	}
	if a.TokenURI == "" {
		a.TokenURI = "https://oauth2.googleapis.com/token"
	}
	p.akun = &a
	return p
}

// SimpanPerangkat mengikat token HP ke akun tenant; token yang sama dipindah bila HP dipakai akun lain.
func (p *Pusher) SimpanPerangkat(kid, tenantID, userID int64, token string) error {
	_, err := p.db.Exec(`INSERT INTO perangkat_push (klien_id,tenant_id,user_id,token,dibuat_at) VALUES (?,?,?,?,?)
		ON DUPLICATE KEY UPDATE klien_id=VALUES(klien_id), tenant_id=VALUES(tenant_id), user_id=VALUES(user_id), dibuat_at=VALUES(dibuat_at)`,
		kid, tenantID, userID, token, time.Now())
	return err
}

func (p *Pusher) HapusPerangkat(kid int64, token string) error {
	_, err := p.db.Exec("DELETE FROM perangkat_push WHERE klien_id=? AND token=?", kid, token)
	return err
}

// KeTenant dijalankan di goroutine sendiri supaya Kirim tidak menunggu Google.
func (p *Pusher) KeTenant(kid, tenantID int64, pesan PesanPush) {
	if p == nil || p.akun == nil {
		return
	}
	go func() {
		rows, err := p.db.Query("SELECT token FROM perangkat_push WHERE klien_id=? AND tenant_id=?", kid, tenantID)
		if err != nil {
			log.Error().Err(err).Msg("push: gagal membaca perangkat")
			return
		}
		var tokens []string
		for rows.Next() {
			var t string
			if rows.Scan(&t) == nil {
				tokens = append(tokens, t)
			}
		}
		rows.Close()
		for _, t := range tokens {
			if err := p.kirim(t, pesan); err != nil {
				log.Warn().Err(err).Int64("tenant", tenantID).Msg("push gagal")
			}
		}
	}()
}

func (p *Pusher) kirim(token string, pesan PesanPush) error {
	akses, err := p.tokenAkses()
	if err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]any{"message": map[string]any{
		"token":        token,
		"notification": map[string]string{"title": pesan.Judul, "body": pesan.Isi},
		"android": map[string]any{
			"priority":     "high",
			"notification": map[string]string{"channel_id": "pesanan", "sound": "default"},
		},
	}})
	req, _ := http.NewRequest(http.MethodPost, "https://fcm.googleapis.com/v1/projects/"+p.akun.ProjectID+"/messages:send", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+akses)
	req.Header.Set("Content-Type", "application/json")
	res, err := p.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		return nil
	}
	isi, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
	// aplikasi dicopot atau token kedaluwarsa: buang supaya tidak dicoba terus
	if res.StatusCode == http.StatusNotFound || strings.Contains(string(isi), "UNREGISTERED") {
		_, _ = p.db.Exec("DELETE FROM perangkat_push WHERE token=?", token)
	}
	return fmt.Errorf("fcm %d: %s", res.StatusCode, isi)
}

// tokenAkses menukar JWT bertanda tangan service account dengan access token OAuth2 (berlaku 1 jam, disimpan).
func (p *Pusher) tokenAkses() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.akses != "" && time.Now().Before(p.kadaluw) {
		return p.akses, nil
	}
	kunci, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(p.akun.PrivateKey))
	if err != nil {
		return "", fmt.Errorf("private_key service account: %w", err)
	}
	now := time.Now()
	assertion, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":   p.akun.ClientEmail,
		"scope": "https://www.googleapis.com/auth/firebase.messaging",
		"aud":   p.akun.TokenURI,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Hour).Unix(),
	}).SignedString(kunci)
	if err != nil {
		return "", err
	}
	res, err := p.http.PostForm(p.akun.TokenURI, url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {assertion},
	})
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil || out.AccessToken == "" {
		return "", fmt.Errorf("gagal mendapat access token FCM (status %d)", res.StatusCode)
	}
	p.akses, p.kadaluw = out.AccessToken, now.Add(time.Duration(out.ExpiresIn-60)*time.Second)
	return p.akses, nil
}
