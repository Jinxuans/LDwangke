package pluginruntime

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
	ordermodule "go-api/internal/modules/order"

	"github.com/gorilla/websocket"
)

type HZWSocketClient struct {
	serverURL string

	mu          sync.RWMutex
	supplierMap map[int]string
	conn        *websocket.Conn
	running     int32
	stopCh      chan struct{}
	stopOnce    sync.Once

	updateCount int64
	skipCount   int64
}

type HZWSocketMessage struct {
	UID           interface{} `json:"uid"`
	OID           interface{} `json:"oid"`
	Status        string      `json:"status"`
	Process       string      `json:"process"`
	Remarks       string      `json:"remarks"`
	ExamStartTime string      `json:"examStartTime"`
	ExamEndTime   string      `json:"examEndTime"`
}

var globalHZWSocket *HZWSocketClient

func newHZWSocketClient(serverURL string) *HZWSocketClient {
	return &HZWSocketClient{
		serverURL:   serverURL,
		supplierMap: make(map[int]string),
		stopCh:      make(chan struct{}),
	}
}

func (c *HZWSocketClient) Start() {
	if !atomic.CompareAndSwapInt32(&c.running, 0, 1) {
		return
	}

	c.loadSuppliers()

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-c.stopCh:
				return
			case <-ticker.C:
				c.loadSuppliers()
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-c.stopCh:
				return
			case <-ticker.C:
				updates := atomic.LoadInt64(&c.updateCount)
				skips := atomic.LoadInt64(&c.skipCount)
				if updates > 0 || skips > 0 {
					log.Printf("[HZW-Socket] 统计: 已更新 %d 条，跳过 %d 条", updates, skips)
				}
			}
		}
	}()

	go c.connectLoop()
}

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

func (c *HZWSocketClient) connectLoop() {
	backoff := 3 * time.Second
	maxBackoff := 30 * time.Second

	for atomic.LoadInt32(&c.running) == 1 {
		if err := c.connect(); err != nil {
			log.Printf("[HZW-Socket] 连接失败: %v，%v 后重试", err, backoff)
			if !c.sleep(backoff) {
				return
			}
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		backoff = 3 * time.Second
		c.readLoop()
		log.Printf("[HZW-Socket] 连接断开，3秒后重连")
		if !c.sleep(3 * time.Second) {
			return
		}
	}
}

func (c *HZWSocketClient) sleep(d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-c.stopCh:
		return false
	case <-timer.C:
		return true
	}
}

func (c *HZWSocketClient) buildWSURL() (string, error) {
	u, err := url.Parse(c.serverURL)
	if err != nil {
		return "", err
	}
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}
	u.Path = "/socket.io/"
	q := u.Query()
	q.Set("EIO", "4")
	q.Set("transport", "websocket")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (c *HZWSocketClient) connect() error {
	wsURL, err := c.buildWSURL()
	if err != nil {
		return fmt.Errorf("构造 WebSocket URL 失败: %w", err)
	}

	log.Printf("[HZW-Socket] 连接中: %s", wsURL)
	dialer := websocket.Dialer{HandshakeTimeout: 15 * time.Second}
	header := http.Header{}
	header.Set("Origin", c.serverURL)

	conn, _, err := dialer.Dial(wsURL, header)
	if err != nil {
		return fmt.Errorf("WebSocket 连接失败: %w", err)
	}
	c.conn = conn

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

	var openData struct {
		SID          string `json:"sid"`
		PingInterval int    `json:"pingInterval"`
	}
	if err := json.Unmarshal([]byte(msgStr[1:]), &openData); err == nil {
		log.Printf("[HZW-Socket] 握手成功 sid=%s pingInterval=%dms", openData.SID, openData.PingInterval)
	}

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
	pingInterval := 25 * time.Second
	if openData.PingInterval > 0 {
		pingInterval = time.Duration(openData.PingInterval) * time.Millisecond
	}
	go c.heartbeat(pingInterval)
	return nil
}

func (c *HZWSocketClient) heartbeat(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		if c.conn == nil {
			return
		}
		if err := c.conn.WriteMessage(websocket.TextMessage, []byte("2")); err != nil {
			return
		}
	}
}

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
			_ = conn.WriteMessage(websocket.TextMessage, []byte("3"))
		case msgStr == "3":
		case strings.HasPrefix(msgStr, "42"):
			c.handleEvent(msgStr[2:])
		case strings.HasPrefix(msgStr, "41"):
			log.Printf("[HZW-Socket] 服务端断开连接")
			return
		}
	}
}

func (c *HZWSocketClient) handleEvent(payload string) {
	var raw []json.RawMessage
	if err := json.Unmarshal([]byte(payload), &raw); err != nil || len(raw) < 2 {
		return
	}

	var eventName string
	if err := json.Unmarshal(raw[0], &eventName); err != nil || eventName != "message" {
		return
	}

	var msgData HZWSocketMessage
	var rawStr string
	if err := json.Unmarshal(raw[1], &rawStr); err == nil {
		if err := json.Unmarshal([]byte(rawStr), &msgData); err != nil {
			return
		}
	} else if err := json.Unmarshal(raw[1], &msgData); err != nil {
		return
	}

	c.processMessage(msgData)
}

func (c *HZWSocketClient) processMessage(msg HZWSocketMessage) {
	remoteUID := fmt.Sprintf("%v", msg.UID)
	remoteOID := fmt.Sprintf("%v", msg.OID)

	c.mu.RLock()
	suppliers := c.supplierMap
	c.mu.RUnlock()

	matched := false
	for hid, sourceUID := range suppliers {
		if sourceUID != remoteUID {
			continue
		}
		matched = true
		result, err := database.DB.Exec(
			"UPDATE qingka_wangke_order SET `status`=?, `process`=?, `remarks`=?, `examStartTime`=?, `examEndTime`=? WHERE `yid`=? AND `hid`=?",
			msg.Status, msg.Process, msg.Remarks, msg.ExamStartTime, msg.ExamEndTime, remoteOID, hid,
		)
		if err != nil {
			log.Printf("[HZW-Socket] DB更新失败 hid=%d yid=%s: %v", hid, remoteOID, err)
			continue
		}
		if affected, _ := result.RowsAffected(); affected > 0 {
			atomic.AddInt64(&c.updateCount, 1)
			log.Printf("[HZW-Socket] 更新 hid=%d yid=%s 状态=%s 进度=%s", hid, remoteOID, msg.Status, msg.Process)
			oidInt, _ := strconv.Atoi(remoteOID)
			if oidInt > 0 {
				var localOID int
				_ = database.DB.QueryRow("SELECT oid FROM qingka_wangke_order WHERE yid=? AND hid=?", remoteOID, hid).Scan(&localOID)
				if localOID > 0 {
					ordermodule.NotifyOrderStatusChange(localOID, msg.Status, msg.Process, msg.Remarks)
				}
			}
		}
	}

	if !matched {
		atomic.AddInt64(&c.skipCount, 1)
	}
}

func (c *HZWSocketClient) Stop() {
	atomic.StoreInt32(&c.running, 0)
	c.stopOnce.Do(func() { close(c.stopCh) })
	if c.conn != nil {
		c.conn.Close()
	}
}

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

func SetHZWSocketURL(socketURL string) error {
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_config (v, k, skey, svalue) VALUES ('hzw_socket_url', '', 'hzw_socket_url', ?) ON DUPLICATE KEY UPDATE svalue = ?",
		socketURL, socketURL,
	)
	return err
}

func StartHZWSocket() {
	socketURL := GetHZWSocketURL()
	if socketURL == "" {
		log.Printf("[HZW-Socket] 未配置 Socket URL，跳过启动。请在对接中心设置。")
		return
	}
	globalHZWSocket = newHZWSocketClient(socketURL)
	globalHZWSocket.Start()
}

func RestartHZWSocket() {
	if globalHZWSocket != nil {
		globalHZWSocket.Stop()
		globalHZWSocket = nil
	}
	StartHZWSocket()
}

func StopHZWSocket() {
	if globalHZWSocket != nil {
		globalHZWSocket.Stop()
		globalHZWSocket = nil
	}
}
