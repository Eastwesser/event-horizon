// WS Connection Management: локальная map user→conn + sticky «сервер».
package main

import (
	"fmt"
	"sync"
)

type Conn struct {
	UserID string
	Server string
}

type Hub struct {
	mu   sync.RWMutex
	conns map[string]*Conn // userID → conn
}

func NewHub() *Hub { return &Hub{conns: make(map[string]*Conn)} }

func (h *Hub) Connect(userID, server string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[userID] = &Conn{UserID: userID, Server: server}
}

func (h *Hub) Disconnect(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns, userID)
}

func (h *Hub) SendLocal(userID, msg string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	c, ok := h.conns[userID]
	if !ok {
		return false
	}
	fmt.Printf("[%s] → %s: %s\n", c.Server, userID, msg)
	return true
}

func main() {
	h := NewHub()
	h.Connect("u1", "ws-2") // sticky: SERVERID=ws-2
	h.Connect("u2", "ws-2")
	h.SendLocal("u1", "hello")
	h.Disconnect("u1")
	fmt.Println("after disconnect:", h.SendLocal("u1", "ping"))
}
