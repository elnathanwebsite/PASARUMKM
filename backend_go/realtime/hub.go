package realtime

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

// Client for WebSocket
type WSClient struct {
	ID   string
	Conn *websocket.Conn
	Send chan []byte
}

// Client for SSE
type SSEClient struct {
	ID   string
	Send chan []byte
}

type Hub struct {
	WSClients  map[*WSClient]bool
	SSEClients map[*SSEClient]bool

	BroadcastWS  chan []byte
	BroadcastSSE chan []byte

	RegisterWS   chan *WSClient
	UnregisterWS chan *WSClient

	RegisterSSE   chan *SSEClient
	UnregisterSSE chan *SSEClient

	mu sync.Mutex
}

var AppHub *Hub

func InitHub() {
	AppHub = &Hub{
		WSClients:     make(map[*WSClient]bool),
		SSEClients:    make(map[*SSEClient]bool),
		BroadcastWS:   make(chan []byte),
		BroadcastSSE:  make(chan []byte),
		RegisterWS:    make(chan *WSClient),
		UnregisterWS:  make(chan *WSClient),
		RegisterSSE:   make(chan *SSEClient),
		UnregisterSSE: make(chan *SSEClient),
	}
	go AppHub.Run()
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.RegisterWS:
			h.mu.Lock()
			h.WSClients[client] = true
			h.mu.Unlock()
			log.Println("New WebSocket client connected. Total:", len(h.WSClients))
		case client := <-h.UnregisterWS:
			h.mu.Lock()
			if _, ok := h.WSClients[client]; ok {
				delete(h.WSClients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Println("WebSocket client disconnected. Total:", len(h.WSClients))

		case client := <-h.RegisterSSE:
			h.mu.Lock()
			h.SSEClients[client] = true
			h.mu.Unlock()
			log.Println("New SSE client connected. Total:", len(h.SSEClients))
		case client := <-h.UnregisterSSE:
			h.mu.Lock()
			if _, ok := h.SSEClients[client]; ok {
				delete(h.SSEClients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Println("SSE client disconnected. Total:", len(h.SSEClients))

		case message := <-h.BroadcastWS:
			h.mu.Lock()
			for client := range h.WSClients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.WSClients, client)
				}
			}
			h.mu.Unlock()

		case message := <-h.BroadcastSSE:
			h.mu.Lock()
			for client := range h.SSEClients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.SSEClients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// BroadcastData sends JSON data to all WS and SSE clients
func BroadcastData(event string, payload interface{}) {
	if AppHub == nil {
		return
	}
	
	msg := map[string]interface{}{
		"event": event,
		"data":  payload,
	}
	
	bytes, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshalling broadcast message:", err)
		return
	}
	
	AppHub.BroadcastWS <- bytes
	AppHub.BroadcastSSE <- bytes
}
