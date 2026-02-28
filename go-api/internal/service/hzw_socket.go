package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go-api/internal/database"

	"github.com/gorilla/websocket"
)

// HZWSocketClient 是 HZW 平台实时进度推送的 Socket.IO v4 客户端
type HZWSocketClient struct {
	serverURL string // e.g. "http://socket.biedawo.org"

	mu          sync.RWMutex
	supplierMap map[int]string // hid -> HZW平台UID (huoyuan.user)
	conn        *websocket.Conn
	running     int32

	// 统计
	updateCount int64
	skipCount   int64
}

// HZWSocketMessage 是 HZW 推送的订单状态消息
type HZWSocketMessage struct {
	UID           interface{} `json:"uid"`
	OID           interface{} `json:"oid"`
	Status        string      `json:"status"`
	Process       string      `json:"process"`
	Remarks       string      `json:"remarks"`
	ExamStartTime string      `json:"examStartTime"`
	ExamEndTime   string      `json:"examEndTime"`
}

var GlobalHZWSocket *HZWSocketClient

// NewHZWSocketClient 创建客户端实例
func NewHZWSocketClient(serverURL string) *HZWSocketClient {
	return &HZWSocketClient{
		serverURL:   serverURL,
		supplierMap: make(map[int]string),
	}
}

// Start 启动客户端（后台协程）
func (c *HZWSocketClient) Start() {
	if !atomic.CompareAndSwapInt32(&c.running, 0, 1) {
		return
	}

	// 先加载货源
	c.loadSuppliers()

	// 定时刷新货源（每5分钟）
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			c.loadSuppliers()
		}
	}()

	// 定时打印统计（每分钟）
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			u := atomic.LoadInt64(&c.updateCount)
			s := atomic.LoadInt64(&c.skipCount)
			if u > 0 || s > 0 {
				log.Printf("[HZW-Socket] 统计: 已更新 %d 条，跳过 %d 条", u, s)
			}
		}
	}()

	// 主连接循环（自动重连）
	go c.connectLoop()
}

// loadSuppliers 从 huoyuan 表加载 HZW 货源
func (c *HZWSocketClient) loadSuppliers() {
	rows, err := database.DB.Query(
		"SELECT hid, `user` FROM qingka_wangke_huoyuan WHERE pt = 'hzw' AND status = '1'",
	)
	if err != nil {
		log.Printf("[HZW-Socket] 加载货源失败: %v", err)
		return
	}
	defer rows.Close()

	newMap := make(map[int]string)
	for rows.Next() {
		var hid int
		var user string
		if err := rows.Scan(&hid, &user); err == nil {
			newMap[hid] = user
		}
	}

	c.mu.Lock()
	c.supplierMap = newMap
	c.mu.Unlock()

	log.Printf("[HZW-Socket] 已加载 %d 个 HZW 货源", len(newMap))
}

// connectLoop 持续连接，断线自动重连
func (c *HZWSocketClient) connectLoop() {
	backoff := 3 * time.Second
	maxBackoff := 30 * time.Second

	for atomic.LoadInt32(&c.running) == 1 {
		err := c.connect()
		if err != nil {
			log.Printf("[HZW-Socket] 连接失败: %v，%v 后重试", err, backoff)
			time.Sleep(backoff)
			backoff = backoff * 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		// 连接成功，重置退避
		backoff = 3 * time.Second

		// 进入消息读取循环
		c.readLoop()

		log.Printf("[HZW-Socket] 连接断开，3秒后重连")
		time.Sleep(3 * time.Second)
	}
}

// buildWSURL 将 http(s) URL 转换为 WebSocket URL，拼接 Socket.IO 路径
func (c *HZWSocketClient) buildWSURL() (string, error) {
	u, err := url.Parse(c.serverURL)
	if err != nil {
		return "", err
	}

	// http -> ws, https -> wss
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	default:
		u.Scheme = "ws"
	}

	u.Path = "/socket.io/"
	q := u.Query()
	q.Set("EIO", "4")
	q.Set("transport", "websocket")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// connect 建立 WebSocket 连接并完成 Socket.IO v4 握手
func (c *HZWSocketClient) connect() error {
	wsURL, err := c.buildWSURL()
	if err != nil {
		return fmt.Errorf("构造 WebSocket URL 失败: %w", err)
	}

	log.Printf("[HZW-Socket] 连接中: %s", wsURL)

	dialer := websocket.Dialer{
		HandshakeTimeout: 15 * time.Second,
	}

	header := http.Header{}
	header.Set("Origin", c.serverURL)

	conn, _, err := dialer.Dial(wsURL, header)
	if err != nil {
		return fmt.Errorf("WebSocket 连接失败: %w", err)
	}

	c.conn = conn

	// 读取 Engine.IO open 包 (type 0)
	_, msg, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return fmt.Errorf("读取握手失败: %w", err)
	}

	msgStr := string(msg)
	if len(msgStr) == 0 || msgStr[0] != '0' {
		conn.Close()
		return fmt.Errorf("非预期的握手响应: %s", msgStr)
	}

	// 解析 pingInterval
	var openData struct {
		SID          string `json:"sid"`
		PingInterval int    `json:"pingInterval"`
		PingTimeout  int    `json:"pingTimeout"`
	}
	if err := json.Unmarshal([]byte(msgStr[1:]), &openData); err == nil {
		log.Printf("[HZW-Socket] 握手成功 sid=%s pingInterval=%dms", openData.SID, openData.PingInterval)
	}

	// 等待 Socket.IO connect 确认 (40 或 40{...})
	_, msg, err = conn.ReadMessage()
	if err != nil {
		conn.Close()
		return fmt.Errorf("等待 Socket.IO 连接确认失败: %w", err)
	}

	msgStr = string(msg)
	if !strings.HasPrefix(msgStr, "40") {
		conn.Close()
		return fmt.Errorf("非预期的 Socket.IO 连接响应: %s", msgStr)
	}

	log.Printf("[HZW-Socket] 已连接到 HZW Socket 服务")

	// 启动心跳协程
	pingInterval := 25 * time.Second
	if openData.PingInterval > 0 {
		pingInterval = time.Duration(openData.PingInterval) * time.Millisecond
	}
	go c.heartbeat(pingInterval)

	return nil
}

// heartbeat 处理 Engine.IO 心跳
func (c *HZWSocketClient) heartbeat(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		conn := c.conn
		if conn == nil {
			return
		}
		// 发送 Engine.IO ping (type 2)
		if err := conn.WriteMessage(websocket.TextMessage, []byte("2")); err != nil {
			return
		}
	}
}

// readLoop 读取消息并处理
func (c *HZWSocketClient) readLoop() {
	conn := c.conn
	if conn == nil {
		return
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[HZW-Socket] 读取消息错误: %v", err)
			return
		}

		msgStr := string(msg)
		if len(msgStr) == 0 {
			continue
		}

		switch {
		case msgStr == "2":
			// Engine.IO ping -> 回复 pong
			conn.WriteMessage(websocket.TextMessage, []byte("3"))
		case msgStr == "3":
			// Engine.IO pong，忽略
		case strings.HasPrefix(msgStr, "42"):
			// Socket.IO event: 42["event_name", data]
			c.handleEvent(msgStr[2:])
		case strings.HasPrefix(msgStr, "41"):
			// Socket.IO disconnect
			log.Printf("[HZW-Socket] 服务端断开连接")
			return
		}
	}
}

// handleEvent 处理 Socket.IO 事件
func (c *HZWSocketClient) handleEvent(payload string) {
	// 解析 JSON 数组: ["message", {...}]
	var raw []json.RawMessage
	if err := json.Unmarshal([]byte(payload), &raw); err != nil {
		return
	}
	if len(raw) < 2 {
		return
	}

	// 检查事件名是否是 "message"
	var eventName string
	if err := json.Unmarshal(raw[0], &eventName); err != nil {
		return
	}
	if eventName != "message" {
		return
	}

	// 消息体可能是字符串(JSON序列化)或直接对象
	var msgData HZWSocketMessage
	var rawStr string
	if err := json.Unmarshal(raw[1], &rawStr); err == nil {
		// 消息是字符串，再解析一次
		if err := json.Unmarshal([]byte(rawStr), &msgData); err != nil {
			return
		}
	} else {
		// 直接是对象
		if err := json.Unmarshal(raw[1], &msgData); err != nil {
			return
		}
	}

	c.processMessage(msgData)
}

// processMessage 处理收到的订单状态更新
func (c *HZWSocketClient) processMessage(msg HZWSocketMessage) {
	remoteUID := fmt.Sprintf("%v", msg.UID)
	remoteOID := fmt.Sprintf("%v", msg.OID)

	c.mu.RLock()
	suppliers := c.supplierMap
	c.mu.RUnlock()

	matched := false
	for hid, sourceUID := range suppliers {
		if sourceUID == remoteUID {
			matched = true
			result, err := database.DB.Exec(
				"UPDATE qingka_wangke_order SET `status`=?, `process`=?, `remarks`=?, `examStartTime`=?, `examEndTime`=? WHERE `yid`=? AND `hid`=?",
				msg.Status, msg.Process, msg.Remarks, msg.ExamStartTime, msg.ExamEndTime,
				remoteOID, hid,
			)
			if err != nil {
				log.Printf("[HZW-Socket] DB更新失败 hid=%d yid=%s: %v", hid, remoteOID, err)
				continue
			}
			if affected, _ := result.RowsAffected(); affected > 0 {
				atomic.AddInt64(&c.updateCount, 1)
				log.Printf("[HZW-Socket] 更新 hid=%d yid=%s 状态=%s 进度=%s", hid, remoteOID, msg.Status, msg.Process)

				// 触发推送通知
				oidInt, _ := strconv.Atoi(remoteOID)
				if oidInt > 0 {
					// 查本地 oid
					var localOID int
					database.DB.QueryRow("SELECT oid FROM qingka_wangke_order WHERE yid=? AND hid=?", remoteOID, hid).Scan(&localOID)
					if localOID > 0 {
						NotifyOrderStatusChange(localOID, msg.Status, msg.Process, msg.Remarks)
					}
				}
			}
		}
	}

	if !matched {
		atomic.AddInt64(&c.skipCount, 1)
	}
}

// Stop 停止客户端
func (c *HZWSocketClient) Stop() {
	atomic.StoreInt32(&c.running, 0)
	if c.conn != nil {
		c.conn.Close()
	}
}

// GetHZWSocketURL 从数据库 qingka_wangke_config 读取 HZW Socket URL
func GetHZWSocketURL() string {
	var val string
	err := database.DB.QueryRow(
		"SELECT svalue FROM qingka_wangke_config WHERE skey = 'hzw_socket_url' LIMIT 1",
	).Scan(&val)
	if err != nil || val == "" {
		return ""
	}
	return val
}

// SetHZWSocketURL 保存 HZW Socket URL 到 qingka_wangke_config
func SetHZWSocketURL(socketURL string) error {
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_config (skey, svalue) VALUES ('hzw_socket_url', ?) ON DUPLICATE KEY UPDATE svalue = ?",
		socketURL, socketURL,
	)
	return err
}

// StartHZWSocket 启动 HZW Socket 客户端（从配置读取 URL）
func StartHZWSocket() {
	socketURL := GetHZWSocketURL()
	if socketURL == "" {
		log.Printf("[HZW-Socket] 未配置 Socket URL，跳过启动。请在对接中心设置。")
		return
	}

	GlobalHZWSocket = NewHZWSocketClient(socketURL)
	GlobalHZWSocket.Start()
}

// RestartHZWSocket 重启客户端（配置变更后调用）
func RestartHZWSocket() {
	if GlobalHZWSocket != nil {
		GlobalHZWSocket.Stop()
		GlobalHZWSocket = nil
	}
	StartHZWSocket()
}
