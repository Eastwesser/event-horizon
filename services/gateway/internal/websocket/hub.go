package websocket

import (
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

// Hub управляет всеми WebSocket соединениями
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			n := len(h.clients)
			h.mu.Unlock()
			slog.Info("websocket client connected", "total", n)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			n := len(h.clients)
			h.mu.Unlock()
			slog.Info("websocket client disconnected", "total", n)

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast отправляет сообщение всем клиентам
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}
