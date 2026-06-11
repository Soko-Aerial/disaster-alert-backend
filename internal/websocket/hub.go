package websocket

import (
	"context"
	"encoding/json"
	"sync"

	coderws "github.com/coder/websocket"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
	}
}

func (h *Hub) AddClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true
}

func (h *Hub) RemoveClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.clients, client)
	_ = client.Conn.Close(coderws.StatusNormalClosure, "connection closed")
}

func (h *Hub) SendToAdmins(event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.Role == "admin" {
			send(client, event)
		}
	}
}

func (h *Hub) SendToUser(userID string, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.UserID == userID {
			send(client, event)
		}
	}
}

func (h *Hub) SendToCountry(country string, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.Country == country {
			send(client, event)
		}
	}
}

func (h *Hub) SendToAllUsers(event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.Role == "user" {
			send(client, event)
		}
	}
}

func send(client *Client, event Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	_ = client.Conn.Write(context.Background(), coderws.MessageText, payload)
}