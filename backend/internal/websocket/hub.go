package websocket

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	UserID uuid.UUID
	Data   []byte
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered Clients.
	Clients map[*Client]bool

	// Inbound messages from the clients.
	Broadcast chan Message

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client

	// Last updated
	LastUpdated time.Time
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:   make(chan Message),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Clients:     make(map[*Client]bool),
		LastUpdated: time.Now().Add(-3 * time.Second),
	}
}

func (hub *Hub) Run() {
	for {
		select {
		case client := <-hub.Register:
			hub.Clients[client] = true
		case client := <-hub.Unregister:
			if _, ok := hub.Clients[client]; ok {
				delete(hub.Clients, client)
				close(client.Send)
			}
		case message := <-hub.Broadcast:
			for client := range hub.Clients {
				if client.ID != message.UserID && message.UserID != uuid.Nil {
					continue
				}
				select {
				case client.Send <- message.Data:
				default:
					close(client.Send)
					delete(hub.Clients, client)
				}
			}
		}
	}
}
