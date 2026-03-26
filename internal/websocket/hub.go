package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type Client struct {
	Conn     *websocket.Conn
	Room     string
	Send     chan []byte
}

type Hub struct {
	rooms      map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan Message
	mu         sync.RWMutex
}

type Message struct {
	Room    string
	Payload []byte
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Message, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.rooms[client.Room]; !ok {
				h.rooms[client.Room] = make(map[*Client]bool)
			}
			h.rooms[client.Room][client] = true
			h.mu.Unlock()
			log.Debug().Str("room", client.Room).Msg("Client joined room")

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.rooms[client.Room]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.rooms, client.Room)
					}
				}
			}
			h.mu.Unlock()
			log.Debug().Str("room", client.Room).Msg("Client left room")

		case msg := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.rooms[msg.Room]; ok {
				for client := range clients {
					select {
					case client.Send <- msg.Payload:
					default:
						close(client.Send)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) BroadcastToRoom(room string, payload []byte) {
	h.broadcast <- Message{Room: room, Payload: payload}
}

func WritePump(client *Client) {
	defer client.Conn.Close()
	for msg := range client.Send {
		if err := client.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}
