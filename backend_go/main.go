package main

import (
	"log"
	"os"

	"backend/database"
	"backend/routers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
)

func main() {
	// Initialize database
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:       "Pasar UMKM API",
		CaseSensitive: true,
		StrictRouting: true,
	})

	// Configure CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "*",
		AllowMethods:     "*",
		AllowCredentials: true,
	}))

	// Base routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Welcome to Pasar UMKM API (Golang version)"})
	})

	app.Get("/api/health", func(c *fiber.Ctx) error {
		// Simple query to check DB connection
		var test int
		err := database.DB.QueryRow("SELECT 1 as test").Scan(&test)
		if err != nil {
			return c.JSON(fiber.Map{
				"status": "unhealthy",
				"error":  err.Error(),
			})
		}
		return c.JSON(fiber.Map{
			"status":   "healthy",
			"database": "connected",
		})
	})

	// WebSocket Upgrade Middleware
	app.Use("/api/realtime/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// API Groups
	api := app.Group("/api")

	products := api.Group("/products")
	routers.SetupProductRoutes(products)

	users := api.Group("/users")
	routers.SetupUserRoutes(users)

	orders := api.Group("/orders")
	routers.SetupOrderRoutes(orders)

	subscription := api.Group("/subscription")
	routers.SetupSubscriptionRoutes(subscription)

	admin := api.Group("/admin")
	routers.SetupAdminRoutes(admin)

	push := api.Group("/push")
	routers.SetupPushRoutes(push)

	rt := api.Group("/realtime")
	routers.SetupRealtimeRoutes(rt)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Printf("Starting server on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
