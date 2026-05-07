package ws

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	obslogger "go-api/internal/observability/logger"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return isAllowedOrigin(r)
	},
}

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

func isAllowedOrigin(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}

	originURL, err := url.Parse(origin)
	if err != nil {
		return false
	}

	host := strings.ToLower(strings.TrimSpace(originURL.Hostname()))
	if host == "" {
		return false
	}

	if host == normalizeHost(r.Host) {
		return true
	}

	if host == "localhost" || host == "127.0.0.1" {
		return true
	}

	if strings.HasSuffix(host, ".29.colnt.com") || host == "29.colnt.com" {
		return true
	}

	return false
}

func normalizeHost(hostPort string) string {
	hostPort = strings.ToLower(strings.TrimSpace(hostPort))
	if host, _, err := net.SplitHostPort(hostPort); err == nil {
		return host
	}
	return strings.Trim(hostPort, "[]")
}

func HandlePush(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt("uid")
		if uid == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			obslogger.L().Warn("WebSocket 升级失败", "error", err)
			return
		}

		client := &Client{
			UID:  uid,
			Conn: conn,
			Send: make(chan []byte, 256),
		}

		hub.Register(client)

		go writePump(conn, client, hub)
		go readPump(conn, client, hub)
	}
}

func writePump(conn *websocket.Conn, client *Client, hub *Hub) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func readPump(conn *websocket.Conn, client *Client, hub *Hub) {
	defer func() {
		hub.Unregister(client)
		conn.Close()
	}()

	conn.SetReadLimit(512)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		// 推送通道不处理客户端消息，只维持心跳
	}
}
