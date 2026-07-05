package routers

import (
	"backend/realtime"
	"bufio"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// SetupRealtimeRoutes configures WebSocket and SSE endpoints
func SetupRealtimeRoutes(router fiber.Router) {
	// Initialize the Realtime Hub
	realtime.InitHub()

	// WebSocket Endpoint
	// Requires middleware check in main.go
	router.Get("/ws", websocket.New(func(c *websocket.Conn) {
		client := &realtime.WSClient{
			ID:   c.Query("uid", "guest"),
			Conn: c,
			Send: make(chan []byte, 256),
		}
		
		realtime.AppHub.RegisterWS <- client

		// Goroutine to send messages to this WS client
		go func() {
			defer func() {
				_ = c.Close()
			}()
			for {
				message, ok := <-client.Send
				if !ok {
					_ = c.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				
				if err := c.WriteMessage(websocket.TextMessage, message); err != nil {
					return
				}
			}
		}()

		// Block and read incoming messages
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				break
			}
			
			// Process incoming WS message (Optional: echo back or broadcast)
			log.Printf("Received WS message from %s: %s", client.ID, msg)
			
			// Auto-respond for test
			var data map[string]interface{}
			if json.Unmarshal(msg, &data) == nil {
				if event, ok := data["event"].(string); ok && event == "ping" {
					client.Send <- []byte(`{"event":"pong", "message":"Hello from Go WebSocket!"}`)
				}
			}
		}
		
		realtime.AppHub.UnregisterWS <- client
	}))

	// SSE (Server-Sent Events) Endpoint
	router.Get("/sse", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")

		// Create SSE Client
		client := &realtime.SSEClient{
			ID:   c.Query("uid", "guest"),
			Send: make(chan []byte, 256),
		}
		realtime.AppHub.RegisterSSE <- client

		c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			defer func() {
				realtime.AppHub.UnregisterSSE <- client
			}()

			// Send initial connected message
			fmt.Fprintf(w, "event: connected\ndata: {\"message\":\"SSE Connected to Go\"}\n\n")
			_ = w.Flush()

			for {
				message, ok := <-client.Send
				if !ok {
					return
				}
				
				// Standard SSE format: "data: {JSON}\n\n"
				fmt.Fprintf(w, "data: %s\n\n", message)
				if err := w.Flush(); err != nil {
					log.Printf("SSE Client %s disconnected", client.ID)
					return
				}
			}
		})

		return nil
	})
	
	// Test Endpoint to trigger broadcast manually
	router.Get("/trigger", func(c *fiber.Ctx) error {
		realtime.BroadcastData("notification", map[string]string{
			"message": "Update realtime cepat dari Golang!",
			"time":    "sekarang",
		})
		return c.SendString("Triggered broadcast to all WS & SSE clients")
	})
}
