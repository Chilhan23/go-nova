package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

type EventMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type Hub struct {
	clients    map[*Client]bool
	ticketRoom map[int]map[*Client]bool
	Broadcast  chan *EventMessage
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		ticketRoom: make(map[int]map[*Client]bool),
		Broadcast:  make(chan *EventMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client] = true
			if client.TicketID > 0 {
				if _, ok := h.ticketRoom[client.TicketID]; !ok {
					h.ticketRoom[client.TicketID] = make(map[*Client]bool)
				}
				h.ticketRoom[client.TicketID][client] = true
			}
			h.mu.Unlock()
			log.Printf("🔌 [WS Hub] Client registered: Ticket %d | Total: %d", client.TicketID, len(h.clients))

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				if client.TicketID > 0 && h.ticketRoom[client.TicketID] != nil {
					delete(h.ticketRoom[client.TicketID], client)
					if len(h.ticketRoom[client.TicketID]) == 0 {
						delete(h.ticketRoom, client.TicketID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("🔌 [WS Hub] Client unregistered: Ticket %d", client.TicketID)

		case message := <-h.Broadcast:
			dataBytes, err := json.Marshal(message)
			if err != nil {
				continue
			}

			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- dataBytes:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastToTicket(ticketID int, event string, data interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	roomClients, exists := h.ticketRoom[ticketID]
	if !exists || len(roomClients) == 0 {
		return
	}

	payload := EventMessage{
		Event: event,
		Data:  data,
	}

	dataBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	for client := range roomClients {
		select {
		case client.Send <- dataBytes:
		default:
			log.Printf("[WS Hub] Dropping message to slow client: ticket %d", ticketID)
		}
	}
}
