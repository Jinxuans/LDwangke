package ws

import (
	"encoding/json"
	"log"
	"sync"
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
	clients    map[int]*Client // uid -> client
	register   chan *Client
	unregister chan *Client
	done       chan struct{}
	stopOnce   sync.Once
}

var GlobalHub *Hub

func NewHub() *Hub {
	hub := &Hub{
		clients:    make(map[int]*Client),
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
			for uid, client := range h.clients {
				delete(h.clients, uid)
				close(client.Send)
			}
			h.mu.Unlock()
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UID] = client
			h.mu.Unlock()
			log.Printf("WebSocket 客户端上线: uid=%d, 当前在线: %d", client.UID, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UID]; ok {
				delete(h.clients, client.UID)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("WebSocket 客户端下线: uid=%d, 当前在线: %d", client.UID, len(h.clients))
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

	h.mu.RLock()
	client, ok := h.clients[uid]
	h.mu.RUnlock()

	if !ok {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("序列化推送消息失败: %v", err)
		return
	}

	select {
	case client.Send <- data:
	default:
		// 通道满了，丢弃
		log.Printf("推送消息到 uid=%d 失败: 通道已满", uid)
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
		log.Printf("序列化广播消息失败: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		select {
		case client.Send <- data:
		default:
		}
	}
}

// OnlineCount 在线人数
func (h *Hub) OnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
