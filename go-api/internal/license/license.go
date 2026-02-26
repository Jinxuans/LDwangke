package license

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"go-api/internal/config"
)

// ===== 密钥（由宝塔插件部署时生成 .secrets 文件） =====

type secrets struct {
	ClientSecret string `json:"client_secret"` // 与授权站 config.toml 中 client_secret 一致
	AESKey       string `json:"aes_key"`       // 离线缓存加密密钥（hex，32字节=64位hex）
	HMACKey      string `json:"hmac_key"`      // 离线缓存签名密钥（hex）
}

var (
	clientSecret string
	cacheAESKey  []byte
	cacheHMACKey []byte
)

// loadSecrets 从 .secrets 文件加载密钥（宝塔插件部署时生成）
func loadSecrets(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取密钥文件失败: %v", err)
	}

	var s secrets
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("解析密钥文件失败: %v", err)
	}

	if s.ClientSecret == "" {
		return fmt.Errorf("密钥文件缺少 client_secret")
	}
	clientSecret = s.ClientSecret

	// AES 密钥：从 hex 解码，确保 32 字节
	if s.AESKey != "" {
		aesBytes, err := hex.DecodeString(s.AESKey)
		if err != nil {
			return fmt.Errorf("AES 密钥 hex 解码失败: %v", err)
		}
		if len(aesBytes) < 32 {
			padded := make([]byte, 32)
			copy(padded, aesBytes)
			cacheAESKey = padded
		} else {
			cacheAESKey = aesBytes[:32]
		}
	} else {
		return fmt.Errorf("密钥文件缺少 aes_key")
	}

	// HMAC 密钥：从 hex 解码
	if s.HMACKey != "" {
		hmacBytes, err := hex.DecodeString(s.HMACKey)
		if err != nil {
			return fmt.Errorf("HMAC 密钥 hex 解码失败: %v", err)
		}
		cacheHMACKey = hmacBytes
	} else {
		return fmt.Errorf("密钥文件缺少 hmac_key")
	}

	return nil
}

// ===== 授权状态 =====

type Status int

const (
	StatusOK       Status = iota // 授权正常
	StatusOffline                // 离线模式（缓存有效）
	StatusWarning                // 警告模式（离线超24h）
	StatusDegraded               // 功能降级（离线超72h 或授权无效）
)

func (s Status) String() string {
	switch s {
	case StatusOK:
		return "正常"
	case StatusOffline:
		return "离线模式"
	case StatusWarning:
		return "离线警告"
	case StatusDegraded:
		return "功能降级"
	default:
		return "未知"
	}
}

// ===== 授权站地址（硬编码，不可通过配置修改） =====

const licenseServerURL = "https://qingka.top"

// ===== 缓存数据结构 =====

type cacheData struct {
	LicenseKey string `json:"k"`
	MachineID  string `json:"m"`
	VerifyTime int64  `json:"t"`
	ExpireAt   string `json:"e"`
	Plan       string `json:"p"`
	MaxUsers   int    `json:"u"`
	MaxAgents  int    `json:"a"`
	Sign       string `json:"s"`
}

// ===== 授权站 API 响应 =====

type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type verifyData struct {
	Valid     bool   `json:"valid"`
	Plan      string `json:"plan"`
	ExpireAt  string `json:"expire_at"`
	MaxUsers  int    `json:"max_users"`
	MaxAgents int    `json:"max_agents"`
	Message   string `json:"message"`
}

// ===== Manager =====

type Manager struct {
	mu             sync.RWMutex
	status         Status
	lastVerifyTime time.Time
	licenseKey     string
	machineID      string
	domain         string
	plan           string
	expireAt       string
	maxUsers       int
	maxAgents      int
	cacheFile      string
	failCount      int
	lastError      string
	stopCh         chan struct{}
}

var Global *Manager

// NewManager 创建授权管理器
func NewManager(cfg config.LicenseConfig) *Manager {
	m := &Manager{
		status:    StatusDegraded,
		domain:    cfg.Domain,
		cacheFile: cfg.CacheFile,
		stopCh:    make(chan struct{}),
	}

	if m.cacheFile == "" {
		m.cacheFile = ".sys_state"
	}

	// 加载密钥文件（宝塔插件部署时生成）
	secretsFile := cfg.SecretsFile
	if secretsFile == "" {
		secretsFile = ".secrets"
	}
	if err := loadSecrets(secretsFile); err != nil {
		log.Printf("[License] 加载密钥文件失败: %v（授权功能不可用）", err)
		m.lastError = fmt.Sprintf("密钥文件加载失败: %v", err)
	}

	// 读取授权码：优先配置文件中的 license_key，其次从 key_file 读
	m.licenseKey = strings.TrimSpace(cfg.LicenseKey)
	if m.licenseKey == "" && cfg.KeyFile != "" {
		if data, err := os.ReadFile(cfg.KeyFile); err == nil {
			m.licenseKey = strings.TrimSpace(string(data))
		}
	}

	// 读取机器码
	m.machineID = readMachineID()

	return m
}

// Start 启动授权验证（启动验证 + 后台心跳）
func (m *Manager) Start() {
	if m.licenseKey == "" {
		log.Println("[License] 未配置授权码，系统将以降级模式运行")
		m.mu.Lock()
		m.status = StatusDegraded
		m.lastError = "未配置授权码"
		m.mu.Unlock()
		return
	}

	keyPreview := m.licenseKey
	if len(keyPreview) > 8 {
		keyPreview = keyPreview[:8]
	}
	midPreview := m.machineID
	if len(midPreview) > 8 {
		midPreview = midPreview[:8]
	}
	log.Printf("[License] 授权码: %s****，机器码: %s...", keyPreview, midPreview)

	// 启动时验证
	if err := m.doVerify(); err != nil {
		log.Printf("[License] 在线验证失败: %v，尝试离线缓存...", err)
		if cacheErr := m.loadCache(); cacheErr != nil {
			log.Printf("[License] 离线缓存也无效: %v，降级运行", cacheErr)
			m.mu.Lock()
			m.status = StatusDegraded
			m.lastError = fmt.Sprintf("在线: %v / 离线: %v", err, cacheErr)
			m.mu.Unlock()
		}
	}

	// 后台心跳协程
	go m.heartbeatLoop()

	log.Printf("[License] 授权状态: %s", m.GetStatus())
}

// Stop 停止后台心跳
func (m *Manager) Stop() {
	close(m.stopCh)
}

// GetStatus 获取当前授权状态
func (m *Manager) GetStatus() Status {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

// GetStatusInfo 获取授权详情（供管理后台展示）
func (m *Manager) GetStatusInfo() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	info := map[string]interface{}{
		"status":      m.status.String(),
		"status_code": int(m.status),
		"license_key": maskKey(m.licenseKey),
		"plan":        m.plan,
		"expire_at":   m.expireAt,
		"max_users":   m.maxUsers,
		"max_agents":  m.maxAgents,
		"last_verify": m.lastVerifyTime.Format("2006-01-02 15:04:05"),
		"fail_count":  m.failCount,
		"last_error":  m.lastError,
	}
	if !m.lastVerifyTime.IsZero() {
		info["offline_duration"] = time.Since(m.lastVerifyTime).String()
	}
	return info
}

// IsAllowed 检查当前授权是否允许正常使用
func (m *Manager) IsAllowed() bool {
	s := m.GetStatus()
	return s == StatusOK || s == StatusOffline || s == StatusWarning
}

// ===== 在线验证 =====

func (m *Manager) doVerify() error {
	ts := time.Now().Unix()
	signStr := fmt.Sprintf("domain=%s&license_key=%s&machine_id=%s&timestamp=%d&version=",
		m.domain, m.licenseKey, m.machineID, ts)
	sign := hmacSHA256(clientSecret, signStr)

	body := map[string]interface{}{
		"license_key": m.licenseKey,
		"domain":      m.domain,
		"machine_id":  m.machineID,
		"version":     "",
		"timestamp":   ts,
		"sign":        sign,
	}

	respData, err := m.callAPI("/api/v1/license/verify", body)
	if err != nil {
		return err
	}

	var vd verifyData
	if err := json.Unmarshal(respData, &vd); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if !vd.Valid {
		m.mu.Lock()
		m.status = StatusDegraded
		m.lastError = vd.Message
		m.mu.Unlock()
		return fmt.Errorf("授权无效: %s", vd.Message)
	}

	now := time.Now()
	m.mu.Lock()
	m.status = StatusOK
	m.lastVerifyTime = now
	m.plan = vd.Plan
	m.expireAt = vd.ExpireAt
	m.maxUsers = vd.MaxUsers
	m.maxAgents = vd.MaxAgents
	m.failCount = 0
	m.lastError = ""
	m.mu.Unlock()

	m.saveCache(now)

	log.Printf("[License] 在线验证成功，套餐: %s，到期: %s", vd.Plan, vd.ExpireAt)
	return nil
}

// ===== 心跳 =====

func (m *Manager) doHeartbeat() error {
	ts := time.Now().Unix()
	signStr := fmt.Sprintf("license_key=%s&machine_id=%s&timestamp=%d&version=",
		m.licenseKey, m.machineID, ts)
	sign := hmacSHA256(clientSecret, signStr)

	body := map[string]interface{}{
		"license_key": m.licenseKey,
		"machine_id":  m.machineID,
		"domain":      m.domain,
		"version":     "",
		"timestamp":   ts,
		"sign":        sign,
	}

	_, err := m.callAPI("/api/v1/license/heartbeat", body)
	return err
}

func (m *Manager) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			if m.licenseKey == "" {
				continue
			}

			if err := m.doVerify(); err != nil {
				if hbErr := m.doHeartbeat(); hbErr != nil {
					m.mu.Lock()
					m.failCount++
					m.lastError = fmt.Sprintf("verify: %v / heartbeat: %v", err, hbErr)
					m.mu.Unlock()
					log.Printf("[License] 心跳失败(第%d次): %v", m.failCount, hbErr)
				} else {
					m.mu.Lock()
					m.failCount = 0
					m.mu.Unlock()
				}
			}

			m.updateOfflineStatus()
		}
	}
}

func (m *Manager) updateOfflineStatus() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status == StatusOK {
		return
	}

	if m.lastVerifyTime.IsZero() {
		m.status = StatusDegraded
		return
	}

	elapsed := time.Since(m.lastVerifyTime)

	switch {
	case elapsed < 24*time.Hour:
		if m.status != StatusOK {
			m.status = StatusOffline
		}
	case elapsed < 72*time.Hour:
		m.status = StatusWarning
		log.Printf("[License] 离线警告：已 %s 未与授权站通信", elapsed.Round(time.Minute))
	default:
		m.status = StatusDegraded
		m.lastError = fmt.Sprintf("离线超过72小时 (%s)", elapsed.Round(time.Minute))
		log.Printf("[License] 功能降级：已 %s 未与授权站通信", elapsed.Round(time.Minute))
	}
}

// ===== HTTP 请求 =====

func (m *Manager) callAPI(path string, body map[string]interface{}) (json.RawMessage, error) {
	jsonBody, _ := json.Marshal(body)
	url := licenseServerURL + path

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "QingkaGoAPI/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("网络请求失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var apiResp apiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		truncated := string(respBody)
		if len(truncated) > 200 {
			truncated = truncated[:200]
		}
		return nil, fmt.Errorf("解析响应失败: %v (body: %s)", err, truncated)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("授权站返回错误(%d): %s", apiResp.Code, apiResp.Message)
	}

	return apiResp.Data, nil
}

// ===== 离线缓存 =====

func (m *Manager) saveCache(verifyTime time.Time) {
	c := cacheData{
		LicenseKey: m.licenseKey,
		MachineID:  m.machineID,
		VerifyTime: verifyTime.Unix(),
		ExpireAt:   m.expireAt,
		Plan:       m.plan,
		MaxUsers:   m.maxUsers,
		MaxAgents:  m.maxAgents,
	}

	dataForSign, _ := json.Marshal(map[string]interface{}{
		"k": c.LicenseKey, "m": c.MachineID, "t": c.VerifyTime,
		"e": c.ExpireAt, "p": c.Plan, "u": c.MaxUsers, "a": c.MaxAgents,
	})
	c.Sign = hmacSHA256(string(cacheHMACKey), string(dataForSign))

	plaintext, _ := json.Marshal(c)

	encrypted, err := aesGCMEncrypt(cacheAESKey, plaintext)
	if err != nil {
		log.Printf("[License] 缓存加密失败: %v", err)
		return
	}

	if err := os.WriteFile(m.cacheFile, encrypted, 0600); err != nil {
		log.Printf("[License] 缓存写入失败: %v", err)
	}
}

func (m *Manager) loadCache() error {
	encrypted, err := os.ReadFile(m.cacheFile)
	if err != nil {
		return fmt.Errorf("读取缓存文件失败: %v", err)
	}

	plaintext, err := aesGCMDecrypt(cacheAESKey, encrypted)
	if err != nil {
		return fmt.Errorf("缓存解密失败: %v", err)
	}

	var c cacheData
	if err := json.Unmarshal(plaintext, &c); err != nil {
		return fmt.Errorf("缓存解析失败: %v", err)
	}

	dataForSign, _ := json.Marshal(map[string]interface{}{
		"k": c.LicenseKey, "m": c.MachineID, "t": c.VerifyTime,
		"e": c.ExpireAt, "p": c.Plan, "u": c.MaxUsers, "a": c.MaxAgents,
	})
	expectedSign := hmacSHA256(string(cacheHMACKey), string(dataForSign))
	if c.Sign != expectedSign {
		return fmt.Errorf("缓存签名验证失败（文件被篡改）")
	}

	if c.LicenseKey != m.licenseKey {
		return fmt.Errorf("缓存授权码不匹配")
	}
	if c.MachineID != m.machineID {
		return fmt.Errorf("缓存机器码不匹配")
	}

	verifyTime := time.Unix(c.VerifyTime, 0)
	elapsed := time.Since(verifyTime)
	if elapsed > 72*time.Hour {
		return fmt.Errorf("缓存已过期（%s 前验证）", elapsed.Round(time.Minute))
	}

	m.mu.Lock()
	m.lastVerifyTime = verifyTime
	m.plan = c.Plan
	m.expireAt = c.ExpireAt
	m.maxUsers = c.MaxUsers
	m.maxAgents = c.MaxAgents

	if elapsed < 24*time.Hour {
		m.status = StatusOffline
	} else {
		m.status = StatusWarning
	}
	m.lastError = ""
	m.mu.Unlock()

	log.Printf("[License] 离线缓存有效（%s 前验证），状态: %s", elapsed.Round(time.Minute), m.status)
	return nil
}

// ===== 工具函数 =====

func readMachineID() string {
	if data, err := os.ReadFile("/etc/machine-id"); err == nil {
		mid := strings.TrimSpace(string(data))
		if mid != "" {
			return mid
		}
	}
	if data, err := os.ReadFile("/var/lib/dbus/machine-id"); err == nil {
		mid := strings.TrimSpace(string(data))
		if mid != "" {
			return mid
		}
	}
	if name, err := os.Hostname(); err == nil {
		h := sha256.Sum256([]byte(name))
		return hex.EncodeToString(h[:16])
	}
	return "unknown"
}

func hmacSHA256(secret, data string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

func aesGCMEncrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func aesGCMDecrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("密文太短")
	}
	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ct, nil)
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:8] + "****"
}
