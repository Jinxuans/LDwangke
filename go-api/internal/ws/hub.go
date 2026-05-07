package ws

import (
	"encoding/json"
	"sync"

	obslogger "go-api/internal/observability/logger"
)

type PushMessage struct {
	Type    string      `json:"type"` // order_status / payment / system / chat_notify
	Title   string      `json:"title"`
	Content string      `json:"content"`
	Data    interface{} `json:"data,omitempty"`
}

type Client struct {
	UID  int
	Conn interface{ WriteMessage(int, []byte) error }
	Send chan []byte
}

type Hub struct {
	mu         sync.RWMutex
	clients    map[int]map[*Client]struct{} // uid -> clients
	register   chan *Client
	unregister chan *Client
	done       chan struct{}
	stopOnce   sync.Once
}

var GlobalHub *Hub

func NewHub() *Hub {
	hub := &Hub{
		clients:    make(map[int]map[*Client]struct{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		done:       make(chan struct{}),
	}
	GlobalHub = hub
	return hub
}

func (h *Hub) Run() {
	for {
		select {
		case <-h.done:
			h.mu.Lock()
			for uid, clients := range h.clients {
				delete(h.clients, uid)
				for client := range clients {
					close(client.Send)
				}
			}
			h.mu.Unlock()
			return

		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.UID] == nil {
				h.clients[client.UID] = make(map[*Client]struct{})
			}
			h.clients[client.UID][client] = struct{}{}
			count := h.onlineCountLocked()
			h.mu.Unlock()
			obslogger.L().Info("WebSocket 客户端上线", "uid", client.UID, "online", count)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
				}
				if len(clients) == 0 {
					delete(h.clients, client.UID)
				}
			}
			count := h.onlineCountLocked()
			h.mu.Unlock()
			obslogger.L().Info("WebSocket 客户端下线", "uid", client.UID, "online", count)
		}
	}
}

func (h *Hub) Register(client *Client) {
	select {
	case <-h.done:
		return
	case h.register <- client:
	}
}

func (h *Hub) Unregister(client *Client) {
	select {
	case <-h.done:
		return
	case h.unregister <- client:
	}
}

func (h *Hub) Stop() {
	if h == nil {
		return
	}
	h.stopOnce.Do(func() {
		close(h.done)
	})
}

// PushToUser 向指定用户推送消息
func (h *Hub) PushToUser(uid int, msg PushMessage) {
	select {
	case <-h.done:
		return
	default:
	}

	data, err := json.Marshal(msg)
	if err != nil {
		obslogger.L().Warn("序列化推送消息失败", "error", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	clients := h.clients[uid]
	if len(clients) == 0 {
		return
	}

	for client := range clients {
		select {
		case client.Send <- data:
		default:
			// 通道满了，丢弃
			obslogger.L().Warn("推送消息失败：通道已满", "uid", uid)
		}
	}
}

// Broadcast 向所有在线用户广播
func (h *Hub) Broadcast(msg PushMessage) {
	select {
	case <-h.done:
		return
	default:
	}

	data, err := json.Marshal(msg)
	if err != nil {
		obslogger.L().Warn("序列化广播消息失败", "error", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, clients := range h.clients {
		for client := range clients {
			select {
			case client.Send <- data:
			default:
			}
		}
	}
}

// OnlineCount 在线人数
func (h *Hub) OnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.onlineCountLocked()
}

func (h *Hub) onlineCountLocked() int {
	total := 0
	for _, clients := range h.clients {
		total += len(clients)
	}
	return total
}
